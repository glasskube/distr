package db_test

import (
	"testing"

	"github.com/glasskube/distr/internal/db"
	. "github.com/onsi/gomega"
)

func TestTemplateApplicationLink(t *testing.T) {
	tests := []struct {
		name        string
		link        string
		envFileData []byte
		valuesYaml  []byte
		want        string
		wantErr     bool
	}{
		{
			name:        "empty link returns empty string",
			link:        "",
			envFileData: nil,
			valuesYaml:  nil,
			want:        "",
			wantErr:     false,
		},
		{
			name:        "static link without template variables",
			link:        "https://example.com",
			envFileData: nil,
			valuesYaml:  nil,
			want:        "https://example.com",
			wantErr:     false,
		},
		{
			name: "template with env variables",
			link: "https://{{ .Env.HELLO_DISTR_HOST }}",
			envFileData: []byte(`# mandatory values:
HELLO_DISTR_HOST=localhost
HELLO_DISTR_DB_NAME=hello-distr
HELLO_DISTR_DB_USER=distr
HELLO_DISTR_DB_PASSWORD=distr123`),
			valuesYaml: nil,
			want:       "https://localhost",
			wantErr:    false,
		},
		{
			name: "template with env variables and comment at the end",
			link: "https://{{ .Env.HELLO_DISTR_HOST }}",
			envFileData: []byte(`# mandatory values:
HELLO_DISTR_HOST=localhost # add localhost
HELLO_DISTR_DB_NAME=hello-distr
HELLO_DISTR_DB_USER=distr
HELLO_DISTR_DB_PASSWORD=distr123`),
			valuesYaml: nil,
			want:       "https://localhost",
			wantErr:    false,
		},
		{
			name: "template with multiple env variables",
			link: "postgres://{{ .Env.HELLO_DISTR_DB_USER }}:{{ .Env.HELLO_DISTR_DB_PASSWORD }}" +
				"@{{ .Env.HELLO_DISTR_HOST }}/{{ .Env.HELLO_DISTR_DB_NAME }}",
			envFileData: []byte(`# mandatory values:
HELLO_DISTR_HOST=localhost
HELLO_DISTR_DB_NAME=hello-distr
HELLO_DISTR_DB_USER=distr
HELLO_DISTR_DB_PASSWORD=distr123`),
			valuesYaml: nil,
			want:       "postgres://distr:distr123@localhost/hello-distr",
			wantErr:    false,
		},
		{
			name:        "template with YAML values",
			link:        "https://{{ .Values.app.ingress.hosts }}",
			envFileData: nil,
			valuesYaml: []byte(`app:
  ingress:
    enabled: true
    hosts:
      - host: hostname.local`),
			want:    "https://[map[host:hostname.local]]",
			wantErr: false,
		},
		{
			name:        "template with nested YAML values",
			link:        "https://{{ index .Values.app.ingress.hosts 0 \"host\" }}",
			envFileData: nil,
			valuesYaml: []byte(`app:
  ingress:
    enabled: true
    hosts:
      - host: hostname.local`),
			want:    "https://hostname.local",
			wantErr: false,
		},
		{
			name: "template with both env and values",
			link: "https://{{ index .Values.app.ingress.hosts 0 \"host\" }}:{{ .Env.PORT }}",
			envFileData: []byte(`PORT=8080
DATABASE_URL=postgres://localhost/db`),
			valuesYaml: []byte(`app:
  ingress:
    enabled: true
    hosts:
      - host: hostname.local`),
			want:    "https://hostname.local:8080",
			wantErr: false,
		},
		{
			name: "env file with comments and empty lines",
			link: "https://{{ .Env.HOST }}:{{ .Env.PORT }}",
			envFileData: []byte(`# This is a comment
HOST=example.com

# Another comment
PORT=3000

# End of file`),
			valuesYaml: nil,
			want:       "https://example.com:3000",
			wantErr:    false,
		},
		{
			name:        "env file with values containing equals sign",
			link:        "{{ .Env.CONNECTION_STRING }}",
			envFileData: []byte(`CONNECTION_STRING=postgres://user:pass=word@localhost/db`),
			valuesYaml:  nil,
			want:        "postgres://user:pass=word@localhost/db",
			wantErr:     false,
		},
		{
			name:        "invalid template syntax",
			link:        "https://{{ .Env.HOST",
			envFileData: []byte("HOST=localhost"),
			valuesYaml:  nil,
			want:        "",
			wantErr:     true,
		},
		{
			name:        "invalid YAML",
			link:        "https://{{ .Values.host }}",
			envFileData: nil,
			valuesYaml:  []byte(`invalid: yaml: content: [[[`),
			want:        "",
			wantErr:     true,
		},
		{
			name:        "accessing non-existent env variable",
			link:        "https://{{ .Env.NONEXISTENT }}",
			envFileData: []byte("HOST=localhost"),
			valuesYaml:  nil,
			want:        "https://<no value>",
			wantErr:     false,
		},
		{
			name: "complex template with conditionals",
			link: "https://{{ if .Values.app.ingress.enabled }}" +
				"{{ index .Values.app.ingress.hosts 0 \"host\" }}{{ else }}localhost{{ end }}",
			envFileData: nil,
			valuesYaml: []byte(`app:
  ingress:
    enabled: true
    hosts:
      - host: hostname.local`),
			want:    "https://hostname.local",
			wantErr: false,
		},
		{
			name:        "template with secrets placeholder",
			link:        "https://{{ .Env.HOST }}:{{ .Secrets.API_KEY }}",
			envFileData: []byte(`HOST=api.example.com`),
			valuesYaml:  nil,
			want:        "https://api.example.com:<no value>",
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)
			got, err := db.TemplateApplicationLink(tt.link, tt.envFileData, tt.valuesYaml)
			if tt.wantErr {
				g.Expect(err).To(HaveOccurred())
			} else {
				g.Expect(err).NotTo(HaveOccurred())
				g.Expect(got).To(Equal(tt.want))
			}
		})
	}
}
