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
	"go.uber.org/zap"
)

//ObjectStorageSessionFactory is a factory for mocked object storage sessions
type ObjectStorageSessionFactory struct {
	//FailCheckBucketAccess ...
	FailCheckBucketAccess bool
	//FailCreateBucket ...
	FailCreateBucket bool
	//FailCreateBucket with specific error msg...
	FailCreateBucketErrMsg string
	//FailDeleteBucket ...
	FailDeleteBucket bool
	//CheckObjectPathExistenceError ...
	CheckObjectPathExistenceError bool
	//CheckObjectPathExistencePathNotFound ...
	CheckObjectPathExistencePathNotFound bool
	//FailUpdateBucketFirewallRules ...
	FailUpdateBucketFirewallRules bool
	//PassUpdateBucketFirewalRules ...
	PassUpdateBucketFirewalRules bool

	// LastEndpoint holds the endpoint of the last created session
	LastEndpoint string
	// LastRegion holds the region of the last created session
	LastRegion string
	// LastCredentials holds the credentials of the last created session
	LastCredentials *backend.ObjectStorageCredentials
	// LastCheckedBucket stores the name of the last bucket that was checked
	LastCheckedBucket string
	// LastCreatedBucket stores the name of the last bucket that was created
	LastCreatedBucket string
	// LastDeletedBucket stores the name of the last bucket that was deleted
	LastDeletedBucket string
}

type fakeObjectStorageSession struct {
	factory *ObjectStorageSessionFactory
}

// NewObjectStorageSession method creates a new fake object store session
func (f *ObjectStorageSessionFactory) NewObjectStorageSession(endpoint, region string, creds *backend.ObjectStorageCredentials, logger *zap.Logger) backend.ObjectStorageSession {
	f.LastEndpoint = endpoint
	f.LastRegion = region
	f.LastCredentials = creds
	return &fakeObjectStorageSession{
		factory: f,
	}
}

// ResetStats clears the details about previous sessions
func (f *ObjectStorageSessionFactory) ResetStats() {
	f.LastEndpoint = ""
	f.LastRegion = ""
	f.LastCredentials = &backend.ObjectStorageCredentials{}
	f.LastCheckedBucket = ""
	f.LastCreatedBucket = ""
	f.LastDeletedBucket = ""
}

func (s *fakeObjectStorageSession) CheckBucketAccess(bucket string) error {
	s.factory.LastCheckedBucket = bucket
	if s.factory.FailCheckBucketAccess {
		return errors.New("")
	}
	return nil
}

func (s *fakeObjectStorageSession) CheckObjectPathExistence(bucket, objectpath string) (bool, error) {
	if s.factory.CheckObjectPathExistenceError {
		return false, errors.New("")
	} else if s.factory.CheckObjectPathExistencePathNotFound {
		return false, nil
	}
	return true, nil
}

func (s *fakeObjectStorageSession) CreateBucket(bucket string) (string, error) {
	s.factory.LastCreatedBucket = bucket
	if s.factory.FailCreateBucket {
		return "", errors.New(s.factory.FailCreateBucketErrMsg)
	}
	return "", nil
}

func (s *fakeObjectStorageSession) DeleteBucket(bucket string) error {
	s.factory.LastDeletedBucket = bucket
	if s.factory.FailDeleteBucket {
		return errors.New("")
	}
	return nil
}

func (s *fakeObjectStorageSession) UpdateBucketFirewallRules(bucket string) error {
	if s.factory.FailUpdateBucketFirewallRules {
		return errors.New("")
	}
	return nil
}

func (s *fakeObjectStorageSession) UpdateBucketFirewalRules(bucket string) (string, error) {
	if s.factory.PassUpdateBucketFirewalRules {
		return "true", nil
	}
	return "", nil
}
