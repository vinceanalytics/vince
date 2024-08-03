package kv

import (
	"errors"
	"net/http"
	"regexp"
	"strings"

	v1 "github.com/gernest/len64/gen/go/len64/v1"
	"github.com/gernest/len64/internal/assert"
	"github.com/google/uuid"
	"go.etcd.io/bbolt"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/protobuf/proto"
)

var (
	ErrNotFound = errors.New("key not found")
)

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
func Domains(tx *bbolt.Tx, f func(domain string)) error {
	if b := tx.Bucket(sdm); b != nil {
		return b.ForEach(func(k, v []byte) error {
			f(string(k))
			return nil
		})
	}
	return nil
}

func (u *User) ID() (o uuid.UUID) {
	copy(o[:], u.Id)
	return
}

func (u *User) save(tx *bbolt.Tx, data []byte) {
	with(tx, uid).Put(u.Id, data)
	with(tx, ue).Put(u.Id, data)
	for _, s := range u.Sites {
		if s.Role != v1.ROLE_owner {
			continue
		}
		with(tx, sid).Put(s.Id, data)
		with(tx, sdm).Put([]byte(s.Domain), data)
	}
}

func with(tx *bbolt.Tx, buckets ...[]byte) *bbolt.Bucket {
	b := tx.Bucket(buckets[0])
	if b == nil {
		var err error
		b, err = tx.CreateBucket(buckets[0])
		assert.Assert(err == nil, "creating root bucket", "bucket", string(buckets[0]), "err", err)
	}
	for _, n := range buckets[1:] {
		x := b.Bucket(n)
		if x == nil {
			var err error
			x, err = b.CreateBucket(n)
			assert.Assert(err == nil, "creating nested bucket", "bucket", string(n), "err", err)
		}
		b = x
	}
	return b
}

func (u *User) Save(tx *bbolt.Tx) error {
	data, err := proto.Marshal(u)
	if err != nil {
		return err
	}
	u.save(tx, data)
	return nil
}

func (u *User) ByID(tx *bbolt.Tx, id uuid.UUID) error {
	return u.get(tx, uid, uid)
}

func (u *User) ByEmail(tx *bbolt.Tx, email string) error {
	return u.get(tx, ue, []byte(email))
}

func (u *User) BySite(tx *bbolt.Tx, siteId []byte) error {
	return u.get(tx, sdm, siteId)
}

func (u *User) ByDomain(tx *bbolt.Tx, domain string) error {
	return u.get(tx, sdm, []byte(domain))
}

func (u *User) CreateGoal(tx *bbolt.Tx, domain, event, path string) error {
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
			return u.Save(tx)
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

func (u *User) CreateSite(tx *bbolt.Tx, domain string, public bool) (id uuid.UUID, err error) {
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
	err = u.Save(tx)
	return
}

func (u *User) Site(domain string) (site *v1.Site) {
	for _, m := range u.Sites {
		if m.Domain == domain {
			return m
		}
	}
	return
}

func (u *User) SiteOwner(tx *bbolt.Tx, siteId []byte) error {
	return u.BySite(tx, siteId)
}

func (u *User) OwnedSites() int {
	return len(u.Sites)
}

func (u *User) get(tx *bbolt.Tx, bucket, key []byte) error {
	b := tx.Bucket(bucket)
	if b != nil {
		if data := b.Get(key); data != nil {
			return proto.Unmarshal(data, u)
		}
	}
	return ErrNotFound
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

func ValidateSiteDomain(tx *bbolt.Tx, domain string) (good, bad string) {
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
	if b := tx.Bucket(sdm); b != nil && (b.Get([]byte(domain))) != nil {
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
