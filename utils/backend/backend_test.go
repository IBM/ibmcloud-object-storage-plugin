/*******************************************************************************
 * IBM Confidential
 * OCO Source Materials
 * IBM Cloud Container Service, 5737-D43
 * (C) Copyright IBM Corp. 2017, 2018 All Rights Reserved.
 * The source code for this program is not  published or otherwise divested of
 * its trade secrets, irrespective of what has been deposited with
 * the U.S. Copyright Office.
 ******************************************************************************/

package backend

import (
	"errors"
	"github.com/IBM/ibm-cos-sdk-go/aws/awserr"
	"github.com/IBM/ibm-cos-sdk-go/service/s3"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"strings"
	"testing"
)

type fakeS3API struct {
	ErrHeadBucket   error
	ErrCreateBucket error
	ErrListObjects  error
	ErrDeleteObject error
	ErrDeleteBucket error
	ObjectPath      string
}

const (
	errFooMsg             = "foo"
	testBucket            = "test-bucket"
	testObjectPath        = "/test/object-path"
	testEndpoint          = "test-endpoint"
	testRegion            = "test-region"
	testAccessKey         = "akey"
	testSecretKey         = "skey"
	testAPIKey            = "apikey"
	testServiceInstanceID = "sid"
	testIAMEndpoint       = "https://test-iam-endpoint"
)

var (
	testObject = "test-object"
	errFoo     = errors.New(errFooMsg)
)

func (a *fakeS3API) HeadBucket(input *s3.HeadBucketInput) (*s3.HeadBucketOutput, error) {
	return nil, a.ErrHeadBucket
}

func (a *fakeS3API) CreateBucket(input *s3.CreateBucketInput) (*s3.CreateBucketOutput, error) {
	return nil, a.ErrCreateBucket
}

func (a *fakeS3API) ListObjects(input *s3.ListObjectsInput) (*s3.ListObjectsOutput, error) {
	return &s3.ListObjectsOutput{
		Contents: []*s3.Object{{Key: &testObject}},
	}, a.ErrListObjects
}

func (a *fakeS3API) DeleteObject(input *s3.DeleteObjectInput) (*s3.DeleteObjectOutput, error) {
	return nil, a.ErrDeleteObject
}

func (a *fakeS3API) DeleteBucket(input *s3.DeleteBucketInput) (*s3.DeleteBucketOutput, error) {
	return nil, a.ErrDeleteBucket
}

func getSession(svc s3API) ObjectStorageSession {
	return &COSSession{
		logger: zap.NewNop(),
		svc:    svc,
	}
}

func Test_NewObjectStorageSession_Positive(t *testing.T) {
	f := &COSSessionFactory{}
	sess := f.NewObjectStorageSession(testEndpoint, testRegion, &ObjectStorageCredentials{AccessKey: testAccessKey, SecretKey: testSecretKey}, zap.NewNop())
	assert.NotNil(t, sess)
}

func Test_NewObjectStorageIAMSession_Positive(t *testing.T) {
	f := &COSSessionFactory{}
	sess := f.NewObjectStorageSession(testEndpoint, testRegion,
		&ObjectStorageCredentials{ServiceInstanceID: testServiceInstanceID, APIKey: testAPIKey, IAMEndpoint: testIAMEndpoint}, zap.NewNop())
	assert.NotNil(t, sess)
}

func Test_CheckBucketAccess_Error(t *testing.T) {
	sess := getSession(&fakeS3API{ErrHeadBucket: errFoo})
	err := sess.CheckBucketAccess(testBucket)
	if assert.Error(t, err) {
		assert.EqualError(t, err, errFooMsg)
	}
}

func Test_CheckBucketAccess_Positive(t *testing.T) {
	sess := getSession(&fakeS3API{})
	err := sess.CheckBucketAccess(testBucket)
	assert.NoError(t, err)
}

func Test_CheckObjectPathExistence_Positive(t *testing.T) {
	testObject = strings.TrimPrefix(testObjectPath, "/")
	testObject = testObject + "/"
	sess := getSession(&fakeS3API{ObjectPath: testObject})
	exist, err := sess.CheckObjectPathExistence(testBucket, testObjectPath)
	assert.NoError(t, err)
	assert.Equal(t, exist, true)
}

func Test_CheckObjectPathExistence_WithoutSuffix(t *testing.T) {
	testObject = strings.TrimPrefix(testObjectPath, "/")
	sess := getSession(&fakeS3API{ObjectPath: testObject})
	exist, err := sess.CheckObjectPathExistence(testBucket, testObjectPath)
	assert.NoError(t, err)
	assert.Equal(t, exist, false)
}

func Test_CheckObjectPathExistence_PathNotFound(t *testing.T) {
	sess := getSession(&fakeS3API{ObjectPath: "test/object-path-xxxx"})
	testObject = "test/object-path-xxxx"
	exist, err := sess.CheckObjectPathExistence(testBucket, testObjectPath)
	assert.NoError(t, err)
	assert.Equal(t, exist, false)
}

func Test_CheckObjectPathExistence_Error(t *testing.T) {
	sess := getSession(&fakeS3API{ErrListObjects: errFoo})
	_, err := sess.CheckObjectPathExistence(testBucket, testObjectPath)
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "cannot list bucket")
	}
}

func Test_CreateBucketAccess_Error(t *testing.T) {
	sess := getSession(&fakeS3API{ErrCreateBucket: errFoo})
	_, err := sess.CreateBucket(testBucket)
	if assert.Error(t, err) {
		assert.EqualError(t, err, errFooMsg)
	}
}

func Test_CreateBucketAccess_BucketAlreadyExists_Positive(t *testing.T) {
	sess := getSession(&fakeS3API{ErrCreateBucket: awserr.New("BucketAlreadyOwnedByYou", "", errFoo)})
	_, err := sess.CreateBucket(testBucket)
	assert.NoError(t, err)
}

func Test_CreateBucket_Positive(t *testing.T) {
	sess := getSession(&fakeS3API{})
	_, err := sess.CreateBucket(testBucket)
	assert.NoError(t, err)
}

func Test_DeleteBucket_BucketAlreadyDeleted_Positive(t *testing.T) {
	sess := getSession(&fakeS3API{ErrListObjects: awserr.New("NoSuchBucket", "", errFoo)})
	err := sess.DeleteBucket(testBucket)
	assert.NoError(t, err)
}

func Test_DeleteBucket_ListObjectsError(t *testing.T) {
	sess := getSession(&fakeS3API{ErrListObjects: errFoo})
	err := sess.DeleteBucket(testBucket)
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "cannot list bucket")
	}
}

func Test_DeleteBucket_DeleteObjectError(t *testing.T) {
	sess := getSession(&fakeS3API{ErrDeleteObject: errFoo})
	err := sess.DeleteBucket(testBucket)
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "cannot delete object")
	}
}

func Test_DeleteBucket_Error(t *testing.T) {
	sess := getSession(&fakeS3API{ErrDeleteBucket: errFoo})
	err := sess.DeleteBucket(testBucket)
	if assert.Error(t, err) {
		assert.EqualError(t, err, errFooMsg)
	}
}

func Test_DeleteBucket_Positive(t *testing.T) {
	sess := getSession(&fakeS3API{})
	err := sess.DeleteBucket(testBucket)
	assert.NoError(t, err)
}
