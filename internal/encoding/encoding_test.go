package encoding

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
	v1 "github.com/vinceanalytics/vince/gen/go/vince/v1"
	"github.com/vinceanalytics/vince/internal/keys"
)

func TestPrefix(t *testing.T) {
	type T struct {
		key    func() []byte
		name   string
		prefix []byte
	}

	samples := []T{
		{
			name:   "bitmap data",
			prefix: keys.DataPrefix,
			key: func() []byte {
				var k Key
				k.WriteData(Global, v1.Field_domain, 0, 0, 0)
				return k.Bytes()
			},
		},
		{
			name:   "bitmap existence",
			prefix: keys.DataExistsPrefix,
			key: func() []byte {
				var k Key
				k.WriteExistence(Global, v1.Field_domain, 0, 0, 0)
				return k.Bytes()
			},
		},
		{
			name:   "site",
			prefix: keys.SitePrefix,
			key: func() []byte {
				return Site([]byte("test"))
			},
		},
		{
			name:   "api key name",
			prefix: keys.APIKeyNamePrefix,
			key: func() []byte {
				return APIKeyName([]byte("test"))
			},
		},
		{
			name:   "api key hash",
			prefix: keys.APIKeyHashPrefix,
			key: func() []byte {
				return APIKeyHash([]byte("test"))
			},
		},
		{
			name:   "acme",
			prefix: keys.AcmePrefix,
			key: func() []byte {
				return ACME([]byte("test"))
			},
		},
		{
			name:   "translate Key",
			prefix: keys.TranslateKeyPrefix,
			key: func() []byte {
				return TranslateKey(v1.Field_domain, []byte("test"))
			},
		},
		{
			name:   "translate id",
			prefix: keys.TranslateIDPrefix,
			key: func() []byte {
				return TranslateID(v1.Field_domain, 0)
			},
		},
	}

	for _, s := range samples {
		t.Run(s.name, func(t *testing.T) {
			require.True(t, bytes.HasPrefix(s.key(), s.prefix))
		})
	}
}
