// Code generated by counterfeiter. DO NOT EDIT.
package controllersfakes

import (
	"context"
	"sync"

	"cloud.google.com/go/gkehub/apiv1beta1/gkehubpb"
	"sigs.k8s.io/cluster-api-provider-gcp/api/v1beta1"

	"github.com/giantswarm/fleet-membership-operator-gcp/controllers"
)

type FakeGKEMembershipClient struct {
	RegisterMembershipStub        func(context.Context, *v1beta1.GCPCluster, []byte) (*gkehubpb.Membership, error)
	registerMembershipMutex       sync.RWMutex
	registerMembershipArgsForCall []struct {
		arg1 context.Context
		arg2 *v1beta1.GCPCluster
		arg3 []byte
	}
	registerMembershipReturns struct {
		result1 *gkehubpb.Membership
		result2 error
	}
	registerMembershipReturnsOnCall map[int]struct {
		result1 *gkehubpb.Membership
		result2 error
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *FakeGKEMembershipClient) RegisterMembership(arg1 context.Context, arg2 *v1beta1.GCPCluster, arg3 []byte) (*gkehubpb.Membership, error) {
	var arg3Copy []byte
	if arg3 != nil {
		arg3Copy = make([]byte, len(arg3))
		copy(arg3Copy, arg3)
	}
	fake.registerMembershipMutex.Lock()
	ret, specificReturn := fake.registerMembershipReturnsOnCall[len(fake.registerMembershipArgsForCall)]
	fake.registerMembershipArgsForCall = append(fake.registerMembershipArgsForCall, struct {
		arg1 context.Context
		arg2 *v1beta1.GCPCluster
		arg3 []byte
	}{arg1, arg2, arg3Copy})
	stub := fake.RegisterMembershipStub
	fakeReturns := fake.registerMembershipReturns
	fake.recordInvocation("RegisterMembership", []interface{}{arg1, arg2, arg3Copy})
	fake.registerMembershipMutex.Unlock()
	if stub != nil {
		return stub(arg1, arg2, arg3)
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	return fakeReturns.result1, fakeReturns.result2
}

func (fake *FakeGKEMembershipClient) RegisterMembershipCallCount() int {
	fake.registerMembershipMutex.RLock()
	defer fake.registerMembershipMutex.RUnlock()
	return len(fake.registerMembershipArgsForCall)
}

func (fake *FakeGKEMembershipClient) RegisterMembershipCalls(stub func(context.Context, *v1beta1.GCPCluster, []byte) (*gkehubpb.Membership, error)) {
	fake.registerMembershipMutex.Lock()
	defer fake.registerMembershipMutex.Unlock()
	fake.RegisterMembershipStub = stub
}

func (fake *FakeGKEMembershipClient) RegisterMembershipArgsForCall(i int) (context.Context, *v1beta1.GCPCluster, []byte) {
	fake.registerMembershipMutex.RLock()
	defer fake.registerMembershipMutex.RUnlock()
	argsForCall := fake.registerMembershipArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2, argsForCall.arg3
}

func (fake *FakeGKEMembershipClient) RegisterMembershipReturns(result1 *gkehubpb.Membership, result2 error) {
	fake.registerMembershipMutex.Lock()
	defer fake.registerMembershipMutex.Unlock()
	fake.RegisterMembershipStub = nil
	fake.registerMembershipReturns = struct {
		result1 *gkehubpb.Membership
		result2 error
	}{result1, result2}
}

func (fake *FakeGKEMembershipClient) RegisterMembershipReturnsOnCall(i int, result1 *gkehubpb.Membership, result2 error) {
	fake.registerMembershipMutex.Lock()
	defer fake.registerMembershipMutex.Unlock()
	fake.RegisterMembershipStub = nil
	if fake.registerMembershipReturnsOnCall == nil {
		fake.registerMembershipReturnsOnCall = make(map[int]struct {
			result1 *gkehubpb.Membership
			result2 error
		})
	}
	fake.registerMembershipReturnsOnCall[i] = struct {
		result1 *gkehubpb.Membership
		result2 error
	}{result1, result2}
}

func (fake *FakeGKEMembershipClient) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.registerMembershipMutex.RLock()
	defer fake.registerMembershipMutex.RUnlock()
	copiedInvocations := map[string][][]interface{}{}
	for key, value := range fake.invocations {
		copiedInvocations[key] = value
	}
	return copiedInvocations
}

func (fake *FakeGKEMembershipClient) recordInvocation(key string, args []interface{}) {
	fake.invocationsMutex.Lock()
	defer fake.invocationsMutex.Unlock()
	if fake.invocations == nil {
		fake.invocations = map[string][][]interface{}{}
	}
	if fake.invocations[key] == nil {
		fake.invocations[key] = [][]interface{}{}
	}
	fake.invocations[key] = append(fake.invocations[key], args)
}

var _ controllers.GKEMembershipClient = new(FakeGKEMembershipClient)