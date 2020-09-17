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
	"errors"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"net"
	"testing"
	"time"
)

const (
	errMsg = "establishing grpc-client connection failed"
)

var (
	errMsgString  = errors.New(errMsg)
	testString    = "test_endpoint"
	sockeEndpoint = &testString
)

type fakeClientConn1 struct {
	fcc1 fakeClConn1
}

type fakeClConn1 interface {
	Connect(target string, opts ...grpc.DialOption) (*(grpc.ClientConn), error)
	//Close() error
}

func (gs *fakeClientConn1) Connect(target string, opts ...grpc.DialOption) (*(grpc.ClientConn), error) {
	var err error
	fakeConn := grpc.ClientConn{}
	return &fakeConn, err
}

//func (gs *fakeClientConn1) Close() error {
//	return nil
//}

type fakeClientConn2 struct {
	fcc2 fakeClConn2
}

type fakeClConn2 interface {
	Connect(target string, opts ...grpc.DialOption) (*(grpc.ClientConn), error)
	//Close() error
}

func (gs *fakeClientConn2) Connect(target string, opts ...grpc.DialOption) (*(grpc.ClientConn), error) {
	return nil, errMsgString
}

//func (gs *fakeClientConn2) Close() error {
//	return nil
//}

func getFakeGrpcSession(gcon *grpc.ClientConn, conn ClientConn) GrpcSession {
	return &GrpcSes{
		conn: gcon,
		cc:   conn,
	}
}

func Test_NewGrpcSession_Positive(t *testing.T) {
	f := &ConnObjFactory{}
	grpcSess := f.NewGrpcSession()
	assert.NotNil(t, grpcSess)
}

var cc1 fakeClConn1 = &fakeClientConn1{}
var cc2 fakeClConn2 = &fakeClientConn2{}
var gcon = &grpc.ClientConn{}

func Test_GrpcDial_Positive(t *testing.T) {
	grSess := getFakeGrpcSession(gcon, &fakeClientConn1{fcc1: cc1})
	_, err := grSess.GrpcDial(cc1, *sockeEndpoint, grpc.WithInsecure(), grpc.WithBlock(), grpc.WithDialer(UnixConnect))
	assert.NoError(t, err)
}

func Test_GrpcDial_Error(t *testing.T) {
	grSess := getFakeGrpcSession(gcon, &fakeClientConn2{fcc2: cc2})
	_, err := grSess.GrpcDial(cc2, *sockeEndpoint, grpc.WithInsecure(), grpc.WithBlock(), grpc.WithDialer(UnixConnect))

	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), errMsg)
	}
}

func UnixConnect(addr string, t time.Duration) (net.Conn, error) {
	unix_addr, err := net.ResolveUnixAddr("unix", addr)
	conn, err := net.DialUnix("unix", nil, unix_addr)
	return conn, err
}
