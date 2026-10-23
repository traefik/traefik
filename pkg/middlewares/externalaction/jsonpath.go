package externalaction

import (
	"strconv"
	"strings"
)

func selectJSON(value any, path string) any {
	tokens, ok := parseJSONPath(path)
	if !ok {
		return nil
	}
	if len(tokens) == 0 {
		return value
	}

	current := value
	for _, token := range tokens {
		if token.index == nil {
			object, ok := current.(map[string]any)
			if !ok {
				return nil
			}
			current = object[token.key]
		} else {
			array, ok := current.([]any)
			if !ok || *token.index < 0 || *token.index >= len(array) {
				return nil
			}
			current = array[*token.index]
		}
	}
	return current
}

type jsonPathToken struct {
	key   string
	index *int
}

func parseJSONPath(value string) ([]jsonPathToken, bool) {
	normalizedPath := strings.TrimSpace(value)
	if normalizedPath == "" || normalizedPath == "$" {
		return nil, true
	}
	if normalizedPath[0] == '$' {
		normalizedPath = normalizedPath[1:]
	}

	var result []jsonPathToken
	for position := 0; position < len(normalizedPath); {
		switch normalizedPath[position] {
		case '.':
			position++
			if position == len(normalizedPath) || normalizedPath[position] == '.' || normalizedPath[position] == '[' {
				return nil, false
			}
		case '[':
			token, next, ok := parseJSONPathBracket(normalizedPath, position)
			if !ok {
				return nil, false
			}
			result = append(result, token)
			position = next
			continue
		}

		start := position
		for position < len(normalizedPath) && normalizedPath[position] != '.' && normalizedPath[position] != '[' {
			position++
		}
		if start == position {
			return nil, false
		}
		result = append(result, jsonPathToken{key: normalizedPath[start:position]})
	}
	return result, true
}

func parseJSONPathBracket(value string, position int) (jsonPathToken, int, bool) {
	position++
	if position >= len(value) {
		return jsonPathToken{}, 0, false
	}

	if value[position] == '"' || value[position] == '\'' {
		quote := value[position]
		position++
		var key strings.Builder
		for position < len(value) {
			character := value[position]
			if character == quote {
				position++
				if position >= len(value) || value[position] != ']' {
					return jsonPathToken{}, 0, false
				}
				return jsonPathToken{key: key.String()}, position + 1, true
			}
			if character == '\\' {
				position++
				if position >= len(value) {
					return jsonPathToken{}, 0, false
				}
				escaped := value[position]
				switch escaped {
				case '\\', '"', '\'':
					key.WriteByte(escaped)
				case 'n':
					key.WriteByte('\n')
				case 'r':
					key.WriteByte('\r')
				case 't':
					key.WriteByte('\t')
				default:
					return jsonPathToken{}, 0, false
				}
				position++
				continue
			}
			key.WriteByte(character)
			position++
		}
		return jsonPathToken{}, 0, false
	}

	start := position
	for position < len(value) && value[position] >= '0' && value[position] <= '9' {
		position++
	}
	if start == position || position >= len(value) || value[position] != ']' {
		return jsonPathToken{}, 0, false
	}
	index, err := strconv.Atoi(value[start:position])
	if err != nil {
		return jsonPathToken{}, 0, false
	}
	return jsonPathToken{index: &index}, position + 1, true
}
