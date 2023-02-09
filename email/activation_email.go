package email

import (
	"bytes"
	"context"

	"github.com/emersion/go-message/mail"
	"github.com/gernest/vince/assets/ui/templates"
)

func SendActivation(ctx context.Context, mailer Mailer, from *mail.Address) error {
	rCtx := templates.New(ctx)
	var b bytes.Buffer
	err := compose(&b, templates.ActivationEmail, rCtx, from,
		rCtx.Code+" is your Vince email verification code",
	)
	if err != nil {
		return err
	}
	return mailer.SendMail(from.Address, []string{rCtx.CurrentUser.Email}, &b)
}
