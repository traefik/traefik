package govultr

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

// BlockStorageService is the interface to interact with Block-Storage endpoint on the Vultr API
// Link: https://www.vultr.com/api/#block
type BlockStorageService interface {
	Attach(ctx context.Context, blockID, InstanceID string) error
	Create(ctx context.Context, regionID, size int, label string) (*BlockStorage, error)
	Delete(ctx context.Context, blockID string) error
	Detach(ctx context.Context, blockID string) error
	SetLabel(ctx context.Context, blockID, label string) error
	List(ctx context.Context) ([]BlockStorage, error)
	Get(ctx context.Context, blockID string) (*BlockStorage, error)
	Resize(ctx context.Context, blockID string, size int) error
}

// BlockStorageServiceHandler handles interaction with the block-storage methods for the Vultr API
type BlockStorageServiceHandler struct {
	client *Client
}

// BlockStorage represents Vultr Block-Storage
type BlockStorage struct {
	BlockStorageID string `json:"SUBID"`
	DateCreated    string `json:"date_created"`
	CostPerMonth   string `json:"cost_per_month"`
	Status         string `json:"status"`
	SizeGB         int    `json:"size_gb"`
	RegionID       int    `json:"DCID"`
	InstanceID     string `json:"attached_to_SUBID"`
	Label          string `json:"label"`
}

// UnmarshalJSON implements json.Unmarshaller on BlockStorage to handle the inconsistent types returned from the Vultr v1 API.
func (b *BlockStorage) UnmarshalJSON(data []byte) (err error) {
	if b == nil {
		*b = BlockStorage{}
	}

	var v map[string]interface{}
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}

	b.BlockStorageID, err = b.unmarshalStr(fmt.Sprintf("%v", v["SUBID"]))
	if err != nil {
		return err
	}

	b.RegionID, err = b.unmarshalInt(fmt.Sprintf("%v", v["DCID"]))
	if err != nil {
		return err
	}

	b.SizeGB, err = b.unmarshalInt(fmt.Sprintf("%v", v["size_gb"]))
	if err != nil {
		return err
	}

	b.InstanceID, err = b.unmarshalStr(fmt.Sprintf("%v", v["attached_to_SUBID"]))
	if err != nil {
		return err
	}

	b.CostPerMonth, err = b.unmarshalStr(fmt.Sprintf("%v", v["cost_per_month"]))
	if err != nil {
		return err
	}

	date := fmt.Sprintf("%v", v["date_created"])
	if date == "<nil>" {
		date = ""
	}
	b.DateCreated = date

	status := fmt.Sprintf("%v", v["status"])
	if status == "<nil>" {
		status = ""
	}
	b.Status = status

	b.Label = fmt.Sprintf("%v", v["label"])

	return nil
}

func (b *BlockStorage) unmarshalInt(value string) (int, error) {
	if len(value) == 0 || value == "<nil>" {
		value = "0"
	}

	i, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return 0, err
	}

	return int(i), nil
}

func (b *BlockStorage) unmarshalStr(value string) (string, error) {
	if len(value) == 0 || value == "<nil>" || value == "0" || value == "false" {
		return "", nil
	}

	f, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return "", err
	}

	return strconv.FormatFloat(f, 'f', -1, 64), nil
}

// Attach will link a given block storage to a given Vultr vps
func (b *BlockStorageServiceHandler) Attach(ctx context.Context, blockID, InstanceID string) error {

	uri := "/v1/block/attach"

	values := url.Values{
		"SUBID":           {blockID},
		"attach_to_SUBID": {InstanceID},
	}

	req, err := b.client.NewRequest(ctx, http.MethodPost, uri, values)

	if err != nil {
		return err
	}

	err = b.client.DoWithContext(ctx, req, nil)

	if err != nil {
		return err
	}

	return nil
}

// Create builds out a block storage
func (b *BlockStorageServiceHandler) Create(ctx context.Context, regionID, sizeGB int, label string) (*BlockStorage, error) {

	uri := "/v1/block/create"

	values := url.Values{
		"DCID":    {strconv.Itoa(regionID)},
		"size_gb": {strconv.Itoa(sizeGB)},
		"label":   {label},
	}

	req, err := b.client.NewRequest(ctx, http.MethodPost, uri, values)

	if err != nil {
		return nil, err
	}

	blockStorage := new(BlockStorage)

	err = b.client.DoWithContext(ctx, req, blockStorage)

	if err != nil {
		return nil, err
	}

	blockStorage.RegionID = regionID
	blockStorage.Label = label
	blockStorage.SizeGB = sizeGB

	return blockStorage, nil
}

// Delete will remove block storage instance from your Vultr account
func (b *BlockStorageServiceHandler) Delete(ctx context.Context, blockID string) error {

	uri := "/v1/block/delete"

	values := url.Values{
		"SUBID": {blockID},
	}

	req, err := b.client.NewRequest(ctx, http.MethodPost, uri, values)

	if err != nil {
		return err
	}

	err = b.client.DoWithContext(ctx, req, nil)

	if err != nil {
		return err
	}

	return nil
}

// Detach will de-link a given block storage to the Vultr vps it is attached to
func (b *BlockStorageServiceHandler) Detach(ctx context.Context, blockID string) error {

	uri := "/v1/block/detach"

	values := url.Values{
		"SUBID": {blockID},
	}

	req, err := b.client.NewRequest(ctx, http.MethodPost, uri, values)

	if err != nil {
		return err
	}

	err = b.client.DoWithContext(ctx, req, nil)

	if err != nil {
		return err
	}

	return nil
}

// SetLabel allows you to set/update the label on your Vultr Block storage
func (b *BlockStorageServiceHandler) SetLabel(ctx context.Context, blockID, label string) error {
	uri := "/v1/block/label_set"

	values := url.Values{
		"SUBID": {blockID},
		"label": {label},
	}

	req, err := b.client.NewRequest(ctx, http.MethodPost, uri, values)

	if err != nil {
		return err
	}

	err = b.client.DoWithContext(ctx, req, nil)

	if err != nil {
		return err
	}

	return nil
}

// List returns a list of all block storage instances on your Vultr Account
func (b *BlockStorageServiceHandler) List(ctx context.Context) ([]BlockStorage, error) {

	uri := "/v1/block/list"

	req, err := b.client.NewRequest(ctx, http.MethodGet, uri, nil)

	if err != nil {
		return nil, err
	}

	var blockStorage []BlockStorage
	err = b.client.DoWithContext(ctx, req, &blockStorage)

	if err != nil {
		return nil, err
	}

	return blockStorage, nil
}

// Get returns a single block storage instance based ony our blockID you provide from your Vultr Account
func (b *BlockStorageServiceHandler) Get(ctx context.Context, blockID string) (*BlockStorage, error) {

	uri := "/v1/block/list"

	req, err := b.client.NewRequest(ctx, http.MethodGet, uri, nil)

	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	q.Add("SUBID", blockID)
	req.URL.RawQuery = q.Encode()

	blockStorage := new(BlockStorage)
	err = b.client.DoWithContext(ctx, req, blockStorage)

	if err != nil {
		return nil, err
	}

	return blockStorage, nil
}

// Resize allows you to resize your Vultr block storage instance
func (b *BlockStorageServiceHandler) Resize(ctx context.Context, blockID string, sizeGB int) error {

	uri := "/v1/block/resize"

	values := url.Values{
		"SUBID":   {blockID},
		"size_gb": {strconv.Itoa(sizeGB)},
	}

	req, err := b.client.NewRequest(ctx, http.MethodPost, uri, values)

	if err != nil {
		return err
	}

	err = b.client.DoWithContext(ctx, req, nil)

	if err != nil {
		return err
	}

	return nil
}
