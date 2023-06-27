package email

import (
	"bytes"
	"context"
	"net/mail"

	"github.com/vinceanalytics/vince/internal/models"
	"github.com/vinceanalytics/vince/internal/templates"
)

var passwordReset = templates.Email("email/password_reset_email.html")

func SendPasswordReset(ctx context.Context, usr *models.User, link string) error {
	mailer := Get(ctx)
	from := mailer.From()
	var b bytes.Buffer
	subject := "Vince password reset"
	err := Compose(ctx, &b, passwordReset, from,
		&mail.Address{Name: usr.Name, Address: usr.Email}, subject, func(ctx *templates.Context) {
			ctx.Recipient = usr.Name
			ctx.ResetLink = link
		})
	if err != nil {
		return err
	}
	return mailer.SendMail(from.Address, []string{usr.Email}, &b)
}
