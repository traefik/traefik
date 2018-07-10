/*
Copyright 2017 The go-marathon Authors All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package marathon

import (
	"encoding/json"
	"fmt"
)

// PodAlias aliases the Pod struct so that it will be marshaled/unmarshaled automatically
type PodAlias Pod

// UnmarshalJSON unmarshals the given Pod JSON as expected except for environment variables and secrets.
// Environment variables are stored in the Env field. Secrets, including the environment variable part,
// are stored in the Secrets field.
func (p *Pod) UnmarshalJSON(b []byte) error {
	aux := &struct {
		*PodAlias
		Env     map[string]interface{} `json:"environment"`
		Secrets map[string]TmpSecret   `json:"secrets"`
	}{
		PodAlias: (*PodAlias)(p),
	}
	if err := json.Unmarshal(b, aux); err != nil {
		return fmt.Errorf("malformed pod definition %v", err)
	}
	env := map[string]string{}
	secrets := map[string]Secret{}

	for envName, genericEnvValue := range aux.Env {
		switch envValOrSecret := genericEnvValue.(type) {
		case string:
			env[envName] = envValOrSecret
		case map[string]interface{}:
			for secret, secretStore := range envValOrSecret {
				if secStore, ok := secretStore.(string); ok && secret == "secret" {
					secrets[secStore] = Secret{EnvVar: envName}
					break
				}
				return fmt.Errorf("unexpected secret field %v of value type %T", secret, envValOrSecret[secret])
			}
		default:
			return fmt.Errorf("unexpected environment variable type %T", envValOrSecret)
		}
	}
	p.Env = env
	for k, v := range aux.Secrets {
		tmp := secrets[k]
		tmp.Source = v.Source
		secrets[k] = tmp
	}
	p.Secrets = secrets
	return nil
}

// MarshalJSON marshals the given Pod as expected except for environment variables and secrets,
// which are marshaled from specialized structs.  The environment variable piece of the secrets and other
// normal environment variables are combined and marshaled to the env field.  The secrets and the related
// source are marshaled into the secrets field.
func (p *Pod) MarshalJSON() ([]byte, error) {
	env := make(map[string]interface{})
	secrets := make(map[string]TmpSecret)

	if p.Env != nil {
		for k, v := range p.Env {
			env[string(k)] = string(v)
		}
	}
	if p.Secrets != nil {
		for k, v := range p.Secrets {
			// Only add it to the root level pod environment if it's used
			// Otherwise it's likely in one of the container environments
			if v.EnvVar != "" {
				env[v.EnvVar] = TmpEnvSecret{Secret: k}
			}
			secrets[k] = TmpSecret{v.Source}
		}
	}
	aux := &struct {
		*PodAlias
		Env     map[string]interface{} `json:"environment,omitempty"`
		Secrets map[string]TmpSecret   `json:"secrets,omitempty"`
	}{PodAlias: (*PodAlias)(p), Env: env, Secrets: secrets}

	return json.Marshal(aux)
}
