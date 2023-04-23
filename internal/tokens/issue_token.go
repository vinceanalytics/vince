package tokens

import (
	"context"
	"strconv"
	"strings"

	"github.com/gernest/vince/config"
	"github.com/gernest/vince/log"
	"github.com/gernest/vince/models"
	"github.com/gernest/vince/timex"
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
)

func Issue(ctx context.Context, key *models.APIKey) string {
	if key.ID != 0 {
		log.Get(ctx).Warn().Msg("re issuing of tokens is not supported")
		return ""
	}
	// we first save this model to obtain its id we dot his within a transaction
	// so we can rollback if we fail to create the token.
	db := models.Get(ctx).Begin()
	err := db.Create(key).Error
	if err != nil {
		models.DBE(ctx, err, "failed to save API Key")
		db.Rollback()
		return ""
	}
	today := timex.Today()
	expires := today.AddDate(1, 0, 0)
	token := jwt.NewWithClaims(&jwt.SigningMethodEd25519{}, jwt.RegisteredClaims{
		Issuer:    "vinceanalytics",
		Subject:   strconv.FormatUint(key.UserID, 10),
		Audience:  strings.Split(key.Scopes, ","),
		ExpiresAt: jwt.NewNumericDate(expires),
		NotBefore: jwt.NewNumericDate(today),
		IssuedAt:  jwt.NewNumericDate(today),
		ID:        strconv.FormatUint(key.ID, 10),
	})
	tokenString, err := token.SignedString(config.SECURITY.Private)
	if err != nil {
		db.Rollback()
		log.Get(ctx).Err(err).Msg("failed to generate token")
		return ""
	}
	err = db.Model(key).Update("key_prefix", tokenString[:6]).Error
	if err != nil {
		db.Rollback()
		models.DBE(ctx, err, "failed to update key prefix for api")
		return ""
	}
	err = db.Commit().Error
	if err != nil {
		db.Rollback()
		models.DBE(ctx, err, "failed to commit token issuance transaction")
		return ""
	}
	return tokenString

}

func Validate(ctx context.Context, tokenString string) (*jwt.RegisteredClaims, bool) {
	token, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(t *jwt.Token) (interface{}, error) {
		return config.SECURITY.Public, nil
	})
	if err != nil {
		log.Get(ctx).Err(err).Msg("failed to parse jwt token")
		return nil, false
	}
	claims := token.Claims.(*jwt.RegisteredClaims)
	// we need to make sure the token belongs to a valid api key in our database
	return claims, models.Exists(ctx, func(db *gorm.DB) *gorm.DB {
		return db.Model(&models.APIKey{}).Where(
			"user_id = ?", claims.Subject,
			"id = ?", claims.ID,
		)
	})
}
