// Code generated by counterfeiter. DO NOT EDIT.
package controllersfakes

import (
	"context"
	"sync"

	"sigs.k8s.io/cluster-api-provider-gcp/api/v1beta1"

	"github.com/giantswarm/fleet-membership-operator-gcp/controllers"
	"github.com/giantswarm/fleet-membership-operator-gcp/types"
)

type FakeGKEMembershipClient struct {
	RegisterStub        func(context.Context, *v1beta1.GCPCluster, []byte) (types.MembershipData, error)
	registerMutex       sync.RWMutex
	registerArgsForCall []struct {
		arg1 context.Context
		arg2 *v1beta1.GCPCluster
		arg3 []byte
	}
	registerReturns struct {
		result1 types.MembershipData
		result2 error
	}
	registerReturnsOnCall map[int]struct {
		result1 types.MembershipData
		result2 error
	}
	UnregisterStub        func(context.Context, *v1beta1.GCPCluster) error
	unregisterMutex       sync.RWMutex
	unregisterArgsForCall []struct {
		arg1 context.Context
		arg2 *v1beta1.GCPCluster
	}
	unregisterReturns struct {
		result1 error
	}
	unregisterReturnsOnCall map[int]struct {
		result1 error
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *FakeGKEMembershipClient) Register(arg1 context.Context, arg2 *v1beta1.GCPCluster, arg3 []byte) (types.MembershipData, error) {
	var arg3Copy []byte
	if arg3 != nil {
		arg3Copy = make([]byte, len(arg3))
		copy(arg3Copy, arg3)
	}
	fake.registerMutex.Lock()
	ret, specificReturn := fake.registerReturnsOnCall[len(fake.registerArgsForCall)]
	fake.registerArgsForCall = append(fake.registerArgsForCall, struct {
		arg1 context.Context
		arg2 *v1beta1.GCPCluster
		arg3 []byte
	}{arg1, arg2, arg3Copy})
	stub := fake.RegisterStub
	fakeReturns := fake.registerReturns
	fake.recordInvocation("Register", []interface{}{arg1, arg2, arg3Copy})
	fake.registerMutex.Unlock()
	if stub != nil {
		return stub(arg1, arg2, arg3)
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	return fakeReturns.result1, fakeReturns.result2
}

func (fake *FakeGKEMembershipClient) RegisterCallCount() int {
	fake.registerMutex.RLock()
	defer fake.registerMutex.RUnlock()
	return len(fake.registerArgsForCall)
}

func (fake *FakeGKEMembershipClient) RegisterCalls(stub func(context.Context, *v1beta1.GCPCluster, []byte) (types.MembershipData, error)) {
	fake.registerMutex.Lock()
	defer fake.registerMutex.Unlock()
	fake.RegisterStub = stub
}

func (fake *FakeGKEMembershipClient) RegisterArgsForCall(i int) (context.Context, *v1beta1.GCPCluster, []byte) {
	fake.registerMutex.RLock()
	defer fake.registerMutex.RUnlock()
	argsForCall := fake.registerArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2, argsForCall.arg3
}

func (fake *FakeGKEMembershipClient) RegisterReturns(result1 types.MembershipData, result2 error) {
	fake.registerMutex.Lock()
	defer fake.registerMutex.Unlock()
	fake.RegisterStub = nil
	fake.registerReturns = struct {
		result1 types.MembershipData
		result2 error
	}{result1, result2}
}

func (fake *FakeGKEMembershipClient) RegisterReturnsOnCall(i int, result1 types.MembershipData, result2 error) {
	fake.registerMutex.Lock()
	defer fake.registerMutex.Unlock()
	fake.RegisterStub = nil
	if fake.registerReturnsOnCall == nil {
		fake.registerReturnsOnCall = make(map[int]struct {
			result1 types.MembershipData
			result2 error
		})
	}
	fake.registerReturnsOnCall[i] = struct {
		result1 types.MembershipData
		result2 error
	}{result1, result2}
}

func (fake *FakeGKEMembershipClient) Unregister(arg1 context.Context, arg2 *v1beta1.GCPCluster) error {
	fake.unregisterMutex.Lock()
	ret, specificReturn := fake.unregisterReturnsOnCall[len(fake.unregisterArgsForCall)]
	fake.unregisterArgsForCall = append(fake.unregisterArgsForCall, struct {
		arg1 context.Context
		arg2 *v1beta1.GCPCluster
	}{arg1, arg2})
	stub := fake.UnregisterStub
	fakeReturns := fake.unregisterReturns
	fake.recordInvocation("Unregister", []interface{}{arg1, arg2})
	fake.unregisterMutex.Unlock()
	if stub != nil {
		return stub(arg1, arg2)
	}
	if specificReturn {
		return ret.result1
	}
	return fakeReturns.result1
}

func (fake *FakeGKEMembershipClient) UnregisterCallCount() int {
	fake.unregisterMutex.RLock()
	defer fake.unregisterMutex.RUnlock()
	return len(fake.unregisterArgsForCall)
}

func (fake *FakeGKEMembershipClient) UnregisterCalls(stub func(context.Context, *v1beta1.GCPCluster) error) {
	fake.unregisterMutex.Lock()
	defer fake.unregisterMutex.Unlock()
	fake.UnregisterStub = stub
}

func (fake *FakeGKEMembershipClient) UnregisterArgsForCall(i int) (context.Context, *v1beta1.GCPCluster) {
	fake.unregisterMutex.RLock()
	defer fake.unregisterMutex.RUnlock()
	argsForCall := fake.unregisterArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2
}

func (fake *FakeGKEMembershipClient) UnregisterReturns(result1 error) {
	fake.unregisterMutex.Lock()
	defer fake.unregisterMutex.Unlock()
	fake.UnregisterStub = nil
	fake.unregisterReturns = struct {
		result1 error
	}{result1}
}

func (fake *FakeGKEMembershipClient) UnregisterReturnsOnCall(i int, result1 error) {
	fake.unregisterMutex.Lock()
	defer fake.unregisterMutex.Unlock()
	fake.UnregisterStub = nil
	if fake.unregisterReturnsOnCall == nil {
		fake.unregisterReturnsOnCall = make(map[int]struct {
			result1 error
		})
	}
	fake.unregisterReturnsOnCall[i] = struct {
		result1 error
	}{result1}
}

func (fake *FakeGKEMembershipClient) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.registerMutex.RLock()
	defer fake.registerMutex.RUnlock()
	fake.unregisterMutex.RLock()
	defer fake.unregisterMutex.RUnlock()
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
