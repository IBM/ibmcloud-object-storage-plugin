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
	"github.com/IBM/go-sdk-core/v3/core"
	rc "github.com/IBM/ibm-cos-sdk-go-config/resourceconfigurationv1"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

const (
	errTestMsg    = "updating bucket configuration failed"
	testBucket2   = "test-bucket"
	allowedIps    = "test_ip"
	resConfApiKey = "test_api_key"
)

var (
	errTest    = errors.New(errTestMsg)
	byteArray  []byte
	httpHeader = http.Header{}
	result     interface{}
	statusCode = 200
)

var dresponse core.DetailedResponse

type fakeResourceConfigurationV1 struct {
	frc1 fakeRCV1
}

type fakeRCV1 interface {
	UpdateBucketConfig(*rc.ResourceConfigurationV1, *rc.UpdateBucketConfigOptions) (*core.DetailedResponse, error)
}

func (rc *fakeResourceConfigurationV1) UpdateBucketConfig(service *rc.ResourceConfigurationV1, options *rc.UpdateBucketConfigOptions) (*core.DetailedResponse, error) {
	// refer https://github.com/IBM/go-sdk-core/blob/3485263179566e258883cc8ce55144b5b99fa308/v4/core/detailed_response.go#L24 for response struct
	dresponse = core.DetailedResponse{StatusCode: statusCode, Headers: httpHeader, Result: result, RawResult: byteArray}
	//return &core.DetailedResponse{statusCode, httpHeader, result, byteArray}, nil
	return &dresponse, nil
}

type fakeResourceConfigurationV1Fail struct {
	frc2 fakeRCV2
}

type fakeRCV2 interface {
	UpdateBucketConfig(*rc.ResourceConfigurationV1, *rc.UpdateBucketConfigOptions) (*core.DetailedResponse, error)
}

func (rc *fakeResourceConfigurationV1Fail) UpdateBucketConfig(service *rc.ResourceConfigurationV1, options *rc.UpdateBucketConfigOptions) (*core.DetailedResponse, error) {
	return nil, errTest
}

func getFakeAccessPolicySession(r ResourceConfigurationV1) AccessPolicy {
	return &UpdateAPObj{rcv1: r}
}

func Test_NewAccessPolicy_Positive(t *testing.T) {
	f := &UpdateAPFactory{}
	rcSess := f.NewAccessPolicy()
	assert.NotNil(t, rcSess)
}

var rc1 fakeRCV1 = &fakeResourceConfigurationV1{}

func Test_UpdateAccessPolicy_Positive(t *testing.T) {
	rcSess := getFakeAccessPolicySession(&fakeResourceConfigurationV1{frc1: rc1})
	err := rcSess.UpdateAccessPolicy(allowedIps, resConfApiKey, testBucket, rc1)
	assert.NoError(t, err)
}

var rc2 fakeRCV2 = &fakeResourceConfigurationV1Fail{}

func Test_UpdateAccessPolicy_Error(t *testing.T) {
	rcSess := getFakeAccessPolicySession(&fakeResourceConfigurationV1Fail{frc2: rc2})
	err := rcSess.UpdateAccessPolicy(allowedIps, resConfApiKey, testBucket2, rc2)
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), errTestMsg)
	}
}
