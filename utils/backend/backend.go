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
	"fmt"
	"strings"

	"github.com/IBM/ibm-cos-sdk-go/aws"
	"github.com/IBM/ibm-cos-sdk-go/aws/awserr"
	"github.com/IBM/ibm-cos-sdk-go/aws/credentials"
	"github.com/IBM/ibm-cos-sdk-go/aws/credentials/ibmiam"
	"github.com/IBM/ibm-cos-sdk-go/aws/session"
	"github.com/IBM/ibm-cos-sdk-go/service/s3"
	"go.uber.org/zap"
)

// ObjectStorageCredentials holds credentials for accessing an object storage service
type ObjectStorageCredentials struct {
	// AccessKey is the account identifier in AWS authentication
	AccessKey string
	// SecretKey is the "password" in AWS authentication
	SecretKey string
	// APIKey is the "password" in IBM IAM authentication
	APIKey string
	// ServiceInstanceID is the account identifier in IBM IAM authentication
	ServiceInstanceID string
	//IAMEndpoint ...
	IAMEndpoint string
}

// ObjectStorageSessionFactory is an interface of an object store session factory
type ObjectStorageSessionFactory interface {

	// NewObjectStorageBackend method creates a new object store session
	NewObjectStorageSession(endpoint, region string, creds *ObjectStorageCredentials, logger *zap.Logger) ObjectStorageSession
}

// ObjectStorageSession is an interface of an object store session
type ObjectStorageSession interface {

	// CheckBucketAccess method check that a bucket can be accessed
	CheckBucketAccess(bucket string) error

	// CheckObjectPathExistence method checks that object-path exists inside bucket
	CheckObjectPathExistence(bucket, objectpath string) (bool, error)

	// CreateBucket methods creates a new bucket
	CreateBucket(bucket, locationConstraint string, kpRootKeyCrn string) (string, error)

	// DeleteBucket methods deletes a bucket (with all of its objects)
	DeleteBucket(bucket string) error

	// SetBucketVersioning sets the versioning state of a bucket
	SetBucketVersioning(bucket string, enabled bool) error
}

// COSSessionFactory represents a COS (S3) session factory
type COSSessionFactory struct{}

type s3API interface {
	HeadBucket(input *s3.HeadBucketInput) (*s3.HeadBucketOutput, error)
	CreateBucket(input *s3.CreateBucketInput) (*s3.CreateBucketOutput, error)
	ListObjects(input *s3.ListObjectsInput) (*s3.ListObjectsOutput, error)
	//ListObjectsV2(input *s3.ListObjectsV2Input) (*s3.ListObjectsV2Output, error)
	DeleteObject(input *s3.DeleteObjectInput) (*s3.DeleteObjectOutput, error)
	DeleteBucket(input *s3.DeleteBucketInput) (*s3.DeleteBucketOutput, error)
	PutBucketVersioning(input *s3.PutBucketVersioningInput) (*s3.PutBucketVersioningOutput, error)
}

// COSSession represents a COS (S3) session
type COSSession struct {
	svc    s3API
	logger *zap.Logger
}

const (
	KPEncryptionAlgorithm = "AES256" // https://github.com/IBM/ibm-cos-sdk-go/blob/master/service/s3/api.go#L8509-L8511
)

// NewObjectStorageSession method creates a new object store session
func (s *COSSessionFactory) NewObjectStorageSession(endpoint, region string, creds *ObjectStorageCredentials, logger *zap.Logger) ObjectStorageSession {
	var sdkCreds *credentials.Credentials
	if creds.APIKey != "" {
		sdkCreds = ibmiam.NewStaticCredentials(aws.NewConfig(), creds.IAMEndpoint+"/identity/token", creds.APIKey, creds.ServiceInstanceID)
	} else {
		sdkCreds = credentials.NewStaticCredentials(creds.AccessKey, creds.SecretKey, "")
	}
	sess, _ := session.NewSession(&aws.Config{
		S3ForcePathStyle: aws.Bool(true),
		Endpoint:         aws.String(endpoint),
		Credentials:      sdkCreds,
		Region:           aws.String(region),
	})

	return &COSSession{
		svc:    s3.New(sess),
		logger: logger,
	}
}

// CheckBucketAccess method check that a bucket can be accessed
func (s *COSSession) CheckBucketAccess(bucket string) error {
	_, err := s.svc.HeadBucket(&s3.HeadBucketInput{
		Bucket: aws.String(bucket),
	})
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok && aerr.Code() == "RequestError" && strings.Contains(err.Error(), "Credential=") {
			s.logger.Warn(fmt.Sprintf("Check your secret access key for bucket %s", bucket))
			return fmt.Errorf("AccessKey/SecretKey is wrong")
		}
	}

	return err
}

// CheckObjectPathExistence method checks that object-path exists inside bucket
func (s *COSSession) CheckObjectPathExistence(bucket, objectpath string) (bool, error) {
	if strings.HasPrefix(objectpath, "/") {
		objectpath = strings.TrimPrefix(objectpath, "/")
	}
	if !strings.HasSuffix(objectpath, "/") {
		objectpath = objectpath + "/"
	}

	resp, err := s.svc.ListObjects(&s3.ListObjectsInput{
		Bucket:  aws.String(bucket),
		MaxKeys: aws.Int64(1),
		Prefix:  aws.String(objectpath),
	})

	if err != nil {
		return false, fmt.Errorf("cannot list bucket '%s': %v", bucket, err)
	}

	if len(resp.Contents) == 1 {
		object := *(resp.Contents[0].Key)
		if (object == objectpath) || (strings.TrimSuffix(object, "/") == objectpath) {
			return true, nil
		}
	}
	return false, nil
}

// CreateBucket methods creates a new bucket
func (s *COSSession) CreateBucket(bucket, locationConstraint string, kpRootKeyCrn string) (string, error) {
	var err error
	if kpRootKeyCrn != "" {
		_, err = s.svc.CreateBucket(&s3.CreateBucketInput{
			Bucket: aws.String(bucket),
			CreateBucketConfiguration: &s3.CreateBucketConfiguration{
				LocationConstraint: aws.String(locationConstraint),
			},
			IBMSSEKPCustomerRootKeyCrn:  aws.String(kpRootKeyCrn),
			IBMSSEKPEncryptionAlgorithm: aws.String(KPEncryptionAlgorithm),
		})
	} else {
		_, err = s.svc.CreateBucket(&s3.CreateBucketInput{
			Bucket: aws.String(bucket),
			CreateBucketConfiguration: &s3.CreateBucketConfiguration{
				LocationConstraint: aws.String(locationConstraint),
			},
		})
	}

	if err != nil {
		if aerr, ok := err.(awserr.Error); ok && aerr.Code() == "BucketAlreadyOwnedByYou" {
			s.logger.Warn(fmt.Sprintf("bucket '%s' already exists", bucket))
			return fmt.Sprintf("bucket '%s' already exists", bucket), nil
		} else if aerr, ok := err.(awserr.Error); ok && aerr.Code() == "RequestError" && strings.Contains(err.Error(), "Credential=") {
			s.logger.Warn(fmt.Sprintf("Check your secret access key for bucket %s", bucket))
			return fmt.Sprintf("Check your secret access key for bucket %s", bucket), fmt.Errorf("AccessKey/SecretKey is wrong")
		}
		return "", err
	}
	return "", nil
}

// DeleteBucket methods deletes a bucket (with all of its objects)
func (s *COSSession) DeleteBucket(bucket string) error {
	resp, err := s.svc.ListObjects(&s3.ListObjectsInput{
		Bucket: aws.String(bucket),
	})

	if err != nil {
		if aerr, ok := err.(awserr.Error); ok && aerr.Code() == "NoSuchBucket" {
			s.logger.Warn(fmt.Sprintf("bucket %s is already deleted", bucket))
			return nil
		}

		return fmt.Errorf("cannot list bucket '%s': %v", bucket, err)
	}

	for _, key := range resp.Contents {
		_, err = s.svc.DeleteObject(&s3.DeleteObjectInput{
			Bucket: aws.String(bucket),
			Key:    key.Key,
		})

		if err != nil {
			return fmt.Errorf("cannot delete object %s/%s: %v", bucket, *key.Key, err)
		}
	}

	_, err = s.svc.DeleteBucket(&s3.DeleteBucketInput{
		Bucket: aws.String(bucket),
	})
	return err
}

func (s *COSSession) SetBucketVersioning(bucket string, enabled bool) error {
	var status string

	// Set the versioning status based on whether it's enabled or suspended.
	if enabled {
		status = s3.BucketVersioningStatusEnabled
	} else {
		status = s3.BucketVersioningStatusSuspended
	}

	// Make the API call to set the versioning status of the bucket.
	out, err := s.svc.PutBucketVersioning(&s3.PutBucketVersioningInput{
		Bucket: aws.String(bucket),
		VersioningConfiguration: &s3.VersioningConfiguration{
			Status: aws.String(status),
		},
	})

	// Check for errors from the AWS SDK call
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			s.logger.Error("Versioning operation failed",
				zap.String("bucket", bucket),
				zap.String("status", status),
				zap.Error(aerr))
		}
		return fmt.Errorf("failed to set versioning status to %s: %w", status, err)
	}

	// Log the output for debugging or confirmation
	s.logger.Info("Bucket versioning response",
		zap.String("bucket", bucket),
		zap.String("status", status),
		zap.Any("response", out))

	return nil
}
