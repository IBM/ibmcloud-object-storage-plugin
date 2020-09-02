/*******************************************************************************
 * IBM Confidential
 * OCO Source Materials
 * IBM Cloud Container Service, 5737-D43
 * (C) Copyright IBM Corp. 2017, 2018 All Rights Reserved.
 * The source code for this program is not  published or otherwise divested of
 * its trade secrets, irrespective of what has been deposited with
 * the U.S. Copyright Office.
 ******************************************************************************/

package mock

import (
	"context"
	"errors"
	"github.com/IBM/ibmcloud-object-storage-plugin/ibm-provider/provider"
	"google.golang.org/grpc"
)

const (
	clusterTypeVpcG2   = "vpc-gen2"
	clusterTypeClassic = "classic"
	clusterTypeOther   = "other"
	testSvcEndpoint    = "10.10.10.10"
	emptySvcEndpoint   = ""
)

type IBMProviderClientFactory struct {
	ClusterTypeVpcG2      bool
	ClusterTypeClassic    bool
	ClusterTypeOther      bool
	FailClusterType       bool
	FailClusterTypeErrMsg string
	FailSvcEndpoint       bool
	FailSvcEndpointErrMsg string
	TestSvcEndpoint       bool
	EmptySvcEndpoint      bool
}

type mockIBMProviderClient struct {
	provider *IBMProviderClientFactory
}

func (pc *IBMProviderClientFactory) NewIBMProviderClient(cc grpc.ClientConnInterface) provider.IBMProviderClient {
	return &mockIBMProviderClient{provider: pc}
}

func (c *mockIBMProviderClient) GetProviderType(
	ctx context.Context, in *provider.ProviderTypeRequest,
	opts ...grpc.CallOption,
) (*provider.ProviderTypeReply, error) {
	var reply provider.ProviderTypeReply
	if c.provider.ClusterTypeVpcG2 {
		reply = provider.ProviderTypeReply{Type: clusterTypeVpcG2}
	} else if c.provider.ClusterTypeClassic {
		reply = provider.ProviderTypeReply{Type: clusterTypeClassic}
	} else if c.provider.ClusterTypeOther {
		reply = provider.ProviderTypeReply{Type: clusterTypeOther}
	} else if c.provider.FailClusterType {
		return &reply, errors.New(c.provider.FailClusterTypeErrMsg)
	}
	out := &reply
	return out, nil
}

func (c *mockIBMProviderClient) GetVPCSvcEndpoint(
	ctx context.Context, in *provider.VPCSvcEndpointRequest,
	opts ...grpc.CallOption,
) (*provider.VPCSvcEndpointReply, error) {
	var reply provider.VPCSvcEndpointReply
	if c.provider.ClusterTypeVpcG2 && c.provider.TestSvcEndpoint {
		reply = provider.VPCSvcEndpointReply{Cse: testSvcEndpoint}
	} else if c.provider.ClusterTypeVpcG2 && c.provider.EmptySvcEndpoint {
		reply = provider.VPCSvcEndpointReply{Cse: emptySvcEndpoint}
	} else if c.provider.ClusterTypeVpcG2 && c.provider.FailSvcEndpoint {
		return &reply, errors.New(c.provider.FailSvcEndpointErrMsg)
	} else {
		return &reply, errors.New(c.provider.FailSvcEndpointErrMsg)
	}
	out := &reply
	return out, nil
}
