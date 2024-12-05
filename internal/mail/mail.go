package mail

type Mail struct {
	To       []string
	Subject  string
	HtmlBody string
	TextBody string
}

type mailOpt func(mail *Mail)

func To(to []string) mailOpt {
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
		mail.HtmlBody = body
	}
}
func TextBody(body string) mailOpt {
	return func(mail *Mail) {
		mail.TextBody = body
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
