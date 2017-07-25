package cmd

import vultr "github.com/JamesClonk/vultr/lib"

// GetClient returns a new lib.Client
func GetClient() *vultr.Client {
	return vultr.NewClient(*apiKey, nil)
}
