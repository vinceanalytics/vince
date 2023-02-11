package auth

import (
	"context"
	"database/sql"
	"time"

	"github.com/gernest/vince/models"
	"gorm.io/gorm"
)

func IssueEmailVerification(db *gorm.DB, usr *models.User) (uint64, error) {
	err := db.Model(&models.EmailVerificationCode{}).Where("user_id=?", usr.ID).Update("user_id", nil).Error
	if err != nil {
		return 0, err
	}
	var code models.EmailVerificationCode
	err = db.Model(&models.EmailVerificationCode{}).First(&code).Error
	if err != nil {
		return 0, err
	}
	code.UpdatedAt = time.Now()
	code.UserID = sql.NullInt64{
		Int64: int64(usr.ID),
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
