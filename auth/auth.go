package auth

import (
	"context"
	"database/sql"

	"github.com/gernest/vince/models"
)

func IssueEmailVerification(ctx context.Context, usr *models.User) (uint64, error) {
	db := models.Get(ctx)
	err := db.Model(&models.EmailVerificationCode{}).Where("user_id=?", usr.ID).Update("user_id", nil).Error
	if err != nil {
		return 0, err
	}
	var code models.EmailVerificationCode
	err = db.Where("user_id is null").First(&code).Error
	if err != nil {
		return 0, err
	}
	code.UserID = sql.NullInt64{
		Int64: int64(usr.ID),
		Valid: true,
	}
	err = db.Save(&code).Error
	if err != nil {
		return 0, err
	}
	return code.Code, nil
}

type activationCodeKey struct{}

func SetActivationCode(ctx context.Context, code uint64) context.Context {
	return context.WithValue(ctx, activationCodeKey{}, code)
}

func GetActivationCode(ctx context.Context) uint64 {
	if v := ctx.Value(activationCodeKey{}); v != nil {
		return v.(uint64)
	}
	return 0
}
