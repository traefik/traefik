package config

func merge(existing, value interface{}) interface{} {
	// append strings
	if left, lok := existing.([]interface{}); lok {
		if right, rok := value.([]interface{}); rok {
			return append(left, right...)
		}
	}

	//merge maps
	if left, lok := existing.(map[interface{}]interface{}); lok {
		if right, rok := value.(map[interface{}]interface{}); rok {
			newLeft := make(map[interface{}]interface{})
			for k, v := range left {
				newLeft[k] = v
			}
			for k, v := range right {
				newLeft[k] = v
			}
			return newLeft
		}
	}

	return value
}

func clone(in RawService) RawService {
	result := RawService{}
	for k, v := range in {
		result[k] = v
	}

	return result
}

func asString(obj interface{}) string {
	if v, ok := obj.(string); ok {
		return v
	}
	return ""
}
