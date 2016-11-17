package mozillatlsconfig

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/containous/traefik/log"
)

// Cipher mapping from OpenSSL name to GO const
// From https://testssl.sh/openssl-rfc.mappping.html
var cipherSuites = map[string]uint16{
	"ECDHE-ECDSA-AES256-GCM-SHA384": 			tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
	"ECDHE-RSA-AES256-GCM-SHA384":   			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
	//"ECDHE-ECDSA-CHACHA20-POLY1305":    tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256,
	//"ECDHE-RSA-CHACHA20-POLY1305"       tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256,
	"ECDHE-ECDSA-AES128-GCM-SHA256": 			tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
	"ECDHE-RSA-AES128-GCM-SHA256":   			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
	//"ECDHE-ECDSA-AES256-SHA384":        tls.TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA384,
	//"ECDHE-RSA-AES256-SHA384":          tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA384,
	//"ECDHE-ECDSA-AES128-SHA256":        tls.TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA256,
	//"ECDHE-RSA-AES128-SHA256":          tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA256,

	//"DHE-RSA-AES128-GCM-SHA256":        tls.TLS_DHE_RSA_WITH_AES_128_GCM_SHA256,
	//"DHE-RSA-AES256-GCM-SHA384":        tls.TLS_DHE_RSA_WITH_AES_256_GCM_SHA384,
	"ECDHE-ECDSA-AES128-SHA": 						tls.TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA,
	"ECDHE-RSA-AES128-SHA":  							tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA,
	"ECDHE-ECDSA-AES256-SHA": 						tls.TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA,
	"ECDHE-RSA-AES256-SHA":   						tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
	//"DHE-RSA-AES128-SHA256":            tls.TLS_DHE_RSA_WITH_AES_128_CBC_SHA256,
	//"DHE-RSA-AES128-SHA":               tls.TLS_DHE_RSA_WITH_AES_128_CBC_SHA,
	//"DHE-RSA-AES256-SHA256":            tls.TLS_DHE_RSA_WITH_AES_256_CBC_SHA256,
	//"DHE-RSA-AES256-SHA":               tls.TLS_DHE_RSA_WITH_AES_256_CBC_SHA,
	//"ECDHE-ECDSA-DES-CBC3-SHA":         tls.TLS_ECDHE_ECDSA_WITH_3DES_EDE_CBC_SHA,
	"ECDHE-RSA-DES-CBC3-SHA": 						tls.TLS_ECDHE_RSA_WITH_3DES_EDE_CBC_SHA,
	//"EDH-RSA-DES-CBC3-SHA":             tls.TLS_DHE_RSA_WITH_3DES_EDE_CBC_SHA,
	"AES128-GCM-SHA256": 									tls.TLS_RSA_WITH_AES_128_GCM_SHA256,
	"AES256-GCM-SHA384": 									tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
	//"AES128-SHA256":                    tls.TLS_RSA_WITH_AES_128_CBC_SHA256,
	//"AES256-SHA256":                    tls.TLS_RSA_WITH_AES_256_CBC_SHA256,
	"AES128-SHA":   											tls.TLS_RSA_WITH_AES_128_CBC_SHA,
	"AES256-SHA":   											tls.TLS_RSA_WITH_AES_256_CBC_SHA,
	"DES-CBC3-SHA": 											tls.TLS_RSA_WITH_3DES_EDE_CBC_SHA,

	//"DHE-DSS-AES128-GCM-SHA256":        tls.TLS_DHE_DSS_WITH_AES_128_GCM_SHA256,
	//"DHE-DSS-AES256-GCM-SHA384":        tls.TLS_DHE_DSS_WITH_AES_256_GCM_SHA384,
	//"DHE-DSS-AES128-SHA256":            tls.TLS_DHE_DSS_WITH_AES_128_CBC_SHA256,
	//"DHE-DSS-AES256-SHA256":            tls.TLS_DHE_DSS_WITH_AES_256_CBC_SHA256,
	//"DHE-DSS-AES128-SHA":               tls.TLS_DHE_DSS_WITH_AES_128_CBC_SHA,
	//"ECDHE-RSA-CAMELLIA256-SHA384":     tls.TLS_ECDHE_RSA_WITH_CAMELLIA_256_CBC_SHA384,
	//"ECDHE-ECDSA-CAMELLIA256-SHA384":
	//"DHE-RSA-CAMELLIA256-SHA256":
	//"DHE-DSS-CAMELLIA256-SHA256":
	//"DHE-RSA-CAMELLIA256-SHA":          tls.TLS_DHE_RSA_WITH_CAMELLIA_256_CBC_SHA,
	//"DHE-DSS-CAMELLIA256-SHA":          tls.TLS_DHE_DSS_WITH_CAMELLIA_256_CBC_SHA,
	//"CAMELLIA256-SHA256":
	//"CAMELLIA256-SHA":                  tls.TLS_RSA_WITH_CAMELLIA_256_CBC_SHA,
	//"ECDHE-RSA-CAMELLIA128-SHA256":     tls.TLS_ECDHE_RSA_WITH_CAMELLIA_128_CBC_SHA256,
	//"ECDHE-ECDSA-CAMELLIA128-SHA256":   tls.TLS_ECDHE_ECDSA_WITH_CAMELLIA_128_CBC_SHA256,
	//"DHE-RSA-CAMELLIA128-SHA256":       tls.TLS_DHE_RSA_WITH_CAMELLIA_128_CBC_SHA256,
	//"DHE-DSS-CAMELLIA128-SHA256":       tls.TLS_DHE_DSS_WITH_CAMELLIA_128_CBC_SHA256,
	//"DHE-RSA-CAMELLIA128-SHA":          tls.TLS_DHE_RSA_WITH_CAMELLIA_128_CBC_SHA,
	//"DHE-DSS-CAMELLIA128-SHA":          tls.TLS_DHE_DSS_WITH_CAMELLIA_128_CBC_SHA,
	//"CAMELLIA128-SHA256":               tls.TLS_RSA_WITH_CAMELLIA_128_CBC_SHA256,
	//"CAMELLIA128-SHA":                  tls.TLS_RSA_WITH_CAMELLIA_128_CBC_SHA,
	//"DHE-RSA-SEED-SHA":                 tls.TLS_DHE_RSA_WITH_SEED_CBC_SHA,
	//"DHE-DSS-SEED-SHA":                 tls.TLS_DHE_DSS_WITH_SEED_CBC_SHA,
	//"SEED-SHA":                         tls.TLS_RSA_WITH_SEED_CBC_SHA,

}

// TLS Versios const indexed by strings
var tlsVersions = map[string]uint16{
	"TLSv1.2": tls.VersionTLS12,
	"TLSv1.1": tls.VersionTLS11,
	"TLSv1":   tls.VersionTLS10,
	"SSLv3":   tls.VersionSSL30,
}

func ApplyMozillaRecommendedTLSConfig(tlsConfig *tls.Config, configurationName string) error {
	mozConfigs, err := serverSideTLSConfiguration()
	if err != nil {
		return err
	}
	mozConfig, ok := mozConfigs.Configurations[configurationName]
	if !ok {
		return errors.New("Could not find Mozilla configuration with name : " + configurationName)
	}
	log.Infof("Applying Mozilla recommended configuration %s, version %f", configurationName, mozConfigs.Version)

	// Ciphers
	tlsConfig.PreferServerCipherSuites = true
	tlsConfig.CipherSuites = make([]uint16, 0)
	for _, cipher := range mozConfig.Ciphersuites {
		if cipherConst, exists := cipherSuites[cipher]; exists {
			log.Debugf("Adding cipher with name : %s", cipher)
			tlsConfig.CipherSuites = append(tlsConfig.CipherSuites, cipherConst)
		} else {
			log.Debugf("Could not find cipher with name : %s, skipping it", cipher)
		}
	}
	if len(tlsConfig.CipherSuites) == 0 {
		return errors.New("No recommended cipher could be found")
	}

	// TLS Version
	tlsMinVersion := mozConfig.TLSVersions[len(mozConfig.TLSVersions)-1]
	tlsMaxVersion := mozConfig.TLSVersions[0]
	log.Debugf("TLS minimum version : %s", tlsMinVersion)
	tlsConfig.MinVersion = tlsVersions[tlsMinVersion]
	log.Debugf("TLS maximum version : %s", tlsMaxVersion)
	tlsConfig.MaxVersion = tlsVersions[tlsMaxVersion]

	return nil
}

// Code snippets copied from https://github.com/mozilla/tls-observatory/tree/master/worker/mozillaEvaluationWorker
// Ideally we'd like to reuse code from this package but evrything is private
// Copied for POC purpose and see if we can do something cleaner LoadCertificateForDomains

var sstlsURL = "https://statics.tls.security.mozilla.org/server-side-tls-conf.json"
var sstls ServerSideTLSJson

func serverSideTLSConfiguration() (ServerSideTLSJson, error) {
	err := getConffromURL(sstlsURL)
	if err != nil {
		log.Error(err)
		log.Error("Could not get tls confs from url - fallback to locally saved configurations")
		// Try to continue with the locally hosted TLS configurations
		err = json.Unmarshal([]byte(ServerSideTLSConfiguration), &sstls)
		if err != nil {
			log.Error(err)
			log.Error("Could not load Server Side TLS configuration. Evaluation Worker not available")
		}
	}
	return sstls, err
}

// getConffromURL retrieves the json containing the TLS configurations from the specified URL.
func getConffromURL(url string) error {
	log.Infof("Downloading Mozilla TLS configuration from %s", url)
	r, err := http.Get(url)
	if err != nil {
		return err
	}
	defer r.Body.Close()

	err = json.NewDecoder(r.Body).Decode(&sstls)
	if err != nil {
		return err
	}

	return nil
}

// ServerSideTLSJson contains all the available configurations and the version of the current document.
type ServerSideTLSJson struct {
	Configurations map[string]Configuration `json:"configurations"`
	Version        float64                  `json:"version"`
}

// Configuration represents configurations levels declared by the Mozilla server-side-tls
// see https://wiki.mozilla.org/Security/Server_Side_TLS
type Configuration struct {
	OpenSSLCiphersuites   string   `json:"openssl_ciphersuites"`
	Ciphersuites          []string `json:"ciphersuites"`
	TLSVersions           []string `json:"tls_versions"`
	TLSCurves             []string `json:"tls_curves"`
	CertificateTypes      []string `json:"certificate_types"`
	CertificateCurves     []string `json:"certificate_curves"`
	CertificateSignatures []string `json:"certificate_signatures"`
	RsaKeySize            float64  `json:"rsa_key_size"`
	DHParamSize           float64  `json:"dh_param_size"`
	ECDHParamSize         float64  `json:"ecdh_param_size"`
	HstsMinAge            float64  `json:"hsts_min_age"`
	OldestClients         []string `json:"oldest_clients"`
}

// EvaluationResults contains the results of the mozillaEvaluationWorker
type EvaluationResults struct {
	Level    string              `json:"level"`
	Failures map[string][]string `json:"failures"`
}

// saved TLS configurations, used as backup if
// online version is not available
var ServerSideTLSConfiguration = `{
    "configurations": {
        "modern": {
            "openssl_ciphersuites": "ECDHE-ECDSA-AES256-GCM-SHA384:ECDHE-RSA-AES256-GCM-SHA384:ECDHE-ECDSA-CHACHA20-POLY1305:ECDHE-RSA-CHACHA20-POLY1305:ECDHE-ECDSA-AES128-GCM-SHA256:ECDHE-RSA-AES128-GCM-SHA256:ECDHE-ECDSA-AES256-SHA384:ECDHE-RSA-AES256-SHA384:ECDHE-ECDSA-AES128-SHA256:ECDHE-RSA-AES128-SHA256",
            "ciphersuites": [
                "ECDHE-ECDSA-AES256-GCM-SHA384",
                "ECDHE-RSA-AES256-GCM-SHA384",
                "ECDHE-ECDSA-CHACHA20-POLY1305",
                "ECDHE-RSA-CHACHA20-POLY1305",
                "ECDHE-ECDSA-AES128-GCM-SHA256",
                "ECDHE-RSA-AES128-GCM-SHA256",
                "ECDHE-ECDSA-AES256-SHA384",
                "ECDHE-RSA-AES256-SHA384",
                "ECDHE-ECDSA-AES128-SHA256",
                "ECDHE-RSA-AES128-SHA256"
            ],
            "tls_versions": ["TLSv1.2" ],
            "tls_curves": [ "prime256v1", "secp384r1", "secp521r1" ],
            "certificate_types": ["ecdsa"],
            "certificate_curves": ["prime256v1", "secp384r1", "secp521r1"],
            "certificate_signatures": ["sha256WithRSAEncryption", "ecdsa-with-SHA256", "ecdsa-with-SHA384", "ecdsa-with-SHA512"],
            "rsa_key_size": 2048,
            "dh_param_size": null,
            "ecdh_param_size": 256,
            "hsts_min_age": 15768000,
            "oldest_clients": [ "Firefox 27", "Chrome 30", "IE 11 on Windows 7", "Edge 1", "Opera 17", "Safari 9", "Android 5.0", "Java 8"]
        },
        "intermediate": {
            "openssl_ciphersuites": "ECDHE-ECDSA-CHACHA20-POLY1305:ECDHE-RSA-CHACHA20-POLY1305:ECDHE-ECDSA-AES128-GCM-SHA256:ECDHE-RSA-AES128-GCM-SHA256:ECDHE-ECDSA-AES256-GCM-SHA384:ECDHE-RSA-AES256-GCM-SHA384:DHE-RSA-AES128-GCM-SHA256:DHE-RSA-AES256-GCM-SHA384:ECDHE-ECDSA-AES128-SHA256:ECDHE-RSA-AES128-SHA256:ECDHE-ECDSA-AES128-SHA:ECDHE-RSA-AES256-SHA384:ECDHE-RSA-AES128-SHA:ECDHE-ECDSA-AES256-SHA384:ECDHE-ECDSA-AES256-SHA:ECDHE-RSA-AES256-SHA:DHE-RSA-AES128-SHA256:DHE-RSA-AES128-SHA:DHE-RSA-AES256-SHA256:DHE-RSA-AES256-SHA:ECDHE-ECDSA-DES-CBC3-SHA:ECDHE-RSA-DES-CBC3-SHA:EDH-RSA-DES-CBC3-SHA:AES128-GCM-SHA256:AES256-GCM-SHA384:AES128-SHA256:AES256-SHA256:AES128-SHA:AES256-SHA:DES-CBC3-SHA:!DSS",
            "ciphersuites": [
                "ECDHE-ECDSA-CHACHA20-POLY1305",
                "ECDHE-RSA-CHACHA20-POLY1305",
                "ECDHE-ECDSA-AES128-GCM-SHA256",
                "ECDHE-RSA-AES128-GCM-SHA256",
                "ECDHE-ECDSA-AES256-GCM-SHA384",
                "ECDHE-RSA-AES256-GCM-SHA384",
                "DHE-RSA-AES128-GCM-SHA256",
                "DHE-RSA-AES256-GCM-SHA384",
                "ECDHE-ECDSA-AES128-SHA256",
                "ECDHE-RSA-AES128-SHA256",
                "ECDHE-ECDSA-AES128-SHA",
                "ECDHE-RSA-AES256-SHA384",
                "ECDHE-RSA-AES128-SHA",
                "ECDHE-ECDSA-AES256-SHA384",
                "ECDHE-ECDSA-AES256-SHA",
                "ECDHE-RSA-AES256-SHA",
                "DHE-RSA-AES128-SHA256",
                "DHE-RSA-AES128-SHA",
                "DHE-RSA-AES256-SHA256",
                "DHE-RSA-AES256-SHA",
                "ECDHE-ECDSA-DES-CBC3-SHA",
                "ECDHE-RSA-DES-CBC3-SHA",
                "EDH-RSA-DES-CBC3-SHA",
                "AES128-GCM-SHA256",
                "AES256-GCM-SHA384",
                "AES128-SHA256",
                "AES256-SHA256",
                "AES128-SHA",
                "AES256-SHA",
                "DES-CBC3-SHA"
            ],
            "tls_versions": ["TLSv1.2", "TLSv1.1", "TLSv1" ],
            "tls_curves": [ "secp256r1", "secp384r1", "secp521r1" ],
            "certificate_types": ["rsa"],
            "certificate_curves": null,
            "certificate_signatures": ["sha256WithRSAEncryption"],
            "rsa_key_size": 2048,
            "dh_param_size": 2048,
            "ecdh_param_size": 256,
            "hsts_min_age": 15768000,
            "oldest_clients": [ "Firefox 1", "Chrome 1", "IE 7", "Opera 5", "Safari 1", "Windows XP IE8", "Android 2.3", "Java 7" ]
        },
        "old": {
            "openssl_ciphersuites": "ECDHE-ECDSA-CHACHA20-POLY1305:ECDHE-RSA-CHACHA20-POLY1305:ECDHE-RSA-AES128-GCM-SHA256:ECDHE-ECDSA-AES128-GCM-SHA256:ECDHE-RSA-AES256-GCM-SHA384:ECDHE-ECDSA-AES256-GCM-SHA384:DHE-RSA-AES128-GCM-SHA256:DHE-DSS-AES128-GCM-SHA256:kEDH+AESGCM:ECDHE-RSA-AES128-SHA256:ECDHE-ECDSA-AES128-SHA256:ECDHE-RSA-AES128-SHA:ECDHE-ECDSA-AES128-SHA:ECDHE-RSA-AES256-SHA384:ECDHE-ECDSA-AES256-SHA384:ECDHE-RSA-AES256-SHA:ECDHE-ECDSA-AES256-SHA:DHE-RSA-AES128-SHA256:DHE-RSA-AES128-SHA:DHE-DSS-AES128-SHA256:DHE-RSA-AES256-SHA256:DHE-DSS-AES256-SHA:DHE-RSA-AES256-SHA:ECDHE-RSA-DES-CBC3-SHA:ECDHE-ECDSA-DES-CBC3-SHA:EDH-RSA-DES-CBC3-SHA:AES128-GCM-SHA256:AES256-GCM-SHA384:AES128-SHA256:AES256-SHA256:AES128-SHA:AES256-SHA:AES:DES-CBC3-SHA:HIGH:SEED:!aNULL:!eNULL:!EXPORT:!DES:!RC4:!MD5:!PSK:!RSAPSK:!aDH:!aECDH:!EDH-DSS-DES-CBC3-SHA:!KRB5-DES-CBC3-SHA:!SRP",
            "ciphersuites": [
                "ECDHE-ECDSA-CHACHA20-POLY1305",
                "ECDHE-RSA-CHACHA20-POLY1305",
                "ECDHE-RSA-AES128-GCM-SHA256",
                "ECDHE-ECDSA-AES128-GCM-SHA256",
                "ECDHE-RSA-AES256-GCM-SHA384",
                "ECDHE-ECDSA-AES256-GCM-SHA384",
                "DHE-RSA-AES128-GCM-SHA256",
                "DHE-DSS-AES128-GCM-SHA256",
                "DHE-DSS-AES256-GCM-SHA384",
                "DHE-RSA-AES256-GCM-SHA384",
                "ECDHE-RSA-AES128-SHA256",
                "ECDHE-ECDSA-AES128-SHA256",
                "ECDHE-RSA-AES128-SHA",
                "ECDHE-ECDSA-AES128-SHA",
                "ECDHE-RSA-AES256-SHA384",
                "ECDHE-ECDSA-AES256-SHA384",
                "ECDHE-RSA-AES256-SHA",
                "ECDHE-ECDSA-AES256-SHA",
                "DHE-RSA-AES128-SHA256",
                "DHE-RSA-AES128-SHA",
                "DHE-DSS-AES128-SHA256",
                "DHE-RSA-AES256-SHA256",
                "DHE-DSS-AES256-SHA",
                "DHE-RSA-AES256-SHA",
                "ECDHE-RSA-DES-CBC3-SHA",
                "ECDHE-ECDSA-DES-CBC3-SHA",
                "EDH-RSA-DES-CBC3-SHA",
                "AES128-GCM-SHA256",
                "AES256-GCM-SHA384",
                "AES128-SHA256",
                "AES256-SHA256",
                "AES128-SHA",
                "AES256-SHA",
                "DHE-DSS-AES256-SHA256",
                "DHE-DSS-AES128-SHA",
                "DES-CBC3-SHA",
                "DHE-RSA-CHACHA20-POLY1305",
                "ECDHE-RSA-CAMELLIA256-SHA384",
                "ECDHE-ECDSA-CAMELLIA256-SHA384",
                "DHE-RSA-CAMELLIA256-SHA256",
                "DHE-DSS-CAMELLIA256-SHA256",
                "DHE-RSA-CAMELLIA256-SHA",
                "DHE-DSS-CAMELLIA256-SHA",
                "CAMELLIA256-SHA256",
                "CAMELLIA256-SHA",
                "ECDHE-RSA-CAMELLIA128-SHA256",
                "ECDHE-ECDSA-CAMELLIA128-SHA256",
                "DHE-RSA-CAMELLIA128-SHA256",
                "DHE-DSS-CAMELLIA128-SHA256",
                "DHE-RSA-CAMELLIA128-SHA",
                "DHE-DSS-CAMELLIA128-SHA",
                "CAMELLIA128-SHA256",
                "CAMELLIA128-SHA",
                "DHE-RSA-SEED-SHA",
                "DHE-DSS-SEED-SHA",
                "SEED-SHA"
            ],
            "tls_versions": ["TLSv1.2", "TLSv1.1", "TLSv1", "SSLv3" ],
            "tls_curves": [ "secp256r1", "secp384r1", "secp521r1" ],
            "certificate_types": ["rsa"],
            "certificate_curves": null,
            "certificate_signatures": ["sha1WithRSAEncryption"],
            "rsa_key_size": 2048,
            "dh_param_size": 1024,
            "ecdh_param_size": 160,
            "hsts_min_age": 15768000,
            "oldest_clients": [ "Firefox 1", "Chrome 1", "Windows XP IE 6", "Opera 4", "Safari 1", "Java 6" ]
        }
    },
    "version": 4.0
}`
