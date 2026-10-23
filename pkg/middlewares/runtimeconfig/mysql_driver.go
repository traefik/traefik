package runtimeconfig

import (
	"errors"

	mysql "github.com/go-sql-driver/mysql"
)

func mysqlDriverConfig(dsn string) (*mysql.Config, error) {
	config, err := mysql.ParseDSN(dsn)
	if err != nil {
		// Driver parsing errors can include malformed DSN parameter values.
		return nil, errors.New("invalid MySQL DSN")
	}
	// Middleware statements never need batches, client-side interpolation or
	// arbitrary local-file uploads. Enforce these boundaries on every rotation.
	config.MultiStatements = false
	config.InterpolateParams = false
	config.AllowAllFiles = false
	return config, nil
}
