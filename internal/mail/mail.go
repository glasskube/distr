package mail

import (
	"bytes"
	"html/template"
)

type Mail struct {
	To           []string
	Subject      string
	HtmlBodyFunc func() (string, error)
	TextBodyFunc func() (string, error)
}

type mailOpt func(mail *Mail)

func To(to ...string) mailOpt {
	return func(mail *Mail) {
		mail.To = append(mail.To, to...)
	}
}

func Subject(subject string) mailOpt {
	return func(mail *Mail) {
		mail.Subject = subject
	}
}

func HtmlBody(body string) mailOpt {
	return func(mail *Mail) {
		mail.HtmlBodyFunc = func() (string, error) { return body, nil }
	}
}

func HtmlBodyTemplate(tmpl *template.Template, data any) mailOpt {
	return func(mail *Mail) {
		mail.HtmlBodyFunc = func() (string, error) {
			var b bytes.Buffer
			err := tmpl.Execute(&b, data)
			return b.String(), err
		}
	}
}

func TextBody(body string) mailOpt {
	return func(mail *Mail) {
		mail.TextBodyFunc = func() (string, error) { return body, nil }
	}
}

type mailOpts []mailOpt

func (opts mailOpts) Apply(mail *Mail) {
	for _, fn := range opts {
		fn(mail)
	}
}

func (opts mailOpts) Create() (mail Mail) {
	opts.Apply(&mail)
	return
}

func New(opts ...mailOpt) Mail {
	return mailOpts(opts).Create()
}
