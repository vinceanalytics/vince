package email

import (
	"bytes"
	"context"
	"fmt"
	"net/mail"

	"github.com/vinceanalytics/vince/internal/models"
	"github.com/vinceanalytics/vince/templates"
)

func SendActivation(ctx context.Context, code uint64) error {
	mailer := Get(ctx)
	from := mailer.From()
	usr := models.GetUser(ctx)
	var b bytes.Buffer
	subject := fmt.Sprintf("%d is your Vince email verification code", code)
	err := Compose(ctx, &b, templates.ActivationEmail, from,
		&mail.Address{Name: usr.Name, Address: usr.Email}, subject, func(ctx *templates.Context) {
			ctx.Code = code
			ctx.Recipient = usr.Name
		})
	if err != nil {
		return err
	}
	return mailer.SendMail(from.Address, []string{usr.Email}, &b)
}
