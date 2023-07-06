package email

import (
	"bytes"
	"context"
	"fmt"
	"html/template"

	"github.com/emersion/go-message/mail"
	"github.com/vinceanalytics/vince/internal/config"
	"github.com/vinceanalytics/vince/internal/core"
	"github.com/vinceanalytics/vince/internal/models"
	"github.com/vinceanalytics/vince/internal/templates"
	"github.com/vinceanalytics/vince/pkg/log"
	"github.com/vinceanalytics/vince/pkg/schema"
)

var existingUserTpl = templates.Email("email/existing_user_invitation.html")

var newUserTpl = templates.Email("email/new_user_invitation.html")

func ExistingUserInvite(ctx context.Context, invite *schema.Invitation) {
	err := sendInvite(ctx, existingUserTpl, invite)
	if err != nil {
		log.Get().Err(err).Msg("failed to send invite to existing user")
	}
}

func NewUserInvite(ctx context.Context, invite *schema.Invitation) {
	err := sendInvite(ctx, newUserTpl, invite)
	if err != nil {
		log.Get().Err(err).Msg("failed to send invite to existing user")
	}
}

func sendInvite(ctx context.Context, tpl *template.Template, invitation *schema.Invitation) error {
	var h mail.Header
	var out bytes.Buffer
	h.SetDate(core.Now(ctx))
	x := Get(ctx)
	h.SetAddressList("From", []*mail.Address{x.From()})
	h.SetAddressList("To", []*mail.Address{
		{Address: invitation.Email},
	})
	h.SetSubject(fmt.Sprintf("[Vince Analytics] You've been invited to %s", invitation.Site.Domain))
	mw, err := mail.CreateWriter(&out, h)
	if err != nil {
		return err
	}
	// Create a text part
	var th mail.InlineHeader
	th.Set("Content-Type", "text/html")
	w, err := mw.CreateSingleInline(th)
	if err != nil {
		return err
	}
	err = tpl.Execute(&out, map[string]any{
		"owner":      models.UserByID(ctx, invitation.Site.UserID),
		"invitation": invitation,
		"host":       config.Get(ctx).URL,
	})
	if err != nil {
		return err
	}
	w.Close()
	mw.Close()
	return x.SendMail(x.From().Address, []string{invitation.Email}, &out)
}
