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
	"github.com/IBM/go-sdk-core/v3/core"
	rc "github.com/IBM/ibm-cos-sdk-go-config/resourceconfigurationv1"
	"strconv"
	"strings"
)

// PrivateServiceURL to make service requests to.
const PrivateResourceConfigEP = "https://config.private.cloud-object-storage.cloud.ibm.com/v1"
const PrivateIAMEPForVPC = "https://private.iam.cloud.ibm.com"

type AccessPolicyFactory interface {
	NewAccessPolicy() AccessPolicy
}

type AccessPolicy interface {
	UpdateAccessPolicy(allowedIps, apiKey, bucketName string, rcc ResourceConfigurationV1) error
}

type UpdateAPFactory struct{}

type ResourceConfigurationV1 interface {
	// UpdateBucketConfig updates the bucket access policy configuration with given ips
	UpdateBucketConfig(*rc.ResourceConfigurationV1, *rc.UpdateBucketConfigOptions) (*core.DetailedResponse, error)
}

type UpdateAPObj struct {
	rcv1 ResourceConfigurationV1
}

func (uc *UpdateAPObj) UpdateBucketConfig(service *rc.ResourceConfigurationV1, options *rc.UpdateBucketConfigOptions) (res *core.DetailedResponse, err error) {
	return service.UpdateBucketConfig(options)
}

func (c *UpdateAPFactory) NewAccessPolicy() AccessPolicy {

	return &UpdateAPObj{}
}

var rcc ResourceConfigurationV1 = &UpdateAPObj{}

// UpdateAccessPolicy updates the bucket access policy configuration with given ips
func (c *UpdateAPObj) UpdateAccessPolicy(allowedIps, apiKey, bucketName string, rcc ResourceConfigurationV1) error {

	allowedIPs := strings.Split(allowedIps, ",")
	for i := range allowedIPs {
		allowedIPs[i] = strings.TrimSpace(allowedIPs[i])
	}

	authenticator := &core.IamAuthenticator{
		ApiKey: apiKey,
		URL:    PrivateIAMEPForVPC,
	}

	service, _ := rc.NewResourceConfigurationV1(&rc.ResourceConfigurationV1Options{
		Authenticator: authenticator,
		URL:           PrivateResourceConfigEP,
	})

	updateConfigOptions := &rc.UpdateBucketConfigOptions{
		Bucket: core.StringPtr(bucketName),
		Firewall: &rc.Firewall{
			AllowedIp: allowedIPs,
		},
	}

	response, err := rcc.UpdateBucketConfig(service, updateConfigOptions)
	if response != nil {
		fmt.Println("UpdateAccessPolicy Response ", strconv.Itoa(response.StatusCode))
	}
	return err
}
