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
		name   string
		prefix []byte
		key    func() []byte
	}

	samples := []T{
		{
			"bitmap data",
			keys.DataPrefix,
			func() []byte {
				var k Key
				k.WriteData(Global, v1.Field_domain, 0, 0, 0)
				return k.Bytes()
			},
		},
		{
			"bitmap existence",
			keys.DataExistsPrefix,
			func() []byte {
				var k Key
				k.WriteExistence(Global, v1.Field_domain, 0, 0, 0)
				return k.Bytes()
			},
		},
		{
			"site",
			keys.SitePrefix,
			func() []byte {
				return Site([]byte("test"))
			},
		},
		{
			"api key name",
			keys.APIKeyNamePrefix,
			func() []byte {
				return APIKeyName([]byte("test"))
			},
		},
		{
			"api key hash",
			keys.APIKeyHashPrefix,
			func() []byte {
				return APIKeyHash([]byte("test"))
			},
		},
		{
			"acme",
			keys.AcmePrefix,
			func() []byte {
				return ACME([]byte("test"))
			},
		},
		{
			"translate Key",
			keys.TranslateKeyPrefix,
			func() []byte {
				return TranslateKey(v1.Field_domain, []byte("test"))
			},
		},
		{
			"translate id",
			keys.TranslateIDPrefix,
			func() []byte {
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
