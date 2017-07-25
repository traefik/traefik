package vmutils

import (
	vm "github.com/Azure/azure-sdk-for-go/management/virtualmachine"
)

func updateOrAddConfig(configs []vm.ConfigurationSet, configType vm.ConfigurationSetType, update func(*vm.ConfigurationSet)) []vm.ConfigurationSet {
	config := findConfig(configs, configType)
	if config == nil {
		configs = append(configs, vm.ConfigurationSet{ConfigurationSetType: configType})
		config = findConfig(configs, configType)
	}
	update(config)

	return configs
}

func findConfig(configs []vm.ConfigurationSet, configType vm.ConfigurationSetType) *vm.ConfigurationSet {
	for i, config := range configs {
		if config.ConfigurationSetType == configType {
			// need to return a pointer to the original set in configs,
			// not the copy made by the range iterator
			return &configs[i]
		}
	}

	return nil
}
