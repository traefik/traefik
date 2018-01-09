package model

// // GetQPSStats returns current queries per second (QPS) for the account
// func (c APIClient) GetQPSStats() (v float64, err error) {
// 	var s map[string]float64
// 	_, err = c.doHTTPUnmarshal("GET", "https://api.nsone.net/v1/stats/qps", nil, &s)
// 	if err != nil {
// 		return v, err
// 	}
// 	v, found := s["qps"]
// 	if !found {
// 		return v, errors.New("Could not find 'qps' key in returned data")
// 	}
// 	return v, nil
// }
