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
	"github.com/vinceanalytics/vince/internal/alicia"
	"github.com/vinceanalytics/vince/internal/config"
	"github.com/vinceanalytics/vince/internal/domains"
	"github.com/vinceanalytics/vince/internal/features"
	"github.com/vinceanalytics/vince/internal/license"
	"google.golang.org/protobuf/proto"
)

func (db *DB) checkLicense(ctx context.Context) {

	key := alicia.Get()
	defer key.Release()

	// we use an update transaction to make sure we can update system for the
	// first time
	err := db.db.Update(func(txn *badger.Txn) error {
		sys := key.System()
		it, err := txn.Get(sys)
		if err != nil {
			if !errors.Is(err, badger.ErrKeyNotFound) {
				return err
			}

			if config.C.License != "" {
				// first time user. First we try to seee if we have a valid license
				// provided as a commmandline argument
				data, err := licenseData(config.C.License)
				if err != nil {
					slog.Error("reading license key", "err", err)
					os.Exit(1)
				}
				ls, err := license.Verify(data)
				if err != nil {
					slog.Error("failed validation", "err", err)
					os.Exit(1)
				}
				features.Expires.Store(ls.Expiry)
				features.Email.Store(ls.Email)

			} else {
				features.Expires.Store(uint64(time.Now().UTC().Add(30 * 24 * time.Hour).UnixMilli()))
				features.Email.Store(config.C.Admin.Email)
			}
			ds, _ := proto.Marshal(&v1.License{
				Expiry: features.Expires.Load(),
				Email:  features.Email.Load().(string),
			})
			return txn.Set(sys, ds)
		}
		return it.Value(func(val []byte) error {
			var ls v1.License
			err := proto.Unmarshal(val, &ls)
			if err != nil {
				return err
			}
			features.Expires.Store(ls.Expiry)
			features.Email.Store(ls.Email)
			return nil
		})
	})
	if err != nil {
		slog.Error("failed setup license", "err", err)
		os.Exit(1)
	}
	if !features.Validate() {
		slog.Error("Invalid license, visit https://vinceanalytics.com/#pricing")
		os.Exit(1)
	}

	ts := time.NewTicker(time.Minute)
	defer ts.Stop()
	last := features.Validate()
	if last {
		domains.Load(db.Domains)
	}
	slog.Info("starting license check loop")
	for {
		select {
		case <-ctx.Done():
			return
		case <-ts.C:
			ok := features.Validate()
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
	ls, err := license.Verify(licenseKey)
	if err != nil {
		return err
	}
	features.Apply(ls)
	if features.Validate() {
		return db.Update(func(tx *Tx) error {
			data, _ := proto.Marshal(ls)
			return tx.tx.Set(
				tx.get().System(), data,
			)
		})
	}
	return errors.New("invalid license key")
}

func (db *DB) LockSites(locked bool) error {
	return db.Update(func(tx *Tx) error {
		it := tx.tx.NewIterator(badger.IteratorOptions{
			Prefix: tx.get().Site(""),
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
				tx.get().Site(ls.Domain),
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
