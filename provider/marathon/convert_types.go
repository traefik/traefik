package marathon

func stringValueMap(mp *map[string]string) map[string]string {
	if mp != nil {
		return *mp
	}
	return make(map[string]string)
}
