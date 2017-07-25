package virtualmachine

import (
	"testing"

	"github.com/Azure/azure-sdk-for-go/management/testutils"
)

func TestAzureGetResourceExtensions(t *testing.T) {
	client := testutils.GetTestClient(t)

	list, err := NewClient(client).GetResourceExtensions()
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("Found %d extensions", len(list))
	if len(list) == 0 {
		t.Fatal("Huh, no resource extensions at all? Something must be wrong.")
	}

	for _, extension := range list {
		if extension.Name == "" {
			t.Fatalf("Resource with empty name? Something must have gone wrong with serialization: %+v", extension)
		}
	}
}
