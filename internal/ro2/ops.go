package ro2

import (
	"cmp"
	"crypto/sha512"
	"errors"
	"fmt"
	"log/slog"
	"regexp"
	"slices"
	"strings"

	"github.com/dgraph-io/badger/v4"
	"github.com/dgraph-io/badger/v4/y"
	gonanoid "github.com/matoous/go-nanoid/v2"
	v1 "github.com/vinceanalytics/vince/gen/go/vince/v1"
	"github.com/vinceanalytics/vince/internal/keys"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/protobuf/proto"
)

func (db *DB) CreateAPIKey(name string, key string) error {
	prefix := key[:6]
	hash := sha512.Sum512([]byte(key))
	data, _ := proto.Marshal(&v1.APIKey{
		Name:   name,
		Prefix: prefix,
		Hash:   hash[:],
	})
	return db.Update(func(tx *Tx) error {
		return errors.Join(
			tx.tx.Set(tx.enc.APIKeyName([]byte(name)), data),
			tx.tx.Set(tx.enc.APIKeyHash(hash[:]), []byte{}),
		)
	})
}

func (db *DB) ValidAPIKkey(key string) bool {
	hash := sha512.Sum512([]byte(key))
	err := db.View(func(tx *Tx) error {
		_, err := tx.tx.Get(tx.enc.APIKeyHash(hash[:]))
		return err
	})
	return err == nil
}

func (db *DB) DeleteAPIKey(name string) error {
	return db.Update(func(tx *Tx) error {
		nameKey := tx.enc.APIKeyName([]byte(name))
		it, err := tx.tx.Get(nameKey)
		if err != nil {
			return err
		}
		var a v1.APIKey
		err = it.Value(func(val []byte) error {
			return proto.Unmarshal(val, &a)
		})
		if err != nil {
			return err
		}
		return errors.Join(
			tx.tx.Delete(nameKey),
			tx.tx.Delete(tx.enc.APIKeyHash(a.Hash)),
		)
	})

}

func (db *DB) APIKeys() (ls []*v1.APIKey, err error) {
	err = db.db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.IteratorOptions{})
		defer it.Close()

		for it.Seek(keys.APIKeyNamePrefix); it.ValidForPrefix(keys.APIKeyNamePrefix); it.Next() {
			var a v1.APIKey
			err := it.Item().Value(func(val []byte) error {
				return proto.Unmarshal(val, &a)
			})
			if err != nil {
				return fmt.Errorf("decoding api key message%w", err)
			}
			ls = append(ls, &a)
		}
		return nil
	})
	return
}

func (db *DB) SetupDomains(domains []string) {
	db.Update(func(tx *Tx) error {
		for _, n := range domains {
			key := tx.enc.Site([]byte(n))
			_, err := tx.tx.Get(key)
			if errors.Is(err, badger.ErrKeyNotFound) {
				data, _ := proto.Marshal(&v1.Site{Domain: n})
				err := tx.tx.Set(key, data)
				if err != nil {
					return err
				}
			}
		}
		return nil
	})
}
func (db *DB) Domains(f func(*v1.Site)) {
	err := db.View(func(tx *Tx) error {
		it := tx.Iter()
		prefix := keys.SitePrefix
		var s v1.Site
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			it.Item().Value(func(val []byte) error {
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
	err := db.View(func(tx *Tx) error {
		it, err := tx.tx.Get(tx.enc.Site([]byte(domain)))
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

func (db *DB) CreateSite(domain string, public bool) (err error) {
	if s := db.Site(domain); s != nil {
		return nil
	}
	return db.Save(&v1.Site{
		Domain: domain,
		Public: public,
	})
}

func (db *DB) DeleteDomain(domain string) (err error) {
	return db.Update(func(tx *Tx) error {
		return tx.tx.Delete(
			tx.enc.Site([]byte(domain)),
		)
	})
}

func (db *DB) SeenFirstStats(domain string) (ok bool) {
	db.View(func(tx *Tx) error {
		_, ok = tx.ID(v1.Field_domain, []byte(domain))
		return nil
	})
	return
}

func (db *DB) EditSharedLink(site *v1.Site, slug, name string) error {
	i, ok := slices.BinarySearchFunc(site.Shares, &v1.Share{Id: slug}, compareShare)
	if !ok {
		return nil
	}
	site.Shares[i].Name = name
	return db.Save(site)
}

func (db *DB) DeleteSharedLink(site *v1.Site, slug string) error {
	site.Shares = slices.DeleteFunc(site.Shares, func(s *v1.Share) bool {
		return s.Id == slug
	})
	return db.Save(site)
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
		y.Check(err)
		share.Password = b
	}
	site.Shares = append(site.Shares, share)
	err := db.Save(site)
	y.Check(err)
	return
}

func (db *DB) Save(u *v1.Site) error {
	slices.SortFunc(u.Shares, compareShare)
	data, err := proto.Marshal(u)
	if err != nil {
		return err
	}
	return db.Update(func(tx *Tx) error {
		err = tx.tx.Set(tx.enc.Site([]byte(u.Domain)), data)
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

func compareShare(a, b *v1.Share) int {
	return cmp.Compare(a.Id, b.Id)
}
