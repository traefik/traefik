package edgegrid

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/go-ini/ini"
	"github.com/mitchellh/go-homedir"
)

// Config struct provides all the necessary fields to
// create authorization header, debug is optional
type Config struct {
	Host         string   `ini:"host"`
	ClientToken  string   `ini:"client_token"`
	ClientSecret string   `ini:"client_secret"`
	AccessToken  string   `ini:"access_token"`
	HeaderToSign []string `ini:"headers_to_sign"`
	MaxBody      int      `ini:"max_body"`
	Debug        bool     `ini:"debug"`
}

// Init initializes by first attempting to use ENV vars, with .edgerc as a fallback
//
// See: InitEnv()
// See: InitEdgeRc()
func Init(filepath string, section string) (Config, error) {
	if section == "" {
		section = defaultSection
	} else {
		section = strings.ToUpper(section)
	}

	_, exists := os.LookupEnv("AKAMAI_" + section + "_HOST")
	if !exists && section == defaultSection {
		_, exists := os.LookupEnv("AKAMAI_HOST")

		if exists {
			return InitEnv("")
		}
	}

	if exists {
		return InitEnv(section)
	}

	c, err := InitEdgeRc(filepath, strings.ToLower(section))

	if err == nil {
		return c, nil
	}

	if section != defaultSection {
		_, ok := os.LookupEnv("AKAMAI_HOST")
		if ok {
			return InitEnv("")
		}
	}

	return c, fmt.Errorf("Unable to create instance using environment or .edgerc file")
}

// InitEdgeRc initializes using a configuration file in standard INI format
//
// By default, it uses the .edgerc found in the users home directory, and the
// "default" section.
func InitEdgeRc(filepath string, section string) (Config, error) {
	var (
		c               Config
		requiredOptions = []string{"host", "client_token", "client_secret", "access_token"}
		missing         []string
	)

	// Check if filepath is empty
	if filepath == "" {
		filepath = "~/.edgerc"
	}

	// Check if section is empty
	if section == "" {
		section = "default"
	}

	// Tilde seems to be not working when passing ~/.edgerc as file
	// Takes current user and use home dir instead

	path, err := homedir.Expand(filepath)

	if err != nil {
		return c, fmt.Errorf(errorMap[ErrHomeDirNotFound], err)
	}

	edgerc, err := ini.Load(path)
	if err != nil {
		return c, fmt.Errorf(errorMap[ErrConfigFile], err)
	}
	err = edgerc.Section(section).MapTo(&c)
	if err != nil {
		return c, fmt.Errorf(errorMap[ErrConfigFileSection], err)
	}
	for _, opt := range requiredOptions {
		if !(edgerc.Section(section).HasKey(opt)) {
			missing = append(missing, opt)
		}
	}
	if len(missing) > 0 {
		return c, fmt.Errorf(errorMap[ErrConfigMissingOptions], missing)
	}
	if c.MaxBody == 0 {
		c.MaxBody = 131072
	}
	return c, nil
}

// InitEnv initializes using the Environment (ENV)
//
// By default, it uses AKAMAI_HOST, AKAMAI_CLIENT_TOKEN, AKAMAI_CLIENT_SECRET,
// AKAMAI_ACCESS_TOKEN, and AKAMAI_MAX_BODY variables.
//
// You can define multiple configurations by prefixing with the section name specified, e.g.
// passing "ccu" will cause it to look for AKAMAI_CCU_HOST, etc.
//
// If AKAMAI_{SECTION} does not exist, it will fall back to just AKAMAI_.
func InitEnv(section string) (Config, error) {
	var (
		c               Config
		requiredOptions = []string{"HOST", "CLIENT_TOKEN", "CLIENT_SECRET", "ACCESS_TOKEN"}
		missing         []string
		prefix          string
	)

	// Check if section is empty
	if section == "" {
		section = defaultSection
	} else {
		section = strings.ToUpper(section)
	}

	prefix = "AKAMAI_"
	_, ok := os.LookupEnv("AKAMAI_" + section + "_HOST")
	if ok {
		prefix = "AKAMAI_" + section + "_"
	}

	for _, opt := range requiredOptions {
		val, ok := os.LookupEnv(prefix + opt)
		if !ok {
			missing = append(missing, prefix+opt)
		} else {
			switch {
			case opt == "HOST":
				c.Host = val
			case opt == "CLIENT_TOKEN":
				c.ClientToken = val
			case opt == "CLIENT_SECRET":
				c.ClientSecret = val
			case opt == "ACCESS_TOKEN":
				c.AccessToken = val
			}
		}
	}

	if len(missing) > 0 {
		return c, fmt.Errorf(errorMap[ErrMissingEnvVariables], missing)
	}

	c.MaxBody = 0

	val, ok := os.LookupEnv(prefix + "MAX_BODY")
	if i, err := strconv.Atoi(val); err == nil {
		c.MaxBody = i
	}

	if !ok || c.MaxBody == 0 {
		c.MaxBody = 131072
	}

	return c, nil
}
