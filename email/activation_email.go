package email

import (
	"bytes"
	"context"
	"fmt"

	"github.com/gernest/vince/assets/ui/templates"
)

func SendActivation(ctx context.Context, mailer Mailer) error {
	from := mailer.From()
	rCtx := templates.New(ctx)
	var b bytes.Buffer
	err := compose(&b, templates.ActivationEmail, rCtx, from,
		fmt.Sprintf("%d is your Vince email verification code", rCtx.Code),
	)
	if err != nil {
		return err
	}
	return mailer.SendMail(from.Address, []string{rCtx.CurrentUser.Email}, &b)
}
