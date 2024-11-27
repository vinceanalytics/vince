package encoding

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/vinceanalytics/vince/internal/keys"
	"github.com/vinceanalytics/vince/internal/models"
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
				k.WriteData(Global, models.Field_domain, 0, 0, 0)
				return k.Bytes()
			},
		},
		{
			"bitmap existence",
			keys.DataExistsPrefix,
			func() []byte {
				var k Key
				k.WriteExistence(Global, models.Field_domain, 0, 0, 0)
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
				return TranslateKey(models.Field_domain, []byte("test"))
			},
		},
		{
			"translate id",
			keys.TranslateIDPrefix,
			func() []byte {
				return TranslateID(models.Field_domain, 0)
			},
		},
	}

	for _, s := range samples {
		t.Run(s.name, func(t *testing.T) {
			require.True(t, bytes.HasPrefix(s.key(), s.prefix))
		})
	}
}
