package ops

import (
	"bytes"
	"cmp"
	"crypto/sha512"
	"errors"
	"fmt"
	"regexp"
	"slices"
	"strings"
	"sync"

	"filippo.io/age"
	"github.com/cockroachdb/pebble"
	gonanoid "github.com/matoous/go-nanoid/v2"
	"github.com/vinceanalytics/vince/fb"
	"github.com/vinceanalytics/vince/fb/admin"
	v1 "github.com/vinceanalytics/vince/gen/go/vince/v1"
	"github.com/vinceanalytics/vince/internal/encoding"
	"github.com/vinceanalytics/vince/internal/keys"
	"github.com/vinceanalytics/vince/internal/models"
	"github.com/vinceanalytics/vince/internal/util/assert"
	"github.com/vinceanalytics/vince/internal/util/data"
	"github.com/vinceanalytics/vince/internal/util/hash"
	"github.com/vinceanalytics/vince/internal/util/translation"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/protobuf/proto"
)

type Ops struct {
	db    *pebble.DB
	tr    translation.Translator
	admin struct {
		name     string
		password []byte
	}
	sites struct {
		mu      sync.RWMutex
		domains map[uint64]*v1.Site
	}
}

func New(db *pebble.DB, tr translation.Translator, sites ...string) *Ops {
	o := &Ops{db: db, tr: tr}
	name, passwd, err := loadAdmin(db)
	assert.Nil(err, "checking admin")
	o.admin.name = name
	o.admin.password = passwd
	o.sites.domains = make(map[uint64]*v1.Site)
	domains(db, func(s *v1.Site) {
		o.sites.domains[hash.String(s.Domain)] = s
	})
	for i := range sites {
		o.CreateSite(sites[i], false)
	}
	return o
}

func (db *Ops) CreateAPIKey(name string, key string) error {
	prefix := key[:6]
	hash := sha512.Sum512([]byte(key))
	data, _ := proto.Marshal(&v1.APIKey{
		Name:   name,
		Prefix: prefix,
		Hash:   hash[:],
	})
	return errors.Join(
		db.db.Set(encoding.APIKeyName([]byte(name)), data, nil),
		db.db.Set(encoding.APIKeyHash(hash[:]), []byte{}, nil),
	)
}

func (db *Ops) ValidAPIKkey(key string) bool {
	hash := sha512.Sum512([]byte(key))
	hk := encoding.APIKeyHash(hash[:])
	err := data.Get(db.db, hk, func(val []byte) error { return nil })
	return err == nil
}

func (db *Ops) DeleteAPIKey(name string) error {
	nameKey := encoding.APIKeyName([]byte(name))
	var a v1.APIKey
	err := data.Get(db.db, nameKey, func(val []byte) error {
		return proto.Unmarshal(val, &a)
	})
	if err != nil {
		return err
	}
	return errors.Join(
		db.db.Delete(nameKey, nil),
		db.db.Delete(encoding.APIKeyHash(a.Hash), nil),
	)
}

func (db *Ops) APIKeys() (ls []*v1.APIKey, err error) {
	err = data.Prefix(db.db, keys.APIKeyNamePrefix, func(key, value []byte) error {
		var a v1.APIKey
		err := proto.Unmarshal(value, &a)
		if err != nil {
			return err
		}
		ls = append(ls, &a)
		return nil
	})
	return
}

func domains(db *pebble.DB, f func(*v1.Site)) {
	err := data.Prefix(db, keys.SitePrefix, func(key, value []byte) error {
		var s v1.Site
		err := proto.Unmarshal(value, &s)
		if err != nil {
			return err
		}
		f(&s)
		return nil
	})
	assert.Nil(err, "loading domains")
}

func (db *Ops) HasSite(domain string) (ok bool) {
	db.sites.mu.RLock()
	_, ok = db.sites.domains[hash.String(domain)]
	db.sites.mu.RUnlock()
	return
}

func (db *Ops) Site(domain string) (u *v1.Site) {
	db.sites.mu.RLock()
	sx := db.sites.domains[hash.String(domain)]
	if sx != nil {
		u = proto.Clone(sx).(*v1.Site)
	}
	db.sites.mu.RUnlock()
	return
}

func (db *Ops) Domains(f func(*v1.Site)) {
	db.sites.mu.RLock()
	for _, s := range db.sites.domains {
		f(s)
	}
	db.sites.mu.RUnlock()
}

func (db *Ops) CreateSite(domain string, public bool) {
	if db.HasSite(domain) {
		return
	}
	db.Save(&v1.Site{
		Domain: domain,
		Public: public,
	})
}

func (db *Ops) DeleteDomain(domain string) {
	err := db.db.Delete(encoding.Site([]byte(domain)), nil)
	assert.Nil(err, "deleting domain")
	db.sites.mu.Lock()
	delete(db.sites.domains, hash.String(domain))
	db.sites.mu.Unlock()
}

func (db *Ops) SeenFirstStats(domain string) (ok bool) {
	key := encoding.TranslateID(models.Field_domain, db.tr.Translate(models.Field_domain, []byte(domain)))
	return data.Has(db.db, key)
}

func (db *Ops) EditSharedLink(site *v1.Site, slug, name string) {
	i, ok := slices.BinarySearchFunc(site.Shares, &v1.Share{Id: slug}, compareShare)
	if !ok {
		return
	}
	site.Shares[i].Name = name
	db.Save(site)
}

func (db *Ops) DeleteSharedLink(site *v1.Site, slug string) {
	site.Shares = slices.DeleteFunc(site.Shares, func(s *v1.Share) bool {
		return s.Id == slug
	})
	db.Save(site)
}

func (db *Ops) FindOrCreateCreateSharedLink(domain string, name, password string) (share *v1.Share) {
	site := db.Site(domain)

	for _, s := range site.Shares {
		if s.Name == name {
			share = s
			return
		}
	}
	id := gonanoid.Must(16)
	share = &v1.Share{Id: id, Name: name}
	if password != "" {
		b, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		assert.Nil(err, "hashing password with bcrypt")
		share.Password = b
	}
	site.Shares = append(site.Shares, share)
	db.Save(site)
	return
}

func (db *Ops) Web() (secret *age.X25519Identity, err error) {
	err = data.Get(db.db, keys.Cookie, func(val []byte) error {
		secret, err = age.ParseX25519Identity(string(val))
		return err
	})
	if errors.Is(err, pebble.ErrNotFound) {
		secret, err = age.GenerateX25519Identity()
		if err == nil {
			err = db.db.Set(keys.Cookie, []byte(secret.String()), nil)
		}
	}
	return
}

func CreateAdmin(db *pebble.DB, name string, password string) error {
	// Hash Password
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hashing admin password %w", err)
	}
	err = db.Set(keys.AdminPrefix, fb.SerializeAdmin([]byte(name), hashed), nil)
	if err != nil {
		return fmt.Errorf("saving admin %w", err)
	}
	return nil
}

func (db *Ops) VerifyPassword(password string) (match bool) {
	err := bcrypt.CompareHashAndPassword(db.admin.password, []byte(password))
	return err == nil
}

func (db *Ops) Admin() (name string) {
	return db.admin.name
}

func loadAdmin(db *pebble.DB) (name string, password []byte, err error) {
	err = data.Get(db, keys.AdminPrefix, func(val []byte) error {
		usr := admin.GetRootAsAdmin(val, 0)
		name = string(usr.NameBytes())
		password = bytes.Clone(usr.PasswordBytes())
		return nil
	})
	if errors.Is(err, pebble.ErrNotFound) {
		err = errors.New("admin account not found")
	}
	return
}

func (db *Ops) Save(u *v1.Site) {
	slices.SortFunc(u.Shares, compareShare)
	sd, err := proto.Marshal(u)
	assert.Nil(err, "marshal site")
	err = db.db.Set(encoding.Site([]byte(u.Domain)), sd, nil)
	assert.Nil(err, "saving site")
	db.sites.mu.Lock()
	db.sites.domains[hash.String(u.Domain)] = u
	db.sites.mu.Unlock()
}

var domainRe = regexp.MustCompile(`(?P<domain>(?:[a-z0-9]+(?:-[a-z0-9]+)*\.)+[a-z]{2,})`)

func (db *Ops) ValidateSiteDomain(domain string) (good, bad string) {
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
	if db.Site(domain) != nil {
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

func compareShare(a, b *v1.Share) int {
	return cmp.Compare(a.Id, b.Id)
}
