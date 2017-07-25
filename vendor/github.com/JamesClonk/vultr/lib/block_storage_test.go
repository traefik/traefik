package lib

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_BlockStorage_GetBlockStorages_Error(t *testing.T) {
	server, client := getTestServerAndClient(http.StatusNotAcceptable, `{error}`)
	defer server.Close()

	storages, err := client.GetBlockStorages()
	assert.Nil(t, storages)
	if assert.NotNil(t, err) {
		assert.Equal(t, `{error}`, err.Error())
	}
}

func Test_BlockStorage_GetBlockStorages_NoKeys(t *testing.T) {
	server, client := getTestServerAndClient(http.StatusOK, `[]`)
	defer server.Close()

	storages, err := client.GetBlockStorages()
	if err != nil {
		t.Error(err)
	}
	assert.Nil(t, storages)
}

func Test_BlockStorage_GetBlockStorages_OK(t *testing.T) {
	server, client := getTestServerAndClient(http.StatusOK, `[
{"SUBID":"2","label":"beta","DCID":33,"size_gb":100,"cost_per_month":10,"date_created":"2014-12-31 13:34:56"},
{"SUBID":"1","label":"alpha","attached_to_SUBID":123,"size_gb":10,"status":"pending","date_created":"2011-11-11 11:11:11"}]`)
	defer server.Close()

	storages, err := client.GetBlockStorages()
	if err != nil {
		t.Error(err)
	}
	if assert.NotNil(t, storages) {
		assert.Equal(t, 2, len(storages))

		assert.Equal(t, "1", storages[0].ID)
		assert.Equal(t, "alpha", storages[0].Name)
		assert.Equal(t, "2011-11-11 11:11:11", storages[0].Created)
		assert.Equal(t, "123", storages[0].AttachedTo)
		assert.Equal(t, 10, storages[0].SizeGB)
		assert.Equal(t, "pending", storages[0].Status)

		assert.Equal(t, "2", storages[1].ID)
		assert.Equal(t, "beta", storages[1].Name)
		assert.Equal(t, "2014-12-31 13:34:56", storages[1].Created)
		assert.Equal(t, 33, storages[1].RegionID)
		assert.Equal(t, 100, storages[1].SizeGB)
		assert.Equal(t, "10", storages[1].Cost)
	}
}

func Test_BlockStorage_CreateBlockStorage_Error(t *testing.T) {
	server, client := getTestServerAndClient(http.StatusNotAcceptable, `{error}`)
	defer server.Close()

	storage, err := client.CreateBlockStorage("delta", 33, 150)
	assert.Equal(t, BlockStorage{}, storage)
	if assert.NotNil(t, err) {
		assert.Equal(t, `{error}`, err.Error())
	}
}

func Test_BlockStorage_CreateBlockStorage_NoKey(t *testing.T) {
	server, client := getTestServerAndClient(http.StatusOK, `[]`)
	defer server.Close()

	storage, err := client.CreateBlockStorage("delta", 33, 150)
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, "", storage.ID)
}

func Test_BlockStorage_CreateBlockStorage_OK(t *testing.T) {
	server, client := getTestServerAndClient(http.StatusOK, `{"SUBID":"5671234"}`)
	defer server.Close()

	storage, err := client.CreateBlockStorage("delta", 33, 150)
	if err != nil {
		t.Error(err)
	}
	if assert.NotNil(t, storage) {
		assert.Equal(t, "5671234", storage.ID)
		assert.Equal(t, "delta", storage.Name)
		assert.Equal(t, 33, storage.RegionID)
		assert.Equal(t, 150, storage.SizeGB)
	}
}

func Test_BlockStorage_AttachBlockStorage_Error(t *testing.T) {
	server, client := getTestServerAndClient(http.StatusNotAcceptable, `{error}`)
	defer server.Close()

	err := client.AttachBlockStorage("555", "666")
	if assert.NotNil(t, err) {
		assert.Equal(t, `{error}`, err.Error())
	}
}

func Test_BlockStorage_ResizeBlockStorage_Error(t *testing.T) {
	server, client := getTestServerAndClient(http.StatusNotAcceptable, `{error}`)
	defer server.Close()

	err := client.ResizeBlockStorage("123", 150)
	if assert.NotNil(t, err) {
		assert.Equal(t, `{error}`, err.Error())
	}
}

func Test_BlockStorage_ResizeBlockStorage_OK(t *testing.T) {
	server, client := getTestServerAndClient(http.StatusOK, `{no-response?!}`)
	defer server.Close()

	assert.Nil(t, client.ResizeBlockStorage("123", 150))
}

func Test_BlockStorage_LabelBlockStorage_Error(t *testing.T) {
	server, client := getTestServerAndClient(http.StatusNotAcceptable, `{error}`)
	defer server.Close()

	err := client.LabelBlockStorage("123", "test")
	if assert.NotNil(t, err) {
		assert.Equal(t, `{error}`, err.Error())
	}
}

func Test_BlockStorage_LabelBlockStorage_OK(t *testing.T) {
	server, client := getTestServerAndClient(http.StatusOK, `{no-response?!}`)
	defer server.Close()

	assert.Nil(t, client.LabelBlockStorage("123", "test"))
}

func Test_BlockStorage_AttachBlockStorage_OK(t *testing.T) {
	server, client := getTestServerAndClient(http.StatusOK, `{no-response?!}`)
	defer server.Close()

	assert.Nil(t, client.AttachBlockStorage("555", "666"))
}

func Test_BlockStorage_DetachBlockStorage_Error(t *testing.T) {
	server, client := getTestServerAndClient(http.StatusNotAcceptable, `{error}`)
	defer server.Close()

	err := client.DetachBlockStorage("id-1")
	if assert.NotNil(t, err) {
		assert.Equal(t, `{error}`, err.Error())
	}
}

func Test_BlockStorage_DetachBlockStorage_OK(t *testing.T) {
	server, client := getTestServerAndClient(http.StatusOK, `{no-response?!}`)
	defer server.Close()

	assert.Nil(t, client.DetachBlockStorage("id-1"))
}

func Test_BlockStorage_DeleteBlockStorage_Error(t *testing.T) {
	server, client := getTestServerAndClient(http.StatusNotAcceptable, `{error}`)
	defer server.Close()

	err := client.DeleteBlockStorage("id-1")
	if assert.NotNil(t, err) {
		assert.Equal(t, `{error}`, err.Error())
	}
}

func Test_BlockStorage_DeleteBlockStorage_OK(t *testing.T) {
	server, client := getTestServerAndClient(http.StatusOK, `{no-response?!}`)
	defer server.Close()

	assert.Nil(t, client.DeleteBlockStorage("id-1"))
}
