/*******************************************************************************
 * IBM Confidential
 * OCO Source Materials
 * IBM Cloud Kubernetes Service, 5737-D43
 * (C) Copyright IBM Corp. 2017, 2023 All Rights Reserved.
 * The source code for this program is not published or otherwise divested of
 * its trade secrets, irrespective of what has been deposited with
 * the U.S. Copyright Office.
 ******************************************************************************/

package fake

import (
	"errors"
	"github.com/IBM/ibmcloud-object-storage-plugin/utils/backend"
)

// FakeAccessPolicyFactory implements backend.AccessPolicyFactory
type FakeAccessPolicyFactory struct {
	//FailUpdateAccessPolicy ...
	FailUpdateAccessPolicy bool
	//FailUpdateAccessPolicyErrMsg with specific error msg...
	FailUpdateAccessPolicyErrMsg string
	//PassUpdateAccessPolicy ...
	PassUpdateAccessPolicy bool
	//FailUpdateAccessPolicy ...
	FailUpdateQuotaLimit bool
	//FailUpdateAccessPolicyErrMsg with specific error msg...
	FailUpdateQuotaLimitErrMsg string
	//PassUpdateAccessPolicy ...
	PassUpdateQuotaLimit bool
}

var _ backend.AccessPolicyFactory = (*FakeAccessPolicyFactory)(nil)

// fakeAccessPolicy implements backend.AccessPolicy
type fakeAccessPolicy struct {
	rcv1 *FakeAccessPolicyFactory
}

// NewAccessPolicy method creates a new fakeAccessPolicy session
func (c *FakeAccessPolicyFactory) NewAccessPolicy() backend.AccessPolicy {
	return &fakeAccessPolicy{
		rcv1: c,
	}
}

// UpdateAccessPolicy method creates a fake updateBucketConfig call
func (c *fakeAccessPolicy) UpdateAccessPolicy(allowedIps, apiKey, bucketName string, rcc backend.ResourceConfigurationV1) error {
	if c.rcv1.FailUpdateAccessPolicy {
		return errors.New(c.rcv1.FailUpdateAccessPolicyErrMsg)
	}
	return nil
}

// UpdateQuotaLimit method creates a fake updateQuotaLimit call
func (c *fakeAccessPolicy) UpdateQuotaLimit(quota int64, apiKey, bucketName, osEndpoint, iamEndpoint string, rcc backend.ResourceConfigurationV1) error {
	if c.rcv1.FailUpdateAccessPolicy {
		return errors.New(c.rcv1.FailUpdateAccessPolicyErrMsg)
	}
	return nil
}
