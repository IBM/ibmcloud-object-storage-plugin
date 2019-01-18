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
	"errors"
	"fmt"
	"github.com/IBM/ibmcloud-object-storage-plugin/driver"
	"github.com/IBM/ibmcloud-object-storage-plugin/utils/backend"
	"github.com/IBM/ibmcloud-object-storage-plugin/utils/logger"
	"github.com/IBM/ibmcloud-object-storage-plugin/utils/parser"
	"github.com/IBM/ibmcloud-object-storage-plugin/utils/uuid"
	"github.com/kubernetes-incubator/external-storage/lib/controller"
	"go.uber.org/zap"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"strconv"
	"strings"
)

// PVC annotations
type pvcAnnotations struct {
	AutoCreateBucket       bool   `json:"ibm.io/auto-create-bucket,string"`
	AutoDeleteBucket       bool   `json:"ibm.io/auto-delete-bucket,string"`
	Bucket                 string `json:"ibm.io/bucket"`
	Endpoint               string `json:"ibm.io/endpoint"`
	Region                 string `json:"ibm.io/region"`
	SecretName             string `json:"ibm.io/secret-name"`
	ChunkSizeMB            string `json:"ibm.io/chunk-size-mb,omitempty"`
	ParallelCount          string `json:"ibm.io/parallel-count,omitempty"`
	MultiReqMax            string `json:"ibm.io/multireq-max,omitempty"`
	StatCacheSize          string `json:"ibm.io/stat-cache-size,omitempty"`
	S3FSFUSERetryCount     string `json:"ibm.io/s3fs-fuse-retry-count,omitempty"`
	StatCacheExpireSeconds string `json:"ibm.io/stat-cache-expire-seconds,omitempty"`
}

// PV annotations
type pvAnnotations struct {
	pvcAnnotations
	SecretNamespace string `json:"ibm.io/secret-namespace"`
}

// Storage Class options
type scOptions struct {
	ChunkSizeMB        int    `json:"ibm.io/chunk-size-mb,string"`
	ParallelCount      int    `json:"ibm.io/parallel-count,string"`
	MultiReqMax        int    `json:"ibm.io/multireq-max,string"`
	StatCacheSize      int    `json:"ibm.io/stat-cache-size,string"`
	TLSCipherSuite     string `json:"ibm.io/tls-cipher-suite,omitempty"`
	DebugLevel         string `json:"ibm.io/debug-level"`
	CurlDebug          bool   `json:"ibm.io/curl-debug,string,omitempty"`
	KernelCache        bool   `json:"ibm.io/kernel-cache,string,omitempty"`
	S3FSFUSERetryCount int    `json:"ibm.io/s3fs-fuse-retry-count,string,omitempty"`
}

const (
	driverName           = "ibm/ibmc-s3fs"
	autoBucketNamePrefix = "tmp-s3fs-"
	fsType               = ""
)

// IBMS3fsProvisioner is a dynamic provisioner of persistent volumes backed by Object Storage via s3fs
type IBMS3fsProvisioner struct {
	// Backend is the object store session factory
	Backend backend.ObjectStorageSessionFactory
	// Logger will be used for logging
	Logger *zap.Logger
	// Client is the Kubernetes Go-Client that will be used to fetch user credentials
	Client kubernetes.Interface
	// UUIDGenerator is a UUID generator that will be used to generate bucket names
	UUIDGenerator uuid.Generator
}

var _ controller.Provisioner = &IBMS3fsProvisioner{}

func parseSecret(secret *v1.Secret, keyName string) (string, error) {
	bytesVal, ok := secret.Data[keyName]
	if !ok {
		return "", fmt.Errorf("%s secret missing", keyName)
	}

	return string(bytesVal), nil
}

func (p *IBMS3fsProvisioner) getCredentials(secretName, secretNamespace string) (*backend.ObjectStorageCredentials, error) {
	secrets, err := p.Client.Core().Secrets(secretNamespace).Get(secretName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("cannot get secret %s: %v", secretName, err)
	}

	var accessKey, secretKey, apiKey, serviceInstanceID string

	apiKey, err = parseSecret(secrets, driver.SecretAPIKey)
	if err != nil {
		accessKey, err = parseSecret(secrets, driver.SecretAccessKey)
		if err != nil {
			return nil, err
		}

		secretKey, err = parseSecret(secrets, driver.SecretSecretKey)
		if err != nil {
			return nil, err
		}
	} else {
		serviceInstanceID, err = parseSecret(secrets, driver.SecretServiceInstanceID)
	}

	return &backend.ObjectStorageCredentials{
		AccessKey:         accessKey,
		SecretKey:         secretKey,
		APIKey:            apiKey,
		ServiceInstanceID: serviceInstanceID,
	}, nil

}

// Provision provisions a new persistent volume
func (p *IBMS3fsProvisioner) Provision(options controller.VolumeOptions) (*v1.PersistentVolume, error) {
	var pvc pvcAnnotations
	var sc scOptions
	var pvcName = options.PVC.Name
	var msg string
	contextLogger, _ := logger.GetZapDefaultContextLogger()
	contextLogger.Info(pvcName + ":Provisioning storage with these spec")
	contextLogger.Info(pvcName+":PVC Details: ", zap.String("pvc", options.PVName))

	err := parser.UnmarshalMap(&options.PVC.Annotations, &pvc)
	if err != nil {
		return nil, fmt.Errorf(pvcName+":cannot unmarshal PVC annotations: %v", err)
	}
	err = parser.UnmarshalMap(&options.Parameters, &sc)
	if err != nil {
		return nil, fmt.Errorf(pvcName+":cannot unmarshal storage class parameters: %v", err)
	}

	if !(strings.HasPrefix(pvc.Endpoint, "https://") || strings.HasPrefix(pvc.Endpoint, "http://")) {
		return nil, fmt.Errorf(pvcName+":Bad value for ibm.io/endpoint \"%v\": scheme is missing. "+
			"Must be of the form http://<hostname> or https://<hostname>", pvc.Endpoint)
	}

	//Override value of s3fs-fuse-retry-count defined in storageclass
	if pvc.S3FSFUSERetryCount != "" {
		if sc.S3FSFUSERetryCount, err = strconv.Atoi(pvc.S3FSFUSERetryCount); err != nil {
			return nil, fmt.Errorf(pvcName+":Cannot convert value of s3fs-fuse-retry-count into integer: %v", err)
		}
	}

	//Override value of chunk-size-mb defined in storageclass
	if pvc.ChunkSizeMB != "" {
		if sc.ChunkSizeMB, err = strconv.Atoi(pvc.ChunkSizeMB); err != nil {
			return nil, fmt.Errorf(pvcName+":Cannot convert value of chunk-size-mb into integer: %v", err)
		}
	}

	//Override value of parallel-count defined in storageclass
	if pvc.ParallelCount != "" {
		if sc.ParallelCount, err = strconv.Atoi(pvc.ParallelCount); err != nil {
			return nil, fmt.Errorf(pvcName+":Cannot convert value of parallel-count into integer: %v", err)
		}
	}

	//Override value of multireq-max defined in storageclass
	if pvc.MultiReqMax != "" {
		if sc.MultiReqMax, err = strconv.Atoi(pvc.MultiReqMax); err != nil {
			return nil, fmt.Errorf(pvcName+":Cannot convert value of multireq-max into integer: %v", err)
		}
	}

	//Override value of stat-cache-size defined in storageclass
	if pvc.StatCacheSize != "" {
		if sc.StatCacheSize, err = strconv.Atoi(pvc.StatCacheSize); err != nil {
			return nil, fmt.Errorf(pvcName+":Cannot convert value of stat-cache-size into integer: %v", err)
		}
	}

	//Check if value of stat-cache-expire-seconds parameter can be converted to integer
	if pvc.StatCacheExpireSeconds != "" {
		if _, err := strconv.Atoi(pvc.StatCacheExpireSeconds); err != nil {
			return nil, fmt.Errorf(pvcName+":Cannot convert value of stat-cache-expire-seconds into integer: %v", err)
		}
	}

	if pvc.AutoDeleteBucket {
		if !pvc.AutoCreateBucket {
			return nil, errors.New(pvcName + ":bucket auto-create must be enabled when bucket auto-delete is enabled")
		}

		if pvc.Bucket != "" {
			return nil, fmt.Errorf(pvcName+":bucket cannot be set when auto-delete is enabled, got: %s", pvc.Bucket)
		}

		id, err := p.UUIDGenerator.New()
		if err != nil {
			return nil, fmt.Errorf(pvcName+":cannot create UUID for bucket name: %v", err)
		}

		pvc.Bucket = autoBucketNamePrefix + id
	}

	creds, err := p.getCredentials(pvc.SecretName, options.PVC.Namespace)
	if err != nil {
		return nil, fmt.Errorf(pvcName+":cannot get credentials: %v", err)
	}

	sess := p.Backend.NewObjectStorageSession(pvc.Endpoint, pvc.Region, creds, p.Logger)

	if pvc.AutoCreateBucket {
		if creds.APIKey != "" && creds.ServiceInstanceID == "" {
			return nil, errors.New(pvcName + ":cannot create bucket using API key without service-instance-id")
		}
		msg, err = sess.CreateBucket(pvc.Bucket)
		if msg != "" {
			contextLogger.Info(pvcName + ":" + msg)
		}
		if err != nil {
			return nil, fmt.Errorf(pvcName+":cannot create bucket %s: %v", pvc.Bucket, err)
		}
	}

	err = sess.CheckBucketAccess(pvc.Bucket)
	if err != nil {
		return nil, fmt.Errorf(pvcName+":cannot access bucket %s: %v", pvc.Bucket, err)
	}

	driverOptions, err := parser.MarshalToMap(&driver.Options{
		ChunkSizeMB:            sc.ChunkSizeMB,
		ParallelCount:          sc.ParallelCount,
		MultiReqMax:            sc.MultiReqMax,
		StatCacheSize:          sc.StatCacheSize,
		TLSCipherSuite:         sc.TLSCipherSuite,
		CurlDebug:              sc.CurlDebug,
		KernelCache:            sc.KernelCache,
		DebugLevel:             sc.DebugLevel,
		S3FSFUSERetryCount:     strconv.Itoa(sc.S3FSFUSERetryCount),
		StatCacheExpireSeconds: pvc.StatCacheExpireSeconds,
		Endpoint:               pvc.Endpoint,
		Region:                 pvc.Region,
		Bucket:                 pvc.Bucket,
	})
	if err != nil {
		return nil, fmt.Errorf(pvcName+":cannot marshal driver options: %v", err)
	}

	pvAnnots, err := parser.MarshalToMap(&pvAnnotations{
		pvcAnnotations:  pvc,
		SecretNamespace: options.PVC.Namespace,
	})
	if err != nil {
		return nil, fmt.Errorf(pvcName+":cannot marshal pv options: %v", err)
	}

	return &v1.PersistentVolume{
		ObjectMeta: metav1.ObjectMeta{
			Name:        options.PVName,
			Annotations: pvAnnots,
		},
		Spec: v1.PersistentVolumeSpec{
			PersistentVolumeReclaimPolicy: options.PersistentVolumeReclaimPolicy,
			AccessModes:                   options.PVC.Spec.AccessModes,
			Capacity: v1.ResourceList{
				v1.ResourceStorage: options.PVC.Spec.Resources.Requests[v1.ResourceName(v1.ResourceStorage)],
			},
			PersistentVolumeSource: v1.PersistentVolumeSource{
				FlexVolume: &v1.FlexVolumeSource{
					Driver:    driverName,
					FSType:    fsType,
					SecretRef: &v1.LocalObjectReference{Name: pvc.SecretName},
					ReadOnly:  false,
					Options:   driverOptions,
				},
			},
		},
	}, nil
}

// Delete deletes a persistent volume
func (p *IBMS3fsProvisioner) Delete(pv *v1.PersistentVolume) error {
	var pvAnnots pvAnnotations

	contextLogger, _ := logger.GetZapDefaultContextLogger()
	contextLogger.Info("Deleting the pvc..")

	err := parser.UnmarshalMap(&pv.Annotations, &pvAnnots)
	if err != nil {
		return fmt.Errorf("cannot unmarshal PV annotations: %v", err)
	}

	if pvAnnots.AutoDeleteBucket {
		err = p.deleteBucket(&pvAnnots)
		if err != nil {
			return fmt.Errorf("cannot delete bucket: %v", err)
		}
	}

	return nil
}

func (p *IBMS3fsProvisioner) deleteBucket(pvAnnots *pvAnnotations) error {
	creds, err := p.getCredentials(pvAnnots.SecretName, pvAnnots.SecretNamespace)
	if err != nil {
		return fmt.Errorf("cannot get credentials: %v", err)
	}

	sess := p.Backend.NewObjectStorageSession(pvAnnots.Endpoint, pvAnnots.Region, creds, p.Logger)

	return sess.DeleteBucket(pvAnnots.Bucket)
}
