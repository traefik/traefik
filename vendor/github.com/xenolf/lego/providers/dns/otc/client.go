package otc

type recordset struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Type        string   `json:"type"`
	TTL         int      `json:"ttl"`
	Records     []string `json:"records"`
}

type nameResponse struct {
	Name string `json:"name"`
}

type userResponse struct {
	Name     string       `json:"name"`
	Password string       `json:"password"`
	Domain   nameResponse `json:"domain"`
}

type passwordResponse struct {
	User userResponse `json:"user"`
}

type identityResponse struct {
	Methods  []string         `json:"methods"`
	Password passwordResponse `json:"password"`
}

type scopeResponse struct {
	Project nameResponse `json:"project"`
}

type authResponse struct {
	Identity identityResponse `json:"identity"`
	Scope    scopeResponse    `json:"scope"`
}

type loginResponse struct {
	Auth authResponse `json:"auth"`
}

type endpointResponse struct {
	Token struct {
		Catalog []struct {
			Type      string `json:"type"`
			Endpoints []struct {
				URL string `json:"url"`
			} `json:"endpoints"`
		} `json:"catalog"`
	} `json:"token"`
}

type zoneItem struct {
	ID string `json:"id"`
}

type zonesResponse struct {
	Zones []zoneItem `json:"zones"`
}

type recordSet struct {
	ID string `json:"id"`
}

type recordSetsResponse struct {
	RecordSets []recordSet `json:"recordsets"`
}
