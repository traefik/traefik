package spnego

import (
	"context"
	"errors"
	"net/http"
	"os"

	"github.com/jcmturner/gokrb5/v8/client"
	"github.com/jcmturner/gokrb5/v8/config"
	"github.com/jcmturner/gokrb5/v8/credentials"
	"github.com/jcmturner/gokrb5/v8/keytab"
	"github.com/jcmturner/gokrb5/v8/spnego"
	"github.com/traefik/traefik/v2/pkg/config/dynamic"
	"github.com/traefik/traefik/v2/pkg/log"
)

// Spnego is a component to make outgoing SPNEGO calls.
type Spnego struct {
	next         http.Handler
	config       dynamic.Spnego
	client       *client.Client
	spnOverrides map[string]string
}

// New creates an Spnego middleware.
func New(ctx context.Context, next http.Handler, config dynamic.Spnego, name string) (http.Handler, error) {
	logger := log.FromContext(ctx)
	logger.Debug("Creating Spnego middleware")

	var err error
	// convert array to map.
	spnOverrides := make(map[string]string)
	for _, v := range config.SpnOverrides {
		spnOverrides[v.DomainName] = v.Spn
	}

	spnego := &Spnego{
		next:         next,
		config:       config,
		spnOverrides: spnOverrides,
	}

	return spnego, err
}

func (s *Spnego) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	logger := log.FromContext(req.Context())
	spn := ""

	hostName := req.Host
	if s.config.Host != "" {
		hostName = s.config.Host
	}

	if value, ok := s.spnOverrides[hostName]; ok {
		spn = value
	}
	logger.Debugf("hostName: %s, spn override: %s", hostName, spn)
	req.URL.Host = hostName

	// SetSPNEGOHeader fails if the ticket is expired.
	// call refreshTicket() only once if it fails.
	for i := 0; i < 2; i++ {
		if s.client != nil {
			err := spnego.SetSPNEGOHeader(s.client, req, spn)
			if err == nil {
				break
			}
			logger.Warnf("Error setting SPNEGO Header. Refreshing ticket. err: %+v", err)
		}
		err := s.refreshTicket(logger)
		if err != nil {
			logger.Errorf("spnego.refreshTicket failed. err: %+v", err)
		}
	}

	s.next.ServeHTTP(rw, req)
}

func (s *Spnego) refreshTicket(logger log.Logger) error {
	var err error
	var kt *keytab.Keytab
	var ccache *credentials.CCache
	var c *config.Config
	var cl *client.Client
	var krb5ConfReader *os.File

	var krbConfPath string
	if s.config.KrbConfPath != "" {
		krbConfPath = s.config.KrbConfPath
	} else {
		krbConfPath = "/etc/krb5.conf"
	}

	krb5ConfReader, err = os.Open(krbConfPath)
	if err != nil {
		return err
	}
	defer krb5ConfReader.Close()

	c, err = config.NewFromReader(krb5ConfReader)
	if err != nil {
		return err
	}
	c.LibDefaults.NoAddresses = true

	switch {
	case s.config.KeytabPath != "":
		logger.Debugf("Using Keytab %s", s.config.KeytabPath)
		user := s.config.User
		if user == "" {
			user = os.Getenv("USER")
		}
		realm := s.config.Realm
		if realm == "" {
			realm = c.LibDefaults.DefaultRealm
		}
		kt, err = keytab.Load(s.config.KeytabPath)
		if err != nil {
			return err
		}
		cl = client.NewWithKeytab(user, realm, kt, c)
	case s.config.CcachePath != "":
		logger.Debugf("Using Ccache %s", s.config.CcachePath)
		ccache, err = credentials.LoadCCache(s.config.CcachePath)
		if err != nil {
			return err
		}
		cl, err = client.NewFromCCache(ccache, c)
		if err != nil {
			return err
		}
	default:
		msg := "either KeytabPath or CcachePath must be specified"
		logger.Error(msg)
		return errors.New(msg)
	}

	s.client = cl
	return nil
}
