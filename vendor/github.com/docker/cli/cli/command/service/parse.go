package service

import (
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	swarmtypes "github.com/docker/docker/api/types/swarm"
	"github.com/docker/docker/client"
	"github.com/pkg/errors"
	"golang.org/x/net/context"
)

// ParseSecrets retrieves the secrets with the requested names and fills
// secret IDs into the secret references.
func ParseSecrets(client client.SecretAPIClient, requestedSecrets []*swarmtypes.SecretReference) ([]*swarmtypes.SecretReference, error) {
	secretRefs := make(map[string]*swarmtypes.SecretReference)
	ctx := context.Background()

	for _, secret := range requestedSecrets {
		if _, exists := secretRefs[secret.File.Name]; exists {
			return nil, errors.Errorf("duplicate secret target for %s not allowed", secret.SecretName)
		}
		secretRef := new(swarmtypes.SecretReference)
		*secretRef = *secret
		secretRefs[secret.File.Name] = secretRef
	}

	args := filters.NewArgs()
	for _, s := range secretRefs {
		args.Add("name", s.SecretName)
	}

	secrets, err := client.SecretList(ctx, types.SecretListOptions{
		Filters: args,
	})
	if err != nil {
		return nil, err
	}

	foundSecrets := make(map[string]string)
	for _, secret := range secrets {
		foundSecrets[secret.Spec.Annotations.Name] = secret.ID
	}

	addedSecrets := []*swarmtypes.SecretReference{}

	for _, ref := range secretRefs {
		id, ok := foundSecrets[ref.SecretName]
		if !ok {
			return nil, errors.Errorf("secret not found: %s", ref.SecretName)
		}

		// set the id for the ref to properly assign in swarm
		// since swarm needs the ID instead of the name
		ref.SecretID = id
		addedSecrets = append(addedSecrets, ref)
	}

	return addedSecrets, nil
}

// ParseConfigs retrieves the configs from the requested names and converts
// them to config references to use with the spec
func ParseConfigs(client client.ConfigAPIClient, requestedConfigs []*swarmtypes.ConfigReference) ([]*swarmtypes.ConfigReference, error) {
	configRefs := make(map[string]*swarmtypes.ConfigReference)
	ctx := context.Background()

	for _, config := range requestedConfigs {
		if _, exists := configRefs[config.File.Name]; exists {
			return nil, errors.Errorf("duplicate config target for %s not allowed", config.ConfigName)
		}

		configRef := new(swarmtypes.ConfigReference)
		*configRef = *config
		configRefs[config.File.Name] = configRef
	}

	args := filters.NewArgs()
	for _, s := range configRefs {
		args.Add("name", s.ConfigName)
	}

	configs, err := client.ConfigList(ctx, types.ConfigListOptions{
		Filters: args,
	})
	if err != nil {
		return nil, err
	}

	foundConfigs := make(map[string]string)
	for _, config := range configs {
		foundConfigs[config.Spec.Annotations.Name] = config.ID
	}

	addedConfigs := []*swarmtypes.ConfigReference{}

	for _, ref := range configRefs {
		id, ok := foundConfigs[ref.ConfigName]
		if !ok {
			return nil, errors.Errorf("config not found: %s", ref.ConfigName)
		}

		// set the id for the ref to properly assign in swarm
		// since swarm needs the ID instead of the name
		ref.ConfigID = id
		addedConfigs = append(addedConfigs, ref)
	}

	return addedConfigs, nil
}
