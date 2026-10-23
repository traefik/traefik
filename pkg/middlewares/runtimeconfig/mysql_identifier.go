package runtimeconfig

import (
	"errors"
	"regexp"
)

var mysqlIdentifierPattern = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9_]{0,63}$`)

// QuoteMySQLIdentifier validates and quotes a table or column name.
// SQL placeholders bind values, not identifiers; never interpolate an
// unchecked identifier into a statement.
func QuoteMySQLIdentifier(value string) (string, error) {
	if !mysqlIdentifierPattern.MatchString(value) {
		return "", errors.New("mysql identifier must match " + mysqlIdentifierPattern.String())
	}
	return "`" + value + "`", nil
}
