package gateway

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/gateway-api/apis/v1alpha2"
)

func Test_allocatedPortBlockAccepts(t *testing.T) {
	tests := []struct {
		desc               string
		listeners          []v1alpha2.Listener
		shareableProtocols []v1alpha2.ProtocolType
		wantErr            bool
	}{
		{
			desc: "test multi-listener unshareable protocol",
			listeners: []v1alpha2.Listener{
				{
					Protocol: v1alpha2.HTTPSProtocolType,
				},
				{
					Protocol: v1alpha2.HTTPSProtocolType,
				},
			},
			shareableProtocols: []v1alpha2.ProtocolType{
				v1alpha2.HTTPProtocolType,
			},
			wantErr: true,
		},
		{
			desc: "test multi-listener protocol mismatch",
			listeners: []v1alpha2.Listener{
				{
					Protocol: v1alpha2.HTTPSProtocolType,
				},
				{
					Protocol: v1alpha2.HTTPProtocolType,
				},
			},
			shareableProtocols: []v1alpha2.ProtocolType{
				v1alpha2.HTTPSProtocolType,
				v1alpha2.HTTPProtocolType,
			},
			wantErr: true,
		},
		{
			desc: "test multi-listener hostname collision",
			listeners: []v1alpha2.Listener{
				{
					Protocol: v1alpha2.HTTPSProtocolType,
					Hostname: hostnamePtr("*.example.com"),
				},
				{
					Protocol: v1alpha2.HTTPSProtocolType,
					Hostname: hostnamePtr("*.example.com"),
				},
			},
			shareableProtocols: []v1alpha2.ProtocolType{
				v1alpha2.HTTPSProtocolType,
			},
			wantErr: true,
		},
		{
			desc: "test multi-listener hostname edge cases",
			listeners: []v1alpha2.Listener{
				{
					Protocol: v1alpha2.HTTPSProtocolType,
					Hostname: nil,
				},
				{
					Protocol: v1alpha2.HTTPSProtocolType,
					Hostname: hostnamePtr("*.example.com"),
				},
			},
			shareableProtocols: []v1alpha2.ProtocolType{
				v1alpha2.HTTPSProtocolType,
			},
			wantErr: false,
		},
		{
			desc: "test multi-listener hostname",
			listeners: []v1alpha2.Listener{
				{
					Protocol: v1alpha2.HTTPSProtocolType,
					Hostname: hostnamePtr("*.example.com"),
				},
				{
					Protocol: v1alpha2.HTTPSProtocolType,
					Hostname: hostnamePtr("sub1.example.com"),
				},
				{
					Protocol: v1alpha2.HTTPSProtocolType,
					Hostname: hostnamePtr("sub2.example.com"),
				},
				{
					Protocol: v1alpha2.HTTPSProtocolType,
					Hostname: hostnamePtr("sub1.example.com"),
				},
			},
			shareableProtocols: []v1alpha2.ProtocolType{
				v1alpha2.HTTPSProtocolType,
			},
			wantErr: false,
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			blocks := make(allocatedPortBlock, 0)

			hasErr := false
			for _, l := range test.listeners {
				if !blocks.accepts(l, test.shareableProtocols) {
					hasErr = true
				} else {
					blocks = append(blocks, l)
				}
			}
			assert.Equal(t, test.wantErr, hasErr)
		})
	}
}

func Test_allocatedListeners(t *testing.T) {
	tests := []struct {
		desc               string
		allocated          allocatedListeners
		listeners          []v1alpha2.Listener
		shareableProtocols []v1alpha2.ProtocolType
		wantErr            bool
		wantListeners      int
	}{
		{
			desc: "test listener tracking hostnames ok",
			allocated: allocatedListeners{
				443: {
					{
						Protocol: v1alpha2.HTTPSProtocolType,
						Hostname: hostnamePtr("*.example.com"),
					},
				},
			},
			listeners: []v1alpha2.Listener{
				{
					Port:     443,
					Protocol: v1alpha2.HTTPSProtocolType,
					Hostname: hostnamePtr("sub1.example.com"),
				},
				{
					Port:     443,
					Protocol: v1alpha2.HTTPSProtocolType,
					Hostname: hostnamePtr("sub2.example.com"),
				},
			},
			shareableProtocols: []v1alpha2.ProtocolType{
				v1alpha2.HTTPSProtocolType,
			},
			wantErr:       false,
			wantListeners: 3,
		},
		{
			desc: "test listener tracking hostnames with outlier",
			allocated: allocatedListeners{
				443: {
					{
						Protocol: v1alpha2.HTTPSProtocolType,
						Hostname: hostnamePtr("*.example.com"),
					},
				},
			},
			listeners: []v1alpha2.Listener{
				{
					Port:     443,
					Protocol: v1alpha2.HTTPSProtocolType,
					Hostname: hostnamePtr("sub1.example.com"),
				},
				{
					Port:     443,
					Protocol: v1alpha2.HTTPSProtocolType,
					Hostname: hostnamePtr("sub1.example.com"),
				},
			},
			shareableProtocols: []v1alpha2.ProtocolType{
				v1alpha2.HTTPSProtocolType,
			},
			wantErr:       true,
			wantListeners: 2,
		},
		{
			desc: "test listener tracking and shareable protocols mismatch",
			allocated: allocatedListeners{
				443: {
					{
						Protocol: v1alpha2.HTTPSProtocolType,
						Hostname: hostnamePtr("*.example.com"),
					},
				},
			},
			listeners: []v1alpha2.Listener{
				{
					Port:     443,
					Protocol: v1alpha2.HTTPSProtocolType,
					Hostname: hostnamePtr("sub1.example.com"),
				},
				{
					Port:     443,
					Protocol: v1alpha2.HTTPSProtocolType,
					Hostname: hostnamePtr("sub1.example.com"),
				},
			},
			shareableProtocols: []v1alpha2.ProtocolType{},
			wantErr:            true,
			wantListeners:      1,
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			if test.allocated == nil {
				test.allocated = newAllocatedListeners()
			}

			hasErr := false
			for _, l := range test.listeners {
				if !test.allocated.accepts(l, test.shareableProtocols) {
					hasErr = true
				} else {
					test.allocated.add(l)
				}
			}
			assert.Equal(t, test.wantErr, hasErr)
			assert.Equal(t, test.wantListeners, test.allocated.count())
		})
	}
}
