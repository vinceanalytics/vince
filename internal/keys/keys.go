package keys

import (
	"path"

	v1 "github.com/vinceanalytics/vince/proto/v1"
)

// Returns a key that stores Site object in the badger database with the given
// domain.
func Site(domain string) string {
	return path.Join((&v1.Site_Key{
		Store: &v1.StoreKey{
			Prefix: v1.StorePrefix_SITES,
		},
		Domain: domain,
	}).Parts()...)
}

// Returns key which stores a block metadata in badger db.
func BlockMeta(domain, uid string) string {
	return path.Join((&v1.Block_Key{
		Store: &v1.StoreKey{
			Prefix: v1.StorePrefix_BLOCKS,
		},
		Kind:   v1.Block_Key_METADATA,
		Domain: domain,
		Uid:    uid,
	}).Parts()...)
}

// Returns key which stores a block index in badger db.
func BlockIndex(domain, uid string) string {
	return path.Join((&v1.Block_Key{
		Store: &v1.StoreKey{
			Prefix: v1.StorePrefix_BLOCKS,
		},
		Kind:   v1.Block_Key_INDEX,
		Domain: domain,
		Uid:    uid,
	}).Parts()...)
}

// Returns a key which stores account object in badger db.
func Account(name string) string {
	return path.Join((&v1.Account_Key{
		Store: &v1.StoreKey{
			Prefix: v1.StorePrefix_ACCOUNT,
		},
		Name: name,
	}).Parts()...)
}
