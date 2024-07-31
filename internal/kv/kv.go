package kv

import (
	"bytes"
	"errors"
	"net/http"
	"regexp"

	v1 "github.com/gernest/len64/gen/go/len64/v1"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/protobuf/proto"
)

var (
	ErrNotFound = errors.New("key not found")
)

type KeyValue interface {
	Set(key, value []byte) error
	Get(key []byte, value func(val []byte) error) error
	Prefix(key []byte, value func(val []byte) error) error
}

type User struct {
	v1.User
}

// by pre fixing with / ensures we never crash with timeseries namespace. We
// will have to figure out compaction scheme for this namespace though.
var (
	uid = []byte("/uid/")
	ue  = []byte("/uem/")
	sid = []byte("/sid/")
	sdm = []byte("/sdm/")
	iid = []byte("/iid/")
)

// Domains iterate over all registered domains
func Domains(db KeyValue, f func(domain string)) error {
	return db.Prefix(sdm, func(val []byte) error {
		f(string(val))
		return nil
	})
}

func (u *User) ID() (o uuid.UUID) {
	copy(o[:], u.Id)
	return
}

// Keys returns all key space with the user data.
func (u *User) Keys() [][]byte {
	base := [][]byte{
		append(uid, u.Id...),
		append(ue, []byte(u.Email)...),
	}
	for _, s := range u.Sites {
		if s.Role != v1.ROLE_OWNER {
			continue
		}
		base = append(base, append(sid, s.Id...))
		base = append(base, append(sdm, []byte(s.Domain)...))
	}
	for _, s := range u.Sites {
		base = append(base, append(sid, s.Id...))
		base = append(base, append(sid, []byte(s.Domain)...))
	}
	for _, s := range u.Invitations {
		base = append(base, append(iid, s.Id...))
	}
	return base
}

func (u *User) Save(db KeyValue) error {
	data, err := proto.Marshal(u)
	if err != nil {
		return err
	}
	for _, key := range u.Keys() {
		err = db.Set(key, data)
		if err != nil {
			return err
		}
	}
	return nil
}

func (u *User) ByID(id uuid.UUID, db KeyValue) error {
	return u.get(
		append(uid, id[:]...), db,
	)
}

func (u *User) ByEmail(email string, db KeyValue) error {
	return u.get(
		append(ue, []byte(email)...), db,
	)
}

func (u *User) BySite(siteId []byte, db KeyValue) error {
	return u.get(
		append(sid, siteId...), db,
	)
}

func (u *User) ByInvitation(inviteId []byte, db KeyValue) error {
	return u.get(
		append(iid, inviteId...), db,
	)
}

func (u *User) ByDomain(domain string, db KeyValue) error {
	return u.get(
		append(sdm, []byte(domain)...), db,
	)
}

func (u *User) CreateGoal(db KeyValue, domain, event, path string) error {
	for _, s := range u.Sites {
		if s.Domain == domain {
			for _, g := range s.Goals {
				if g.EventName == event && g.PagePath == path {
					return nil
				}
			}
			id := id()
			s.Goals = append(s.Goals, &v1.Goal{
				Id:        id[:],
				EventName: event,
				PagePath:  path,
			})
			return u.Save(db)
		}
	}
	return nil
}

func (u *User) NewUser(r *http.Request) (validation map[string]any, err error) {
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
		return nil, err
	}
	u.Password = b
	return
}

func (u *User) PasswordMatch(pwd string) bool {
	return bcrypt.CompareHashAndPassword(u.Password, []byte(pwd)) == nil
}

func (u *User) CreateSite(db KeyValue, domain string, public bool) (id uuid.UUID, err error) {
	for _, s := range u.Sites {
		if s.Domain == domain {
			copy(id[:], s.Id)
			return
		}
	}
	id = uuid.Must(uuid.NewV7())
	u.Sites = append(u.Sites, &v1.Site{
		Id:     id[:],
		Domain: domain,
		Public: public,
	})
	err = u.Save(db)
	return
}

func (u *User) Role(sid uuid.UUID) (role v1.ROLE, ok bool) {
	for _, m := range u.Sites {
		if bytes.Equal(m.Id, sid[:]) {
			role = m.Role
			ok = true
			break
		}
	}
	return
}

func (u *User) SiteOwner(siteId []byte, db KeyValue) error {
	return u.BySite(siteId, db)
}

func (u *User) OwnedSites() int {
	return len(u.Sites)
}

func (u *User) get(key []byte, db KeyValue) error {
	return db.Get(key, func(val []byte) error {
		return proto.Unmarshal(val, u)
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
