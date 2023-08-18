package tokens

import (
	"context"
	"crypto/ed25519"
	"errors"
	"time"

	"github.com/dgraph-io/badger/v4"
	"github.com/golang-jwt/jwt/v5"
	"github.com/oklog/ulid/v2"
	"github.com/vinceanalytics/vince/internal/core"
	"github.com/vinceanalytics/vince/internal/keys"
	"github.com/vinceanalytics/vince/internal/must"
	v1 "github.com/vinceanalytics/vince/proto/v1"
	"google.golang.org/protobuf/proto"
)

func Generate(ctx context.Context, key ed25519.PrivateKey,
	issuer v1.Token_Issuer,
	account string, expires time.Time) (string, *jwt.RegisteredClaims) {
	claims := &jwt.RegisteredClaims{
		Issuer:    issuer.String(),
		Subject:   account,
		ID:        ulid.Make().String(),
		ExpiresAt: jwt.NewNumericDate(expires),
		IssuedAt:  jwt.NewNumericDate(core.Now(ctx)),
	}
	token := jwt.NewWithClaims(&jwt.SigningMethodEd25519{}, claims)

	return must.Must(token.SignedString(key))(
		"failed signing jwt token",
	), claims

}

func Valid(db *badger.DB, token string) bool {
	return db.View(func(txn *badger.Txn) error {
		key := keys.Token(token)
		it, err := txn.Get([]byte(key))
		if err != nil {
			return err
		}
		return it.Value(func(val []byte) error {
			var tpub v1.Token
			err := proto.Unmarshal(val, &tpub)
			if err != nil {
				return err
			}
			t, err := jwt.Parse(token, func(t *jwt.Token) (interface{}, error) {
				return ed25519.PublicKey(tpub.PubKey), nil
			})
			if err != nil {
				return err
			}
			if !t.Valid {
				return errors.New("invalid token")
			}
			return nil
		})
	}) == nil
}
