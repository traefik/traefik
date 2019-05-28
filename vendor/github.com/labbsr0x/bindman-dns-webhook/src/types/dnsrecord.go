package types

import (
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
)

// DNSRecord defines what we understand as a DNSRecord
type DNSRecord struct {
	// Name the DNS host name
	Name string `json:"name"`

	// Value the value of this record
	Value string `json:"value"`

	// Type the record type
	Type string `json:"type"`
}

// Check verifies if the DNS record satisfies certain conditions
func (record *DNSRecord) Check() []string {
	logrus.Infof("Record to check: '%v'", record)
	emptyValueErrorMessage := "the value of field '%s' cannot be empty"
	var errs []string

	if strings.TrimSpace(record.Name) == "" {
		errs = append(errs, fmt.Sprintf(emptyValueErrorMessage, "name"))
	}

	if strings.TrimSpace(record.Value) == "" {
		errs = append(errs, fmt.Sprintf(emptyValueErrorMessage, "value"))
	}

	if strings.TrimSpace(record.Type) == "" {
		errs = append(errs, fmt.Sprintf(emptyValueErrorMessage, "type"))
	}
	return errs
}
