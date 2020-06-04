/*******************************************************************************
 * IBM Confidential
 * OCO Source Materials
 * IBM Cloud Container Service, 5737-D43
 * (C) Copyright IBM Corp. 2017, 2018 All Rights Reserved.
 * The source code for this program is not  published or otherwise divested of
 * its trade secrets, irrespective of what has been deposited with
 * the U.S. Copyright Office.
 ******************************************************************************/

package provisioner

import (
  "fmt"
  "github.com/IBM/go-sdk-core/v3/core"
  rc "github.com/IBM/ibm-cos-sdk-go-config/resourceconfigurationv1"
	// "github.com/IBM/ibmcloud-object-storage-plugin/utils/logger"
	// "go.uber.org/zap"
	"strings"
)

func UpdateFirewallRules(allowed_ips, apiKey, bucketName string) error {
  allowedIPs := strings.Split(allowed_ips, ",")

  authenticator := &core.IamAuthenticator{
        ApiKey:  apiKey,
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
  response, error := service.UpdateBucketConfig(updateConfigOptions)
  fmt.Println(response)
  return error
}
