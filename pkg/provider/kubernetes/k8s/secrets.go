package k8s

import (
	"fmt"
	"slices"
	"time"

	corev1 "k8s.io/api/core/v1"
	kerror "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	kschema "k8s.io/apimachinery/pkg/runtime/schema"
	kinformers "k8s.io/client-go/informers"
	kclientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

// secretNameWildcard allows all the Secrets of a namespace.
// It cannot clash with a real Secret name, as names must be valid DNS subdomains.
const secretNameWildcard = "*"

// SecretInformers manages the informers watching the Secrets.
// When no restriction is configured, one informer watches all the Secrets of each namespace.
// Otherwise, one informer is started per allowed Secret,
// so the credentials used by Traefik only need access to these Secrets.
// A namespace allowing all its Secrets (secretNameWildcard) is watched by a single informer.
type SecretInformers struct {
	factories  map[string]kinformers.SharedInformerFactory
	restricted bool
}

// NewSecretInformers creates the informers for the given namespaces.
// The allowed Secret names are grouped by namespace,
// and a namespace allowing "*" has all its Secrets watched.
// When allowed is empty, all the Secrets (not owned by Helm) of the given namespaces are watched.
func NewSecretInformers(client kclientset.Interface, resync time.Duration, namespaces []string, allowed map[string][]string, handler cache.ResourceEventHandler) (*SecretInformers, error) {
	s := &SecretInformers{
		factories:  make(map[string]kinformers.SharedInformerFactory),
		restricted: len(allowed) > 0,
	}

	notOwnedByHelm := func(opts *metav1.ListOptions) {
		opts.LabelSelector = "owner!=helm"
	}

	addFactory := func(key, namespace string, tweak func(*metav1.ListOptions)) error {
		factory := kinformers.NewSharedInformerFactoryWithOptions(client, resync, kinformers.WithNamespace(namespace), kinformers.WithTweakListOptions(tweak), kinformers.WithTransform(StripManagedFields))
		if _, err := factory.Core().V1().Secrets().Informer().AddEventHandler(handler); err != nil {
			return err
		}

		s.factories[key] = factory

		return nil
	}

	for _, ns := range namespaces {
		if !s.restricted {
			if err := addFactory(ns, ns, notOwnedByHelm); err != nil {
				return nil, err
			}

			continue
		}

		for allowedNamespace, names := range allowed {
			if ns != metav1.NamespaceAll && ns != allowedNamespace {
				continue
			}

			if slices.Contains(names, secretNameWildcard) {
				key := secretKey(allowedNamespace, secretNameWildcard)
				if _, exists := s.factories[key]; exists {
					continue
				}

				if err := addFactory(key, allowedNamespace, notOwnedByHelm); err != nil {
					return nil, err
				}

				continue
			}

			for _, name := range names {
				key := secretKey(allowedNamespace, name)
				if _, exists := s.factories[key]; exists {
					continue
				}

				tweak := func(opts *metav1.ListOptions) {
					opts.LabelSelector = "owner!=helm"
					opts.FieldSelector = fields.OneTermEqualSelector("metadata.name", name).String()
				}

				if err := addFactory(key, allowedNamespace, tweak); err != nil {
					return nil, err
				}
			}
		}
	}

	return s, nil
}

// Start starts all the informers.
func (s *SecretInformers) Start(stopCh <-chan struct{}) {
	for _, factory := range s.factories {
		factory.Start(stopCh)
	}
}

// WaitForCacheSync waits for all the informer caches to be synced.
func (s *SecretInformers) WaitForCacheSync(stopCh <-chan struct{}) error {
	keys := make([]string, 0, len(s.factories))
	for key := range s.factories {
		keys = append(keys, key)
	}
	slices.Sort(keys)

	for _, key := range keys {
		for t, ok := range s.factories[key].WaitForCacheSync(stopCh) {
			if !ok {
				return fmt.Errorf("timed out waiting for controller caches to sync %s for secrets %q", t.String(), key)
			}
		}
	}

	return nil
}

// Get returns the named Secret from the given namespace.
// The lookupNamespace is the namespace the informer has been registered with (metav1.NamespaceAll if all namespaces are watched).
func (s *SecretInformers) Get(lookupNamespace, namespace, name string) (*corev1.Secret, error) {
	if s.restricted {
		factory, ok := s.factories[secretKey(namespace, name)]
		if !ok {
			factory, ok = s.factories[secretKey(namespace, secretNameWildcard)]
		}
		if !ok {
			return nil, kerror.NewNotFound(kschema.GroupResource{Resource: "secrets"}, name)
		}

		return factory.Core().V1().Secrets().Lister().Secrets(namespace).Get(name)
	}

	return s.factories[lookupNamespace].Core().V1().Secrets().Lister().Secrets(namespace).Get(name)
}

func secretKey(namespace, name string) string {
	return namespace + "/" + name
}
