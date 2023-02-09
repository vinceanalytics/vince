package email

import (
	"html/template"
	"io"
	"time"

	"github.com/emersion/go-message/mail"
	"github.com/emersion/go-smtp"
	"github.com/gernest/vince/assets/ui/templates"
)

func compose(out io.Writer, tpl *template.Template, ctx *templates.Context, from *mail.Address, subject string) error {
	var h mail.Header
	h.SetDate(time.Now())
	h.SetAddressList("From", []*mail.Address{from})
	h.SetAddressList("To", []*mail.Address{ctx.CurrentUser.Address()})
	h.SetSubject(subject)
	mw, err := mail.CreateWriter(out, h)
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
	err = tpl.Execute(out, ctx)
	if err != nil {
		return err
	}
	w.Close()
	mw.Close()
	return nil
}

type Mailer interface {
	SendMail(from string, to []string, msg io.Reader) error
	io.Closer
}

var _ Mailer = (*MailHog)(nil)

type MailHog struct {
	*smtp.Client
}

func NewMailHog() (*MailHog, error) {
	x, err := smtp.Dial("localhost:1025")
	if err != nil {
		return nil, err
	}
	return &MailHog{Client: x}, nil
}
