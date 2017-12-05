package servicefabric

type queryParamsFunc func(params []string) []string

func withContinue(token string) queryParamsFunc {
	if len(token) == 0 {
		return noOp
	}
	return withParam("continue", token)
}

func withParam(name, value string) queryParamsFunc {
	return func(params []string) []string {
		return append(params, name+"="+value)
	}
}

func noOp(params []string) []string {
	return params
}
