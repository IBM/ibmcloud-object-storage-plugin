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
	"errors"
	"fmt"
	"github.com/IBM/ibmcloud-object-storage-plugin/driver/interfaces"
	"github.com/IBM/ibmcloud-object-storage-plugin/utils/backend"
	"github.com/IBM/ibmcloud-object-storage-plugin/utils/parser"
	"go.uber.org/zap"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"syscall"
)

const (
	dataRootPath          = "/var/lib/ibmc-s3fs"
	passwordFileName      = "passwd"
	cacheDirectoryName    = "cache"
	defaultTLSCipherSuite = "AES"
	caPath                = "/tmp"
	// SecretAccessKey is the key name for the AWS Access Key
	SecretAccessKey = "access-key"
	// SecretSecretKey is the key name for the AWS Secret Key
	SecretSecretKey = "secret-key"
	// SecretAPIKey is the key name for the IBM API Key (IAM Authentication)
	SecretAPIKey = "api-key"
	// SecretServiceInstanceID is the key name for the service instance ID (IAM Authentication)
	SecretServiceInstanceID = "service-instance-id"
	// defaultIAMEndPoint is the default URL of the IBM IAM endpoint
	defaultIAMEndPoint = "https://iam.bluemix.net"
	// CrtBundle is the base64 encoded crt bundle
	CrtBundle = "ca-bundle-crt"
)

var (
	command            = exec.Command
	stat               = os.Stat
	unmount            = syscall.Unmount
	mount              = syscall.Mount
	writeFile          = ioutil.WriteFile
	mkdirAll           = os.MkdirAll
	removeAll          = os.RemoveAll
	hostname, anyerror = os.Hostname()
)

// buildVersion holds the driver version string
var buildVersion = ""
var podUID = ""

// Options are the FlexVolume driver options
type Options struct {
	ChunkSizeMB             int    `json:"chunk-size-mb,string"`
	ParallelCount           int    `json:"parallel-count,string"`
	MultiReqMax             int    `json:"multireq-max,string"`
	StatCacheSize           int    `json:"stat-cache-size,string"`
	FSGroup                 string `json:"kubernetes.io/fsGroup,omitempty"`
	Endpoint                string `json:"endpoint,omitempty"` //Will be deprecated
	Region                  string `json:"region,omitempty"`   //Will be deprecated
	Bucket                  string `json:"bucket"`
	ObjectPath              string `json:"object-path,omitempty"`
	DebugLevel              string `json:"debug-level"`
	CurlDebug               bool   `json:"curl-debug,string"`
	KernelCache             bool   `json:"kernel-cache,string,omitempty"`
	TLSCipherSuite          string `json:"tls-cipher-suite,omitempty"`
	S3FSFUSERetryCount      string `json:"s3fs-fuse-retry-count,omitempty"`
	StatCacheExpireSeconds  string `json:"stat-cache-expire-seconds,omitempty"`
	AccessKeyB64            string `json:"kubernetes.io/secret/access-key,omitempty"`
	SecretKeyB64            string `json:"kubernetes.io/secret/secret-key,omitempty"`
	APIKeyB64               string `json:"kubernetes.io/secret/api-key,omitempty"`
	OSEndpoint              string `json:"object-store-endpoint,omitempty"`
	OSStorageClass          string `json:"object-store-storage-class,omitempty"`
	IAMEndpoint             string `json:"iam-endpoint,omitempty"`
	ConnectTimeoutSeconds   string `json:"connect-timeout,omitempty"`
	ReadwriteTimeoutSeconds string `json:"readwrite-timeout,omitempty"`
	UseXattr                bool   `json:"use-xattr,string,omitempty"`
	AccessMode              string `json:"access-mode,omitempty"`
	ServiceInstanceIDB64    string `json:"kubernetes.io/secret/service-instance-id,omitempty"`
	CAbundleB64             string `json:"kubernetes.io/secret/ca-bundle-crt,omitempty"`
	ServiceIP               string `json:"service-ip,omitempty"`
}

// PathExists returns true if the specified path exists.
func pathExists(path string) (bool, error) {
	if path == "" {
		return false, errors.New("Undefined path")
	}
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	} else if os.IsNotExist(err) {
		return false, nil
	} else if isCorruptedMnt(err) {
		return true, err
	} else {
		return false, err
	}
}

// isCorruptedMnt return true if err is about corrupted mount point
func isCorruptedMnt(err error) bool {
	if err == nil {
		return false
	}
	var underlyingError error
	switch pe := err.(type) {
	case nil:
		return false
	case *os.PathError:
		underlyingError = pe.Err
	case *os.LinkError:
		underlyingError = pe.Err
	case *os.SyscallError:
		underlyingError = pe.Err
	}
	return underlyingError == syscall.ENOTCONN || underlyingError == syscall.ESTALE
}

// S3fsPlugin supports mount & unmount requests of s3fs volumes
type S3fsPlugin struct {
	Backend backend.ObjectStorageSessionFactory
	Logger  *zap.Logger
}

// SetBuildVersion sets the driver version
func SetBuildVersion(version string) {
	buildVersion = version
}

// SetPodUID sets the POD UID
func SetPodUID(poduid string) {
	podUID = poduid
}

// isMountpoint return true if pathname is a mountpoint
func (p *S3fsPlugin) isMountpoint(pathname string) (bool, error) {
	p.Logger.Info(podUID+":"+"Checking if path is mountpoint",
		zap.String("Pathname", pathname))

	out, err := command("mountpoint", pathname).CombinedOutput()
	outStr := strings.TrimSpace(string(out))
	if err != nil {
		if strings.HasSuffix(outStr, "Transport endpoint is not connected") {
			return true, err
		} else {
			return false, err
		}
	}
	//outStr := strings.TrimSpace(string(out))

	if strings.HasSuffix(outStr, "is a mountpoint") {
		p.Logger.Info(podUID+":"+"Path is a mountpoint",
			zap.String("Pathname", pathname))
		return true, nil
	} else if strings.HasSuffix(outStr, "is not a mountpoint") {
		p.Logger.Info(podUID+":"+"Path is NOT a mountpoint",
			zap.String("Pathname", pathname))
		return false, nil
	} else {
		p.Logger.Error(podUID+":"+"Cannot parse mountpoint result",
			zap.String("Error", outStr))
		return false, fmt.Errorf("cannot parse mountpoint result: %s", outStr)
	}
}

func (p *S3fsPlugin) unmountPath(mountPath string, delete bool) error {
	p.Logger.Info(podUID+":"+"Deleting mountpoint",
		zap.String("Mount path", mountPath))

	pathExist, err := pathExists(mountPath)
	if !pathExist {
		//If path does not exist
		if err == nil {
			p.Logger.Info(podUID+":"+"Path does not exist",
				zap.String("Mount path", mountPath))
			return nil
		} else {
			p.Logger.Error(podUID+":"+"Cannot stat directory",
				zap.String("Mount path", mountPath), zap.Error(err))
			return fmt.Errorf("cannot stat directory %s: %v", mountPath, err)
		}
	}

	// directory exists
	isMount, checkMountErr := p.isMountpoint(mountPath)
	if isMount || checkMountErr != nil {
		p.Logger.Info(podUID+":"+"Calling unmount",
			zap.String("Mount path", mountPath))
		err = unmount(mountPath, syscall.MNT_DETACH)
		if err != nil && checkMountErr == nil {
			p.Logger.Error(podUID+":"+"Cannot unmount. Trying force unmount",
				zap.String("Mount path", mountPath), zap.Error(err))
			//Do force unmount
			err = unmount(mountPath, syscall.MNT_FORCE)
			if err != nil {
				p.Logger.Error(podUID+":"+"Cannot force unmount",
					zap.String("Mount path", mountPath), zap.Error(err))
				return fmt.Errorf("cannot force unmount %s: %v", mountPath, err)
			}
		}
	}

	if delete {
		p.Logger.Info(podUID+":"+"Deleting directory",
			zap.String("Mount path", mountPath))
		err = removeAll(mountPath)
		if err != nil {
			p.Logger.Error(podUID+":"+"Cannot remove",
				zap.String("Mount path", mountPath), zap.Error(err))
			return fmt.Errorf("cannot remove %s: %v", mountPath, err)
		}
	}

	return nil
}

func (p *S3fsPlugin) createEmptyMountpoint(mountPath string) error {
	p.Logger.Info(podUID+":"+"Creating empty mountpoint",
		zap.String("mountPath", mountPath))

	err := p.unmountPath(mountPath, true)
	if err != nil {
		return err
	}

	// directory does not exist

	p.Logger.Info(podUID+":"+"Creating directory",
		zap.String("mountPath", mountPath))
	err = mkdirAll(mountPath, 0755)
	if err != nil {
		p.Logger.Error(podUID+":Cannot create directory",
			zap.String("mountPath", mountPath), zap.Error(err))
		return fmt.Errorf("cannot create directory %s: %v", mountPath, err)
	}

	// directory exists and unmounted
	p.Logger.Info(podUID+":"+"Creating tmpfs mountpoint",
		zap.String("mountPath", mountPath))
	err = mount("tmpfs", mountPath, "tmpfs", 0, "size=4k")
	if err != nil {
		p.Logger.Error(podUID+":Cannot create tmpfs mountpoint",
			zap.String("mountPath", mountPath), zap.Error(err))
		return fmt.Errorf("cannot create tmpfs mountpoint %s: %v", mountPath, err)
	}

	return nil
}

// Init method is to initialize the flexvolume, it is a no op right now
func (p *S3fsPlugin) Init() interfaces.FlexVolumeResponse {
	p.Logger.Info(podUID + ":" + "S3fsPlugin-Init()-start")
	defer p.Logger.Info(podUID + ":" + "S3fsPlugin-Init()-end")

	return interfaces.FlexVolumeResponse{
		Status:       interfaces.StatusSuccess,
		Message:      "Plugin init successfully",
		Capabilities: interfaces.CapabilitiesResponse{Attach: false, FSGroup: false},
	}
}

func (p *S3fsPlugin) checkBucket(endpoint, region, bucket string, creds *backend.ObjectStorageCredentials) error {
	p.Logger.Info(podUID+":"+"Checking if bucket exists",
		zap.String("bucket", bucket))
	sess := p.Backend.NewObjectStorageSession(endpoint, region, creds, p.Logger)
	return sess.CheckBucketAccess(bucket)
}

func (p *S3fsPlugin) checkObjectPath(endpoint, region, bucket, objectpath string, creds *backend.ObjectStorageCredentials) (bool, error) {
	p.Logger.Info(podUID+":"+"Checking if object-path exists inside bucket",
		zap.String("bucket", bucket), zap.String("object-path", objectpath))
	sess := p.Backend.NewObjectStorageSession(endpoint, region, creds, p.Logger)
	return sess.CheckObjectPathExistence(bucket, objectpath)
}

func (p *S3fsPlugin) createDirectoryIfNotExists(path string) error {
	p.Logger.Info(podUID+":"+"Checking if directory exists",
		zap.String("path", path))

	_, err := stat(path)
	if err == nil {
		p.Logger.Info(podUID+":"+"Directory exists",
			zap.String("path", path))
	} else {
		if os.IsNotExist(err) {
			p.Logger.Info(podUID+":"+"Creating directory",
				zap.String("path", path))
			err = mkdirAll(path, 0755)
			if err != nil {
				p.Logger.Error(podUID+":"+"Cannot create directory",
					zap.Error(err))
				return fmt.Errorf("cannot create directory: %v", err)
			}
		} else {
			p.Logger.Error(podUID+":"+"Cannot stat directory",
				zap.Error(err))
			return fmt.Errorf("cannot stat directory: %v", err)
		}
	}
	return nil
}

// Mount method allows to mount the volume/fileset to a given location for a pod
func (p *S3fsPlugin) mountInternal(mountRequest interfaces.FlexVolumeMountRequest) error {
	var options Options
	var apiKey, serviceInstanceId, accessKey, secretKey string
	var fInfo os.FileInfo
	var regionValue, endptValue, iamEndpoint string
	var fullBucketPath string

	err := parser.UnmarshalMap(&mountRequest.Opts, &options)
	if err != nil {
		p.Logger.Error(podUID+":"+"Cannot unmarshal driver options",
			zap.Error(err))
		return fmt.Errorf("cannot unmarshal driver options: %v", err)
	}

	// Support both endpoint and object-store-endpoint option
	if options.OSEndpoint != "" {
		endptValue = options.OSEndpoint
	} else {
		endptValue = options.Endpoint
	}
	// Support both region and object-store-storage-class option
	if options.OSStorageClass != "" {
		regionValue = options.OSStorageClass
	} else if options.Region != "" {
		regionValue = options.Region
	} else {
		regionValue = "dummy-object-store-storageclass"
	}

	if !(strings.HasPrefix(endptValue, "https://") || strings.HasPrefix(endptValue, "http://")) {
		p.Logger.Error(podUID+":"+
			"Bad value for object-store-endpoint: scheme is missing."+
			" Must be of the form http://<hostname> or https://<hostname>",
			zap.String("object-store-endpoint", endptValue))
		return fmt.Errorf(podUID+":"+
			"Bad value for object-store-endpoint \"%v\": scheme is missing."+
			" Must be of the form http://<hostname> or https://<hostname>",
			endptValue)
	}

	//Check if value of s3fs-fuse-retry-count parameter can be converted to integer
	if options.S3FSFUSERetryCount != "" {
		retryCount, err := strconv.Atoi(options.S3FSFUSERetryCount)
		if err != nil {
			p.Logger.Error(podUID+":"+
				"Cannot convert value of s3fs-fuse-retry-count into integer",
				zap.Error(err))
			return fmt.Errorf("Cannot convert value of s3fs-fuse-retry-count into integer: %v", err)
		}
		if retryCount < 1 {
			p.Logger.Error(podUID+":"+
				" value of s3fs-fuse-retry-count should be >= 1",
				zap.Error(err))
			return fmt.Errorf("value of s3fs-fuse-retry-count should be >= 1")
		}
	}

	//Check if value of stat-cache-expire-seconds parameter can be converted to integer
	if options.StatCacheExpireSeconds != "" {
		cacheExpireSeconds, err := strconv.Atoi(options.StatCacheExpireSeconds)
		if err != nil {
			p.Logger.Error(podUID+":"+
				" Cannot convert value of stat-cache-expire-seconds into integer",
				zap.Error(err))
			return fmt.Errorf("Cannot convert value of stat-cache-expire-seconds into integer: %v", err)
		} else if cacheExpireSeconds < 0 {
			p.Logger.Error(podUID+":"+
				" value of stat-cache-expire-seconds should be >= 0",
				zap.Error(err))
			return fmt.Errorf("value of stat-cache-expire-seconds should be >= 0")
		}
	}

	//Check if value of connect_timeout parameter can be converted to integer
	if options.ConnectTimeoutSeconds != "" {
		_, err := strconv.Atoi(options.ConnectTimeoutSeconds)
		if err != nil {
			p.Logger.Error(podUID+":"+
				"Cannot convert value of connect-timeout-seconds into integer",
				zap.Error(err))
			return fmt.Errorf("Cannot convert value of connect-timeout-seconds into integer: %v", err)
		}
	}

	//Check if value of connect_timeout parameter can be converted to integer
	if options.ReadwriteTimeoutSeconds != "" {
		_, err := strconv.Atoi(options.ReadwriteTimeoutSeconds)
		if err != nil {
			p.Logger.Error(podUID+":"+
				"Cannot convert value of readwrite-timeout-seconds into integer",
				zap.Error(err))
			return fmt.Errorf("Cannot convert value of readwrite-timeout-seconds into integer: %v", err)
		}
	}

	if options.APIKeyB64 != "" {
		apiKey, err = parser.DecodeBase64(options.APIKeyB64)
		if err != nil {
			p.Logger.Error(podUID+":"+
				" Cannot decode API key",
				zap.Error(err))
			return fmt.Errorf("cannot decode API key: %v", err)
		}
		serviceInstanceId, err = parser.DecodeBase64(options.ServiceInstanceIDB64)
		if err != nil {
			p.Logger.Error(podUID+":"+
				" Cannot decode Service Instance ID",
				zap.Error(err))
			return fmt.Errorf("cannot decode Service Instance ID: %v", err)
		}
	} else {
		accessKey, err = parser.DecodeBase64(options.AccessKeyB64)
		if err != nil {
			p.Logger.Error(podUID+":"+
				" Cannot decode access key",
				zap.Error(err))
			return fmt.Errorf("cannot decode access key: %v", err)
		}

		secretKey, err = parser.DecodeBase64(options.SecretKeyB64)
		if err != nil {
			p.Logger.Error(podUID+":"+
				" Cannot decode secret key",
				zap.Error(err))
			return fmt.Errorf("cannot decode secret key: %v", err)
		}
	}

	if apiKey != "" {
		if options.IAMEndpoint == "" {
			iamEndpoint = defaultIAMEndPoint
		} else {
			if !(strings.HasPrefix(options.IAMEndpoint, "https://") || strings.HasPrefix(options.IAMEndpoint, "http://")) {
				p.Logger.Error(podUID+":"+
					" Bad value for iam-endpoint."+
					" Must be of the form https://<hostname> or http://<hostname>",
					zap.String("iam-endpoint", options.IAMEndpoint))
				return fmt.Errorf(podUID+":"+
					" Bad value for iam-endpoint \"%v\":"+
					" Must be of the form https://<hostname> or http://<hostname>",
					options.IAMEndpoint)
			} else {
				iamEndpoint = options.IAMEndpoint
			}
		}
	}
	if options.CAbundleB64 != "" && options.ServiceIP != "" {
		CaBundleKey, err := parser.DecodeBase64(options.CAbundleB64)
		//caFile := path.Join(mountPath, caFileName)
		caFileName := options.ServiceIP + "_ ca.crt"
		caFile := path.Join(caPath, caFileName)
		err = writeFile(caFile, []byte(CaBundleKey), 0600)
		if err != nil {
			p.Logger.Error(podUID+":"+" Cannot create ca crt file",
				zap.Error(err))
			return fmt.Errorf("cannot create ca crt file: %v", err)
		}
		os.Setenv("CURL_CA_BUNDLE", caFile)
		os.Setenv("AWS_CA_BUNDLE", caFile)
	}
	// check that bucket exists before doing the mount
	err = p.checkBucket(endptValue, regionValue, options.Bucket,
		&backend.ObjectStorageCredentials{
			AccessKey:         accessKey,
			SecretKey:         secretKey,
			APIKey:            apiKey,
			ServiceInstanceID: serviceInstanceId,
			IAMEndpoint:       iamEndpoint})
	if err != nil {
		p.Logger.Error(podUID+":"+" Cannot access bucket",
			zap.Error(err))
		return fmt.Errorf("cannot access bucket: %v", err)
	}

	// check that object-path exists inside bucket before doing the mount
	if options.ObjectPath != "" {
		exist, err := p.checkObjectPath(endptValue, regionValue, options.Bucket, options.ObjectPath,
			&backend.ObjectStorageCredentials{
				AccessKey:         accessKey,
				SecretKey:         secretKey,
				APIKey:            apiKey,
				ServiceInstanceID: serviceInstanceId,
				IAMEndpoint:       iamEndpoint})
		if err != nil {
			p.Logger.Error(podUID+":"+" Cannot access object-path inside bucket",
				zap.String("bucket", options.Bucket), zap.String("object-path", options.ObjectPath), zap.Error(err))
			return fmt.Errorf("cannot access object-path \"%s\" inside bucket %s: %v", options.ObjectPath, options.Bucket, err)
		} else if !exist {
			p.Logger.Error(podUID+":"+" object-path not found inside bucket",
				zap.String("bucket", options.Bucket), zap.String("object-path", options.ObjectPath))
			return fmt.Errorf("object-path \"%s\" not found inside bucket %s", options.ObjectPath, options.Bucket)
		}
	}

	// create target directory
	err = p.createDirectoryIfNotExists(mountRequest.MountDir)
	if err != nil {
		p.Logger.Error(podUID+":"+"Cannot create target directory",
			zap.Error(err))
		return fmt.Errorf("cannot create target directory: %v", err)
	}

	// mount data path
	mountPath := path.Join(dataRootPath, fmt.Sprintf("%x", sha256.Sum256([]byte(mountRequest.MountDir))))
	done := false
	err = p.createEmptyMountpoint(mountPath)
	if err != nil {
		p.Logger.Error(podUID+":"+" Cannot create mount point",
			zap.Error(err))
		return fmt.Errorf("cannot create mount point: %v", err)
	}

	defer func() {
		// try to delete cache upon error or panic
		if !done {
			p.unmountPath(mountPath, true)
		}
	}()

	// create password file
	passwordFile := path.Join(mountPath, passwordFileName)
	if apiKey != "" {
		err = writeFile(passwordFile, []byte(":"+apiKey), 0600)
	} else {
		err = writeFile(passwordFile, []byte(accessKey+":"+secretKey), 0600)
	}
	if err != nil {
		p.Logger.Error(podUID+":"+" Cannot create password file",
			zap.Error(err))
		return fmt.Errorf("cannot create password file: %v", err)
	}
	var tlsCipherSuite string
	if options.TLSCipherSuite != "" {
		tlsCipherSuite = options.TLSCipherSuite
	} else {
		tlsCipherSuite = defaultTLSCipherSuite
	}

	if options.ObjectPath != "" {
		if strings.HasPrefix(options.ObjectPath, "/") {
			fullBucketPath = options.Bucket + ":" + options.ObjectPath
		} else {
			fullBucketPath = options.Bucket + ":/" + options.ObjectPath
		}
	} else {
		fullBucketPath = options.Bucket
	}
	args := []string{fullBucketPath, mountRequest.MountDir,
		"-o", "multireq_max=" + strconv.Itoa(options.MultiReqMax),
		"-o", "cipher_suites=" + tlsCipherSuite,
		"-o", "use_path_request_style",
		"-o", "passwd_file=" + passwordFile,
		"-o", "url=" + endptValue,
		"-o", "endpoint=" + regionValue,
		"-o", "parallel_count=" + strconv.Itoa(options.ParallelCount),
		"-o", "multipart_size=" + strconv.Itoa(options.ChunkSizeMB),
		"-o", "dbglevel=" + options.DebugLevel,
		"-o", "max_stat_cache_size=" + strconv.Itoa(options.StatCacheSize),
		"-o", "allow_other",
		"-o", "max_background=1000",
		"-o", "mp_umask=002",
		"-o", "instance_name=" + mountRequest.MountDir,
	}

	//if options.FSGroup != "" {
	if _, ok := mountRequest.Opts["kubernetes.io/fsGroup"]; ok {
		args = append(args, "-o", "gid="+options.FSGroup)
		args = append(args, "-o", "uid="+options.FSGroup)
	}

	// Check if AccessMode is ReadOnlyMany
	if options.AccessMode == "ReadOnlyMany" {
		args = append(args, "-o", "ro")
	}

	//Number of retries for failed S3 transaction
	if options.S3FSFUSERetryCount != "" {
		args = append(args, "-o", "retries="+options.S3FSFUSERetryCount)
	}

	if options.StatCacheExpireSeconds != "" {
		args = append(args, "-o", "stat_cache_expire="+options.StatCacheExpireSeconds)
	}

	if options.CurlDebug {
		args = append(args, "-o", "curldbg")
	}

	if options.KernelCache {
		args = append(args, "-o", "kernel_cache")
	}

	if apiKey != "" {
		args = append(args, "-o", "ibm_iam_auth")
		args = append(args, "-o", "ibm_iam_endpoint="+iamEndpoint)
	} else {
		args = append(args, "-o", "default_acl=private")
	}

	if options.ConnectTimeoutSeconds != "" {
		args = append(args, "-o", "connect_timeout="+options.ConnectTimeoutSeconds)
	}

	if options.ReadwriteTimeoutSeconds != "" {
		args = append(args, "-o", "readwrite_timeout="+options.ReadwriteTimeoutSeconds)
	}

	if options.UseXattr {
		args = append(args, "-o", "use_xattr")
	}

	fInfo, err = os.Lstat(mountRequest.MountDir)
	if err == nil {
		p.Logger.Info(podUID+":"+"Target directory before-mount: ",
			zap.String("mode:", fInfo.Mode().String()),
			zap.Uint32("uid:", fInfo.Sys().(*syscall.Stat_t).Uid),
			zap.Uint32("gid:", fInfo.Sys().(*syscall.Stat_t).Gid),
			zap.String("path:", mountRequest.MountDir))
	}

	p.Logger.Info(podUID+":"+"Running s3fs",
		zap.Reflect("args", args))

	output, err := command("s3fs", "--version").CombinedOutput()
	if err == nil {
		version := strings.Split(string(output), "\n")
		p.Logger.Info(podUID+":S3FS-Fuse info:", zap.String("Version", version[0]))
	}
	p.Logger.Info(podUID+":S3FS-Driver info:", zap.String("Version", buildVersion))

	out, err := command("s3fs", args...).CombinedOutput()
	if err != nil {
		p.Logger.Error(podUID+":"+"Running s3fs",
			zap.String("Error", string(out)))
		return fmt.Errorf("s3fs mount failed: %s", string(out))
	}

	fInfo, err = os.Lstat(mountRequest.MountDir)
	if err == nil {
		p.Logger.Info(podUID+":"+"Target directory after-mount: ",
			zap.String("mode:", fInfo.Mode().String()),
			zap.Uint32("uid:", fInfo.Sys().(*syscall.Stat_t).Uid),
			zap.Uint32("gid:", fInfo.Sys().(*syscall.Stat_t).Gid),
			zap.String("path:", mountRequest.MountDir))
	}

	done = true
	return nil
}

// Mount method allows to mount the volume/fileset to a given location for a pod
func (p *S3fsPlugin) Mount(mountRequest interfaces.FlexVolumeMountRequest) interfaces.FlexVolumeResponse {
	p.Logger.Info(podUID + ":" + "S3fsPlugin-Mount()-start")
	defer p.Logger.Info(podUID + ":" + "S3fsPlugin-Mount()-end")

	err := p.mountInternal(mountRequest)
	if err != nil {
		p.Logger.Info(podUID+":"+"Error mounting volume",
			zap.Reflect("err", err))

		return interfaces.FlexVolumeResponse{
			Status:  interfaces.StatusFailure,
			Message: fmt.Sprintf("Error mounting volume: %v", err),
		}
	}

	p.Logger.Info(podUID+":"+"Successfully executed mount",
		zap.String("mountRequest.MountDir", mountRequest.MountDir))

	return interfaces.FlexVolumeResponse{
		Status:  interfaces.StatusSuccess,
		Message: fmt.Sprintf("Volume mounted successfully to %s", mountRequest.MountDir),
	}
}

// Unmount methods unmounts the volume/ fileset from the pod
func (p *S3fsPlugin) Unmount(unmountRequest interfaces.FlexVolumeUnmountRequest) interfaces.FlexVolumeResponse {
	p.Logger.Info(podUID + ":" + "S3fsPlugin-Unmount()-start")
	defer p.Logger.Info(podUID + ":" + "S3fsPlugin-Unmount()-end")

	err := p.unmountInternal(unmountRequest)
	if err != nil {
		p.Logger.Info(podUID+":"+"Error unmounting volume",
			zap.Reflect("err", err))

		return interfaces.FlexVolumeResponse{
			Status:  interfaces.StatusFailure,
			Message: fmt.Sprintf("Error unmounting volume: %v", err),
		}
	}

	p.Logger.Info(podUID+":"+"Successfully executed unmount",
		zap.String("mountRequest.MountDir", unmountRequest.MountDir))

	return interfaces.FlexVolumeResponse{
		Status:  interfaces.StatusSuccess,
		Message: "Volume unmounted successfully",
	}
}

// Unmount methods unmounts the volume/ fileset from the pod
func (p *S3fsPlugin) unmountInternal(unmountRequest interfaces.FlexVolumeUnmountRequest) error {
	err := p.unmountPath(unmountRequest.MountDir, false)
	if err != nil {
		p.Logger.Error(podUID+":"+"Cannot unmount s3fs mount point",
			zap.String("Request", unmountRequest.MountDir),
			zap.Error(err))
		return fmt.Errorf("cannot unmount s3fs mount point %s: %v", unmountRequest.MountDir, err)
	}

	mountPath := path.Join(dataRootPath, fmt.Sprintf("%x", sha256.Sum256([]byte(unmountRequest.MountDir))))
	err = p.unmountPath(mountPath, true)
	if err != nil {
		p.Logger.Error(podUID+":"+"Cannot delete data  mount point",
			zap.String("mountpath", mountPath), zap.Error(err))
		return fmt.Errorf("cannot delete data mount point %s: %v", mountPath, err)
	}

	return nil
}
