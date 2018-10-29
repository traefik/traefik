package linodego

import (
	"context"
	"fmt"
)

// LinodeKernel represents a Linode Instance kernel object
type LinodeKernel struct {
	ID           string `json:"id"`
	Label        string `json:"label"`
	Version      string `json:"version"`
	Architecture string `json:"architecture"`
	KVM          bool   `json:"kvm"`
	XEN          bool   `json:"xen"`
	PVOPS        bool   `json:"pvops"`
}

// LinodeKernelsPagedResponse represents a Linode kernels API response for listing
type LinodeKernelsPagedResponse struct {
	*PageOptions
	Data []LinodeKernel `json:"data"`
}

// ListKernels lists linode kernels
func (c *Client) ListKernels(ctx context.Context, opts *ListOptions) ([]LinodeKernel, error) {
	response := LinodeKernelsPagedResponse{}
	err := c.listHelper(ctx, &response, opts)
	if err != nil {
		return nil, err
	}
	return response.Data, nil
}

func (LinodeKernelsPagedResponse) endpoint(c *Client) string {
	endpoint, err := c.Kernels.Endpoint()
	if err != nil {
		panic(err)
	}
	return endpoint
}

func (resp *LinodeKernelsPagedResponse) appendData(r *LinodeKernelsPagedResponse) {
	resp.Data = append(resp.Data, r.Data...)
}

// GetKernel gets the kernel with the provided ID
func (c *Client) GetKernel(ctx context.Context, kernelID string) (*LinodeKernel, error) {
	e, err := c.Kernels.Endpoint()
	if err != nil {
		return nil, err
	}
	e = fmt.Sprintf("%s/%s", e, kernelID)
	r, err := c.R(ctx).
		SetResult(&LinodeKernel{}).
		Get(e)
	if err != nil {
		return nil, err
	}
	return r.Result().(*LinodeKernel), nil
}
