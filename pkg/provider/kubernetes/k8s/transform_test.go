package k8s

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestStripManagedFields(t *testing.T) {
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: "whoami",
			ManagedFields: []metav1.ManagedFieldsEntry{
				{
					Manager:    "kubectl",
					Operation:  metav1.ManagedFieldsOperationApply,
					APIVersion: "v1",
					Time:       &metav1.Time{Time: time.Now()},
				},
			},
		},
	}

	got, err := StripManagedFields(service)
	require.NoError(t, err)

	assert.Same(t, service, got)
	assert.Nil(t, service.ManagedFields)
	assert.Equal(t, "whoami", service.Name)
}

func TestStripManagedFieldsIgnoresNonKubernetesObjects(t *testing.T) {
	obj := struct{ Name string }{Name: "whoami"}

	got, err := StripManagedFields(obj)
	require.NoError(t, err)

	assert.Equal(t, obj, got)
}
