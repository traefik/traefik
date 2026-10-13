package k8s

// SecretURNPrefix is the prefix of the URN referencing the value of a key of a Secret.
const SecretURNPrefix = "urn:k8s:secret:"

// SecretURN returns the URN referencing the value of the given key of the given Secret.
func SecretURN(secretName, key string) string {
	return SecretURNPrefix + secretName + ":" + key
}
