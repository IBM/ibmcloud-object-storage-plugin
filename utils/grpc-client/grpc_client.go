/*******************************************************************************
 * IBM Confidential
 * OCO Source Materials
 * IBM Cloud Container Service, 5737-D43
 * (C) Copyright IBM Corp. 2017, 2018 All Rights Reserved.
 * The source code for this program is not  published or otherwise divested of
 * its trade secrets, irrespective of what has been deposited with
 * the U.S. Copyright Office.
 ******************************************************************************/

package grpc_client

import (
	"google.golang.org/grpc"
)

type GrpcSessionFactory interface {
	NewGrpcSession() GrpcSession
}

type GrpcSession interface {
	GrpcDial(cc ClientConn, target string, opts ...grpc.DialOption) (*grpc.ClientConn, error)
}

type ConnObjFactory struct{}

type ClientConn interface {
	Connect(target string, opts ...grpc.DialOption) (*grpc.ClientConn, error)
	//Close() error
}

type GrpcSes struct {
	conn *grpc.ClientConn
	cc   ClientConn
}

func (gs *GrpcSes) Connect(target string, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	var err error
	gs.conn, err = grpc.Dial(target, opts...)
	return gs.conn, err
}

//func (gs *GrpcSes) Close() error {
//	if gs.conn != nil {
//		return gs.conn.Close()
//	}
//	return nil
//}

func (c *ConnObjFactory) NewGrpcSession() GrpcSession {
	return &GrpcSes{}
}

var cc ClientConn = &GrpcSes{}

// GrpcDial establishes a grpc-client client server connection
func (c *GrpcSes) GrpcDial(cc ClientConn, target string, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	conn, err := cc.Connect(target, opts...)
	if err != nil {
		return nil, err
	}
	return conn, err
}
