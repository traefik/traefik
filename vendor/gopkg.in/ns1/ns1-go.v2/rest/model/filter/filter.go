package filter

// Filter wraps the values of a Record's "filters" attribute
type Filter struct {
	Type     string `json:"filter"`
	Disabled bool   `json:"disabled,omitempty"`
	Config   Config `json:"config"`
}

// Enable a filter.
func (f *Filter) Enable() {
	f.Disabled = false
}

// Disable a filter.
func (f *Filter) Disable() {
	f.Disabled = true
}

// Config is a flat mapping where values are simple (no slices/maps).
type Config map[string]interface{}

// NewSelFirstN returns a filter that eliminates all but the
// first N answers from the list.
func NewSelFirstN(n int) *Filter {
	return &Filter{
		Type:   "select_first_n",
		Config: Config{"N": n},
	}
}

// NewShuffle returns a filter that randomly sorts the answers.
func NewShuffle() *Filter {
	return &Filter{Type: "shuffle", Config: Config{}}
}

// GEOGRAPHICAL FILTERS

// NewSelFirstRegion returns a filter that keeps only the answers
// that are in the same region as the first answer.
func NewSelFirstRegion() *Filter {
	return &Filter{Type: "select_first_n", Config: Config{}}
}

// NewStickyRegion first sorts regions uniquely depending on the IP
// address of the requester, and then groups all answers together by
// region. The same requester always gets the same ordering of regions,
// but answers within each region may be in any order. byNetwork indicates
// whether to apply the 'stickyness' by subnet(not individual IP).
func NewStickyRegion(byNetwork bool) *Filter {
	return &Filter{
		Type:   "sticky_region",
		Config: Config{"sticky_by_network": byNetwork},
	}
}

// NewGeofenceCountry returns a filter that fences using "country",
// "us_state", and "ca_province" metadata fields in answers. Only
// answers in the same country/state/province as the user (or
// answers with no specified location) are returned. rmNoLoc determines
// whether to remove answers without location on any match.
func NewGeofenceCountry(rmNoLoc bool) *Filter {
	return &Filter{
		Type:   "geofence_country",
		Config: Config{"remove_no_location": rmNoLoc},
	}
}

// NewGeofenceRegional returns a filter that restricts to answers in
// same geographical region as requester. rmNoGeo determines whether
// to remove answers without georegion on any match.
func NewGeofenceRegional(rmNoGeo bool) *Filter {
	return &Filter{
		Type:   "geofence_regional",
		Config: Config{"remove_no_georegion": rmNoGeo},
	}
}

// NewGeotargetCountry returns a filter that sorts answers by distance
// to requester by country, US state, and/or Canadian province.
func NewGeotargetCountry() *Filter {
	return &Filter{Type: "geofence_country", Config: Config{}}
}

// NewGeotargetLatLong returns a filter that sorts answers by distance
// to user using lat/long.
func NewGeotargetLatLong() *Filter {
	return &Filter{Type: "geotarget_latlong", Config: Config{}}
}

// NewGeotargetRegional returns a filter that sorts answers by distance
// to user by geographical region.
func NewGeotargetRegional() *Filter {
	return &Filter{Type: "geotarget_regional", Config: Config{}}
}

// NETWORK FILTERS

// NewSticky returns a filter that sorts answers uniquely depending
// on the IP address of the requester. The same requester always
// gets the same ordering of answers. byNetwork indicates whether
// to apply the 'stickyness' by subnet(not individual IP).
func NewSticky(byNetwork bool) *Filter {
	return &Filter{
		Type:   "sticky",
		Config: Config{"sticky_by_network": byNetwork},
	}
}

// NewWeightedSticky returns a filter that shuffles answers randomly
// per-requester based on weight. byNetwork indicates whether to
// apply the 'stickyness' by subnet(not individual IP).
func NewWeightedSticky(byNetwork bool) *Filter {
	return &Filter{
		Type:   "weighted_sticky",
		Config: Config{"sticky_by_network": byNetwork},
	}
}

// NewIPv4PrefixShuffle returns a filter that randomly selects
// IPv4 addresses from prefix list. This filter can only be used
// A records. n is the number of IPs to randomly select per answer.
func NewIPv4PrefixShuffle(n int) *Filter {
	return &Filter{
		Type:   "ipv4_prefix_shuffle",
		Config: Config{"N": n},
	}
}

// NewNetfenceASN returns a filter that restricts to answers where
// the ASN of requester IP matches ASN list. rmNoASN determines
// whether to remove answers without asn list on any match.
func NewNetfenceASN(rmNoASN bool) *Filter {
	return &Filter{
		Type:   "netfence_asn",
		Config: Config{"remove_no_asn": rmNoASN},
	}
}

// NewNetfencePrefix returns a filter that restricts to answers where
// requester IP matches prefix list. rmNoIPPrefix determines
// whether to remove answers without ip prefixes on any match.
func NewNetfencePrefix(rmNoIPPrefix bool) *Filter {
	return &Filter{
		Type:   "netfence_prefix",
		Config: Config{"remove_no_ip_prefixes": rmNoIPPrefix},
	}
}

// STATUS FILTERS

// NewUp returns a filter that eliminates all answers where
// the 'up' metadata field is not true.
func NewUp() *Filter {
	return &Filter{Type: "up", Config: Config{}}
}

// NewPriority returns a filter that fails over according to
// prioritized answer tiers.
func NewPriority() *Filter {
	return &Filter{Type: "priority", Config: Config{}}
}

// NewShedLoad returns a filter that "sheds" traffic to answers
// based on load, using one of several load metrics. You must set
// values for low_watermark, high_watermark, and the configured
// load metric, for each answer you intend to subject to load
// shedding.
func NewShedLoad(metric string) *Filter {
	return &Filter{
		Type:   "shed_load",
		Config: Config{"metric": metric},
	}
}

// TRAFFIC FILTERS

// NewWeightedShuffle returns a filter that shuffles answers
// randomly based on their weight.
func NewWeightedShuffle() *Filter {
	return &Filter{Type: "weighted_shuffle", Config: Config{}}
}
