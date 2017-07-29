package lib

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Regions_GetRegions_Error(t *testing.T) {
	server, client := getTestServerAndClient(http.StatusNotAcceptable, `{error}`)
	defer server.Close()

	regions, err := client.GetRegions()
	assert.Nil(t, regions)
	if assert.NotNil(t, err) {
		assert.Equal(t, `{error}`, err.Error())
	}
}

func Test_Regions_GetRegions_NoRegions(t *testing.T) {
	server, client := getTestServerAndClient(http.StatusOK, `[]`)
	defer server.Close()

	regions, err := client.GetRegions()
	if err != nil {
		t.Error(err)
	}
	assert.Nil(t, regions)
}

func Test_Regions_GetRegions_OK(t *testing.T) {
	server, client := getTestServerAndClient(http.StatusOK, `{
"5":{"DCID":"5","name":"Los Angeles","country":"US","continent":"North America","state":"CA","ddos_protection":true,"regioncode":"LAX"},
"9":{"DCID":"9","name":"Frankfurt","country":"DE","continent":"Europe","state":"","block_storage":false},
"19":{"DCID":"19","name":"Australia","country":"AU","continent":"Australia","state":"","ddos_protection":false,"block_storage":true}}`)
	defer server.Close()

	regions, err := client.GetRegions()
	if err != nil {
		t.Error(err)
	}
	if assert.NotNil(t, regions) {
		assert.Equal(t, 3, len(regions))

		assert.Equal(t, 19, regions[0].ID)
		assert.Equal(t, "AU", regions[0].Country)
		assert.Equal(t, "", regions[0].State)
		assert.Equal(t, "Australia", regions[0].Continent)
		assert.Equal(t, false, regions[0].Ddos)
		assert.Equal(t, true, regions[0].BlockStorage)

		assert.Equal(t, 9, regions[1].ID)
		assert.Equal(t, "Frankfurt", regions[1].Name)
		assert.Equal(t, "DE", regions[1].Country)
		assert.Equal(t, "Europe", regions[1].Continent)
		assert.Equal(t, false, regions[1].Ddos)
		assert.Equal(t, false, regions[1].BlockStorage)
		assert.Equal(t, "", regions[1].Code)

		assert.Equal(t, 5, regions[2].ID)
		assert.Equal(t, "Los Angeles", regions[2].Name)
		assert.Equal(t, "US", regions[2].Country)
		assert.Equal(t, "CA", regions[2].State)
		assert.Equal(t, true, regions[2].Ddos)
		assert.Equal(t, false, regions[2].BlockStorage)
		assert.Equal(t, "LAX", regions[2].Code)
	}
}
