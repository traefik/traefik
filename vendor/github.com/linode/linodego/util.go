package linodego

import "time"

const (
	dateLayout = "2006-01-02T15:04:05"
)

func parseDates(dateStr string) (*time.Time, error) {
	d, err := time.Parse(dateLayout, dateStr)
	if err != nil {
		return nil, err
	}
	return &d, nil
}
