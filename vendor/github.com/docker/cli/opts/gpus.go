package opts

import (
	"encoding/csv"
	"fmt"
	"strconv"
	"strings"

	"github.com/docker/docker/api/types/container"
	"github.com/pkg/errors"
)

// GpuOpts is a Value type for parsing mounts
type GpuOpts struct {
	values []container.DeviceRequest
}

func parseCount(s string) (int, error) {
	if s == "all" {
		return -1, nil
	}
	i, err := strconv.Atoi(s)
	return i, errors.Wrap(err, "count must be an integer")
}

// Set a new mount value
// nolint: gocyclo
func (o *GpuOpts) Set(value string) error {
	csvReader := csv.NewReader(strings.NewReader(value))
	fields, err := csvReader.Read()
	if err != nil {
		return err
	}

	req := container.DeviceRequest{}

	seen := map[string]struct{}{}
	// Set writable as the default
	for _, field := range fields {
		parts := strings.SplitN(field, "=", 2)
		key := parts[0]
		if _, ok := seen[key]; ok {
			return fmt.Errorf("gpu request key '%s' can be specified only once", key)
		}
		seen[key] = struct{}{}

		if len(parts) == 1 {
			seen["count"] = struct{}{}
			req.Count, err = parseCount(key)
			if err != nil {
				return err
			}
			continue
		}

		value := parts[1]
		switch key {
		case "driver":
			req.Driver = value
		case "count":
			req.Count, err = parseCount(value)
			if err != nil {
				return err
			}
		case "device":
			req.DeviceIDs = strings.Split(value, ",")
		case "capabilities":
			req.Capabilities = [][]string{append(strings.Split(value, ","), "gpu")}
		case "options":
			r := csv.NewReader(strings.NewReader(value))
			optFields, err := r.Read()
			if err != nil {
				return errors.Wrap(err, "failed to read gpu options")
			}
			req.Options = ConvertKVStringsToMap(optFields)
		default:
			return fmt.Errorf("unexpected key '%s' in '%s'", key, field)
		}
	}

	if _, ok := seen["count"]; !ok && req.DeviceIDs == nil {
		req.Count = 1
	}
	if req.Options == nil {
		req.Options = make(map[string]string)
	}
	if req.Capabilities == nil {
		req.Capabilities = [][]string{{"gpu"}}
	}

	o.values = append(o.values, req)
	return nil
}

// Type returns the type of this option
func (o *GpuOpts) Type() string {
	return "gpu-request"
}

// String returns a string repr of this option
func (o *GpuOpts) String() string {
	gpus := []string{}
	for _, gpu := range o.values {
		gpus = append(gpus, fmt.Sprintf("%v", gpu))
	}
	return strings.Join(gpus, ", ")
}

// Value returns the mounts
func (o *GpuOpts) Value() []container.DeviceRequest {
	return o.values
}
