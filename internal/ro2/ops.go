package ro2

import (
	"bytes"
	"errors"
	"log/slog"
	"net/http"
	"regexp"
	"strings"

	"github.com/dgraph-io/badger/v4"
	"github.com/google/uuid"
	v1 "github.com/vinceanalytics/vince/gen/go/vince/v1"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/protobuf/proto"
)

func (db *DB) Domains(f func([]byte)) {
	err := db.db.View(func(txn *badger.Txn) error {
		var prefix [2]byte
		prefix[0] = byte(SITE_DOMAIN)
		it := txn.NewIterator(badger.IteratorOptions{
			Prefix: prefix[:],
		})
		defer it.Close()

		for it.Rewind(); it.Valid(); it.Next() {
			f(it.Item().Key()[2:])
		}
		return nil
	})
	if err != nil {
		slog.Error("reading domains", "err", err)
	}
}

func (db *DB) UserByID(id uuid.UUID) (u *v1.User) {
	err := db.db.View(func(txn *badger.Txn) error {
		var err error
		u, err = db.byID(txn, id[:])
		return err
	})
	if err != nil {
		if !errors.Is(err, badger.ErrKeyNotFound) {
			slog.Error("reading user by id", "id", id.String())
		}
		return nil
	}
	return
}

func (db *DB) byID(txn *badger.Txn, id []byte) (*v1.User, error) {
	var u v1.User
	var key [18]byte
	key[0] = byte(USER_ID)
	copy(key[2:], id[:])
	it, err := txn.Get(key[:])
	if err != nil {
		return nil, err
	}
	err = it.Value(func(val []byte) error {
		return proto.Unmarshal(val, &u)
	})
	if err != nil {

		return nil, err
	}
	return &u, nil
}

func (db *DB) UserByEmail(email string) (u *v1.User) {
	err := db.db.View(func(txn *badger.Txn) error {
		key := make([]byte, len(email)+2)
		key[0] = byte(User_EMAIL)
		copy(key[2:], []byte(email))
		it, err := txn.Get(key[:])
		if err != nil {
			return err
		}
		return it.Value(func(val []byte) error {
			u, err = db.byID(txn, val)
			return err
		})
	})
	if err != nil {
		if !errors.Is(err, badger.ErrKeyNotFound) {
			slog.Error("reading user by email", "email", email)
		}
		return nil
	}
	return
}

func (db *DB) UserByDomain(domain string) (u *v1.User) {
	err := db.db.View(func(txn *badger.Txn) error {
		key := make([]byte, len(domain)+2)
		key[0] = byte(SITE_DOMAIN)
		copy(key[2:], []byte(domain))
		it, err := txn.Get(key[:])
		if err != nil {
			return err
		}
		return it.Value(func(val []byte) error {
			u, err = db.byID(txn, val)
			return err
		})
	})
	if err != nil {
		if !errors.Is(err, badger.ErrKeyNotFound) {
			slog.Error("reading user by domain", "domain", domain)
		}
		return nil
	}
	return
}

func (db *DB) BootStrap(u *v1.User) error {
	if db.UserByEmail(u.Email) != nil {
		return nil
	}
	uid := id()
	u.Id = uid[:]
	b, err := bcrypt.GenerateFromPassword(u.Password, bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = b
	u.SuperUser = true
	return db.SaveUser(u)
}

func PasswordMatch(u *v1.User, pwd string) bool {
	return bcrypt.CompareHashAndPassword(u.Password, []byte(pwd)) == nil
}

func (db *DB) CreateSite(u *v1.User, domain string, public bool) (uid uuid.UUID, err error) {
	for _, s := range u.Sites {
		if s.Domain == domain {
			copy(uid[:], s.Id)
			return
		}
	}
	uid = id()
	u.Sites = append(u.Sites, &v1.Site{
		Id:     uid[:],
		Domain: domain,
		Public: public,
	})
	err = db.SaveUser(u)
	return
}

func Site(u *v1.User, domain string) (site *v1.Site) {
	for _, m := range u.Sites {
		if m.Domain == domain {
			return m
		}
	}
	return
}

func NewUser(r *http.Request) (u *v1.User, validation map[string]any, err error) {
	u = &v1.User{}
	uid := id()
	u.Id = uid[:]
	u.Name = r.Form.Get("name")
	u.Email = r.Form.Get("email")
	password := r.Form.Get("password")
	passwordConfirm := r.Form.Get("password_confirmation")
	validation = make(map[string]any)
	validate("name", u.Name, "required", validation, func(s string) bool {
		return s != ""
	})
	validate("email", u.Email, "required", validation, func(s string) bool {
		return s != ""
	})
	validate("email", u.Email, "invalid email", validation, func(s string) bool {
		return emailRRe.MatchString(s)
	})
	validate("password", password, "required", validation, func(s string) bool {
		return s != ""
	})
	validate("password", password, "has to be at least 6 characters", validation, func(s string) bool {
		return len(s) >= 6
	})
	validate("password", password, "cannot be longer than 64 characters", validation, func(s string) bool {
		return len(s) <= 64
	})
	validate("password_confirmation", passwordConfirm, "required", validation, func(s string) bool {
		return s != ""
	})
	validate("password_confirmation", passwordConfirm, "must match password", validation, func(s string) bool {
		return s == password
	})
	if len(validation) != 0 {
		return
	}
	b, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, nil, err
	}
	u.Password = b
	return
}

func (db *DB) SaveUser(u *v1.User) error {
	data, err := proto.Marshal(u)
	if err != nil {
		return err
	}
	return db.db.Update(func(txn *badger.Txn) error {
		uid := make([]byte, 18)
		uid[0] = byte(USER_ID)
		copy(uid[2:], u.Id)
		err := txn.Set(uid, data)
		if err != nil {
			return err
		}
		em := make([]byte, len(u.Email)+2)
		em[0] = byte(User_EMAIL)
		copy(em[2:], []byte(u.Email))
		err = txn.Set(em, u.Id)
		if err != nil {
			return err
		}

		for _, s := range u.Sites {
			em = em[:2]
			em[0] = byte(SITE_DOMAIN)
			em = append(em, []byte(s.Domain)...)
			err = txn.Set(bytes.Clone(em), u.Id)
			if err != nil {
				return err
			}
		}
		return nil
	})
}

const emailRegexString = "^(?:(?:(?:(?:[a-zA-Z]|\\d|[!#\\$%&'\\*\\+\\-\\/=\\?\\^_`{\\|}~]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])+(?:\\.([a-zA-Z]|\\d|[!#\\$%&'\\*\\+\\-\\/=\\?\\^_`{\\|}~]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])+)*)|(?:(?:\\x22)(?:(?:(?:(?:\\x20|\\x09)*(?:\\x0d\\x0a))?(?:\\x20|\\x09)+)?(?:(?:[\\x01-\\x08\\x0b\\x0c\\x0e-\\x1f\\x7f]|\\x21|[\\x23-\\x5b]|[\\x5d-\\x7e]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])|(?:(?:[\\x01-\\x09\\x0b\\x0c\\x0d-\\x7f]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}]))))*(?:(?:(?:\\x20|\\x09)*(?:\\x0d\\x0a))?(\\x20|\\x09)+)?(?:\\x22))))@(?:(?:(?:[a-zA-Z]|\\d|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])|(?:(?:[a-zA-Z]|\\d|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])(?:[a-zA-Z]|\\d|-|\\.|~|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])*(?:[a-zA-Z]|\\d|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])))\\.)+(?:(?:[a-zA-Z]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])|(?:(?:[a-zA-Z]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])(?:[a-zA-Z]|\\d|-|\\.|~|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])*(?:[a-zA-Z]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])))\\.?$"

var emailRRe = regexp.MustCompile(emailRegexString)

func validate(field, value, reason string, m map[string]any, f func(string) bool) {
	if f(value) {
		return
	}
	m["validation_"+field] = reason
}

func id() uuid.UUID {
	return uuid.Must(uuid.NewV7())
}

func FormatID(id []byte) string {
	var u uuid.UUID
	copy(u[:], id)
	return u.String()
}

var domainRe = regexp.MustCompile(`(?P<domain>(?:[a-z0-9]+(?:-[a-z0-9]+)*\.)+[a-z]{2,})`)

func (db *DB) ValidateSiteDomain(domain string) (good, bad string) {
	good = CleanupDOmain(domain)
	if good == "" {
		bad = "is required"
		return
	}
	if !domainRe.MatchString(good) {
		bad = "only letters, numbers, slashes and period allowed"
		return
	}
	if strings.ContainsAny(domain, reservedChars) {
		bad = "must not contain URI reserved characters " + reservedChars
		return
	}
	if db.UserByDomain(domain) != nil {
		bad = " already exists"
	}
	return
}

const reservedChars = `:?#[]@!$&'()*+,;=`

func CleanupDOmain(domain string) string {
	domain = strings.TrimSpace(domain)
	domain = strings.TrimPrefix(domain, "http://")
	domain = strings.TrimPrefix(domain, "https://")
	domain = strings.TrimPrefix(domain, "www.")
	domain = strings.TrimSuffix(domain, "/")
	domain = strings.ToLower(domain)
	return domain
}
