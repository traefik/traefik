package auth

import (
	"fmt"
	"strings"

	"github.com/containous/traefik/types"
)

func parserBasicUsers(basic *types.Basic) (map[string]string, error) {
	var userStrs []string
	if basic.UsersFile != "" {
		var err error
		if userStrs, err = getLinesFromFile(basic.UsersFile); err != nil {
			return nil, err
		}
	}
	userStrs = append(basic.Users, userStrs...)
	userMap := make(map[string]string)
	for _, user := range userStrs {
		split := strings.Split(user, ":")
		if len(split) != 2 {
			return nil, fmt.Errorf("error parsing Authenticator user: %v", user)
		}
		userMap[split[0]] = split[1]
	}
	return userMap, nil
}

func parserDigestUsers(digest *types.Digest) (map[string]string, error) {
	var userStrs []string
	if digest.UsersFile != "" {
		var err error
		if userStrs, err = getLinesFromFile(digest.UsersFile); err != nil {
			return nil, err
		}
	}
	userStrs = append(digest.Users, userStrs...)
	userMap := make(map[string]string)
	for _, user := range userStrs {
		split := strings.Split(user, ":")
		if len(split) != 3 {
			return nil, fmt.Errorf("error parsing Authenticator user: %v", user)
		}
		userMap[split[0]+":"+split[1]] = split[2]
	}
	return userMap, nil
}
