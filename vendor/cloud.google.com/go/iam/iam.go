// Copyright 2016 Google Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package iam supports the resource-specific operations of Google Cloud
// IAM (Identity and Access Management) for the Google Cloud Libraries.
// See https://cloud.google.com/iam for more about IAM.
//
// Users of the Google Cloud Libraries will typically not use this package
// directly. Instead they will begin with some resource that supports IAM, like
// a pubsub topic, and call its IAM method to get a Handle for that resource.
package iam

import (
	"golang.org/x/net/context"
	pb "google.golang.org/genproto/googleapis/iam/v1"
	"google.golang.org/grpc"
)

// A Handle provides IAM operations for a resource.
type Handle struct {
	c        pb.IAMPolicyClient
	resource string
}

// InternalNewHandle is for use by the Google Cloud Libraries only.
//
// InternalNewHandle returns a Handle for resource.
// The conn parameter refers to a server that must support the IAMPolicy service.
func InternalNewHandle(conn *grpc.ClientConn, resource string) *Handle {
	return &Handle{
		c:        pb.NewIAMPolicyClient(conn),
		resource: resource,
	}
}

// Policy retrieves the IAM policy for the resource.
func (h *Handle) Policy(ctx context.Context) (*Policy, error) {
	proto, err := h.c.GetIamPolicy(ctx, &pb.GetIamPolicyRequest{Resource: h.resource})
	if err != nil {
		return nil, err
	}
	return &Policy{InternalProto: proto}, nil
}

// SetPolicy replaces the resource's current policy with the supplied Policy.
//
// If policy was created from a prior call to Get, then the modification will
// only succeed if the policy has not changed since the Get.
func (h *Handle) SetPolicy(ctx context.Context, policy *Policy) error {
	_, err := h.c.SetIamPolicy(ctx, &pb.SetIamPolicyRequest{
		Resource: h.resource,
		Policy:   policy.InternalProto,
	})
	return err
}

// TestPermissions returns the subset of permissions that the caller has on the resource.
func (h *Handle) TestPermissions(ctx context.Context, permissions []string) ([]string, error) {
	res, err := h.c.TestIamPermissions(ctx, &pb.TestIamPermissionsRequest{
		Resource:    h.resource,
		Permissions: permissions,
	})
	if err != nil {
		return nil, err
	}
	return res.Permissions, nil
}

// A RoleName is a name representing a collection of permissions.
type RoleName string

// Common role names.
const (
	Owner  RoleName = "roles/owner"
	Editor RoleName = "roles/editor"
	Viewer RoleName = "roles/viewer"
)

const (
	// AllUsers is a special member that denotes all users, even unauthenticated ones.
	AllUsers = "allUsers"

	// AllAuthenticatedUsers is a special member that denotes all authenticated users.
	AllAuthenticatedUsers = "allAuthenticatedUsers"
)

// A Policy is a list of Bindings representing roles
// granted to members.
//
// The zero Policy is a valid policy with no bindings.
type Policy struct {
	// TODO(jba): when type aliases are available, put Policy into an internal package
	// and provide an exported alias here.

	// This field is exported for use by the Google Cloud Libraries only.
	// It may become unexported in a future release.
	InternalProto *pb.Policy
}

// Members returns the list of members with the supplied role.
// The return value should not be modified. Use Add and Remove
// to modify the members of a role.
func (p *Policy) Members(r RoleName) []string {
	b := p.binding(r)
	if b == nil {
		return nil
	}
	return b.Members
}

// HasRole reports whether member has role r.
func (p *Policy) HasRole(member string, r RoleName) bool {
	return memberIndex(member, p.binding(r)) >= 0
}

// Add adds member member to role r if it is not already present.
// A new binding is created if there is no binding for the role.
func (p *Policy) Add(member string, r RoleName) {
	b := p.binding(r)
	if b == nil {
		if p.InternalProto == nil {
			p.InternalProto = &pb.Policy{}
		}
		p.InternalProto.Bindings = append(p.InternalProto.Bindings, &pb.Binding{
			Role:    string(r),
			Members: []string{member},
		})
		return
	}
	if memberIndex(member, b) < 0 {
		b.Members = append(b.Members, member)
		return
	}
}

// Remove removes member from role r if it is present.
func (p *Policy) Remove(member string, r RoleName) {
	b := p.binding(r)
	i := memberIndex(member, b)
	if i < 0 {
		return
	}
	// Order doesn't matter, so move the last member into the
	// removed spot and shrink the slice.
	// TODO(jba): worry about multiple copies of m?
	last := len(b.Members) - 1
	b.Members[i] = b.Members[last]
	b.Members[last] = ""
	b.Members = b.Members[:last]
}

// Roles returns the names of all the roles that appear in the Policy.
func (p *Policy) Roles() []RoleName {
	if p.InternalProto == nil {
		return nil
	}
	var rns []RoleName
	for _, b := range p.InternalProto.Bindings {
		rns = append(rns, RoleName(b.Role))
	}
	return rns
}

// binding returns the Binding for the suppied role, or nil if there isn't one.
func (p *Policy) binding(r RoleName) *pb.Binding {
	if p.InternalProto == nil {
		return nil
	}
	for _, b := range p.InternalProto.Bindings {
		if b.Role == string(r) {
			return b
		}
	}
	return nil
}

// memberIndex returns the index of m in b's Members, or -1 if not found.
func memberIndex(m string, b *pb.Binding) int {
	if b == nil {
		return -1
	}
	for i, mm := range b.Members {
		if mm == m {
			return i
		}
	}
	return -1
}
