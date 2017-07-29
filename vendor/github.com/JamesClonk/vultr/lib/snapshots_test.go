package lib

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Snapshots_GetSnapshots_Error(t *testing.T) {
	server, client := getTestServerAndClient(http.StatusNotAcceptable, `{error}`)
	defer server.Close()

	snapshots, err := client.GetSnapshots()
	assert.Nil(t, snapshots)
	if assert.NotNil(t, err) {
		assert.Equal(t, `{error}`, err.Error())
	}
}

func Test_Snapshots_GetSnapshots_NoSnapshots(t *testing.T) {
	server, client := getTestServerAndClient(http.StatusOK, `[]`)
	defer server.Close()

	snapshots, err := client.GetSnapshots()
	if err != nil {
		t.Error(err)
	}
	assert.Nil(t, snapshots)
}

func Test_Snapshots_GetSnapshots_OK(t *testing.T) {
	server, client := getTestServerAndClient(http.StatusOK, `{
"5359435d28b9a": {"SNAPSHOTID": "5359435d28b9a","date_created": "2014-04-18 12:40:40",
    "description": "test snapshot","size": "42949672960","status": "complete"},
"5359435dc1df3": {"SNAPSHOTID": "5359435dc1df3","date_created": "2014-04-22 16:11:46",
    "description": "a","size": "10000000","status": "incomplete"}}`)
	defer server.Close()

	snapshots, err := client.GetSnapshots()
	if err != nil {
		t.Error(err)
	}
	if assert.NotNil(t, snapshots) {
		assert.Equal(t, 2, len(snapshots))

		assert.Equal(t, "5359435dc1df3", snapshots[0].ID)
		assert.Equal(t, "a", snapshots[0].Description)
		assert.Equal(t, "10000000", snapshots[0].Size)
		assert.Equal(t, "incomplete", snapshots[0].Status)
		assert.Equal(t, "2014-04-22 16:11:46", snapshots[0].Created)

		assert.Equal(t, "5359435d28b9a", snapshots[1].ID)
		assert.Equal(t, "test snapshot", snapshots[1].Description)
		assert.Equal(t, "42949672960", snapshots[1].Size)
		assert.Equal(t, "complete", snapshots[1].Status)
		assert.Equal(t, "2014-04-18 12:40:40", snapshots[1].Created)
	}
}

func Test_Snapshots_CreateSnapshot_Error(t *testing.T) {
	server, client := getTestServerAndClient(http.StatusNotAcceptable, `{error}`)
	defer server.Close()

	snapshot, err := client.CreateSnapshot("12345", "test snappy")
	assert.Equal(t, Snapshot{}, snapshot)
	if assert.NotNil(t, err) {
		assert.Equal(t, `{error}`, err.Error())
	}
}

func Test_Snapshots_CreateSnapshot_NoSnapshot(t *testing.T) {
	server, client := getTestServerAndClient(http.StatusOK, `[]`)
	defer server.Close()

	snapshot, err := client.CreateSnapshot("12345", "test snappy")
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, "", snapshot.ID)
}

func Test_Snapshots_CreateSnapshot_OK(t *testing.T) {
	server, client := getTestServerAndClient(http.StatusOK, `{"SNAPSHOTID": "544e52f31c706"}`)
	defer server.Close()

	snapshot, err := client.CreateSnapshot("12345", "test snappy")
	if err != nil {
		t.Error(err)
	}
	if assert.NotNil(t, snapshot) {
		assert.Equal(t, "544e52f31c706", snapshot.ID)
		assert.Equal(t, "test snappy", snapshot.Description)
	}
}

func Test_Snapshots_DeleteSnapshot_Error(t *testing.T) {
	server, client := getTestServerAndClient(http.StatusNotAcceptable, `{error}`)
	defer server.Close()

	err := client.DeleteSnapshot("id-1")
	if assert.NotNil(t, err) {
		assert.Equal(t, `{error}`, err.Error())
	}
}

func Test_Snapshots_DeleteSnapshot_OK(t *testing.T) {
	server, client := getTestServerAndClient(http.StatusOK, `{no-response?!}`)
	defer server.Close()

	assert.Nil(t, client.DeleteSnapshot("id-1"))
}
