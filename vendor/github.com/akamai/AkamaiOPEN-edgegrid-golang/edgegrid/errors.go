package edgegrid

// Error constants
const (
	ErrUUIDGenerateFailed   = 500
	ErrHomeDirNotFound      = 501
	ErrConfigFile           = 502
	ErrConfigFileSection    = 503
	ErrConfigMissingOptions = 504
	ErrMissingEnvVariables  = 505
)

var (
	errorMap = map[int]string{
		ErrUUIDGenerateFailed:   "Generate UUID failed: %s",
		ErrHomeDirNotFound:      "Fatal could not find home dir from user: %s",
		ErrConfigFile:           "Fatal error edgegrid file: %s",
		ErrConfigFileSection:    "Could not map section: %s",
		ErrConfigMissingOptions: "Fatal missing required options: %s",
		ErrMissingEnvVariables:  "Fatal missing required environment variables: %s",
	}
)
