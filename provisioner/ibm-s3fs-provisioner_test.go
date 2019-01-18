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
	"bytes"
	"fmt"
	"github.com/IBM/ibmcloud-object-storage-plugin/driver"
	"github.com/IBM/ibmcloud-object-storage-plugin/utils/backend"
	"github.com/IBM/ibmcloud-object-storage-plugin/utils/backend/fake"
	"github.com/IBM/ibmcloud-object-storage-plugin/utils/uuid"
	"github.com/kubernetes-incubator/external-storage/lib/controller"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	k8fake "k8s.io/client-go/kubernetes/fake"
	"strconv"
	"testing"
)

const (
	testSecretName         = "test-secret"
	testNamespace          = "test-namespace"
	testAccessKey          = "akey"
	testSecretKey          = "skey"
	testAPIKey             = "apikey"
	testServiceInstanceID  = "sid"
	testBucket             = "test-bucket"
	testEndpoint           = "https://test-endpoint"
	testRegion             = "test-region"
	testChunkSizeMB        = 2
	testParallelCount      = 3
	testMultiReqMax        = 4
	testStatCacheSize      = 5
	testS3FSFUSERetryCount = 1
	testDebugLevel         = "debug"
	testCurlDebug          = "false"
	testTLSCipherSuite     = "test-tls-cipher-suite"

	annotationBucket                 = "ibm.io/bucket"
	annotationAutoCreateBucket       = "ibm.io/auto-create-bucket"
	annotationAutoDeleteBucket       = "ibm.io/auto-delete-bucket"
	annotationEndpoint               = "ibm.io/endpoint"
	annotationRegion                 = "ibm.io/region"
	annotationSecretName             = "ibm.io/secret-name"
	annotationSecretNamespace        = "ibm.io/secret-namespace"
	annotationStatCacheExpireSeconds = "ibm.io/stat-cache-expire-seconds"

	parameterChunkSizeMB        = "ibm.io/chunk-size-mb"
	parameterParallelCount      = "ibm.io/parallel-count"
	parameterMultiReqMax        = "ibm.io/multireq-max"
	parameterStatCacheSize      = "ibm.io/stat-cache-size"
	parameterS3FSFUSERetryCount = "ibm.io/s3fs-fuse-retry-count"
	parameterTLSCipherSuite     = "ibm.io/tls-cipher-suite"
	parameterDebugLevel         = "ibm.io/debug-level"
	parameterCurlDebug          = "ibm.io/curl-debug"
	parameterKernelCache        = "ibm.io/kernel-cache"

	optionChunkSizeMB            = "chunk-size-mb"
	optionParallelCount          = "parallel-count"
	optionMultiReqMax            = "multireq-max"
	optionStatCacheSize          = "stat-cache-size"
	optionS3FSFUSERetryCount     = "s3fs-fuse-retry-count"
	optionTLSCipherSuite         = "tls-cipher-suite"
	optionDebugLevel             = "debug-level"
	optionCurlDebug              = "curl-debug"
	optionKernelCache            = "kernel-cache"
	optionEndpoint               = "endpoint"
	optionRegion                 = "region"
	optionBucket                 = "bucket"
	optionStatCacheExpireSeconds = "stat-cache-expire-seconds"
)

type clientGoConfig struct {
	missingSecret         bool
	missingAccessKey      bool
	missingSecretKey      bool
	withAPIKey            bool
	withServiceInstanceID bool
}

func getFakeClientGo(cfg *clientGoConfig) kubernetes.Interface {
	objects := []runtime.Object{}
	if !cfg.missingSecret {
		secret := &v1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      testSecretName,
				Namespace: testNamespace,
			},
			Data: make(map[string][]byte),
		}

		if cfg.withAPIKey {
			secret.Data[driver.SecretAPIKey] = []byte(testAPIKey)
		}

		if cfg.withServiceInstanceID {
			secret.Data[driver.SecretServiceInstanceID] = []byte(testServiceInstanceID)
		}

		if !cfg.missingAccessKey {
			secret.Data[driver.SecretAccessKey] = []byte(testAccessKey)
		}

		if !cfg.missingSecretKey {
			secret.Data[driver.SecretSecretKey] = []byte(testSecretKey)
		}
		objects = append(objects, runtime.Object(secret))
	}

	return k8fake.NewSimpleClientset(objects...)
}

func getCustomProvisioner(cfg *clientGoConfig, factory backend.ObjectStorageSessionFactory, uuidGen uuid.Generator) *IBMS3fsProvisioner {
	return &IBMS3fsProvisioner{
		Client:        getFakeClientGo(cfg),
		Logger:        zap.NewNop(),
		UUIDGenerator: uuidGen,
		Backend:       factory,
	}
}

func getFailedUUIDProvisioner() *IBMS3fsProvisioner {
	return getCustomProvisioner(
		&clientGoConfig{},
		&fake.ObjectStorageSessionFactory{},
		&uuid.ReaderGenerator{Reader: bytes.NewReader(nil)},
	)
}

func getFakeClientGoProvisioner(cfg *clientGoConfig) *IBMS3fsProvisioner {
	return getCustomProvisioner(
		cfg,
		&fake.ObjectStorageSessionFactory{},
		uuid.NewCryptoGenerator(),
	)
}

func getFakeBackendProvisioner(factory backend.ObjectStorageSessionFactory) *IBMS3fsProvisioner {
	return getCustomProvisioner(
		&clientGoConfig{},
		factory,
		uuid.NewCryptoGenerator(),
	)
}

func getProvisioner() *IBMS3fsProvisioner {
	return getCustomProvisioner(
		&clientGoConfig{},
		&fake.ObjectStorageSessionFactory{},
		uuid.NewCryptoGenerator(),
	)
}

func getVolumeOptions() controller.VolumeOptions {
	v := controller.VolumeOptions{
		PVC: &v1.PersistentVolumeClaim{
			ObjectMeta: metav1.ObjectMeta{
				Annotations: map[string]string{
					annotationBucket:     testBucket,
					annotationEndpoint:   testEndpoint,
					annotationRegion:     testRegion,
					annotationSecretName: testSecretName,
				},
				Namespace: testNamespace,
			},
		},
		Parameters: map[string]string{
			parameterChunkSizeMB:        strconv.Itoa(testChunkSizeMB),
			parameterParallelCount:      strconv.Itoa(testParallelCount),
			parameterMultiReqMax:        strconv.Itoa(testMultiReqMax),
			parameterStatCacheSize:      strconv.Itoa(testStatCacheSize),
			parameterS3FSFUSERetryCount: strconv.Itoa(testS3FSFUSERetryCount),
			parameterTLSCipherSuite:     testTLSCipherSuite,
			parameterDebugLevel:         testDebugLevel,
		},
	}

	return v
}

func getAutoDeletePersistentVolume() *v1.PersistentVolume {
	return &v1.PersistentVolume{
		ObjectMeta: metav1.ObjectMeta{
			Annotations: map[string]string{
				annotationAutoDeleteBucket: "true",
				annotationSecretName:       testSecretName,
				annotationSecretNamespace:  testNamespace,
			},
		},
	}
}

func Test_Provision_BadPVCAnnotations(t *testing.T) {
	p := getProvisioner()
	v := getVolumeOptions()
	v.PVC.Annotations[annotationAutoCreateBucket] = "non-bool-value"

	_, err := p.Provision(v)
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "cannot unmarshal PVC annotations")
	}
}

func Test_Provision_BadSCParameters(t *testing.T) {
	p := getProvisioner()
	v := getVolumeOptions()
	v.Parameters[parameterParallelCount] = "non-int-value"

	_, err := p.Provision(v)
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "cannot unmarshal storage class parameters")
	}
}

func Test_Provision_BadPVCEndpoint(t *testing.T) {
	p := getProvisioner()
	v := getVolumeOptions()
	v.PVC.Annotations[annotationEndpoint] = "test-endpoint"

	_, err := p.Provision(v)
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), fmt.Sprintf("Bad value for ibm.io/endpoint \"%s\": scheme is missing. "+
			"Must be of the form http://<hostname> or https://<hostname>", v.PVC.Annotations[annotationEndpoint]))
	}
}

func Test_Provision_PVCAnnotations_BadChunkSizeMB(t *testing.T) {
	p := getProvisioner()
	v := getVolumeOptions()
	v.PVC.Annotations["ibm.io/chunk-size-mb"] = "non-int-value"

	_, err := p.Provision(v)
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "Cannot convert value of chunk-size-mb into integer")
	}
}

func Test_Provision_PVCAnnotations_ChunkSizeMB_Positive(t *testing.T) {
	p := getProvisioner()
	v := getVolumeOptions()
	v.PVC.Annotations["ibm.io/chunk-size-mb"] = "20"

	pv, err := p.Provision(v)
	assert.NoError(t, err)
	assert.Equal(t, "20", pv.Spec.FlexVolume.Options[optionChunkSizeMB])
}

func Test_Provision_PVCAnnotations_BadParallelCount(t *testing.T) {
	p := getProvisioner()
	v := getVolumeOptions()
	v.PVC.Annotations["ibm.io/parallel-count"] = "non-int-value"

	_, err := p.Provision(v)
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "Cannot convert value of parallel-count into integer")
	}
}

func Test_Provision_PVCAnnotations_ParallelCount_Positive(t *testing.T) {
	p := getProvisioner()
	v := getVolumeOptions()
	v.PVC.Annotations["ibm.io/parallel-count"] = "30"

	pv, err := p.Provision(v)
	assert.NoError(t, err)
	assert.Equal(t, "30", pv.Spec.FlexVolume.Options[optionParallelCount])
}

func Test_Provision_PVCAnnotations_BadMultiReqMax(t *testing.T) {
	p := getProvisioner()
	v := getVolumeOptions()
	v.PVC.Annotations["ibm.io/multireq-max"] = "non-int-value"

	_, err := p.Provision(v)
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "Cannot convert value of multireq-max into integer")
	}
}

func Test_Provision_PVCAnnotations_MultiReqMax_Positive(t *testing.T) {
	p := getProvisioner()
	v := getVolumeOptions()
	v.PVC.Annotations["ibm.io/multireq-max"] = "40"

	pv, err := p.Provision(v)
	assert.NoError(t, err)
	assert.Equal(t, "40", pv.Spec.FlexVolume.Options[optionMultiReqMax])
}

func Test_Provision_PVCAnnotations_BadStatCacheSize(t *testing.T) {
	p := getProvisioner()
	v := getVolumeOptions()
	v.PVC.Annotations["ibm.io/stat-cache-size"] = "non-int-value"

	_, err := p.Provision(v)
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "Cannot convert value of stat-cache-size into integer")
	}
}

func Test_Provision_PVCAnnotations_StatCacheSize_Positive(t *testing.T) {
	p := getProvisioner()
	v := getVolumeOptions()
	v.PVC.Annotations["ibm.io/stat-cache-size"] = "50"

	pv, err := p.Provision(v)
	assert.NoError(t, err)
	assert.Equal(t, "50", pv.Spec.FlexVolume.Options[optionStatCacheSize])
}

func Test_Provision_PVCAnnotations_BadStatCacheExpireSeconds(t *testing.T) {
	p := getProvisioner()
	v := getVolumeOptions()
	v.PVC.Annotations[annotationStatCacheExpireSeconds] = "non-int-value"

	_, err := p.Provision(v)
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "Cannot convert value of stat-cache-expire-seconds into integer")
	}
}

func Test_Provision_PVCAnnotations_StatCacheExpireSeconds_Positive(t *testing.T) {
	p := getProvisioner()
	v := getVolumeOptions()
	v.PVC.Annotations[annotationStatCacheExpireSeconds] = "6"

	pv, err := p.Provision(v)
	assert.NoError(t, err)
	assert.Equal(t, "6", pv.Spec.FlexVolume.Options[optionStatCacheExpireSeconds])
}

func Test_Provision_PVCAnnotations_BadS3FSFUSERetryCount(t *testing.T) {
	p := getProvisioner()
	v := getVolumeOptions()
	v.PVC.Annotations["ibm.io/s3fs-fuse-retry-count"] = "non-int-value"

	_, err := p.Provision(v)
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "Cannot convert value of s3fs-fuse-retry-count into integer")
	}
}

func Test_Provision_PVCAnnotations_S3FSFUSERetryCount_Positive(t *testing.T) {
	p := getProvisioner()
	v := getVolumeOptions()
	v.PVC.Annotations["ibm.io/s3fs-fuse-retry-count"] = "10"

	pv, err := p.Provision(v)
	assert.NoError(t, err)
	assert.Equal(t, "10", pv.Spec.FlexVolume.Options[optionS3FSFUSERetryCount])
}

func Test_Provision_AutoDeleteBucketWithoutAutoCreateBucket(t *testing.T) {
	p := getProvisioner()
	v := getVolumeOptions()
	v.PVC.Annotations[annotationAutoDeleteBucket] = "true"

	_, err := p.Provision(v)
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "bucket auto-create must be enabled when bucket auto-delete is enabled")
	}
}

func Test_Provision_AutoDeleteBucketWithNonEmptyBucket(t *testing.T) {
	p := getProvisioner()
	v := getVolumeOptions()
	v.PVC.Annotations[annotationAutoDeleteBucket] = "true"
	v.PVC.Annotations[annotationAutoCreateBucket] = "true"

	_, err := p.Provision(v)
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "bucket cannot be set when auto-delete is enabled")
	}
}

func Test_Provision_UUIDGeneratorFailure(t *testing.T) {
	p := getFailedUUIDProvisioner()
	v := getVolumeOptions()
	v.PVC.Annotations[annotationAutoDeleteBucket] = "true"
	v.PVC.Annotations[annotationAutoCreateBucket] = "true"
	delete(v.PVC.Annotations, annotationBucket)

	_, err := p.Provision(v)
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "cannot create UUID for bucket name")
	}
}

func Test_Provision_MissingSecret(t *testing.T) {
	p := getFakeClientGoProvisioner(&clientGoConfig{missingSecret: true})
	v := getVolumeOptions()

	_, err := p.Provision(v)
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "cannot get secret")
	}
}

func Test_Provision_MissingAccessKey(t *testing.T) {
	p := getFakeClientGoProvisioner(&clientGoConfig{missingAccessKey: true})
	v := getVolumeOptions()

	_, err := p.Provision(v)
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), fmt.Sprintf("%s secret missing", driver.SecretAccessKey))
	}
}

func Test_Provision_MissingSecretKey(t *testing.T) {
	p := getFakeClientGoProvisioner(&clientGoConfig{missingSecretKey: true})
	v := getVolumeOptions()

	_, err := p.Provision(v)
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), fmt.Sprintf("%s secret missing", driver.SecretSecretKey))
	}
}

func Test_Provision_APIKeyWithoutServiceInstanceIDInBucketCreation(t *testing.T) {
	p := getCustomProvisioner(
		&clientGoConfig{withAPIKey: true},
		&fake.ObjectStorageSessionFactory{},
		uuid.NewCryptoGenerator(),
	)
	v := getVolumeOptions()
	v.PVC.Annotations[annotationAutoCreateBucket] = "true"

	_, err := p.Provision(v)
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "cannot create bucket using API key without service-instance-id")
	}
}

func Test_Provision_FailCreateBucket(t *testing.T) {
	p := getFakeBackendProvisioner(&fake.ObjectStorageSessionFactory{FailCreateBucket: true})
	v := getVolumeOptions()
	v.PVC.Annotations[annotationAutoCreateBucket] = "true"

	_, err := p.Provision(v)
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "cannot create bucket")
	}
}

func Test_Provision_FailCheckBucketAccess(t *testing.T) {
	p := getFakeBackendProvisioner(&fake.ObjectStorageSessionFactory{FailCheckBucketAccess: true})
	v := getVolumeOptions()

	_, err := p.Provision(v)
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "cannot access bucket")
	}
}

func Test_Provision_Positive(t *testing.T) {
	factory := &fake.ObjectStorageSessionFactory{}
	p := getFakeBackendProvisioner(factory)
	v := getVolumeOptions()

	pv, err := p.Provision(v)
	assert.NoError(t, err)
	assert.Equal(t,
		map[string]string{
			optionChunkSizeMB:        strconv.Itoa(testChunkSizeMB),
			optionParallelCount:      strconv.Itoa(testParallelCount),
			optionMultiReqMax:        strconv.Itoa(testMultiReqMax),
			optionStatCacheSize:      strconv.Itoa(testStatCacheSize),
			optionS3FSFUSERetryCount: strconv.Itoa(testS3FSFUSERetryCount),
			optionTLSCipherSuite:     testTLSCipherSuite,
			optionDebugLevel:         testDebugLevel,
			optionCurlDebug:          testCurlDebug,
			optionEndpoint:           testEndpoint,
			optionRegion:             testRegion,
			optionBucket:             testBucket,
		},
		pv.Spec.FlexVolume.Options,
	)

	assert.Equal(t, testEndpoint, factory.LastEndpoint)
	assert.Equal(t, testRegion, factory.LastRegion)
	assert.Equal(t, testAccessKey, factory.LastCredentials.AccessKey)
	assert.Equal(t, testSecretKey, factory.LastCredentials.SecretKey)
	assert.Equal(t, "", factory.LastCredentials.APIKey)
	assert.Equal(t, testBucket, factory.LastCheckedBucket)
	assert.Equal(t, "", factory.LastCreatedBucket)
	assert.Equal(t, "", factory.LastDeletedBucket)
}

func Test_Provision_CurlDebug_Positive(t *testing.T) {
	p := getProvisioner()
	v := getVolumeOptions()
	v.Parameters[parameterCurlDebug] = "true"

	pv, err := p.Provision(v)
	assert.NoError(t, err)
	assert.Equal(t, "true", pv.Spec.FlexVolume.Options[optionCurlDebug])
}

func Test_Provision_KernelCache_Positive(t *testing.T) {
	p := getProvisioner()
	v := getVolumeOptions()
	v.Parameters[parameterKernelCache] = "true"

	pv, err := p.Provision(v)
	assert.NoError(t, err)
	assert.Equal(t, "true", pv.Spec.FlexVolume.Options[optionKernelCache])
}

func Test_Provision_IAM_Positive(t *testing.T) {
	factory := &fake.ObjectStorageSessionFactory{}
	p := getCustomProvisioner(
		&clientGoConfig{withAPIKey: true, withServiceInstanceID: true},
		factory,
		uuid.NewCryptoGenerator(),
	)
	v := getVolumeOptions()

	_, err := p.Provision(v)
	assert.NoError(t, err)

	assert.Equal(t, testAPIKey, factory.LastCredentials.APIKey)
	assert.Equal(t, testServiceInstanceID, factory.LastCredentials.ServiceInstanceID)
}

func Test_Provision_BucketAutoDelete_Positive(t *testing.T) {
	factory := &fake.ObjectStorageSessionFactory{}
	p := getFakeBackendProvisioner(factory)
	v := getVolumeOptions()
	v.PVC.Annotations[annotationAutoDeleteBucket] = "true"
	v.PVC.Annotations[annotationAutoCreateBucket] = "true"
	delete(v.PVC.Annotations, annotationBucket)

	pv, err := p.Provision(v)
	assert.NoError(t, err)
	assert.Contains(t, pv.Spec.FlexVolume.Options[optionBucket], autoBucketNamePrefix)
	assert.Equal(t, pv.Spec.FlexVolume.Options[optionBucket], factory.LastCreatedBucket)
}

func Test_Delete_BadPVAnnotations(t *testing.T) {
	p := getProvisioner()
	pv := getAutoDeletePersistentVolume()
	pv.Annotations[annotationAutoDeleteBucket] = "non-bool-value"

	err := p.Delete(pv)
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "cannot unmarshal PV annotations")
	}
}

func Test_Delete_MissingSecret(t *testing.T) {
	p := getFakeClientGoProvisioner(&clientGoConfig{missingSecret: true})
	pv := getAutoDeletePersistentVolume()
	err := p.Delete(pv)
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "cannot get secret")
	}
}

func Test_Delete_FailDeleteBucket(t *testing.T) {
	p := getFakeBackendProvisioner(&fake.ObjectStorageSessionFactory{FailDeleteBucket: true})
	pv := getAutoDeletePersistentVolume()
	err := p.Delete(pv)
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "cannot delete bucket")
	}
}

func Test_Provision_Delete_Positive(t *testing.T) {
	factory := &fake.ObjectStorageSessionFactory{}
	p := getFakeBackendProvisioner(factory)
	v := getVolumeOptions()
	v.PVC.Annotations[annotationAutoDeleteBucket] = "true"
	v.PVC.Annotations[annotationAutoCreateBucket] = "true"
	delete(v.PVC.Annotations, annotationBucket)

	pv, err := p.Provision(v)
	assert.NoError(t, err)

	bucketName := factory.LastCreatedBucket

	factory.ResetStats()

	err = p.Delete(pv)
	assert.NoError(t, err)

	assert.Equal(t, testEndpoint, factory.LastEndpoint)
	assert.Equal(t, testRegion, factory.LastRegion)
	assert.Equal(t, testAccessKey, factory.LastCredentials.AccessKey)
	assert.Equal(t, testSecretKey, factory.LastCredentials.SecretKey)
	assert.Equal(t, "", factory.LastCredentials.APIKey)
	assert.Equal(t, bucketName, factory.LastDeletedBucket)
}

func Test_Provision_Delete_IAM_Positive(t *testing.T) {
	factory := &fake.ObjectStorageSessionFactory{}
	p := getCustomProvisioner(
		&clientGoConfig{withAPIKey: true, withServiceInstanceID: true},
		factory,
		uuid.NewCryptoGenerator(),
	)
	v := getVolumeOptions()
	v.PVC.Annotations[annotationAutoDeleteBucket] = "true"
	v.PVC.Annotations[annotationAutoCreateBucket] = "true"
	delete(v.PVC.Annotations, annotationBucket)

	pv, err := p.Provision(v)
	assert.NoError(t, err)

	bucketName := factory.LastCreatedBucket

	factory.ResetStats()

	err = p.Delete(pv)
	assert.NoError(t, err)

	assert.Equal(t, testServiceInstanceID, factory.LastCredentials.ServiceInstanceID)
	assert.Equal(t, testAPIKey, factory.LastCredentials.APIKey)
	assert.Equal(t, bucketName, factory.LastDeletedBucket)
}
