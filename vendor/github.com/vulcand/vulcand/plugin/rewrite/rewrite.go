package rewrite

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	"github.com/vulcand/oxy/utils"
	"github.com/vulcand/vulcand/plugin"
)

const Type = "rewrite"

type Rewrite struct {
	Regexp      string
	Replacement string
	RewriteBody bool
	Redirect    bool
}

func NewRewrite(regex, replacement string, rewriteBody, redirect bool) (*Rewrite, error) {
	return &Rewrite{regex, replacement, rewriteBody, redirect}, nil
}

func (rw *Rewrite) NewHandler(next http.Handler) (http.Handler, error) {
	return newRewriteHandler(next, rw)
}

func (rw *Rewrite) String() string {
	return fmt.Sprintf("regexp=%v, replacement=%v, rewriteBody=%v, redirect=%v",
		rw.Regexp, rw.Replacement, rw.RewriteBody, rw.Redirect)
}

type rewriteHandler struct {
	next        http.Handler
	errHandler  utils.ErrorHandler
	regexp      *regexp.Regexp
	replacement string
	rewriteBody bool
	redirect    bool
}

func newRewriteHandler(next http.Handler, spec *Rewrite) (*rewriteHandler, error) {
	re, err := regexp.Compile(spec.Regexp)
	if err != nil {
		return nil, err
	}
	return &rewriteHandler{
		regexp:      re,
		replacement: spec.Replacement,
		rewriteBody: spec.RewriteBody,
		redirect:    spec.Redirect,
		next:        next,
		errHandler:  utils.DefaultHandler,
	}, nil
}

func (rw *rewriteHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	oldURL := rawURL(req)

	// only continue if the Regexp param matches the URL
	if !rw.regexp.MatchString(oldURL) {
		rw.next.ServeHTTP(w, req)
		return
	}

	// apply a rewrite regexp to the URL
	newURL := rw.regexp.ReplaceAllString(oldURL, rw.replacement)

	// replace any variables that may be in there
	rewrittenURL := &bytes.Buffer{}
	if err := ApplyString(newURL, rewrittenURL, req); err != nil {
		rw.errHandler.ServeHTTP(w, req, err)
		return
	}

	// parse the rewritten URL and replace request URL with it
	parsedURL, err := url.Parse(rewrittenURL.String())
	if err != nil {
		rw.errHandler.ServeHTTP(w, req, err)
		return
	}

	if rw.redirect && newURL != oldURL {
		(&redirectHandler{u: parsedURL}).ServeHTTP(w, req)
		return
	}

	req.URL = parsedURL

	// make sure the request URI corresponds the rewritten URL
	req.RequestURI = req.URL.RequestURI()

	if !rw.rewriteBody {
		rw.next.ServeHTTP(w, req)
		return
	}

	bw := &bufferWriter{header: make(http.Header), buffer: &bytes.Buffer{}}
	newBody := &bytes.Buffer{}

	rw.next.ServeHTTP(bw, req)

	if err := Apply(bw.buffer, newBody, req); err != nil {
		log.Errorf("Failed to rewrite response body: %v", err)
		return
	}

	utils.CopyHeaders(w.Header(), bw.Header())
	w.Header().Set("Content-Length", strconv.Itoa(newBody.Len()))
	w.WriteHeader(bw.code)
	io.Copy(w, newBody)
}

func FromOther(rw Rewrite) (plugin.Middleware, error) {
	return NewRewrite(rw.Regexp, rw.Replacement, rw.RewriteBody, rw.Redirect)
}

func FromCli(c *cli.Context) (plugin.Middleware, error) {
	return NewRewrite(c.String("regexp"), c.String("replacement"), c.Bool("rewriteBody"), c.Bool("redirect"))
}

func GetSpec() *plugin.MiddlewareSpec {
	return &plugin.MiddlewareSpec{
		Type:      Type,
		FromOther: FromOther,
		FromCli:   FromCli,
		CliFlags:  CliFlags(),
	}
}

func CliFlags() []cli.Flag {
	return []cli.Flag{
		cli.StringFlag{
			Name:  "regexp",
			Usage: "regex to match against http request path",
		},
		cli.StringFlag{
			Name:  "replacement",
			Usage: "replacement text into which regex expansions are inserted",
		},
		cli.BoolFlag{
			Name:  "rewriteBody",
			Usage: "if provided, response body is treated as as template and all variables in it are replaced",
		},
		cli.BoolFlag{
			Name:  "redirect",
			Usage: "if provided, request is redirected to the rewritten URL",
		},
	}
}

func rawURL(request *http.Request) string {
	scheme := "http"
	if request.TLS != nil || isXForwardedHTTPS(request) {
		scheme = "https"
	}

	return strings.Join([]string{scheme, "://", request.Host, request.RequestURI}, "")
}

func isXForwardedHTTPS(request *http.Request) bool {
	xForwardedProto := request.Header.Get("X-Forwarded-Proto")

	return len(xForwardedProto) > 0 && xForwardedProto == "https"
}

type redirectHandler struct {
	u *url.URL
}

func (f *redirectHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Location", f.u.String())
	w.WriteHeader(http.StatusFound)
	w.Write([]byte(http.StatusText(http.StatusFound)))
}

type bufferWriter struct {
	header http.Header
	code   int
	buffer *bytes.Buffer
}

func (b *bufferWriter) Close() error {
	return nil
}

func (b *bufferWriter) Header() http.Header {
	return b.header
}

func (b *bufferWriter) Write(buf []byte) (int, error) {
	return b.buffer.Write(buf)
}

// WriteHeader sets rw.Code.
func (b *bufferWriter) WriteHeader(code int) {
	b.code = code
}
