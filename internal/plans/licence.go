package plans

import (
	"crypto/ed25519"
	"strconv"

	"github.com/gernest/vince/timex"
	"github.com/golang-jwt/jwt/v5"
)

func IssueLicense(key ed25519.PrivateKey, plan *Plan, emails []string, lid uint64) (string, error) {
	today := timex.Today()
	yearDays := timex.DaysInAYear(today)
	expires := today.AddDate(0, 0, yearDays)
	token := jwt.NewWithClaims(&jwt.SigningMethodEd25519{}, jwt.RegisteredClaims{
		Issuer:    "vinceanalytics",
		Subject:   strconv.FormatUint(plan.YearlyProductID, 10),
		Audience:  emails,
		ExpiresAt: jwt.NewNumericDate(expires),
		NotBefore: jwt.NewNumericDate(today),
		IssuedAt:  jwt.NewNumericDate(today),
		ID:        strconv.FormatUint(lid, 10),
	})
	return token.SignedString(key)
}
