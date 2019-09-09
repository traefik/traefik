package cloudflare

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/pkg/errors"
)

// WorkersKVNamespaceRequest provides parameters for creating and updating storage namespaces
type WorkersKVNamespaceRequest struct {
	Title string `json:"title"`
}

// WorkersKVNamespaceResponse is the response received when creating storage namespaces
type WorkersKVNamespaceResponse struct {
	Response
	Result WorkersKVNamespace `json:"result"`
}

// WorkersKVNamespace contains the unique identifier and title of a storage namespace
type WorkersKVNamespace struct {
	ID    string `json:"id"`
	Title string `json:"title"`
}

// ListWorkersKVNamespacesResponse contains a slice of storage namespaces associated with an
// account, pagination information, and an embedded response struct
type ListWorkersKVNamespacesResponse struct {
	Response
	Result     []WorkersKVNamespace `json:"result"`
	ResultInfo `json:"result_info"`
}

// StorageKey is a key name used to identify a storage value
type StorageKey struct {
	Name string `json:"name"`
}

// ListStorageKeysResponse contains a slice of keys belonging to a storage namespace,
// pagination information, and an embedded response struct
type ListStorageKeysResponse struct {
	Response
	Result     []StorageKey `json:"result"`
	ResultInfo `json:"result_info"`
}

// CreateWorkersKVNamespace creates a namespace under the given title.
// A 400 is returned if the account already owns a namespace with this title.
// A namespace must be explicitly deleted to be replaced.
//
// API reference: https://api.cloudflare.com/#workers-kv-namespace-create-a-namespace
func (api *API) CreateWorkersKVNamespace(ctx context.Context, req *WorkersKVNamespaceRequest) (WorkersKVNamespaceResponse, error) {
	uri := fmt.Sprintf("/accounts/%s/storage/kv/namespaces", api.OrganizationID)
	res, err := api.makeRequestContext(ctx, http.MethodPost, uri, req)
	if err != nil {
		return WorkersKVNamespaceResponse{}, errors.Wrap(err, errMakeRequestError)
	}

	result := WorkersKVNamespaceResponse{}
	if err := json.Unmarshal(res, &result); err != nil {
		return result, errors.Wrap(err, errUnmarshalError)
	}

	return result, err
}

// ListWorkersKVNamespaces lists storage namespaces
//
// API reference: https://api.cloudflare.com/#workers-kv-namespace-list-namespaces
func (api *API) ListWorkersKVNamespaces(ctx context.Context) (ListWorkersKVNamespacesResponse, error) {
	uri := fmt.Sprintf("/accounts/%s/storage/kv/namespaces", api.OrganizationID)
	res, err := api.makeRequestContext(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return ListWorkersKVNamespacesResponse{}, errors.Wrap(err, errMakeRequestError)
	}

	result := ListWorkersKVNamespacesResponse{}
	if err := json.Unmarshal(res, &result); err != nil {
		return result, errors.Wrap(err, errUnmarshalError)
	}

	return result, err
}

// DeleteWorkersKVNamespace deletes the namespace corresponding to the given ID
//
// API reference: https://api.cloudflare.com/#workers-kv-namespace-remove-a-namespace
func (api *API) DeleteWorkersKVNamespace(ctx context.Context, namespaceID string) (Response, error) {
	uri := fmt.Sprintf("/accounts/%s/storage/kv/namespaces/%s", api.OrganizationID, namespaceID)
	res, err := api.makeRequestContext(ctx, http.MethodDelete, uri, nil)
	if err != nil {
		return Response{}, errors.Wrap(err, errMakeRequestError)
	}

	result := Response{}
	if err := json.Unmarshal(res, &result); err != nil {
		return result, errors.Wrap(err, errUnmarshalError)
	}

	return result, err
}

// UpdateWorkersKVNamespace modifies a namespace's title
//
// API reference: https://api.cloudflare.com/#workers-kv-namespace-rename-a-namespace
func (api *API) UpdateWorkersKVNamespace(ctx context.Context, namespaceID string, req *WorkersKVNamespaceRequest) (Response, error) {
	uri := fmt.Sprintf("/accounts/%s/storage/kv/namespaces/%s", api.OrganizationID, namespaceID)
	res, err := api.makeRequestContext(ctx, http.MethodPut, uri, req)
	if err != nil {
		return Response{}, errors.Wrap(err, errMakeRequestError)
	}

	result := Response{}
	if err := json.Unmarshal(res, &result); err != nil {
		return result, errors.Wrap(err, errUnmarshalError)
	}

	return result, err
}

// WriteWorkersKV writes a value identified by a key.
//
// API reference: https://api.cloudflare.com/#workers-kv-namespace-write-key-value-pair
func (api *API) WriteWorkersKV(ctx context.Context, namespaceID, key string, value []byte) (Response, error) {
	key = url.PathEscape(key)
	uri := fmt.Sprintf("/accounts/%s/storage/kv/namespaces/%s/values/%s", api.OrganizationID, namespaceID, key)
	res, err := api.makeRequestWithAuthTypeAndHeaders(
		ctx, http.MethodPut, uri, value, api.authType, http.Header{"Content-Type": []string{"application/octet-stream"}},
	)
	if err != nil {
		return Response{}, errors.Wrap(err, errMakeRequestError)
	}

	result := Response{}
	if err := json.Unmarshal(res, &result); err != nil {
		return result, errors.Wrap(err, errUnmarshalError)
	}

	return result, err
}

// ReadWorkersKV returns the value associated with the given key in the given namespace
//
// API reference: https://api.cloudflare.com/#workers-kv-namespace-read-key-value-pair
func (api API) ReadWorkersKV(ctx context.Context, namespaceID, key string) ([]byte, error) {
	key = url.PathEscape(key)
	uri := fmt.Sprintf("/accounts/%s/storage/kv/namespaces/%s/values/%s", api.OrganizationID, namespaceID, key)
	res, err := api.makeRequestContext(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return nil, errors.Wrap(err, errMakeRequestError)
	}
	return res, nil
}

// DeleteWorkersKV deletes a key and value for a provided storage namespace
//
// API reference: https://api.cloudflare.com/#workers-kv-namespace-delete-key-value-pair
func (api API) DeleteWorkersKV(ctx context.Context, namespaceID, key string) (Response, error) {
	key = url.PathEscape(key)
	uri := fmt.Sprintf("/accounts/%s/storage/kv/namespaces/%s/values/%s", api.OrganizationID, namespaceID, key)
	res, err := api.makeRequestContext(ctx, http.MethodDelete, uri, nil)
	if err != nil {
		return Response{}, errors.Wrap(err, errMakeRequestError)
	}

	result := Response{}
	if err := json.Unmarshal(res, &result); err != nil {
		return result, errors.Wrap(err, errUnmarshalError)
	}
	return result, err
}

// ListWorkersKVs lists a namespace's keys
//
// API Reference: https://api.cloudflare.com/#workers-kv-namespace-list-a-namespace-s-keys
func (api API) ListWorkersKVs(ctx context.Context, namespaceID string) (ListStorageKeysResponse, error) {
	uri := fmt.Sprintf("/accounts/%s/storage/kv/namespaces/%s/keys", api.OrganizationID, namespaceID)
	res, err := api.makeRequestContext(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return ListStorageKeysResponse{}, errors.Wrap(err, errMakeRequestError)
	}

	result := ListStorageKeysResponse{}
	if err := json.Unmarshal(res, &result); err != nil {
		return result, errors.Wrap(err, errUnmarshalError)
	}
	return result, err
}
