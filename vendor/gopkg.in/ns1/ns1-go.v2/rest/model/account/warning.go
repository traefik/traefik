package account

// UsageWarning wraps an NS1 /account/usagewarnings resource
type UsageWarning struct {
	Records Warning `json:"records"`
	Queries Warning `json:"queries"`
}

// Warning contains alerting toggles and thresholds for overage warning alert messages.
// First thresholds must be smaller than Second ones and all thresholds
// must be percentages between 0 and 100.
type Warning struct {
	Send bool `json:"send_warnings"`

	First  int `json:"warning_1"`
	Second int `json:"warning_2"`
}
