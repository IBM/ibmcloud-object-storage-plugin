/*******************************************************************************
 * IBM Confidential
 * OCO Source Materials
 * IBM Cloud Container Service, 5737-D43
 * (C) Copyright IBM Corp. 2017, 2018 All Rights Reserved.
 * The source code for this program is not  published or otherwise divested of
 * its trade secrets, irrespective of what has been deposited with
 * the U.S. Copyright Office.
 ******************************************************************************/

package interfaces

import ()

const (
	// StatusSuccess returned when the operation succeeded
	StatusSuccess = "Success"
	// StatusFailure returned when the operation failed
	StatusFailure = "Failure"
	// StatusNotSupported returned when the operation is not supported
	StatusNotSupported = "Not supported"
)

// FlexPlugin is a partial interface of the flexvolume volume plugin
type FlexPlugin interface {

	// Init method is to initialize the flexvolume, it is a no op right now
	Init() FlexVolumeResponse

	// Mount method allows to mount the volume/fileset to a given location for a pod
	Mount(mountRequest FlexVolumeMountRequest) FlexVolumeResponse

	// Unmount methods unmounts the volume/ fileset from the pod
	Unmount(unmountRequest FlexVolumeUnmountRequest) FlexVolumeResponse
}

// CapabilitiesResponse represents a capabilities response of the init command
type CapabilitiesResponse struct {
	// Attach value is True/False (depending if the driver implements attach and detach)
	Attach bool `json:"attach"`
}

// FlexVolumeResponse represents a response of the volume plugin
type FlexVolumeResponse struct {
	// Status should be either "Success", "Failure" or "Not supported".
	Status string `json:"status"`
	// Reason for success or failure.
	Message string `json:"message,omitempty"`
	// Capabilities used in Init responses
	Capabilities CapabilitiesResponse `json:"capabilities,omitempty"`
}

// FlexVolumeMountRequest represents a mount request from the volume plugin
type FlexVolumeMountRequest struct {
	// MountDir is the path where the volume should be mounted
	MountDir string `json:"mountDir"`
	// Opts are the plugin options
	Opts map[string]string `json:"opts"`
}

// FlexVolumeUnmountRequest represents an unmount request from the volume plugin
type FlexVolumeUnmountRequest struct {
	// MountDir is the path to the mountpoint of the volume to be unmounted
	MountDir string `json:"mountDir"`
}
