package auth

import (
	"context"
	"database/sql"

	"github.com/gernest/vince/assets/ui/templates"
	"github.com/gernest/vince/email"
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

func SendVerificationEmail(ctx context.Context, usr *models.User) error {
	code, err := IssueEmailVerification(ctx, usr)
	if err != nil {
		return err
	}
	ctx = templates.SetActivationCode(ctx, code)
	return email.SendActivation(ctx)
}
