package securitytxt

import (
	"context"
	"fmt"
	"github.com/Masterminds/sprig"
	"github.com/containous/traefik/v2/pkg/log"
	"github.com/gorilla/mux"
	"text/template"
	"net/http"
)

type Configuration struct {
	Acknowledgements	string		`description:"URL for acknowledgements" json:"acknowledgements,omitempty" toml:"acknowledgements,omitempty" yaml:"acknowledgements,omitempty" export:"true"`
	Canonical 			string		`description:"Canonical URL of the security.txt"`
	Contact 			string		`description:"Contact URI"`
	Encryption 			string		`description:"Encryption key (URI or Fingerprint)"`
	PreferredLanguages	[]string	`description:"List of preferred language for communications"`
}

type SecurityTxt struct {
	*Configuration
}

const fileTpl = `Canonical: {{.Canonical}}
{{- if .Acknowledgements -}}
Acknowledgements: {{.Acknowledgements}}
{{- end}}
{{if .Contact -}}
Contact: {{.Contact}}
{{- end}}
{{if .Encryption -}}
Encryption: {{.Encryption}}
{{- end}}
{{if .PreferredLanguages -}}
Preferred-Languages: {{ .PreferredLanguages | join ", "}}
{{- end}}
`

func (p *SecurityTxt) CreateHandler(notFoundHandler http.Handler) http.Handler {
	router := mux.NewRouter().SkipClean(true)
	router.NotFoundHandler = notFoundHandler

	router.Methods(http.MethodGet).
		Path("/.well-known/security.txt").
		Handler(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {

			ctx := log.With(context.Background(), log.Str(log.ProviderName, "security.txt"))
			logger := log.FromContext(ctx)

			if (req.TLS != nil) {
				req.URL.Scheme = "https"
			} else {
				req.URL.Scheme = "http"
			}
			p.Canonical = fmt.Sprintf("%s://%s%s", req.URL.Scheme, req.Host, req.URL.Path)

			logger.Debugf("Received a security.txt request for %s", req.Host)

			funcs := sprig.TxtFuncMap()
			t, err := template.New("security.txt").Funcs(funcs).Parse(fileTpl)
			if err != nil {
				logger.Error(err)
			}
			t.Execute(rw, p.Configuration)
	}))
	return router
}
