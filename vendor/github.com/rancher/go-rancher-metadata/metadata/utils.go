package metadata

import (
	"time"
)

func testConnection(mdClient Client) error {
	var err error
	maxTime := 20 * time.Second

	for i := 1 * time.Second; i < maxTime; i *= time.Duration(2) {
		if _, err = mdClient.GetVersion(); err != nil {
			time.Sleep(i)
		} else {
			return nil
		}
	}
	return err
}
