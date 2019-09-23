/*******************************************************************************
 * IBM Confidential
 * OCO Source Materials
 * IBM Cloud Container Service, 5737-D43
 * (C) Copyright IBM Corp. 2017, 2018 All Rights Reserved.
 * The source code for this program is not  published or otherwise divested of
 * its trade secrets, irrespective of what has been deposited with
 * the U.S. Copyright Office.
 ******************************************************************************/

package driver

import (
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/IBM/ibmcloud-object-storage-plugin/driver/interfaces"
	"github.com/IBM/ibmcloud-object-storage-plugin/utils/backend/fake"
	"github.com/IBM/ibmcloud-object-storage-plugin/utils/parser"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strconv"
	"testing"
)

const (
	optionChunkSizeMB             = "chunk-size-mb"
	optiontlsCipherSuite          = "tls-cipher-suite"
	optionCurlDebug               = "curl-debug"
	optionKernelCache             = "kernel-cache"
	optionS3FSFUSERetryCount      = "s3fs-fuse-retry-count"
	optionStatCacheExpireSeconds  = "stat-cache-expire-seconds"
	optionObjectPath              = "object-path"
	optionOSEndpoint              = "object-store-endpoint"
	optionOSStorageClass          = "object-store-storage-class"
	optionIAMEndpoint             = "iam-endpoint"
	optionAccessKey               = "kubernetes.io/secret/access-key"
	optionSecretKey               = "kubernetes.io/secret/secret-key"
	optionAPIKey                  = "kubernetes.io/secret/api-key"
	optionConnectTimeoutSeconds   = "connect-timeout"
	optionReadwriteTimeoutSeconds = "readwrite-timeout"
	optionUseXattr                = "use-xattr"
	optionServiceInstanceID       = "kubernetes.io/secret/service-instance-id"
	optionCAbundleB64             = "kubernetes.io/secret/ca-bundle-crt"
	optionServiceIP               = "service-ip"

	testDir            = "/tmp/"
	testChunkSizeMB    = 500
	testParallelCount  = 2
	testMultiReqMax    = 4
	testStatCacheSize  = 5
	testIAMEndpoint    = "https://test-iam-endpoint"
	testOSEndpoint     = "https://test-object-store-endpoint"
	testStorageClass   = "test-object-store-storage-class"
	testBucket         = "test-bucket"
	testObjectPath     = "/test/object-path"
	testAccessKey      = "akey"
	testSecretKey      = "skey"
	testAPIKey         = "apikey"
	testTLSCipherSuite = "test-tls-cipher-suite"
	testDebugLevel     = "debug"
	testCABundle       = "test-ca-bundle"
	testServiceIP      = "1.0.0.0.1"
)

// these are actually constants
var (
	statSuccess     = func(name string) (os.FileInfo, error) { return nil, nil }
	statUnknownErr  = func(name string) (os.FileInfo, error) { return nil, errors.New("") }
	statErrNotExist = func(name string) (os.FileInfo, error) { return nil, os.ErrNotExist }

	mkdirAllSuccess = func(string, os.FileMode) error { return nil }
	mkdirAllError   = func(string, os.FileMode) error { return errors.New("") }

	mountSuccess = func(string, string, string, uintptr, string) error { return nil }
	mountError   = func(string, string, string, uintptr, string) error { return errors.New("") }

	unmountSuccess = func(target string, flags int) error { return nil }
	unmountError   = func(target string, flags int) error { return errors.New("") }

	removeAllSuccess = func(string) error { return nil }
	removeAllError   = func(string) error { return errors.New("") }

	writeFileSuccess = func(string, []byte, os.FileMode) error { return nil }
	writeFileError   = func(string, []byte, os.FileMode) error { return errors.New("") }
)

var commandArgs []string
var commandOutput string
var commandFailure bool

func getPlugin() *S3fsPlugin {
	commandFailure = false
	mkdirAll = mkdirAllSuccess
	mount = mountSuccess
	stat = statSuccess
	removeAll = removeAllSuccess
	unmount = unmountSuccess
	writeFile = writeFileSuccess
	commandArgs = nil
	command = func(cmd string, args ...string) *exec.Cmd {
		commandArgs = args

		cs := []string{"-test.run=TestHelperProcess", "--"}
		cs = append(cs, args...)
		cs = append(cs, commandOutput)

		ret := exec.Command(os.Args[0], cs...)
		ret.Env = []string{"GO_WANT_HELPER_PROCESS=1"}
		if commandFailure {
			ret.Stdout = ioutil.Discard
		}
		return ret
	}
	return &S3fsPlugin{
		Backend: &fake.ObjectStorageSessionFactory{},
		Logger:  zap.NewNop(),
	}
}

func TestHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}
	defer os.Exit(0)
	fmt.Println(os.Args[len(os.Args)-1])
}

func getMountRequest() interfaces.FlexVolumeMountRequest {
	opts := Options{
		ChunkSizeMB:    testChunkSizeMB,
		ParallelCount:  testParallelCount,
		MultiReqMax:    testMultiReqMax,
		StatCacheSize:  testStatCacheSize,
		OSEndpoint:     testOSEndpoint,
		OSStorageClass: testStorageClass,
		IAMEndpoint:    testIAMEndpoint,
		Bucket:         testBucket,
		TLSCipherSuite: testTLSCipherSuite,
		DebugLevel:     testDebugLevel,
		AccessKeyB64:   base64.StdEncoding.EncodeToString([]byte(testAccessKey)),
		SecretKeyB64:   base64.StdEncoding.EncodeToString([]byte(testSecretKey)),
	}
	driverOptions, _ := parser.MarshalToMap(&opts)
	return interfaces.FlexVolumeMountRequest{
		MountDir: testDir,
		Opts:     driverOptions,
	}
}

func getUnmountRequest() interfaces.FlexVolumeUnmountRequest {
	return interfaces.FlexVolumeUnmountRequest{
		MountDir: testDir,
	}
}

func Test_Mount_BadDriverOptions(t *testing.T) {
	p := getPlugin()
	r := getMountRequest()
	r.Opts[optionChunkSizeMB] = "non-int-value"

	resp := p.Mount(r)
	if assert.Equal(t, interfaces.StatusFailure, resp.Status) {
		assert.Contains(t, resp.Message, "cannot unmarshal driver options")
	}
}

func Test_Mount_BadAPIKey(t *testing.T) {
	p := getPlugin()
	r := getMountRequest()
	r.Opts[optionAPIKey] = "illegal-base-64"

	resp := p.Mount(r)
	if assert.Equal(t, interfaces.StatusFailure, resp.Status) {
		assert.Contains(t, resp.Message, "cannot decode API key")
	}
}

func Test_Mount_BadAccessKey(t *testing.T) {
	p := getPlugin()
	r := getMountRequest()
	r.Opts[optionAccessKey] = "illegal-base-64"

	resp := p.Mount(r)
	if assert.Equal(t, interfaces.StatusFailure, resp.Status) {
		assert.Contains(t, resp.Message, "cannot decode access key")
	}
}

func Test_Mount_BadSecretKey(t *testing.T) {
	p := getPlugin()
	r := getMountRequest()
	r.Opts[optionSecretKey] = "illegal-base-64"

	resp := p.Mount(r)
	if assert.Equal(t, interfaces.StatusFailure, resp.Status) {
		assert.Contains(t, resp.Message, "cannot decode secret key")
	}
}

func Test_Mount_BadOSEndpoint(t *testing.T) {
	p := getPlugin()
	r := getMountRequest()
	r.Opts[optionOSEndpoint] = "test-object-store-endpoint"

	resp := p.Mount(r)
	if assert.Equal(t, interfaces.StatusFailure, resp.Status) {
		assert.Contains(t, resp.Message, fmt.Sprintf("Bad value for object-store-endpoint \"%s\": scheme is missing. "+
			"Must be of the form http://<hostname> or https://<hostname>", r.Opts[optionOSEndpoint]))
	}
}

func Test_Mount_BadIAMEndpoint(t *testing.T) {
	p := getPlugin()
	r := getMountRequest()
	r.Opts[optionAPIKey] = base64.StdEncoding.EncodeToString([]byte(testAPIKey))
	r.Opts[optionIAMEndpoint] = "test-iam-endpoint"

	resp := p.Mount(r)
	if assert.Equal(t, interfaces.StatusFailure, resp.Status) {
		assert.Contains(t, resp.Message, fmt.Sprintf("Bad value for iam-endpoint \"%s\":"+
			" Must be of the form https://<hostname> or http://<hostname>", r.Opts[optionIAMEndpoint]))
	}
}

func Test_Mount_BadS3FSFUSERetryCount(t *testing.T) {
	p := getPlugin()
	r := getMountRequest()
	r.Opts[optionS3FSFUSERetryCount] = "non-int-value"

	resp := p.Mount(r)
	if assert.Equal(t, interfaces.StatusFailure, resp.Status) {
		assert.Contains(t, resp.Message, "Cannot convert value of s3fs-fuse-retry-count into integer")
	}
}

func Test_Mount_S3FSFUSERetryCount_Negative(t *testing.T) {
	p := getPlugin()
	r := getMountRequest()
	r.Opts[optionS3FSFUSERetryCount] = "-1"

	resp := p.Mount(r)
	if assert.Equal(t, interfaces.StatusFailure, resp.Status) {
		assert.Contains(t, resp.Message, "value of s3fs-fuse-retry-count should be >= 1")
	}
}

func Test_Mount_BadStatCacheExpireSeconds_NonInt(t *testing.T) {
	p := getPlugin()
	r := getMountRequest()
	r.Opts[optionStatCacheExpireSeconds] = "non-int-value"

	resp := p.Mount(r)
	if assert.Equal(t, interfaces.StatusFailure, resp.Status) {
		assert.Contains(t, resp.Message, "Cannot convert value of stat-cache-expire-seconds into integer")
	}
}

func Test_Mount_BadStatCacheExpireSeconds_NegativeInt(t *testing.T) {
	p := getPlugin()
	r := getMountRequest()
	r.Opts[optionStatCacheExpireSeconds] = "-10"

	resp := p.Mount(r)
	if assert.Equal(t, interfaces.StatusFailure, resp.Status) {
		assert.Contains(t, resp.Message, "value of stat-cache-expire-seconds should be >= 0")
	}
}

func Test_Mount_FailBucketAccess(t *testing.T) {
	p := &S3fsPlugin{
		Backend: &fake.ObjectStorageSessionFactory{FailCheckBucketAccess: true},
		Logger:  zap.NewNop(),
	}
	r := getMountRequest()

	resp := p.Mount(r)
	if assert.Equal(t, interfaces.StatusFailure, resp.Status) {
		assert.Contains(t, resp.Message, "cannot access bucket")
	}
}

func Test_Mount_CheckObjectPath_Error(t *testing.T) {
	p := &S3fsPlugin{
		Backend: &fake.ObjectStorageSessionFactory{CheckObjectPathExistenceError: true},
		Logger:  zap.NewNop(),
	}
	r := getMountRequest()
	r.Opts[optionObjectPath] = testObjectPath

	resp := p.Mount(r)
	if assert.Equal(t, interfaces.StatusFailure, resp.Status) {
		assert.Contains(t, resp.Message, fmt.Sprintf("cannot access object-path \"%s\" inside bucket %s",
			r.Opts[optionObjectPath], r.Opts["bucket"]))
	}
}

func Test_Mount_CheckObjectPath_PathNotFound(t *testing.T) {
	p := &S3fsPlugin{
		Backend: &fake.ObjectStorageSessionFactory{CheckObjectPathExistencePathNotFound: true},
		Logger:  zap.NewNop(),
	}
	r := getMountRequest()
	r.Opts[optionObjectPath] = testObjectPath

	resp := p.Mount(r)
	if assert.Equal(t, interfaces.StatusFailure, resp.Status) {
		assert.Contains(t, resp.Message, fmt.Sprintf("object-path \"%s\" not found inside bucket %s",
			r.Opts[optionObjectPath], r.Opts["bucket"]))
	}
}

func Test_Mount_CannotStatTargetDir(t *testing.T) {
	p := getPlugin()
	r := getMountRequest()
	stat = statUnknownErr

	resp := p.Mount(r)
	if assert.Equal(t, interfaces.StatusFailure, resp.Status) {
		assert.Contains(t, resp.Message, "cannot stat directory")
	}
}

func Test_Mount_CannotCreateTargetDir(t *testing.T) {
	p := getPlugin()
	r := getMountRequest()
	stat = statErrNotExist
	mkdirAll = mkdirAllError

	resp := p.Mount(r)
	if assert.Equal(t, interfaces.StatusFailure, resp.Status) {
		assert.Contains(t, resp.Message, "cannot create directory")
	}
}

func Test_isMountpoint_Error(t *testing.T) {
	p := getPlugin()
	commandFailure = true

	_, err := p.isMountpoint("")
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "Stdout already set")
	}
}

func Test_isMountpoint_IsMountpoint_Positive(t *testing.T) {
	p := getPlugin()
	commandOutput = "... is a mountpoint"

	ret, err := p.isMountpoint("")
	if assert.NoError(t, err) {
		assert.True(t, ret)
	}
}

func Test_isMountpoint_IsNotMountpoint_Positive(t *testing.T) {
	p := getPlugin()
	commandOutput = "... is not a mountpoint"

	ret, err := p.isMountpoint("")
	if assert.NoError(t, err) {
		assert.False(t, ret)
	}
}

func Test_isMountpoint_UnknownOutput(t *testing.T) {
	p := getPlugin()
	commandOutput = "unknown output"

	_, err := p.isMountpoint("")
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "cannot parse mountpoint result")
	}
}

func Test_unmountPath_StatError(t *testing.T) {
	p := getPlugin()
	stat = statUnknownErr
	err := p.unmountPath("", true)
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "cannot stat directory")
	}
}

func Test_unmountPath_PathDoesNotExist_Positive(t *testing.T) {
	p := getPlugin()
	stat = statErrNotExist
	err := p.unmountPath("/s3fs/checkme", true)
	assert.NoError(t, err)
}

func Test_unmountPath_UnmountError(t *testing.T) {
	p := getPlugin()
	commandOutput = "... is a mountpoint"
	unmount = unmountError
	err := p.unmountPath("", true)
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "cannot stat directory")
	}
}

func Test_unmountPath_DeleteError(t *testing.T) {
	p := getPlugin()
	commandOutput = "... is a mountpoint"
	removeAll = removeAllError
	err := p.unmountPath("", true)
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "cannot stat directory")
	}
}

func Test_unmountPath_Positive(t *testing.T) {
	p := getPlugin()
	commandOutput = "... is a mountpoint"
	err := p.unmountPath("/s3fs/checkme", true)
	assert.NoError(t, err)
}

//This test need root access for execution
//func Test_Mount_UnmountDataPathError(t *testing.T) {
//	p := getPlugin()
//	r := getMountRequest()
//	unmount = unmountError
//
//	resp := p.Mount(r)
//	if assert.Equal(t, interfaces.StatusFailure, resp.Status) {
//		assert.Contains(t, resp.Message, "cannot force unmount")
//	}
//}

func Test_Mount_CreateDataPathError(t *testing.T) {
	p := getPlugin()
	r := getMountRequest()
	mkdirAll = mkdirAllError

	resp := p.Mount(r)
	if assert.Equal(t, interfaces.StatusFailure, resp.Status) {
		assert.Contains(t, resp.Message, "cannot create directory")
	}
}

func Test_Mount_MountDataPathError(t *testing.T) {
	p := getPlugin()
	r := getMountRequest()
	mount = mountError

	resp := p.Mount(r)
	if assert.Equal(t, interfaces.StatusFailure, resp.Status) {
		assert.Contains(t, resp.Message, "cannot create tmpfs mountpoint")
	}
}

func Test_Mount_CreatePasswordFileError(t *testing.T) {
	p := getPlugin()
	r := getMountRequest()
	writeFile = writeFileError

	resp := p.Mount(r)
	if assert.Equal(t, interfaces.StatusFailure, resp.Status) {
		assert.Contains(t, resp.Message, "cannot create password file")
	}
}

func Test_Mount_DefaultTLSCipher_Positive(t *testing.T) {
	p := getPlugin()
	r := getMountRequest()
	delete(r.Opts, optiontlsCipherSuite)

	expectedArgs := []string{
		testBucket,
		testDir,
		"-o", "multireq_max=" + strconv.Itoa(testMultiReqMax),
		"-o", "cipher_suites=" + defaultTLSCipherSuite,
		"-o", "use_path_request_style",
		"-o", "passwd_file=" + path.Join(dataRootPath, fmt.Sprintf("%x", sha256.Sum256([]byte(testDir))), passwordFileName),
		"-o", "url=" + testOSEndpoint,
		"-o", "endpoint=" + testStorageClass,
		"-o", "parallel_count=" + strconv.Itoa(testParallelCount),
		"-o", "multipart_size=" + strconv.Itoa(testChunkSizeMB),
		"-o", "dbglevel=" + testDebugLevel,
		"-o", "max_stat_cache_size=" + strconv.Itoa(testStatCacheSize),
		"-o", "allow_other",
		"-o", "max_background=1000",
		"-o", "mp_umask=002",
		"-o", "instance_name=" + testDir,
		"-o", "default_acl=private",
	}

	resp := p.Mount(r)
	if assert.Equal(t, interfaces.StatusSuccess, resp.Status) {
		assert.Equal(t, expectedArgs, commandArgs)
	}
}

func Test_KernelCache_Positive(t *testing.T) {
	p := getPlugin()
	r := getMountRequest()
	r.Opts[optionKernelCache] = "true"

	expectedArgs := []string{
		testBucket,
		testDir,
		"-o", "multireq_max=" + strconv.Itoa(testMultiReqMax),
		"-o", "cipher_suites=" + testTLSCipherSuite,
		"-o", "use_path_request_style",
		"-o", "passwd_file=" + path.Join(dataRootPath, fmt.Sprintf("%x", sha256.Sum256([]byte(testDir))), passwordFileName),
		"-o", "url=" + testOSEndpoint,
		"-o", "endpoint=" + testStorageClass,
		"-o", "parallel_count=" + strconv.Itoa(testParallelCount),
		"-o", "multipart_size=" + strconv.Itoa(testChunkSizeMB),
		"-o", "dbglevel=" + testDebugLevel,
		"-o", "max_stat_cache_size=" + strconv.Itoa(testStatCacheSize),
		"-o", "allow_other",
		"-o", "max_background=1000",
		"-o", "mp_umask=002",
		"-o", "instance_name=" + testDir,
		"-o", "kernel_cache",
		"-o", "default_acl=private",
	}

	resp := p.Mount(r)
	if assert.Equal(t, interfaces.StatusSuccess, resp.Status) {
		assert.Equal(t, expectedArgs, commandArgs)
	}
}

func Test_CurlDebug_Positive(t *testing.T) {
	p := getPlugin()
	r := getMountRequest()
	r.Opts[optionCurlDebug] = "true"

	expectedArgs := []string{
		testBucket,
		testDir,
		"-o", "multireq_max=" + strconv.Itoa(testMultiReqMax),
		"-o", "cipher_suites=" + testTLSCipherSuite,
		"-o", "use_path_request_style",
		"-o", "passwd_file=" + path.Join(dataRootPath, fmt.Sprintf("%x", sha256.Sum256([]byte(testDir))), passwordFileName),
		"-o", "url=" + testOSEndpoint,
		"-o", "endpoint=" + testStorageClass,
		"-o", "parallel_count=" + strconv.Itoa(testParallelCount),
		"-o", "multipart_size=" + strconv.Itoa(testChunkSizeMB),
		"-o", "dbglevel=" + testDebugLevel,
		"-o", "max_stat_cache_size=" + strconv.Itoa(testStatCacheSize),
		"-o", "allow_other",
		"-o", "max_background=1000",
		"-o", "mp_umask=002",
		"-o", "instance_name=" + testDir,
		"-o", "curldbg",
		"-o", "default_acl=private",
	}

	resp := p.Mount(r)
	if assert.Equal(t, interfaces.StatusSuccess, resp.Status) {
		assert.Equal(t, expectedArgs, commandArgs)
	}
}

func Test_S3FSFUSERetryCount_Positive(t *testing.T) {
	p := getPlugin()
	r := getMountRequest()
	r.Opts[optionS3FSFUSERetryCount] = "1"

	expectedArgs := []string{
		testBucket,
		testDir,
		"-o", "multireq_max=" + strconv.Itoa(testMultiReqMax),
		"-o", "cipher_suites=" + testTLSCipherSuite,
		"-o", "use_path_request_style",
		"-o", "passwd_file=" + path.Join(dataRootPath, fmt.Sprintf("%x", sha256.Sum256([]byte(testDir))), passwordFileName),
		"-o", "url=" + testOSEndpoint,
		"-o", "endpoint=" + testStorageClass,
		"-o", "parallel_count=" + strconv.Itoa(testParallelCount),
		"-o", "multipart_size=" + strconv.Itoa(testChunkSizeMB),
		"-o", "dbglevel=" + testDebugLevel,
		"-o", "max_stat_cache_size=" + strconv.Itoa(testStatCacheSize),
		"-o", "allow_other",
		"-o", "max_background=1000",
		"-o", "mp_umask=002",
		"-o", "instance_name=" + testDir,
		"-o", "retries=1",
		"-o", "default_acl=private",
	}

	resp := p.Mount(r)
	if assert.Equal(t, interfaces.StatusSuccess, resp.Status) {
		assert.Equal(t, expectedArgs, commandArgs)
	}
}

func Test_StatCacheExpireSeconds_Positive(t *testing.T) {
	p := getPlugin()
	r := getMountRequest()
	r.Opts[optionStatCacheExpireSeconds] = "1"

	expectedArgs := []string{
		testBucket,
		testDir,
		"-o", "multireq_max=" + strconv.Itoa(testMultiReqMax),
		"-o", "cipher_suites=" + testTLSCipherSuite,
		"-o", "use_path_request_style",
		"-o", "passwd_file=" + path.Join(dataRootPath, fmt.Sprintf("%x", sha256.Sum256([]byte(testDir))), passwordFileName),
		"-o", "url=" + testOSEndpoint,
		"-o", "endpoint=" + testStorageClass,
		"-o", "parallel_count=" + strconv.Itoa(testParallelCount),
		"-o", "multipart_size=" + strconv.Itoa(testChunkSizeMB),
		"-o", "dbglevel=" + testDebugLevel,
		"-o", "max_stat_cache_size=" + strconv.Itoa(testStatCacheSize),
		"-o", "allow_other",
		"-o", "max_background=1000",
		"-o", "mp_umask=002",
		"-o", "instance_name=" + testDir,
		"-o", "stat_cache_expire=1",
		"-o", "default_acl=private",
	}

	resp := p.Mount(r)
	if assert.Equal(t, interfaces.StatusSuccess, resp.Status) {
		assert.Equal(t, expectedArgs, commandArgs)
	}
}

func Test_ObjectPath_Positive_PathWithPrefixForwardSlash(t *testing.T) {
	p := getPlugin()
	r := getMountRequest()
	r.Opts[optionObjectPath] = testObjectPath

	expectedArgs := []string{
		testBucket + ":" + testObjectPath,
		testDir,
		"-o", "multireq_max=" + strconv.Itoa(testMultiReqMax),
		"-o", "cipher_suites=" + testTLSCipherSuite,
		"-o", "use_path_request_style",
		"-o", "passwd_file=" + path.Join(dataRootPath, fmt.Sprintf("%x", sha256.Sum256([]byte(testDir))), passwordFileName),
		"-o", "url=" + testOSEndpoint,
		"-o", "endpoint=" + testStorageClass,
		"-o", "parallel_count=" + strconv.Itoa(testParallelCount),
		"-o", "multipart_size=" + strconv.Itoa(testChunkSizeMB),
		"-o", "dbglevel=" + testDebugLevel,
		"-o", "max_stat_cache_size=" + strconv.Itoa(testStatCacheSize),
		"-o", "allow_other",
		"-o", "max_background=1000",
		"-o", "mp_umask=002",
		"-o", "instance_name=" + testDir,
		"-o", "default_acl=private",
	}

	resp := p.Mount(r)
	if assert.Equal(t, interfaces.StatusSuccess, resp.Status) {
		assert.Equal(t, expectedArgs, commandArgs)
	}
}

func Test_ObjectPath_Positive_PathWithoutPrefixForwardSlash(t *testing.T) {
	p := getPlugin()
	r := getMountRequest()
	r.Opts[optionObjectPath] = "tmp/object-path"

	expectedArgs := []string{
		testBucket + ":/tmp/object-path",
		testDir,
		"-o", "multireq_max=" + strconv.Itoa(testMultiReqMax),
		"-o", "cipher_suites=" + testTLSCipherSuite,
		"-o", "use_path_request_style",
		"-o", "passwd_file=" + path.Join(dataRootPath, fmt.Sprintf("%x", sha256.Sum256([]byte(testDir))), passwordFileName),
		"-o", "url=" + testOSEndpoint,
		"-o", "endpoint=" + testStorageClass,
		"-o", "parallel_count=" + strconv.Itoa(testParallelCount),
		"-o", "multipart_size=" + strconv.Itoa(testChunkSizeMB),
		"-o", "dbglevel=" + testDebugLevel,
		"-o", "max_stat_cache_size=" + strconv.Itoa(testStatCacheSize),
		"-o", "allow_other",
		"-o", "max_background=1000",
		"-o", "mp_umask=002",
		"-o", "instance_name=" + testDir,
		"-o", "default_acl=private",
	}

	resp := p.Mount(r)
	if assert.Equal(t, interfaces.StatusSuccess, resp.Status) {
		assert.Equal(t, expectedArgs, commandArgs)
	}
}

func Test_Mount_fsGroup_Nogroup_Positive(t *testing.T) {
	p := getPlugin()
	r := getMountRequest()
	r.Opts["kubernetes.io/fsGroup"] = "65534"
	expectedArgs := []string{
		testBucket,
		testDir,
		"-o", "multireq_max=" + strconv.Itoa(testMultiReqMax),
		"-o", "cipher_suites=" + testTLSCipherSuite,
		"-o", "use_path_request_style",
		"-o", "passwd_file=" + path.Join(dataRootPath, fmt.Sprintf("%x", sha256.Sum256([]byte(testDir))), passwordFileName),
		"-o", "url=" + testOSEndpoint,
		"-o", "endpoint=" + testStorageClass,
		"-o", "parallel_count=" + strconv.Itoa(testParallelCount),
		"-o", "multipart_size=" + strconv.Itoa(testChunkSizeMB),
		"-o", "dbglevel=" + testDebugLevel,
		"-o", "max_stat_cache_size=" + strconv.Itoa(testStatCacheSize),
		"-o", "allow_other",
		"-o", "max_background=1000",
		"-o", "mp_umask=002",
		"-o", "instance_name=" + testDir,
		"-o", "gid=65534",
		"-o", "uid=65534",
		"-o", "default_acl=private",
	}
	resp := p.Mount(r)
	if assert.Equal(t, interfaces.StatusSuccess, resp.Status) {
		assert.Equal(t, expectedArgs, commandArgs)
	}
}

func Test_Mount_ReadOnly_Positive(t *testing.T) {
	p := getPlugin()
	r := getMountRequest()
	r.Opts["access-mode"] = "ReadOnlyMany"
	expectedArgs := []string{
		testBucket,
		testDir,
		"-o", "multireq_max=" + strconv.Itoa(testMultiReqMax),
		"-o", "cipher_suites=" + testTLSCipherSuite,
		"-o", "use_path_request_style",
		"-o", "passwd_file=" + path.Join(dataRootPath, fmt.Sprintf("%x", sha256.Sum256([]byte(testDir))), passwordFileName),
		"-o", "url=" + testOSEndpoint,
		"-o", "endpoint=" + testStorageClass,
		"-o", "parallel_count=" + strconv.Itoa(testParallelCount),
		"-o", "multipart_size=" + strconv.Itoa(testChunkSizeMB),
		"-o", "dbglevel=" + testDebugLevel,
		"-o", "max_stat_cache_size=" + strconv.Itoa(testStatCacheSize),
		"-o", "allow_other",
		"-o", "max_background=1000",
		"-o", "mp_umask=002",
		"-o", "instance_name=" + testDir,
		"-o", "ro",
		"-o", "default_acl=private",
	}
	resp := p.Mount(r)
	if assert.Equal(t, interfaces.StatusSuccess, resp.Status) {
		assert.Equal(t, expectedArgs, commandArgs)
	}
}

func Test_Mount_DummyOSStorageClass_Positive(t *testing.T) {
	p := getPlugin()
	r := getMountRequest()
	r.Opts[optionOSStorageClass] = ""

	expectedArgs := []string{
		testBucket,
		testDir,
		"-o", "multireq_max=" + strconv.Itoa(testMultiReqMax),
		"-o", "cipher_suites=" + testTLSCipherSuite,
		"-o", "use_path_request_style",
		"-o", "passwd_file=" + path.Join(dataRootPath, fmt.Sprintf("%x", sha256.Sum256([]byte(testDir))), passwordFileName),
		"-o", "url=" + testOSEndpoint,
		"-o", "endpoint=dummy-object-store-storageclass",
		"-o", "parallel_count=" + strconv.Itoa(testParallelCount),
		"-o", "multipart_size=" + strconv.Itoa(testChunkSizeMB),
		"-o", "dbglevel=" + testDebugLevel,
		"-o", "max_stat_cache_size=" + strconv.Itoa(testStatCacheSize),
		"-o", "allow_other",
		"-o", "max_background=1000",
		"-o", "mp_umask=002",
		"-o", "instance_name=" + testDir,
		"-o", "default_acl=private",
	}

	resp := p.Mount(r)
	if assert.Equal(t, interfaces.StatusSuccess, resp.Status) {
		assert.Equal(t, expectedArgs, commandArgs)
	}
}

func Test_Mount_S3fsError(t *testing.T) {
	p := getPlugin()
	r := getMountRequest()
	commandFailure = true

	resp := p.Mount(r)
	if assert.Equal(t, interfaces.StatusFailure, resp.Status) {
		assert.Contains(t, resp.Message, "s3fs mount failed")
	}
}

func Test_Mount_Positive(t *testing.T) {
	p := getPlugin()
	r := getMountRequest()

	expectedArgs := []string{
		testBucket,
		testDir,
		"-o", "multireq_max=" + strconv.Itoa(testMultiReqMax),
		"-o", "cipher_suites=" + testTLSCipherSuite,
		"-o", "use_path_request_style",
		"-o", "passwd_file=" + path.Join(dataRootPath, fmt.Sprintf("%x", sha256.Sum256([]byte(testDir))), passwordFileName),
		"-o", "url=" + testOSEndpoint,
		"-o", "endpoint=" + testStorageClass,
		"-o", "parallel_count=" + strconv.Itoa(testParallelCount),
		"-o", "multipart_size=" + strconv.Itoa(testChunkSizeMB),
		"-o", "dbglevel=" + testDebugLevel,
		"-o", "max_stat_cache_size=" + strconv.Itoa(testStatCacheSize),
		"-o", "allow_other",
		"-o", "max_background=1000",
		"-o", "mp_umask=002",
		"-o", "instance_name=" + testDir,
		"-o", "default_acl=private",
	}

	resp := p.Mount(r)
	if assert.Equal(t, interfaces.StatusSuccess, resp.Status) {
		assert.Equal(t, expectedArgs, commandArgs)
	}
}

func Test_Mount_IAM_Positive(t *testing.T) {
	p := getPlugin()
	r := getMountRequest()
	r.Opts[optionAPIKey] = base64.StdEncoding.EncodeToString([]byte(testAPIKey))

	expectedArgs := []string{
		testBucket,
		testDir,
		"-o", "multireq_max=" + strconv.Itoa(testMultiReqMax),
		"-o", "cipher_suites=" + testTLSCipherSuite,
		"-o", "use_path_request_style",
		"-o", "passwd_file=" + path.Join(dataRootPath, fmt.Sprintf("%x", sha256.Sum256([]byte(testDir))), passwordFileName),
		"-o", "url=" + testOSEndpoint,
		"-o", "endpoint=" + testStorageClass,
		"-o", "parallel_count=" + strconv.Itoa(testParallelCount),
		"-o", "multipart_size=" + strconv.Itoa(testChunkSizeMB),
		"-o", "dbglevel=" + testDebugLevel,
		"-o", "max_stat_cache_size=" + strconv.Itoa(testStatCacheSize),
		"-o", "allow_other",
		"-o", "max_background=1000",
		"-o", "mp_umask=002",
		"-o", "instance_name=" + testDir,
		"-o", "ibm_iam_auth",
		"-o", "ibm_iam_endpoint=" + testIAMEndpoint,
	}

	resp := p.Mount(r)
	if assert.Equal(t, interfaces.StatusSuccess, resp.Status) {
		assert.Equal(t, expectedArgs, commandArgs)
	}
}

func Test_Mount_IAM_Positive_DefaultIAMEndpoint(t *testing.T) {
	p := getPlugin()
	r := getMountRequest()
	r.Opts[optionAPIKey] = base64.StdEncoding.EncodeToString([]byte(testAPIKey))
	r.Opts[optionIAMEndpoint] = ""

	expectedArgs := []string{
		testBucket,
		testDir,
		"-o", "multireq_max=" + strconv.Itoa(testMultiReqMax),
		"-o", "cipher_suites=" + testTLSCipherSuite,
		"-o", "use_path_request_style",
		"-o", "passwd_file=" + path.Join(dataRootPath, fmt.Sprintf("%x", sha256.Sum256([]byte(testDir))), passwordFileName),
		"-o", "url=" + testOSEndpoint,
		"-o", "endpoint=" + testStorageClass,
		"-o", "parallel_count=" + strconv.Itoa(testParallelCount),
		"-o", "multipart_size=" + strconv.Itoa(testChunkSizeMB),
		"-o", "dbglevel=" + testDebugLevel,
		"-o", "max_stat_cache_size=" + strconv.Itoa(testStatCacheSize),
		"-o", "allow_other",
		"-o", "max_background=1000",
		"-o", "mp_umask=002",
		"-o", "instance_name=" + testDir,
		"-o", "ibm_iam_auth",
		"-o", "ibm_iam_endpoint=https://iam.bluemix.net",
	}

	resp := p.Mount(r)
	if assert.Equal(t, interfaces.StatusSuccess, resp.Status) {
		assert.Equal(t, expectedArgs, commandArgs)
	}
}

func Test_Unmount_UnmountS3fsError(t *testing.T) {
	p := getPlugin()
	r := getUnmountRequest()
	unmount = unmountError

	resp := p.Unmount(r)
	if assert.Equal(t, interfaces.StatusFailure, resp.Status) {
		assert.Contains(t, resp.Message, "cannot unmount s3fs mount point")
	}
}

//This test need root access for execution
//func Test_Unmount_DeleteDataDirError(t *testing.T) {
//	p := getPlugin()
//	r := getUnmountRequest()
//	removeAll = removeAllError
//
//	resp := p.Unmount(r)
//	if assert.Equal(t, interfaces.StatusFailure, resp.Status) {
//		assert.Contains(t, resp.Message, "cannot delete data mount point")
//	}
//}

func Test_Unmount_Positive(t *testing.T) {
	p := getPlugin()
	r := getUnmountRequest()

	resp := p.Unmount(r)
	assert.Equal(t, interfaces.StatusSuccess, resp.Status)
}

func Test_Init_Positive(t *testing.T) {
	p := getPlugin()
	resp := p.Init()
	assert.Equal(t, interfaces.StatusSuccess, resp.Status)
}
func Test_ConnectTimeoutSeconds_NonInt(t *testing.T) {
	p := getPlugin()
	r := getMountRequest()
	r.Opts[optionConnectTimeoutSeconds] = "non-int-value"

	resp := p.Mount(r)
	if assert.Equal(t, interfaces.StatusFailure, resp.Status) {
		assert.Contains(t, resp.Message, "Cannot convert value of connect-timeout-seconds into integer")
	}
}

func Test_ConnectTimeoutSeconds_Positive(t *testing.T) {
	p := getPlugin()
	r := getMountRequest()
	r.Opts[optionConnectTimeoutSeconds] = "1"

	expectedArgs := []string{
		testBucket,
		testDir,
		"-o", "multireq_max=" + strconv.Itoa(testMultiReqMax),
		"-o", "cipher_suites=" + testTLSCipherSuite,
		"-o", "use_path_request_style",
		"-o", "passwd_file=" + path.Join(dataRootPath, fmt.Sprintf("%x", sha256.Sum256([]byte(testDir))), passwordFileName),
		"-o", "url=" + testOSEndpoint,
		"-o", "endpoint=" + testStorageClass,
		"-o", "parallel_count=" + strconv.Itoa(testParallelCount),
		"-o", "multipart_size=" + strconv.Itoa(testChunkSizeMB),
		"-o", "dbglevel=" + testDebugLevel,
		"-o", "max_stat_cache_size=" + strconv.Itoa(testStatCacheSize),
		"-o", "allow_other",
		"-o", "max_background=1000",
		"-o", "mp_umask=002",
		"-o", "instance_name=" + testDir,
		"-o", "default_acl=private",
		"-o", "connect_timeout=1",
	}

	resp := p.Mount(r)
	if assert.Equal(t, interfaces.StatusSuccess, resp.Status) {
		assert.Equal(t, expectedArgs, commandArgs)
	}
}

func Test_ReadwriteTimeoutSeconds_NonInt(t *testing.T) {
	p := getPlugin()
	r := getMountRequest()
	r.Opts[optionReadwriteTimeoutSeconds] = "non-int-value"

	resp := p.Mount(r)
	if assert.Equal(t, interfaces.StatusFailure, resp.Status) {
		assert.Contains(t, resp.Message, "Cannot convert value of readwrite-timeout-seconds into integer")
	}
}

func Test_ReadwriteTimeoutSeconds_Positive(t *testing.T) {
	p := getPlugin()
	r := getMountRequest()
	r.Opts[optionReadwriteTimeoutSeconds] = "1"

	expectedArgs := []string{
		testBucket,
		testDir,
		"-o", "multireq_max=" + strconv.Itoa(testMultiReqMax),
		"-o", "cipher_suites=" + testTLSCipherSuite,
		"-o", "use_path_request_style",
		"-o", "passwd_file=" + path.Join(dataRootPath, fmt.Sprintf("%x", sha256.Sum256([]byte(testDir))), passwordFileName),
		"-o", "url=" + testOSEndpoint,
		"-o", "endpoint=" + testStorageClass,
		"-o", "parallel_count=" + strconv.Itoa(testParallelCount),
		"-o", "multipart_size=" + strconv.Itoa(testChunkSizeMB),
		"-o", "dbglevel=" + testDebugLevel,
		"-o", "max_stat_cache_size=" + strconv.Itoa(testStatCacheSize),
		"-o", "allow_other",
		"-o", "max_background=1000",
		"-o", "mp_umask=002",
		"-o", "instance_name=" + testDir,
		"-o", "default_acl=private",
		"-o", "readwrite_timeout=1",
	}

	resp := p.Mount(r)
	if assert.Equal(t, interfaces.StatusSuccess, resp.Status) {
		assert.Equal(t, expectedArgs, commandArgs)
	}
}
func Test_UseXattr_Positive(t *testing.T) {
	p := getPlugin()
	r := getMountRequest()
	r.Opts[optionUseXattr] = "true"

	expectedArgs := []string{
		testBucket,
		testDir,
		"-o", "multireq_max=" + strconv.Itoa(testMultiReqMax),
		"-o", "cipher_suites=" + testTLSCipherSuite,
		"-o", "use_path_request_style",
		"-o", "passwd_file=" + path.Join(dataRootPath, fmt.Sprintf("%x", sha256.Sum256([]byte(testDir))), passwordFileName),
		"-o", "url=" + testOSEndpoint,
		"-o", "endpoint=" + testStorageClass,
		"-o", "parallel_count=" + strconv.Itoa(testParallelCount),
		"-o", "multipart_size=" + strconv.Itoa(testChunkSizeMB),
		"-o", "dbglevel=" + testDebugLevel,
		"-o", "max_stat_cache_size=" + strconv.Itoa(testStatCacheSize),
		"-o", "allow_other",
		"-o", "max_background=1000",
		"-o", "mp_umask=002",
		"-o", "instance_name=" + testDir,
		"-o", "default_acl=private",
		"-o", "use_xattr",
	}

	resp := p.Mount(r)
	if assert.Equal(t, interfaces.StatusSuccess, resp.Status) {
		assert.Equal(t, expectedArgs, commandArgs)
	}
}

func Test_Mount_BadServiceInstanceID(t *testing.T) {
	p := getPlugin()
	r := getMountRequest()
	r.Opts[optionAPIKey] = base64.StdEncoding.EncodeToString([]byte(testAPIKey))
	r.Opts[optionServiceInstanceID] = "illegal-base-64"

	resp := p.Mount(r)
	if assert.Equal(t, interfaces.StatusFailure, resp.Status) {
		assert.Contains(t, resp.Message, "cannot decode Service Instance ID")
	}
}

func Test_Mount_ServiceNameNegative(t *testing.T) {
	p := getPlugin()
	r := getMountRequest()
	r.Opts[optionCAbundleB64] = base64.StdEncoding.EncodeToString([]byte(testCABundle))
	r.Opts[optionServiceIP] = testServiceIP
	writeFile = writeFileError
	resp := p.Mount(r)
	if assert.Equal(t, interfaces.StatusFailure, resp.Status) {
		assert.Contains(t, resp.Message, "cannot create ca crt file")
	}
}

func Test_Mount_ServiceNamePositive(t *testing.T) {
	p := getPlugin()
	r := getMountRequest()
	r.Opts[optionCAbundleB64] = base64.StdEncoding.EncodeToString([]byte(testCABundle))
	r.Opts[optionServiceIP] = testServiceIP
	expectedArgs := []string{
		testBucket,
		testDir,
		"-o", "multireq_max=" + strconv.Itoa(testMultiReqMax),
		"-o", "cipher_suites=" + testTLSCipherSuite,
		"-o", "use_path_request_style",
		"-o", "passwd_file=" + path.Join(dataRootPath, fmt.Sprintf("%x", sha256.Sum256([]byte(testDir))), passwordFileName),
		"-o", "url=" + testOSEndpoint,
		"-o", "endpoint=" + testStorageClass,
		"-o", "parallel_count=" + strconv.Itoa(testParallelCount),
		"-o", "multipart_size=" + strconv.Itoa(testChunkSizeMB),
		"-o", "dbglevel=" + testDebugLevel,
		"-o", "max_stat_cache_size=" + strconv.Itoa(testStatCacheSize),
		"-o", "allow_other",
		"-o", "max_background=1000",
		"-o", "mp_umask=002",
		"-o", "instance_name=" + testDir,
		"-o", "default_acl=private",
	}

	resp := p.Mount(r)
	if assert.Equal(t, interfaces.StatusSuccess, resp.Status) {
		assert.Equal(t, expectedArgs, commandArgs)
	}
}
