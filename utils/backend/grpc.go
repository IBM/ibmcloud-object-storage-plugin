package backend

import (
	"fmt"
	"google.golang.org/grpc"
	"net"
	"time"
)

type GrpcSessionFactory interface {

	// NewObjectStorageBackend method creates a new object store session
	NewGrpcSession() GrpcSession
}

type GrpcSession interface {
	GrpcDial(*string) (*grpc.ClientConn, error)
}

type ConnObjFactory struct{}

// COSSession represents a COS (S3) session
type ConnObj struct {
	conn *grpc.ClientConn
}

// NewObjectStorageSession method creates a new object store session
func (c *ConnObjFactory) NewGrpcSession() GrpcSession {

	return &ConnObj{}
}

func UnixConnect(addr string, t time.Duration) (net.Conn, error) {
	unix_addr, err := net.ResolveUnixAddr("unix", addr)
	conn, err := net.DialUnix("unix", nil, unix_addr)
	return conn, err
}

func (*ConnObj) GrpcDial(SockEndpoint *string) (*grpc.ClientConn, error) {
	conn, err := grpc.Dial(*SockEndpoint, grpc.WithInsecure(), grpc.WithBlock(), grpc.WithDialer(UnixConnect))
	if err != nil {
		return conn, fmt.Errorf("could not not connect to grpc server: %v", err)
	}
	return conn, err
}
