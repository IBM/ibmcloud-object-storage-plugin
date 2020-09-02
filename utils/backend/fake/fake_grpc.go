/*******************************************************************************
 * IBM Confidential
 * OCO Source Materials
 * IBM Cloud Container Service, 5737-D43
 * (C) Copyright IBM Corp. 2017, 2018 All Rights Reserved.
 * The source code for this program is not  published or otherwise divested of
 * its trade secrets, irrespective of what has been deposited with
 * the U.S. Copyright Office.
 ******************************************************************************/

package fake

import (
	"errors"
	"github.com/IBM/ibmcloud-object-storage-plugin/utils/backend"
	"google.golang.org/grpc"
)

//ObjectStorageSessionFactory is a factory for mocked object storage sessions
type GrpcSessionFactory struct {
	//FailGrpcConnection ...
	FailGrpcConnection bool
	//FailGrpcConnectionErr with specific error msg...
	FailGrpcConnectionErr string
	//PassGrpcConnection ...
	PassGrpcConnection bool
}

type fakeGrpcSession struct {
	factory *GrpcSessionFactory
}

// NewObjectStorageSession method creates a new fake object store session
func (f *GrpcSessionFactory) NewGrpcSession() backend.GrpcSession {
	return &fakeGrpcSession{
		factory: f,
	}
}

func (c *fakeGrpcSession) GrpcDial(SockEndpoint *string) (conn *grpc.ClientConn, err error) {
	if c.factory.FailGrpcConnection {
		return conn, errors.New(c.factory.FailGrpcConnectionErr)
	}
	return conn, err
}
