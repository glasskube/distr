package mail

import (
	"bytes"
	"html/template"
	"net/mail"
)

type Mail struct {
	To           []string
	From         *mail.Address
	Bcc          []string
	ReplyTo      string
	Subject      string
	HtmlBodyFunc func() (string, error)
	TextBodyFunc func() (string, error)
}

type MailOpt func(mail *Mail)

func To(to ...string) MailOpt {
	return func(mail *Mail) {
		mail.To = append(mail.To, to...)
	}
}

func From(from mail.Address) MailOpt {
	return func(mail *Mail) {
		mail.From = &from
	}
}

func Bcc(to ...string) MailOpt {
	return func(mail *Mail) {
		mail.Bcc = append(mail.Bcc, to...)
	}
}

func ReplyTo(to string) MailOpt {
	return func(mail *Mail) {
		mail.ReplyTo = to
	}
}

func Subject(subject string) MailOpt {
	return func(mail *Mail) {
		mail.Subject = subject
	}
}

func HtmlBody(body string) MailOpt {
	return func(mail *Mail) {
		mail.HtmlBodyFunc = func() (string, error) { return body, nil }
	}
}

func HtmlBodyTemplate(tmpl *template.Template, data any) MailOpt {
	return func(mail *Mail) {
		mail.HtmlBodyFunc = func() (string, error) {
			var b bytes.Buffer
			err := tmpl.Execute(&b, data)
			return b.String(), err
		}
	}
}

func TextBody(body string) MailOpt {
	return func(mail *Mail) {
		mail.TextBodyFunc = func() (string, error) { return body, nil }
	}
}

type mailOpts []MailOpt

func (opts mailOpts) Apply(mail *Mail) {
	for _, fn := range opts {
		fn(mail)
	}
}

func (opts mailOpts) Create() (mail Mail) {
	opts.Apply(&mail)
	return
}

func New(opts ...MailOpt) Mail {
	return mailOpts(opts).Create()
}
