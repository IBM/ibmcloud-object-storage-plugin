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
	"errors"
	"fmt"
	"github.com/IBM/ibmcloud-object-storage-plugin/driver"
	"github.com/IBM/ibmcloud-object-storage-plugin/utils/backend"
	"github.com/IBM/ibmcloud-object-storage-plugin/utils/backend/fake"
	"github.com/IBM/ibmcloud-object-storage-plugin/utils/uuid"
	"github.com/kubernetes-sigs/sig-storage-lib-external-provisioner/controller"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"k8s.io/client-go/kubernetes"
	k8fake "k8s.io/client-go/kubernetes/fake"
	"os"
	//"k8s.io/client-go/pkg/api/v1"
	"k8s.io/api/core/v1"
	//"k8s.io/client-go/pkg/runtime"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"strconv"
	"testing"
)

const (
	testSecretName        = "test-secret"
	testAccessKey         = "akey"
	testSecretKey         = "skey"
	testAPIKey            = "apikey"
	testServiceInstanceID = "sid"
	testBucket            = "test-bucket"
	testOSEndpoint        = "https://test-object-store-endpoint"
	testIAMEndpoint       = "https://test-iam-endpoint"
	testServiceName       = "test-service"
	testServiceNamespace  = "test-default"
	testCAKey             = "cacrt-key"
	testAllowedNamespace  = "test-allowed-namespace1 test-allowed-namespace2"
	testAllowedIps        = "169.55.124.128/29"
	testResConfApiKey     = "resconfapikey"

	testChunkSizeMB            = 2
	testParallelCount          = 3
	testMultiReqMax            = 4
	testStatCacheSize          = 5
	testS3FSFUSERetryCount     = 1
	testStatCacheExpireSeconds = 1
	testDebugLevel             = "debug"
	testCurlDebug              = "false"
	testTLSCipherSuite         = "test-tls-cipher-suite"
	testStorageClass           = "test-storage-class"
	testObjectPath             = "/test/object-path"
	testValidateBucket         = "yes"

	annotationBucket                  = "ibm.io/bucket"
	annotationObjectPath              = "ibm.io/object-path"
	annotationAutoCreateBucket        = "ibm.io/auto-create-bucket"
	annotationAutoDeleteBucket        = "ibm.io/auto-delete-bucket"
	annotationEndpoint                = "ibm.io/endpoint"
	annotationRegion                  = "ibm.io/region"
	annotationIAMEndpoint             = "ibm.io/iam-endpoint"
	annotationSecretName              = "ibm.io/secret-name"
	annotationSecretNamespace         = "ibm.io/secret-namespace"
	annotationStatCacheExpireSeconds  = "ibm.io/stat-cache-expire-seconds"
	annotationValidateBucket          = "ibm.io/validate-bucket"
	annotationConnectTimeoutSeconds   = "ibm.io/connect-timeout"
	annotationReadwriteTimeoutSeconds = "ibm.io/readwrite-timeout"
	annotationServiceName             = "ibm.io/cos-service"
	annotationServiceNamespace        = "ibm.io/cos-service-ns"
	annotationAllowedIPs              = "ibm.io/allowed_ips"
	annotationConfigureFirewall       = "ibm.io/configure-firewall"

	parameterChunkSizeMB            = "ibm.io/chunk-size-mb"
	parameterParallelCount          = "ibm.io/parallel-count"
	parameterMultiReqMax            = "ibm.io/multireq-max"
	parameterStatCacheSize          = "ibm.io/stat-cache-size"
	parameterS3FSFUSERetryCount     = "ibm.io/s3fs-fuse-retry-count"
	parameterTLSCipherSuite         = "ibm.io/tls-cipher-suite"
	parameterDebugLevel             = "ibm.io/debug-level"
	parameterCurlDebug              = "ibm.io/curl-debug"
	parameterKernelCache            = "ibm.io/kernel-cache"
	parameterOSEndpoint             = "ibm.io/object-store-endpoint"
	parameterIAMEndpoint            = "ibm.io/iam-endpoint"
	parameterStorageClass           = "ibm.io/object-store-storage-class"
	parameterStatCacheExpireSeconds = "ibm.io/stat-cache-expire-seconds"
	parameterAutoCache              = "ibm.io/auto_cache"

	optionChunkSizeMB             = "chunk-size-mb"
	optionParallelCount           = "parallel-count"
	optionMultiReqMax             = "multireq-max"
	optionStatCacheSize           = "stat-cache-size"
	optionS3FSFUSERetryCount      = "s3fs-fuse-retry-count"
	optionTLSCipherSuite          = "tls-cipher-suite"
	optionDebugLevel              = "debug-level"
	optionCurlDebug               = "curl-debug"
	optionKernelCache             = "kernel-cache"
	optionOSEndpoint              = "object-store-endpoint"
	optionRegion                  = "region"
	optionBucket                  = "bucket"
	optionStatCacheExpireSeconds  = "stat-cache-expire-seconds"
	optionObjectPath              = "object-path"
	optionStorageClass            = "object-store-storage-class"
	optionIAMEndpoint             = "iam-endpoint"
	optionReadwriteTimeoutSeconds = "readwrite-timeout"
	optionConnectTimeoutSeconds   = "connect-timeout"
	optionUseXattr                = "use-xattr"
	optionAccessMode              = "access-mode"
	optionServiceIP               = "service-ip"
	optionAutoCache               = "auto_cache"
)

type clientGoConfig struct {
	missingSecret         bool
	missingAccessKey      bool
	missingSecretKey      bool
	withAllowedNamespace  bool
	withAPIKey            bool
	withServiceInstanceID bool
	wrongSecretType       bool
	isTLS                 bool
	withcaBundle          bool
	withResConfApiKey     bool
	withAllowedIPs        bool
}

var (
	writeFileError   = func(string, []byte, os.FileMode) error { return errors.New("") }
	writeFileSuccess = func(string, []byte, os.FileMode) error { return nil }
	testNamespace    = "test-namespace"
)

func getFakeClientGo(cfg *clientGoConfig) kubernetes.Interface {
	objects := []runtime.Object{}
	var secret *v1.Secret
	var svc *v1.Service
	if cfg.isTLS {
		svc = &v1.Service{
			ObjectMeta: metav1.ObjectMeta{Name: testServiceName, Namespace: testServiceNamespace},
			Spec: v1.ServiceSpec{
				Selector: map[string]string{},
				Ports:    []v1.ServicePort{{Port: 80, Protocol: "TCP"}},
			},
		}
		objects = append(objects, runtime.Object(svc))
	}
	if !cfg.missingSecret {
		secret = &v1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      testSecretName,
				Namespace: testNamespace,
			},
			Data: make(map[string][]byte),
		}
		if cfg.wrongSecretType {
			secret.Type = "test-type"
		} else {
			secret.Type = "ibm/ibmc-s3fs"
		}
		if cfg.withcaBundle {
			secret.Data[driver.CrtBundle] = []byte(testCAKey)
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

		if cfg.withAllowedNamespace {
			secret.Data[driver.SecretAllowedNS] = []byte(testAllowedNamespace)
		}

		if cfg.withResConfApiKey {
			secret.Data[driver.ResConfApiKey] = []byte(testResConfApiKey)
		}

		if cfg.withAllowedIPs {
			secret.Data[driver.AllowedIPs] = []byte(testAllowedIps)
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
					annotationSecretName: testSecretName,
				},
				Namespace: testNamespace,
			},
			Spec: v1.PersistentVolumeClaimSpec{
				AccessModes: []v1.PersistentVolumeAccessMode{v1.ReadWriteMany},
			},
		},
		Parameters: map[string]string{
			parameterChunkSizeMB:            strconv.Itoa(testChunkSizeMB),
			parameterParallelCount:          strconv.Itoa(testParallelCount),
			parameterMultiReqMax:            strconv.Itoa(testMultiReqMax),
			parameterStatCacheSize:          strconv.Itoa(testStatCacheSize),
			parameterS3FSFUSERetryCount:     strconv.Itoa(testS3FSFUSERetryCount),
			parameterStatCacheExpireSeconds: strconv.Itoa(testStatCacheExpireSeconds),
			parameterTLSCipherSuite:         testTLSCipherSuite,
			parameterDebugLevel:             testDebugLevel,
			parameterStorageClass:           testStorageClass,
			parameterOSEndpoint:             testOSEndpoint,
			parameterIAMEndpoint:            testIAMEndpoint,
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
		Spec: v1.PersistentVolumeSpec{
			PersistentVolumeSource: v1.PersistentVolumeSource{
				FlexVolume: &v1.FlexPersistentVolumeSource{
					Options: map[string]string{"object-store-endpoint": testOSEndpoint, "object-store-storage-class": testStorageClass},
				},
			},
		},
	}
}

func Test_Provision_BadPVCAnnotations_AutoCreateBucket(t *testing.T) {
	p := getProvisioner()
	v := getVolumeOptions()
	v.PVC.Annotations[annotationAutoCreateBucket] = "non-true-value"

	_, err := p.Provision(v)
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "invalid value for auto-create-bucket, expects true/false")
	}
}

func Test_Provision_BadPVCAnnotations_AutoDeleteBucket(t *testing.T) {
	p := getProvisioner()
	v := getVolumeOptions()
	v.PVC.Annotations[annotationAutoDeleteBucket] = "non-true-value"

	_, err := p.Provision(v)
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "invalid value for auto-delete-bucket, expects true/false")
	}
}

func Test_Provision_BadPVCAnnotations_ConfigureFirewall(t *testing.T) {
	p := getProvisioner()
	v := getVolumeOptions()
	v.PVC.Annotations[annotationConfigureFirewall] = "non-true-value"

	_, err := p.Provision(v)
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "invalid value for configure-firewall, expects true/false")
	}
}

func Test_Provision_Empty_SecretName(t *testing.T) {
	p := getProvisioner()
	v := getVolumeOptions()
	v.PVC.Annotations[annotationSecretName] = ""

	_, err := p.Provision(v)
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "secret-name not specified")
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

func Test_Provision_BadPVCOSEndpoint(t *testing.T) {
	p := getProvisioner()
	v := getVolumeOptions()
	v.PVC.Annotations[annotationEndpoint] = "test-object-store-endpoint"

	_, err := p.Provision(v)
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), fmt.Sprintf("Bad value for ibm.io/object-store-endpoint \"%s\": scheme is missing. "+
			"Must be of the form http://<hostname> or https://<hostname>", v.PVC.Annotations[annotationEndpoint]))
	}
}

func Test_Provision_PVCAnnotations_OSEndpoint_Positive(t *testing.T) {
	p := getProvisioner()
	v := getVolumeOptions()
	v.PVC.Annotations[annotationEndpoint] = "https://test-object-store-endpoint-defined-in-pvc"

	pv, err := p.Provision(v)
	assert.NoError(t, err)
	assert.Equal(t, "https://test-object-store-endpoint-defined-in-pvc", pv.Spec.FlexVolume.Options[optionOSEndpoint])
}

func Test_Provision_PVCAnnotations_StorageClass_Positive(t *testing.T) {
	p := getProvisioner()
	v := getVolumeOptions()
	v.PVC.Annotations[annotationRegion] = "test-storage-class-defined-in-pvc"

	pv, err := p.Provision(v)
	assert.NoError(t, err)
	assert.Equal(t, "test-storage-class-defined-in-pvc", pv.Spec.FlexVolume.Options[optionStorageClass])
}

func Test_Provision_BadPVCIAMEndpoint(t *testing.T) {
	p := getProvisioner()
	v := getVolumeOptions()
	v.PVC.Annotations[annotationIAMEndpoint] = "test-iam-endpoint"

	_, err := p.Provision(v)
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), fmt.Sprintf("Bad value for ibm.io/iam-endpoint \"%s\":"+
			" Must be of the form https://<hostname> or http://<hostname>", v.PVC.Annotations[annotationIAMEndpoint]))
	}
}

func Test_Provision_PVCAnnotations_IAMEndpoint_Positive(t *testing.T) {
	p := getProvisioner()
	v := getVolumeOptions()
	v.PVC.Annotations[annotationIAMEndpoint] = "https://test-iam-endpoint-defined-in-pvc"

	pv, err := p.Provision(v)
	assert.NoError(t, err)
	assert.Equal(t, "https://test-iam-endpoint-defined-in-pvc", pv.Spec.FlexVolume.Options[optionIAMEndpoint])
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

func Test_Provision_PVCAnnotations_BadStatCacheExpireSeconds_NonInt(t *testing.T) {
	p := getProvisioner()
	v := getVolumeOptions()
	v.PVC.Annotations[annotationStatCacheExpireSeconds] = "non-int-value"

	_, err := p.Provision(v)
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "Cannot convert value of stat-cache-expire-seconds into integer")
	}
}

func Test_Provision_PVCAnnotations_BadStatCacheExpireSeconds_NegativeInt(t *testing.T) {
	p := getProvisioner()
	v := getVolumeOptions()
	v.PVC.Annotations[annotationStatCacheExpireSeconds] = "-6"

	_, err := p.Provision(v)
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "value of stat-cache-expire-seconds should be >= 0")
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

func Test_Provision_PVCAnnotations_S3FSFUSERetryCount_Negative(t *testing.T) {
	p := getProvisioner()
	v := getVolumeOptions()
	v.PVC.Annotations["ibm.io/s3fs-fuse-retry-count"] = "-1"

	_, err := p.Provision(v)
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "value of s3fs-fuse-retry-count should be >= 1")
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
	v.PVC.Annotations[annotationAutoCreateBucket] = "false"
	v.PVC.Annotations[annotationAutoDeleteBucket] = "true"

	_, err := p.Provision(v)
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "bucket auto-create must be enabled when bucket auto-delete is enabled")
	}
}

func Test_Provision_SetDefault_AutoCreateBucket_AutoDeleteBucket_BucketName(t *testing.T) {
	p := getProvisioner()
	v := getVolumeOptions()

	_, err := p.Provision(v)
	assert.NoError(t, err)
}

func Test_Provision_SetDefault_AutoCreateBucket_AutoDeleteBucket(t *testing.T) {
	p := getProvisioner()
	v := getVolumeOptions()
	v.PVC.Annotations[annotationBucket] = testBucket

	_, err := p.Provision(v)
	assert.NoError(t, err)
}

func Test_Provision_AutoDeleteBucketWithNonEmptyBucket(t *testing.T) {
	p := getProvisioner()
	v := getVolumeOptions()
	v.PVC.Annotations[annotationAutoDeleteBucket] = "true"
	v.PVC.Annotations[annotationAutoCreateBucket] = "true"
	v.PVC.Annotations[annotationBucket] = testBucket

	_, err := p.Provision(v)
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "bucket cannot be set when auto-delete is enabled")
	}
}

func Test_Provision_AutoDeleteBucketWithEmptyBucket(t *testing.T) {
	p := getProvisioner()
	v := getVolumeOptions()
	v.PVC.Annotations[annotationAutoDeleteBucket] = "true"
	v.PVC.Annotations[annotationAutoCreateBucket] = "true"

	_, err := p.Provision(v)
	assert.NoError(t, err)
}

func Test_Provision_MissingBucket(t *testing.T) {
	p := getProvisioner()
	v := getVolumeOptions()
	v.PVC.Annotations[annotationAutoDeleteBucket] = "false"
	v.PVC.Annotations[annotationAutoCreateBucket] = "false"

	_, err := p.Provision(v)
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "bucket name not specified")
	}
}

func Test_Provision_SetBucketWithoutAutoCreateBucketAndWithoutAutoDeleteBucket(t *testing.T) {
	p := getProvisioner()
	v := getVolumeOptions()
	v.PVC.Annotations[annotationAutoDeleteBucket] = "false"
	v.PVC.Annotations[annotationAutoCreateBucket] = "false"
	v.PVC.Annotations[annotationBucket] = testBucket

	_, err := p.Provision(v)
	assert.NoError(t, err)
}

func Test_Provision_ObjectPathWithAutoCreateBucket(t *testing.T) {
	p := getProvisioner()
	v := getVolumeOptions()
	v.PVC.Annotations[annotationObjectPath] = testObjectPath
	v.PVC.Annotations[annotationAutoCreateBucket] = "true"

	_, err := p.Provision(v)
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "object-path cannot be set when auto-create is enabled")
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
		assert.Contains(t, err.Error(), "cannot retrieve secret")
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

func Test_Provision_PVCNamespaceNotAllowedInSecrets(t *testing.T) {
	p := getFakeClientGoProvisioner(&clientGoConfig{withAllowedNamespace: true})
	v := getVolumeOptions()

	_, err := p.Provision(v)
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "PVC creation in "+v.PVC.Namespace+" namespace is not allowed")
	}
}

func Test_Provision_PVCNamespaceAllowedInSecrets(t *testing.T) {
	testNamespace = "test-allowed-namespace1"
	p := getFakeClientGoProvisioner(&clientGoConfig{withAllowedNamespace: true})
	v := getVolumeOptions()

	_, err := p.Provision(v)

	assert.NoError(t, err)
}

func Test_Provision_Set_ConfigureFirewall(t *testing.T) {
	p := getFakeClientGoProvisioner(&clientGoConfig{withResConfApiKey: true, withAllowedIPs: true})
	v := getVolumeOptions()
	v.PVC.Annotations[annotationConfigureFirewall] = "true"

	_, err := p.Provision(v)
	assert.NoError(t, err)
}

func Test_Provision_Set_ConfigureFirewall_EmptyResConfApiKeyInSecret(t *testing.T) {
	p := getFakeClientGoProvisioner(&clientGoConfig{withAllowedIPs: true})
	v := getVolumeOptions()
	v.PVC.Annotations[annotationConfigureFirewall] = "true"

	_, err := p.Provision(v)
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "cannot configure firewall for bucket. res-conf-apikey is empty")
	}
}

func Test_Provision_Set_ConfigureFirewall_EmptyAllowedIPsInSecret(t *testing.T) {
	p := getCustomProvisioner(
		&clientGoConfig{withResConfApiKey: true},
		&fake.ObjectStorageSessionFactory{PassUpdateBucketFirewalRules: true},
		uuid.NewCryptoGenerator(),
	)
	v := getVolumeOptions()
	v.PVC.Annotations[annotationConfigureFirewall] = "true"
	v.PVC.Annotations[annotationAllowedIPs] = "10.69.208.4/16"

	_, err := p.Provision(v)
	assert.NoError(t, err)
}

func Test_Provision_Set_ConfigureFirewall_EmptyAnnotationAllowedIPs(t *testing.T) {
	p := getFakeClientGoProvisioner(&clientGoConfig{withResConfApiKey: true})
	v := getVolumeOptions()
	v.PVC.Annotations[annotationConfigureFirewall] = "true"
	v.PVC.Annotations[annotationAllowedIPs] = ""

	_, err := p.Provision(v)
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "cannot configure firewall for bucket. allowed_ips is empty")
	}
}

func Test_Provision_Set_ConfigureFirewall_FailUpdateFirewallRules(t *testing.T) {
	factory := &fake.ObjectStorageSessionFactory{FailUpdateBucketFirewallRules: true}
	p := getFakeBackendProvisioner(factory)
	v := getVolumeOptions()
	v.PVC.Annotations[annotationConfigureFirewall] = "true"

	_, err := p.Provision(v)
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "cannot configure firewall for bucket")
	}
}

func Test_Provision_Set_ConfigureFirewall_FailUpdateFirewallRules_FailDeleteBucket(t *testing.T) {
	factory := &fake.ObjectStorageSessionFactory{FailUpdateBucketFirewallRules: true, FailDeleteBucket: true}
	p := getFakeBackendProvisioner(factory)
	v := getVolumeOptions()
	v.PVC.Annotations[annotationConfigureFirewall] = "true"

	_, err := p.Provision(v)
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "cannot configure firewall for bucket")
		assert.Contains(t, err.Error(), "and cannot delete bucket")
	}
}

func Test_Provision_Set_ConfigureFirewall_ExistingBucket(t *testing.T) {
	p := getCustomProvisioner(
		&clientGoConfig{withResConfApiKey: true, withAllowedIPs: true},
		&fake.ObjectStorageSessionFactory{PassUpdateBucketFirewallRules: true},
		uuid.NewCryptoGenerator(),
	)
	v := getVolumeOptions()
	v.PVC.Annotations[annotationConfigureFirewall] = "true"
	v.PVC.Annotations[annotationAutoDeleteBucket] = "false"
	v.PVC.Annotations[annotationAutoCreateBucket] = "false"
	v.PVC.Annotations[annotationBucket] = testBucket

	_, err := p.Provision(v)
	assert.NoError(t, err)
}

func Test_Provision_CreateBucket_BucketAlreadyOwnedByYou_Positive(t *testing.T) {
	p := getFakeBackendProvisioner(&fake.ObjectStorageSessionFactory{FailCreateBucket: true, FailCreateBucketErrMsg: "BucketAlreadyExists"})
	v := getVolumeOptions()
	v.PVC.Annotations[annotationAutoCreateBucket] = "true"

	_, err := p.Provision(v)
	assert.NoError(t, err)
}

func Test_Provision_FailCreateBucket_BucketOwnedByOther(t *testing.T) {
	p := getFakeBackendProvisioner(&fake.ObjectStorageSessionFactory{FailCreateBucket: true, FailCreateBucketErrMsg: "BucketAlreadyExists", FailCheckBucketAccess: true})
	v := getVolumeOptions()
	v.PVC.Annotations[annotationAutoCreateBucket] = "true"

	_, err := p.Provision(v)
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "cannot create bucket")
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

func Test_Provision_PVCAnnotations_ObjectPath_Positive(t *testing.T) {
	factory := &fake.ObjectStorageSessionFactory{}
	p := getFakeBackendProvisioner(factory)
	v := getVolumeOptions()
	v.PVC.Annotations[annotationAutoCreateBucket] = "false"
	v.PVC.Annotations[annotationObjectPath] = testObjectPath
	v.PVC.Annotations[annotationBucket] = testBucket

	pv, err := p.Provision(v)
	assert.NoError(t, err)
	assert.Equal(t, testObjectPath, pv.Spec.FlexVolume.Options[optionObjectPath])
}

func Test_Provision_CheckObjectPathExistence_Error(t *testing.T) {
	p := getFakeBackendProvisioner(&fake.ObjectStorageSessionFactory{CheckObjectPathExistenceError: true})
	v := getVolumeOptions()
	v.PVC.Annotations[annotationAutoCreateBucket] = "false"
	v.PVC.Annotations[annotationObjectPath] = testObjectPath
	v.PVC.Annotations[annotationBucket] = testBucket

	_, err := p.Provision(v)
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), fmt.Sprintf("cannot access object-path \"%s\" inside bucket %s",
			v.PVC.Annotations[annotationObjectPath], v.PVC.Annotations[annotationBucket]))
	}
}

func Test_Provision_CheckObjectPathExistence_PathNotFound(t *testing.T) {
	p := getFakeBackendProvisioner(&fake.ObjectStorageSessionFactory{CheckObjectPathExistencePathNotFound: true})
	v := getVolumeOptions()
	v.PVC.Annotations[annotationAutoCreateBucket] = "false"
	v.PVC.Annotations[annotationObjectPath] = testObjectPath
	v.PVC.Annotations[annotationBucket] = testBucket

	_, err := p.Provision(v)
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), fmt.Sprintf("object-path \"%s\" not found inside bucket %s",
			v.PVC.Annotations[annotationObjectPath], v.PVC.Annotations[annotationBucket]))
	}
}

func Test_Provision_Positive(t *testing.T) {
	factory := &fake.ObjectStorageSessionFactory{}
	p := getFakeBackendProvisioner(factory)
	v := getVolumeOptions()
	v.PVC.Annotations[annotationAutoCreateBucket] = "false"
	v.PVC.Annotations[annotationBucket] = testBucket

	pv, err := p.Provision(v)
	assert.NoError(t, err)
	assert.Equal(t,
		map[string]string{
			optionChunkSizeMB:            strconv.Itoa(testChunkSizeMB),
			optionParallelCount:          strconv.Itoa(testParallelCount),
			optionMultiReqMax:            strconv.Itoa(testMultiReqMax),
			optionStatCacheSize:          strconv.Itoa(testStatCacheSize),
			optionS3FSFUSERetryCount:     strconv.Itoa(testS3FSFUSERetryCount),
			optionStatCacheExpireSeconds: strconv.Itoa(testStatCacheExpireSeconds),
			optionTLSCipherSuite:         testTLSCipherSuite,
			optionDebugLevel:             testDebugLevel,
			optionCurlDebug:              testCurlDebug,
			optionOSEndpoint:             testOSEndpoint,
			optionBucket:                 testBucket,
			optionStorageClass:           testStorageClass,
			optionIAMEndpoint:            testIAMEndpoint,
			optionAccessMode:             "ReadWriteMany",
		},
		pv.Spec.FlexVolume.Options,
	)
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

func Test_Provision_AccessMode_Negative(t *testing.T) {
	p := getProvisioner()
	v := getVolumeOptions()
	v.PVC.Spec.AccessModes = []v1.PersistentVolumeAccessMode{v1.ReadOnlyMany, v1.ReadWriteOnce}

	_, err := p.Provision(v)
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "More that one access mode is not supported")
	}
}

func Test_Provision_AccessMode_ReadWrite_Positive(t *testing.T) {
	p := getProvisioner()
	v := getVolumeOptions()

	pv, err := p.Provision(v)
	assert.NoError(t, err)
	assert.Equal(t, "ReadWriteMany", pv.Spec.FlexVolume.Options[optionAccessMode])
}

func Test_Provision_AccessMode_ReadOnly_Positive(t *testing.T) {
	p := getProvisioner()
	v := getVolumeOptions()
	v.PVC.Spec.AccessModes = []v1.PersistentVolumeAccessMode{v1.ReadOnlyMany}

	pv, err := p.Provision(v)
	assert.NoError(t, err)
	assert.Equal(t, "ReadOnlyMany", pv.Spec.FlexVolume.Options[optionAccessMode])
}

func Test_Provision_AutoBucketCreate_Positive(t *testing.T) {
	factory := &fake.ObjectStorageSessionFactory{}
	p := getFakeBackendProvisioner(factory)
	v := getVolumeOptions()
	v.PVC.Annotations[annotationAutoCreateBucket] = "true"
	v.PVC.Annotations[annotationBucket] = testBucket

	_, err := p.Provision(v)
	assert.NoError(t, err)

	assert.Equal(t, testOSEndpoint, factory.LastEndpoint)
	assert.Equal(t, testStorageClass, factory.LastRegion)
	assert.Equal(t, testAccessKey, factory.LastCredentials.AccessKey)
	assert.Equal(t, testSecretKey, factory.LastCredentials.SecretKey)
	assert.Equal(t, "", factory.LastCredentials.APIKey)
	assert.Equal(t, testIAMEndpoint, factory.LastCredentials.IAMEndpoint)
	assert.Equal(t, testBucket, factory.LastCreatedBucket)
	assert.Equal(t, testBucket, factory.LastCheckedBucket)
	assert.Equal(t, "", factory.LastDeletedBucket)
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
	assert.Equal(t, testIAMEndpoint, factory.LastCredentials.IAMEndpoint)
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
	pv.Annotations[annotationAutoDeleteBucket] = "non-false-value"
	err := p.Delete(pv)
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "invalid value for auto-delete-bucket, expects true/false")
	}
}

func Test_Delete_MissingSecret(t *testing.T) {
	p := getFakeClientGoProvisioner(&clientGoConfig{missingSecret: true})
	pv := getAutoDeletePersistentVolume()
	err := p.Delete(pv)
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "cannot retrieve secret")
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

	assert.Equal(t, testOSEndpoint, factory.LastEndpoint)
	assert.Equal(t, testStorageClass, factory.LastRegion)
	assert.Equal(t, testAccessKey, factory.LastCredentials.AccessKey)
	assert.Equal(t, testSecretKey, factory.LastCredentials.SecretKey)
	assert.Equal(t, "", factory.LastCredentials.APIKey)
	assert.Equal(t, testIAMEndpoint, factory.LastCredentials.IAMEndpoint)
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
	assert.Equal(t, testIAMEndpoint, factory.LastCredentials.IAMEndpoint)
	assert.Equal(t, bucketName, factory.LastDeletedBucket)
}

func Test_Provision_DifferentSecretNS(t *testing.T) {
	p := getProvisioner()
	v := getVolumeOptions()
	v.PVC.Namespace = "pvc-namespace"
	v.PVC.Annotations[annotationSecretNamespace] = testNamespace
	pv, err := p.Provision(v)
	assert.NoError(t, err)
	assert.Equal(t, testNamespace, pv.Annotations[annotationSecretNamespace])
}

func Test_Validate_Bucket_True(t *testing.T) {
	p := getFakeBackendProvisioner(&fake.ObjectStorageSessionFactory{FailCheckBucketAccess: true})
	v := getVolumeOptions()
	v.PVC.Annotations[annotationValidateBucket] = testValidateBucket
	_, err := p.Provision(v)
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "cannot access bucket")
	}
}

func Test_Wrong_Secret_Type_True(t *testing.T) {
	p := getFakeClientGoProvisioner(&clientGoConfig{wrongSecretType: true})
	v := getVolumeOptions()
	_, err := p.Provision(v)
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "Wrong Secret Type")
	}
}
func Test_Provision_PVCAnnotations_ReadwriteTimeoutSeconds_NonInt(t *testing.T) {
	p := getProvisioner()
	v := getVolumeOptions()
	v.PVC.Annotations[annotationReadwriteTimeoutSeconds] = "non-int-value"

	_, err := p.Provision(v)
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "Cannot convert value of readwrite-timeout-seconds into integer")
	}
}

func Test_Provision_PVCAnnotations_ConnectTimeoutSeconds_NonInt(t *testing.T) {
	p := getProvisioner()
	v := getVolumeOptions()
	v.PVC.Annotations[annotationConnectTimeoutSeconds] = "non-int-value"

	_, err := p.Provision(v)
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "Cannot convert value of connect-timeout-seconds into integer")
	}
}

func Test_Provision_PVCAnnotations_ReadwriteTimeoutSeconds_Positive(t *testing.T) {
	p := getProvisioner()
	v := getVolumeOptions()
	v.PVC.Annotations[annotationReadwriteTimeoutSeconds] = "6"

	pv, err := p.Provision(v)
	assert.NoError(t, err)
	assert.Equal(t, "6", pv.Spec.FlexVolume.Options[optionReadwriteTimeoutSeconds])
}

func Test_Provision_PVCAnnotations_ConnectTimeoutSeconds_Positive(t *testing.T) {
	p := getProvisioner()
	v := getVolumeOptions()
	v.PVC.Annotations[annotationConnectTimeoutSeconds] = "6"

	pv, err := p.Provision(v)
	assert.NoError(t, err)
	assert.Equal(t, "6", pv.Spec.FlexVolume.Options[optionConnectTimeoutSeconds])
}
func Test_Provision_PVCAnnotations_UseXattr(t *testing.T) {
	p := getProvisioner()
	v := getVolumeOptions()
	v.PVC.Annotations["ibm.io/use-xattr"] = "true"
	pv, err := p.Provision(v)
	assert.NoError(t, err)
	assert.Equal(t, "true", pv.Spec.FlexVolume.Options[optionUseXattr])
}

func Test_Provision_PVCAnnotations_DebugLevel(t *testing.T) {
	p := getProvisioner()
	v := getVolumeOptions()
	v.PVC.Annotations["ibm.io/debug-level"] = "info"
	pv, err := p.Provision(v)
	assert.NoError(t, err)
	assert.Equal(t, "info", pv.Spec.FlexVolume.Options[optionDebugLevel])
}

func Test_Provision_PVCAnnotations_CurlDebug(t *testing.T) {
	p := getProvisioner()
	v := getVolumeOptions()
	v.PVC.Annotations["ibm.io/curl-debug"] = "true"
	pv, err := p.Provision(v)
	assert.NoError(t, err)
	assert.Equal(t, "true", pv.Spec.FlexVolume.Options[optionCurlDebug])
}

func Test_Provision_PVCAnnotations_TLS(t *testing.T) {
	p := getProvisioner()
	v := getVolumeOptions()
	v.PVC.Annotations["ibm.io/tls-cipher-suite"] = "AESGCM"
	pv, err := p.Provision(v)
	assert.NoError(t, err)
	assert.Equal(t, "AESGCM", pv.Spec.FlexVolume.Options[optionTLSCipherSuite])
}

func Test_Provision_CASNegative(t *testing.T) {
	p := getProvisioner()
	v := getVolumeOptions()
	v.PVC.Annotations[annotationServiceName] = testServiceName
	v.PVC.Annotations[annotationServiceNamespace] = testServiceNamespace

	_, err := p.Provision(v)
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "cannot retrieve service details")
	}
}

func Test_Provision_CACrtSrvcWithDefaultCACert(t *testing.T) {
	p := getFakeClientGoProvisioner(&clientGoConfig{isTLS: true})
	v := getVolumeOptions()
	v.PVC.Annotations[annotationServiceName] = testServiceName
	v.PVC.Annotations[annotationServiceNamespace] = testServiceNamespace

	_, err := p.Provision(v)
	assert.NoError(t, err)
}

func Test_Provision_CACrtSecretPositive(t *testing.T) {
	p := getFakeClientGoProvisioner(&clientGoConfig{isTLS: true, withcaBundle: true})
	v := getVolumeOptions()
	v.PVC.Annotations[annotationServiceName] = testServiceName
	v.PVC.Annotations[annotationServiceNamespace] = testServiceNamespace

	_, err := p.Provision(v)
	assert.NoError(t, err)
}

func Test_Provision_CACrtSrvcNamespaceOptional(t *testing.T) {
	p := getFakeClientGoProvisioner(&clientGoConfig{isTLS: true, withcaBundle: true})
	v := getVolumeOptions()
	v.PVC.Annotations[annotationServiceName] = testServiceName

	_, err := p.Provision(v)
	assert.NoError(t, err)
}

func Test_Provision_CACrtSecretWriteError(t *testing.T) {
	p := getFakeClientGoProvisioner(&clientGoConfig{isTLS: true, withcaBundle: true})
	v := getVolumeOptions()
	v.PVC.Annotations[annotationServiceName] = testServiceName
	v.PVC.Annotations[annotationServiceNamespace] = testServiceNamespace
	writeFile = writeFileError
	_, err := p.Provision(v)
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "cannot retrieve secret")
	}
}

func Test_Delete_TLS_Negative(t *testing.T) {
	p := getFakeClientGoProvisioner(&clientGoConfig{isTLS: true, withcaBundle: true})
	pv := getAutoDeletePersistentVolume()
	pv.Annotations[annotationServiceName] = testServiceName
	pv.Annotations[annotationServiceNamespace] = testServiceNamespace
	err := p.Delete(pv)
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "cannot delete bucket: cannot retrieve secret")
	}
}

func Test_Delete_TLS_Positive(t *testing.T) {
	p := getFakeClientGoProvisioner(&clientGoConfig{isTLS: true, withcaBundle: true})
	pv := getAutoDeletePersistentVolume()
	pv.Annotations[annotationServiceName] = testServiceName
	pv.Annotations[annotationServiceNamespace] = testServiceNamespace
	writeFile = writeFileSuccess
	err := p.Delete(pv)
	assert.NoError(t, err)
}

func Test_Provision_AutoCache_Positive(t *testing.T) {
	p := getProvisioner()
	v := getVolumeOptions()
	v.PVC.Annotations[parameterAutoCache] = "true"

	pv, err := p.Provision(v)
	assert.NoError(t, err)
	assert.Equal(t, "true", pv.Spec.FlexVolume.Options[optionAutoCache])
}
