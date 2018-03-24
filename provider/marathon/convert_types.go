package marathon

func Unp(mp *map[string]string) map[string]string {
	if mp != nil {
		return *mp
	}
	return make(map[string]string)
}
