package ro2

import (
	"cmp"
	"crypto/subtle"
	"errors"
	"log/slog"
	"regexp"
	"slices"
	"strings"

	"github.com/dgraph-io/badger/v4"
	"github.com/google/uuid"
	gonanoid "github.com/matoous/go-nanoid/v2"
	v1 "github.com/vinceanalytics/vince/gen/go/vince/v1"
	"github.com/vinceanalytics/vince/internal/config"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/protobuf/proto"
)

func (db *DB) Domains(f func(*v1.Site)) {
	err := db.db.View(func(txn *badger.Txn) error {
		var prefix [2]byte
		prefix[0] = byte(SITE_DOMAIN)
		it := txn.NewIterator(badger.IteratorOptions{
			Prefix: prefix[:],
		})
		defer it.Close()

		for it.Rewind(); it.Valid(); it.Next() {
			it.Item().Value(func(val []byte) error {
				var s v1.Site
				proto.Unmarshal(val, &s)
				f(&s)
				return nil
			})
		}
		return nil
	})
	if err != nil {
		slog.Error("reading domains", "err", err)
	}
}

func (db *DB) Site(domain string) (u *v1.Site) {
	err := db.db.View(func(txn *badger.Txn) error {
		key := make([]byte, len(domain)+2)
		key[0] = byte(SITE_DOMAIN)
		copy(key[2:], []byte(domain))
		it, err := txn.Get(key[:])
		if err != nil {
			return err
		}
		return it.Value(func(val []byte) error {
			u = &v1.Site{}
			proto.Unmarshal(val, u)
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

func PasswordMatch(pwd string) bool {
	return subtle.ConstantTimeCompare(
		[]byte(config.C.GetAdmin().GetPassword()),
		[]byte(pwd),
	) == 1
}

func (db *DB) CreateSite(domain string, public bool) (err error) {
	if s := db.Site(domain); s != nil {
		return nil
	}
	return db.Save(&v1.Site{
		Domain: domain,
		Public: public,
	})
}

func (db *DB) FindOrCreateCreateSharedLink(domain string, name, password string) (share *v1.Share) {

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
		assert(err == nil, "generating shared password", "err", err)
		share.Password = b
	}
	site.Shares = append(site.Shares, share)
	err := db.Save(site)
	assert(err == nil, "saving user after adding shared link", "err", err)
	return
}

func (db *DB) Save(u *v1.Site) error {
	slices.SortFunc(u.Shares, compareShare)

	data, err := proto.Marshal(u)
	if err != nil {
		return err
	}
	return db.db.Update(func(txn *badger.Txn) error {
		em := make([]byte, len(u.Domain)+2)
		em[0] = byte(SITE_DOMAIN)
		copy(em[2:], []byte(u.Domain))
		err = txn.Set(em, data)
		if err != nil {
			return err
		}
		return nil
	})
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

func ID(id []byte) (o uuid.UUID) {
	copy(o[:], id)
	return
}

func compareSite(a, b *v1.Site) int {
	return cmp.Compare(a.Domain, b.Domain)
}

func compareShare(a, b *v1.Share) int {
	return cmp.Compare(a.Id, b.Id)
}
