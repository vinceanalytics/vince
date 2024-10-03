package ro2

import (
	"context"
	"encoding/base64"
	"errors"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/dgraph-io/badger/v4"
	v1 "github.com/vinceanalytics/vince/gen/go/vince/v1"
	"github.com/vinceanalytics/vince/internal/config"
	"github.com/vinceanalytics/vince/internal/domains"
	"github.com/vinceanalytics/vince/internal/encoding"
	"github.com/vinceanalytics/vince/internal/features"
	"github.com/vinceanalytics/vince/internal/keys"
	"github.com/vinceanalytics/vince/internal/license"
	"google.golang.org/protobuf/proto"
)

func (db *DB) checkLicense(ctx context.Context) {
	var loadedLicense *v1.License
	if config.C.License != "" {
		data, err := licenseData(config.C.License)
		if err != nil {
			slog.Error("reading license key", "err", err)
			os.Exit(1)
		}
		loadedLicense, err = license.Parse(data)
		if err != nil {
			slog.Error("parsing license key", "err", err)
			os.Exit(1)
		}
	}

	// handle license updated on the web UI
	err := db.db.Update(func(txn *badger.Txn) error {
		it, err := txn.Get(keys.OpsPrefix)
		if err != nil {
			if !errors.Is(err, badger.ErrKeyNotFound) {
				return err
			}
			return nil
		}

		return it.Value(func(val []byte) error {
			var ls v1.License
			err := proto.Unmarshal(val, &ls)
			if err != nil {
				return err
			}
			// we have saved a license in the database. Only apply if it is  more
			// recent than the one initialized with vince
			if loadedLicense.GetExpiry() < ls.Expiry {
				loadedLicense = &ls
			}
			return nil
		})

	})

	if err != nil {
		slog.Error("failed setup license", "err", err)
		os.Exit(1)
	}

	if loadedLicense == nil {
		slog.Error("missing a valid license key")
		os.Exit(1)
	}

	features.Expires.Store(loadedLicense.GetExpiry())
	features.Email.Store(loadedLicense.GetEmail())

	if err := features.Validate(); err != nil && !errors.Is(err, features.ErrExpired) {
		// It is allowed to start vince with expired license
		slog.Error("license validation", "err", err)
		os.Exit(1)
	}

	ts := time.NewTicker(time.Minute)
	defer ts.Stop()
	last := features.Valid()

	if len(config.C.Domains) > 0 {
		err = db.Update(func(tx *Tx) error {
			for _, n := range config.C.Domains {
				k := encoding.EncodeSite([]byte(n))
				if _, err := tx.tx.Get(k); errors.Is(err, badger.ErrKeyNotFound) {
					data, _ := proto.Marshal(&v1.Site{
						Domain: n,
						Locked: !last,
					})
					err := tx.tx.Set(k, data)
					if err != nil {
						return err
					}
				}
			}
			return nil
		})
		if err != nil {
			slog.Error("failed setup domains", "err", err)
			os.Exit(1)
		}
	}
	domains.Reload(db.Domains)

	slog.Info("starting license check loop")
	for {
		select {
		case <-ctx.Done():
			return
		case <-ts.C:
			ok := features.Valid()
			if ok != last {
				err := db.LockSites(!ok)
				if err != nil {
					slog.Error("locking site", "locked", !ok, "err", err)
				} else {
					last = ok
					domains.Reload(db.Domains)
				}
			}
		}
	}
}

func (db *DB) ApplyLicense(licenseKey []byte) error {
	ls, err := license.Parse(licenseKey)
	if err != nil {
		return err
	}
	err = features.IsValid(ls.Email, ls.Expiry)
	if err != nil {
		return err
	}
	features.Apply(ls)
	return db.Update(func(tx *Tx) error {
		data, _ := proto.Marshal(ls)
		return tx.tx.Set(
			keys.OpsPrefix, data,
		)
	})
}

func (db *DB) PatchLicense(key *v1.License) error {
	return db.Update(func(tx *Tx) error {
		data, _ := proto.Marshal(key)
		return tx.tx.Set(
			keys.OpsPrefix, data,
		)
	})
}

func (db *DB) LockSites(locked bool) error {
	return db.Update(func(tx *Tx) error {
		it := tx.tx.NewIterator(badger.IteratorOptions{
			Prefix: keys.SitePrefix,
		})
		defer it.Close()
		var ls v1.Site
		for it.Rewind(); it.Valid(); it.Next() {
			err := it.Item().Value(func(val []byte) error {
				return proto.Unmarshal(val, &ls)
			})
			if err != nil {
				return err
			}
			ls.Locked = locked
			data, _ := proto.Marshal(&ls)
			err = tx.tx.Set(
				encoding.EncodeSite([]byte(ls.Domain)),
				data,
			)
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func licenseData(path string) ([]byte, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			d, e := base64.StdEncoding.DecodeString(path)
			if e != nil {
				if strings.Contains(e.Error(), "illegal base64 data at input byte ") {
					// returns the filepath error instead
					return nil, err
				}
				return nil, e
			}
			return d, nil
		}
		return nil, err
	}
	return data, nil
}
