/*******************************************************************************
 * IBM Confidential
 * OCO Source Materials
 * IBM Cloud Container Service, 5737-D43
 * (C) Copyright IBM Corp. 2020  All Rights Reserved.
 * The source code for this program is not  published or otherwise divested of
 * its trade secrets, irrespective of what has been deposited with
 * the U.S. Copyright Office.
 ******************************************************************************
 */

package provisioner

import (
	"fmt"
	"github.com/IBM/go-sdk-core/v3/core"
	rc "github.com/IBM/ibm-cos-sdk-go-config/resourceconfigurationv1"
	"strings"
)

func UpdateFirewallRules(allowed_ips, apiKey, bucketName string) error {

	allowedIPs := strings.Split(allowed_ips, ",")
	for i := range allowedIPs {
		allowedIPs[i] = strings.TrimSpace(allowedIPs[i])
	}

	authenticator := &core.IamAuthenticator{
		ApiKey: apiKey,
	}

	service, _ := rc.NewResourceConfigurationV1(&rc.ResourceConfigurationV1Options{
		Authenticator: authenticator,
	})

	updateConfigOptions := &rc.UpdateBucketConfigOptions{
		Bucket: core.StringPtr(bucketName),
		Firewall: &rc.Firewall{
			AllowedIp: allowedIPs,
		},
	}

	response, err := service.UpdateBucketConfig(updateConfigOptions)
	if response.StatusCode != nil {
		fmt.Println("UpdateFirewallRules: Response", response.StatusCode)
	}
	return err
}
