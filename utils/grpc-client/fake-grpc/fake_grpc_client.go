/*******************************************************************************
 * IBM Confidential
 * OCO Source Materials
 * IBM Cloud Container Service, 5737-D43
 * (C) Copyright IBM Corp. 2017, 2018 All Rights Reserved.
 * The source code for this program is not  published or otherwise divested of
 * its trade secrets, irrespective of what has been deposited with
 * the U.S. Copyright Office.
 ******************************************************************************/

package fake_grpc

import (
	"errors"
	grpcClient "github.com/IBM/ibmcloud-object-storage-plugin/utils/grpc-client"
	"google.golang.org/grpc"
)

// FakeGrpcSessionFactory implements grpcClient.GrpcSessionFactory
type FakeGrpcSessionFactory struct {
	//FailGrpcConnection ...
	FailGrpcConnection bool
	//FailGrpcConnectionErr with specific error msg...
	FailGrpcConnectionErr string
	//PassGrpcConnection ...
	PassGrpcConnection bool
}

var _ grpcClient.GrpcSessionFactory = (*FakeGrpcSessionFactory)(nil)

// fakeGrpcSession implements grpcClient.GrpcSession
type fakeGrpcSession struct {
	factory *FakeGrpcSessionFactory
}

// NewGrpcSession method creates a new fakeGrpcSession session
func (f *FakeGrpcSessionFactory) NewGrpcSession() grpcClient.GrpcSession {
	return &fakeGrpcSession{
		factory: f,
	}
}

// GrpcDial method creates a fake-grpc-client connection
func (c *fakeGrpcSession) GrpcDial(clientConn grpcClient.ClientConn, target string, opts ...grpc.DialOption) (conn *grpc.ClientConn, err error) {
	if c.factory.FailGrpcConnection {
		return conn, errors.New(c.factory.FailGrpcConnectionErr)
	}
	return conn, err
}
