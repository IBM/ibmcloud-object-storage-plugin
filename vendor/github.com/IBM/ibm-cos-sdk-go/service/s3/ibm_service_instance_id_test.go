package s3_test

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/IBM/ibm-cos-sdk-go/aws"
	"github.com/IBM/ibm-cos-sdk-go/awstesting/unit"
	"github.com/IBM/ibm-cos-sdk-go/service/s3"
	"github.com/stretchr/testify/assert"
)

const (
	// Service Instance ID Error Message
	ibmsiiErrMsg = `<Error><Code>ErrorCode</Code><Message>message body</Message>
		<RequestId>requestID</RequestId><HostId>hostID=</HostId></Error>`

	// Default value for Service Instance ID when unreachable
	defaultValueWhenUnreach = `---UNREACH---`
)

// List Bucket when Service Instance ID is present
func TestWhenListBucketsSetHeader_IsPresent(t *testing.T) {
	client, ibmServiceInstanceId := newHeaderGrabberSvc()

	id := genRandomID()

	input := &s3.ListBucketsInput{}
	input.SetIBMServiceInstanceId(id)
	client.ListBuckets(input)

	// Asserts the id in List Buckets to take precedence over
	// Service Instance ID in the client creation time
	assert.Equal(t, id, *ibmServiceInstanceId, "ids do not match")
}

// List Bucket when Service Instance ID is not present
func TestWhenListBucketsNotSetHeader_NotPresent(t *testing.T) {
	client, ibmServiceInstanceId := newHeaderGrabberSvc()

	input := &s3.ListBucketsInput{}
	client.ListBuckets(input)

	// Asserts Service Instance ID in the client creation time
	// is not overwritten by ListBuckets op
	assert.Equal(t, "", *ibmServiceInstanceId, "Found Unexpected Value for ibm-service-instance-id")
}

// Create Bucket when Service Instance ID is present
func TestWhenCreateBucketSetHeader_IsPresent(t *testing.T) {
	client, ibmServiceInstanceId := newHeaderGrabberSvc()

	id := genRandomID()

	input := &s3.CreateBucketInput{}
	input.SetIBMServiceInstanceId(id)
	input.SetBucket("b1")
	client.CreateBucket(input)

	// Asserts the id in List Buckets to take precedence over
	// Service Instance ID in the client creation time
	assert.Equal(t, id, *ibmServiceInstanceId, "ids do not match")
}

// Create Bucket when Service Instance ID is not present
func TestWhenCreateBucketNotSetHeader_NotPresent(t *testing.T) {
	client, ibmServiceInstanceId := newHeaderGrabberSvc()

	input := &s3.CreateBucketInput{}
	input.SetBucket("b1")
	client.CreateBucket(input)

	// Asserts Service Instance ID in the client creation time
	// is not overwritten by CreateBucket op
	assert.Equal(t, "", *ibmServiceInstanceId, "Found Unexpected Value for ibm-service-instance-id")
}

// A helper function to create a server
// Returns:
// 		A new S3 session with Service Instance ID header with
//		SSL disabled and zero retries
func newHeaderGrabberSvc() (*s3.S3, *string) {
	ibmServiceInstanceIdHeader := defaultValueWhenUnreach

	// Creates a new server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ibmServiceInstanceIdHeader = r.Header.Get("ibm-service-instance-id")
		http.Error(w, ibmsiiErrMsg, http.StatusOK)
	}))
	return s3.New(unit.Session, aws.NewConfig().
		WithEndpoint(server.URL).
		WithDisableSSL(true).
		WithMaxRetries(0)), &ibmServiceInstanceIdHeader
}

// Create a random 16-character ID
// Returns:
//		An encoded 16-character ID
func genRandomID() string {
	u := make([]byte, 16)
	_, err := rand.Read(u)
	if err != nil {
		panic(err)
	}
	return hex.EncodeToString(u)
}
