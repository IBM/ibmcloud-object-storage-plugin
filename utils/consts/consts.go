/*******************************************************************************
 * IBM Confidential
 * OCO Source Materials
 * IBM Cloud Container Service, 5737-D43
 * (C) Copyright IBM Corp. 2017, 2018 All Rights Reserved.
 * The source code for this program is not  published or otherwise divested of
 * its trade secrets, irrespective of what has been deposited with
 * the U.S. Copyright Office.
 ******************************************************************************/

package consts

const (
	/**
	context constants
	*/

	// RequestIDLabel is the context key for storing the request ID
	RequestIDLabel = "requestID"

	// TriggerKeyLabel is the context key for storing the trigger key
	TriggerKeyLabel = "triggerKey"

	/**
	deployment file environment variable constants
	*/

	// PodNameEnvVar is the pod name environment variable
	PodNameEnvVar = "POD_NAME"
)
