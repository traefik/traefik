package govultr

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
)

// ISOService is the interface to interact with the ISO endpoints on the Vultr API
// Link: https://www.vultr.com/api/#ISO
type ISOService interface {
	CreateFromURL(ctx context.Context, ISOURL string) (*ISO, error)
	Delete(ctx context.Context, ISOID int) error
	List(ctx context.Context) ([]ISO, error)
	GetPublicList(ctx context.Context) ([]PublicISO, error)
}

// ISOServiceHandler handles interaction with the ISO methods for the Vultr API
type ISOServiceHandler struct {
	Client *Client
}

// ISO represents ISOs currently available on this account.
type ISO struct {
	ISOID       int    `json:"ISOID"`
	DateCreated string `json:"date_created"`
	FileName    string `json:"filename"`
	Size        int    `json:"size"`
	MD5Sum      string `json:"md5sum"`
	SHA512Sum   string `json:"sha512sum"`
	Status      string `json:"status"`
}

// PublicISO represents public ISOs offered in the Vultr ISO library.
type PublicISO struct {
	ISOID       int    `json:"ISOID"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// CreateFromURL will create a new ISO image on your account
func (i *ISOServiceHandler) CreateFromURL(ctx context.Context, ISOURL string) (*ISO, error) {

	uri := "/v1/iso/create_from_url"

	values := url.Values{
		"url": {ISOURL},
	}

	req, err := i.Client.NewRequest(ctx, http.MethodPost, uri, values)

	if err != nil {
		return nil, err
	}

	iso := new(ISO)
	err = i.Client.DoWithContext(ctx, req, iso)

	if err != nil {
		return nil, err
	}

	return iso, nil
}

// Delete will delete an ISO image from your account
func (i *ISOServiceHandler) Delete(ctx context.Context, isoID int) error {

	uri := "/v1/iso/destroy"

	values := url.Values{
		"ISOID": {strconv.Itoa(isoID)},
	}

	req, err := i.Client.NewRequest(ctx, http.MethodPost, uri, values)

	if err != nil {
		return err
	}

	err = i.Client.DoWithContext(ctx, req, nil)

	if err != nil {
		return err
	}

	return nil
}

// List will list all ISOs currently available on your account
func (i *ISOServiceHandler) List(ctx context.Context) ([]ISO, error) {

	uri := "/v1/iso/list"

	req, err := i.Client.NewRequest(ctx, http.MethodGet, uri, nil)

	if err != nil {
		return nil, err
	}

	var ISOMap map[string]ISO
	err = i.Client.DoWithContext(ctx, req, &ISOMap)

	if err != nil {
		return nil, err
	}

	var iso []ISO
	for _, i := range ISOMap {
		iso = append(iso, i)
	}

	return iso, nil
}

// GetPublicList will list public ISOs offered in the Vultr ISO library.
func (i *ISOServiceHandler) GetPublicList(ctx context.Context) ([]PublicISO, error) {

	uri := "/v1/iso/list_public"

	req, err := i.Client.NewRequest(ctx, http.MethodGet, uri, nil)

	if err != nil {
		return nil, err
	}

	var ISOMap map[string]PublicISO
	err = i.Client.DoWithContext(ctx, req, &ISOMap)

	if err != nil {
		return nil, err
	}

	var publicISO []PublicISO

	for _, p := range ISOMap {
		publicISO = append(publicISO, p)
	}

	return publicISO, nil
}
