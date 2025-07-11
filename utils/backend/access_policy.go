/*******************************************************************************
 * IBM Confidential
 * OCO Source Materials
 * IBM Cloud Kubernetes Service, 5737-D43
 * (C) Copyright IBM Corp. 2017, 2023 All Rights Reserved.
 * The source code for this program is not published or otherwise divested of
 * its trade secrets, irrespective of what has been deposited with
 * the U.S. Copyright Office.
 ******************************************************************************/

package backend

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/IBM/go-sdk-core/v5/core"
	rc "github.com/IBM/ibm-cos-sdk-go-config/v2/resourceconfigurationv1"
)

const ResourceConfigEPDirect = "https://config.direct.cloud-object-storage.cloud.ibm.com/v1"
const ResourceConfigEPPrivate = "https://config.private.cloud-object-storage.cloud.ibm.com/v1"
const IAMEPForVPC = "https://private.iam.cloud.ibm.com/identity/token"
const Private = "private"

type AccessPolicyFactory interface {
	NewAccessPolicy() AccessPolicy
}

type AccessPolicy interface {
	UpdateAccessPolicy(allowedIps, apiKey, bucketName string, rcc ResourceConfigurationV1) error
	UpdateQuotaLimit(quota int64, apiKey, bucketName, osEndpoint, iamEndpoint string, rcc ResourceConfigurationV1) error
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

//var rcc ResourceConfigurationV1 = &UpdateAPObj{}

// UpdateAccessPolicy updates the bucket access policy configuration with given ips
func (c *UpdateAPObj) UpdateAccessPolicy(allowedIps, apiKey, bucketName string, rcc ResourceConfigurationV1) error {

	allowedIPs := strings.Split(allowedIps, ",")
	for i := range allowedIPs {
		allowedIPs[i] = strings.TrimSpace(allowedIPs[i])
	}

	authenticator := &core.IamAuthenticator{
		ApiKey: apiKey,
		URL:    IAMEPForVPC,
	}

	service, _ := rc.NewResourceConfigurationV1(&rc.ResourceConfigurationV1Options{
		Authenticator: authenticator,
		URL:           ResourceConfigEPDirect,
	})

	// Create a map to hold the bucket patch
	bucketPatchMap := make(map[string]interface{})

	// Set firewall in the map
	bucketPatchMap["firewall"] = &rc.Firewall{
		AllowedIp: allowedIPs,
	}

	updateConfigOptions := &rc.UpdateBucketConfigOptions{
		Bucket:      core.StringPtr(bucketName),
		BucketPatch: bucketPatchMap,
	}

	response, err := rcc.UpdateBucketConfig(service, updateConfigOptions)
	if response != nil {
		fmt.Println("UpdateAccessPolicy Response ", strconv.Itoa(response.StatusCode))
	}
	return err
}

// UpdateQuotaLimit updates the bucket quota limits
func (c *UpdateAPObj) UpdateQuotaLimit(quota int64, apiKey, bucketName, osEndpoint, iamEndpoint string, rcc ResourceConfigurationV1) error {

	ConfigEP := ""
	IAMEP := iamEndpoint + "/identity/token"

	if strings.Contains(osEndpoint, Private) {
		ConfigEP = ResourceConfigEPPrivate
	} else {
		ConfigEP = ResourceConfigEPDirect
	}

	fmt.Println("ConfigEP used: ", ConfigEP)
	fmt.Println("IAMEndpoint used: ", IAMEP)

	authenticator := &core.IamAuthenticator{
		ApiKey: apiKey,
		URL:    IAMEP,
	}

	service, _ := rc.NewResourceConfigurationV1(&rc.ResourceConfigurationV1Options{
		Authenticator: authenticator,
		URL:           ConfigEP,
	})

	// Create a map to hold the bucket patch
	bucketPatchMap := make(map[string]interface{})

	// Set firewall in the map
	bucketPatchMap["hard_quota"] = core.Int64Ptr(quota)

	updateConfigOptions := &rc.UpdateBucketConfigOptions{
		Bucket:      core.StringPtr(bucketName),
		BucketPatch: bucketPatchMap,
	}

	response, err := rcc.UpdateBucketConfig(service, updateConfigOptions)
	if response != nil {
		fmt.Println("UpdateQuotaLimit Response ", strconv.Itoa(response.StatusCode))
	}
	return err
}
