package email

import (
	"bytes"
	"context"
	"fmt"
	"net/mail"

	"github.com/vinceanalytics/vince/internal/models"
	"github.com/vinceanalytics/vince/internal/templates"
)

func SendInviteToExistingUser(ctx context.Context, by *models.User, site *models.Site, to string) error {
	mailer := Get(ctx)
	from := mailer.From()
	var b bytes.Buffer
	subject := fmt.Sprintf("[Vince Analytics] You've been invited to %s", site.Domain)
	err := Compose(ctx, &b, templates.InviteExistingUser, from,
		&mail.Address{Address: to}, subject, func(ctx *templates.Context) {
			ctx.Invite = &templates.Invite{
				Email: by.Email,
			}
		})
	if err != nil {
		return err
	}
	return mailer.SendMail(from.Address, []string{to}, &b)
}

func SendInviteToNewUser(ctx context.Context, by *models.User, site *models.Site, i *models.Invitation) error {
	mailer := Get(ctx)
	from := mailer.From()
	var b bytes.Buffer
	subject := fmt.Sprintf("[Vince Analytics] You've been invited to %s", site.Domain)
	err := Compose(ctx, &b, templates.InviteNewUser, from,
		&mail.Address{Address: i.Email}, subject, func(ctx *templates.Context) {
			ctx.Invite = &templates.Invite{
				Email: by.Email,
				ID:    i.ID,
			}
		})
	if err != nil {
		return err
	}
	return mailer.SendMail(from.Address, []string{i.Email}, &b)
}
