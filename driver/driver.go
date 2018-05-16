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

	// SecretAccessKey is the key name for the AWS Access Key
	SecretAccessKey = "access-key"
	// SecretSecretKey is the key name for the AWS Secret Key
	SecretSecretKey = "secret-key"
	// SecretAPIKey is the key name for the IBM API Key (IAM Authentication)
	SecretAPIKey = "api-key"
	// SecretServiceInstanceID is the key name for the service instance ID (IAM Authentication)
	SecretServiceInstanceID = "service-instance-id"
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
var buildVersion string

// Options are the FlexVolume driver options
type Options struct {
	ChunkSizeMB        int    `json:"chunk-size-mb,string"`
	ParallelCount      int    `json:"parallel-count,string"`
	MultiReqMax        int    `json:"multireq-max,string"`
	StatCacheSize      int    `json:"stat-cache-size,string"`
	FSGroup            string `json:"kubernetes.io/fsGroup,omitempty"`
	Endpoint           string `json:"endpoint"`
	Region             string `json:"region"`
	Bucket             string `json:"bucket"`
	DebugLevel         string `json:"debug-level"`
	CurlDebug          bool   `json:"curl-debug,string"`
	KernelCache        bool   `json:"kernel-cache,string,omitempty"`
	TLSCipherSuite     string `json:"tls-cipher-suite,omitempty"`
	S3FSFUSERetryCount string `json:"s3fs-fuse-retry-count,omitempty"`
	AccessKeyB64       string `json:"kubernetes.io/secret/access-key,omitempty"`
	SecretKeyB64       string `json:"kubernetes.io/secret/secret-key,omitempty"`
	APIKeyB64          string `json:"kubernetes.io/secret/api-key,omitempty"`
}

// S3fsPlugin supports mount & unmount requests of s3fs volumes
type S3fsPlugin struct {
	Backend backend.ObjectStorageSessionFactory
	Logger  *zap.Logger
}

// SetBuildVersion sets the driver version
func (p *S3fsPlugin) SetBuildVersion(version string) {
	buildVersion = version
}

func (p *S3fsPlugin) isMountpoint(pathname string) (bool, error) {
	p.Logger.Info(hostname+" Component: S3FS Driver, Message: Checking if path is mountpoint",
		zap.String("Pathname", pathname))

	out, err := command("mountpoint", pathname).Output()
	if err != nil {
		return false, err
	}
	outStr := strings.TrimSpace(string(out))

	if strings.HasSuffix(outStr, "is a mountpoint") {
		p.Logger.Info(hostname+" Component: S3FS Driver, Message: Is a mountpoint",
			zap.String("Pathname", pathname))
		return true, nil
	} else if strings.HasSuffix(outStr, "is not a mountpoint") {
		p.Logger.Info(hostname+" Component: S3FS Driver, Message: Is NOT a mountpoint",
			zap.String("Pathname", pathname))
		return false, nil
	} else {
		p.Logger.Error(hostname+" Component: S3FS Driver, Message: Cannot parse mountpoint result",
			zap.String("Error", outStr))
		return false, fmt.Errorf("cannot parse mountpoint result: %s", outStr)
	}
}

func (p *S3fsPlugin) unmountPath(mountPath string, delete bool) error {
	p.Logger.Info(hostname+" Component: S3FS Driver, Message: Deleting mountpoint",
		zap.String("Mount path", mountPath))

	_, err := stat(mountPath)
	if err != nil {
		if os.IsNotExist(err) {
			p.Logger.Info(hostname+" Component: S3FS Driver, Message: Path does not exist",
				zap.String("Mount path", mountPath))
			return nil
		}
		p.Logger.Error(hostname+" Component: S3FS Driver, Message: Cannot stat directory",
			zap.String("Mount path", mountPath), zap.Error(err))
		return fmt.Errorf("cannot stat directory %s: %v", mountPath, err)
	}

	// directory exists

	isMount, checkMountErr := p.isMountpoint(mountPath)
	if isMount || checkMountErr != nil {
		p.Logger.Info(hostname+" Component: S3FS Driver, Message: Calling unmount",
			zap.String("Mount path", mountPath))
		err = unmount(mountPath, 0)
		if err != nil && checkMountErr == nil {
			p.Logger.Error(hostname+" Component: S3FS Driver, Message: Cannot unmount. Trying force unmount",
				zap.String("Mount path", mountPath), zap.Error(err))
			//Do force unmount
			err = unmount(mountPath, syscall.MNT_FORCE)
			if err != nil {
				p.Logger.Error(hostname+" Component: S3FS Driver, Message: Cannot force unmount",
					zap.String("Mount path", mountPath), zap.Error(err))
				return fmt.Errorf("cannot force unmount %s: %v", mountPath, err)
			}
		}
	}

	if delete {
		p.Logger.Info(hostname+" Component: S3FS Driver, Message: Deleting directory",
			zap.String("Mount path", mountPath))
		err = removeAll(mountPath)
		if err != nil {
			p.Logger.Error(hostname+" Component: S3FS Driver, Message: Cannot remove",
				zap.String("Mount path", mountPath), zap.Error(err))
			return fmt.Errorf("cannot remove %s: %v", mountPath, err)
		}
	}

	return nil
}

func (p *S3fsPlugin) createEmptyMountpoint(mountPath string) error {
	p.Logger.Info(hostname+" Component: S3FS Driver, Message: Creating empty mountpoint",
		zap.String("mountPath", mountPath))

	err := p.unmountPath(mountPath, true)
	if err != nil {
		return err
	}

	// directory does not exist

	p.Logger.Info(hostname+" Component: S3FS Driver, Message: Creating directory",
		zap.String("mountPath", mountPath))
	err = mkdirAll(mountPath, 0755)
	if err != nil {
		p.Logger.Error(hostname+" Component: S3FS Driver, Message: Cannot create directory",
			zap.String("mountPath", mountPath), zap.Error(err))
		return fmt.Errorf("cannot create directory %s: %v", mountPath, err)
	}

	// directory exists and unmounted

	p.Logger.Info(hostname+" Component: S3FS Driver, Message: Creating tmpfs mountpoint",
		zap.String("mountPath", mountPath))
	err = mount("tmpfs", mountPath, "tmpfs", 0, "size=4k")
	if err != nil {
		p.Logger.Error(hostname+" Component: S3FS Driver, Message: Cannot create tmpfs mountpoint",
			zap.String("mountPath", mountPath), zap.Error(err))
		return fmt.Errorf("cannot create tmpfs mountpoint %s: %v", mountPath, err)
	}

	return nil
}

// Init method is to initialize the flexvolume, it is a no op right now
func (p *S3fsPlugin) Init() interfaces.FlexVolumeResponse {
	p.Logger.Info(hostname + " Component: S3FS Driver, Message: S3fsPlugin-Init()-start")
	defer p.Logger.Info(hostname + " Component: S3FS Driver, Message: S3fsPlugin-Init()-end")

	return interfaces.FlexVolumeResponse{
		Status:       interfaces.StatusSuccess,
		Message:      "Plugin init successfully",
		Capabilities: interfaces.CapabilitiesResponse{Attach: false},
	}
}

func (p *S3fsPlugin) checkBucket(endpoint, region, bucket string, creds *backend.ObjectStorageCredentials) error {
	p.Logger.Info(hostname+" Component: S3FS Driver, Message: Checking that bucket exists",
		zap.String("bucket", bucket))
	sess := p.Backend.NewObjectStorageSession(endpoint, region, creds, p.Logger)
	return sess.CheckBucketAccess(bucket)
}

func (p *S3fsPlugin) createDirectoryIfNotExists(path string) error {
	p.Logger.Info(hostname+" Component: S3FS Driver, Message: Checking if directory exists",
		zap.String("path", path))

	_, err := stat(path)
	if err == nil {
		p.Logger.Info(hostname+" Component: S3FS Driver, Message: Directory exists",
			zap.String("path", path))
	} else {
		if os.IsNotExist(err) {
			p.Logger.Info(hostname+" Component: S3FS Driver, Message: Creating directory",
				zap.String("path", path))
			err = mkdirAll(path, 0755)
			if err != nil {
				p.Logger.Error(hostname+" Component: S3FS Driver, Message: Cannot create directory",
					zap.Error(err))
				return fmt.Errorf("cannot create directory: %v", err)
			}
		} else {
			p.Logger.Error(hostname+" Component: S3FS Driver, Message: Cannot stat directory",
				zap.Error(err))
			return fmt.Errorf("cannot stat directory: %v", err)
		}
	}
	return nil
}

// Mount method allows to mount the volume/fileset to a given location for a pod
func (p *S3fsPlugin) mountInternal(mountRequest interfaces.FlexVolumeMountRequest) error {
	var options Options
	var apiKey, accessKey, secretKey string
	var fInfo os.FileInfo

	err := parser.UnmarshalMap(&mountRequest.Opts, &options)
	if err != nil {
		p.Logger.Error(hostname+" Component: S3FS Driver, Message: Cannot unmarshal driver options",
			zap.Error(err))
		return fmt.Errorf("cannot unmarshal driver options: %v", err)
	}

	if !(strings.HasPrefix(options.Endpoint, "https://") || strings.HasPrefix(options.Endpoint, "http://")) {
		return fmt.Errorf(hostname+" Component: S3FS Driver, Message: Bad value for endpoint \"%v\": scheme is missing. "+
			"Must be of the form http://<hostname> or https://<hostname>", options.Endpoint)
	}

	//Check if value of s3fs-fuse-retry-count parameter can be converted to integer
	if options.S3FSFUSERetryCount != "" {
		retryCount, err := strconv.Atoi(options.S3FSFUSERetryCount)
		if err != nil {
			p.Logger.Error(hostname+" Component: S3FS Driver, "+
				"Message: Cannot convert value of s3fs-fuse-retry-count into integer",
				zap.Error(err))
			return fmt.Errorf("Cannot convert value of s3fs-fuse-retry-count into integer: %v", err)
		}
		if retryCount == 0 {
			p.Logger.Error(hostname+" Component: S3FS Driver, "+
				"Message: value of s3fs-fuse-retry-count should be non-zero",
				zap.Error(err))
			return fmt.Errorf("value of s3fs-fuse-retry-count should be non-zero")
		}
	}

	if options.APIKeyB64 != "" {
		apiKey, err = parser.DecodeBase64(options.APIKeyB64)
		if err != nil {
			p.Logger.Error(hostname+" Component: S3FS Driver, Message: Cannot decode API key",
				zap.Error(err))
			return fmt.Errorf("cannot decode API key: %v", err)
		}
	} else {
		accessKey, err = parser.DecodeBase64(options.AccessKeyB64)
		if err != nil {
			p.Logger.Error(hostname+" Component: S3FS Driver, Message: Cannot decode access key",
				zap.Error(err))
			return fmt.Errorf("cannot decode access key: %v", err)
		}

		secretKey, err = parser.DecodeBase64(options.SecretKeyB64)
		if err != nil {
			p.Logger.Error(hostname+" Component: S3FS Driver, Message: Cannot decode secret key",
				zap.Error(err))
			return fmt.Errorf("cannot decode secret key: %v", err)
		}
	}

	// check that bucket exists before doing the mount
	err = p.checkBucket(options.Endpoint, options.Region, options.Bucket, &backend.ObjectStorageCredentials{AccessKey: accessKey, SecretKey: secretKey, APIKey: apiKey})
	if err != nil {
		p.Logger.Error(hostname+" Component: S3FS Driver, Message: Cannot access bucket",
			zap.Error(err))
		return fmt.Errorf("cannot access bucket: %v", err)
	}

	// create target directory
	err = p.createDirectoryIfNotExists(mountRequest.MountDir)
	if err != nil {
		p.Logger.Error(hostname+" Component: S3FS Driver, Message: Cannot create target directory",
			zap.Error(err))
		return fmt.Errorf("cannot create target directory: %v", err)
	}

	// mount data path
	mountPath := path.Join(dataRootPath, fmt.Sprintf("%x", sha256.Sum256([]byte(mountRequest.MountDir))))
	done := false
	err = p.createEmptyMountpoint(mountPath)
	if err != nil {
		p.Logger.Error(hostname+" Component: S3FS Driver, Message: Cannot create mount point",
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
		p.Logger.Error(hostname+" Component: S3FS Driver, Message: Cannot create password file",
			zap.Error(err))
		return fmt.Errorf("cannot create password file: %v", err)
	}

	var tlsCipherSuite string
	if options.TLSCipherSuite != "" {
		tlsCipherSuite = options.TLSCipherSuite
	} else {
		tlsCipherSuite = defaultTLSCipherSuite
	}

	args := []string{options.Bucket, mountRequest.MountDir,
		"-o", "multireq_max=" + strconv.Itoa(options.MultiReqMax),
		"-o", "cipher_suites=" + tlsCipherSuite,
		"-o", "use_path_request_style",
		"-o", "passwd_file=" + passwordFile,
		"-o", "url=" + options.Endpoint,
		"-o", "endpoint=" + options.Region,
		"-o", "parallel_count=" + strconv.Itoa(options.ParallelCount),
		"-o", "multipart_size=" + strconv.Itoa(options.ChunkSizeMB),
		"-o", "dbglevel=" + options.DebugLevel,
		"-o", "max_stat_cache_size=" + strconv.Itoa(options.StatCacheSize),
		"-o", "allow_other",
		"-o", "sync_read",
		"-o", "mp_umask=002",
		"-o", "instance_name=" + mountRequest.MountDir,
	}

	//if options.FSGroup != "" {
	if _, ok := mountRequest.Opts["kubernetes.io/fsGroup"]; ok {
		if options.FSGroup == "65534" {
			args = append(args, "-o", "gid="+options.FSGroup)
		}
	}

	//Number of retries for failed S3 transaction
	if options.S3FSFUSERetryCount != "" {
		args = append(args, "-o", "retries="+options.S3FSFUSERetryCount)
	}

	if options.CurlDebug {
		args = append(args, "-o", "curldbg")
	}

	if options.KernelCache {
		args = append(args, "-o", "kernel_cache")
	}

	if apiKey != "" {
		args = append(args, "-o", "ibm_iam_auth")
	} else {
		args = append(args, "-o", "default_acl=")
	}

	fInfo, err = os.Lstat(mountRequest.MountDir)
	if err == nil {
		p.Logger.Info(hostname+" Component: S3FS Driver, Target Directory Before Mount: ",
			zap.String("mode:", fInfo.Mode().String()),
			zap.Uint32("uid:", fInfo.Sys().(*syscall.Stat_t).Uid),
			zap.Uint32("gid:", fInfo.Sys().(*syscall.Stat_t).Gid),
			zap.String("path:", mountRequest.MountDir))
	}

	p.Logger.Info(hostname+" Component: S3FS Driver, Message: Running s3fs",
		zap.Reflect("args", args))

	output, err := command("s3fs", "--version").CombinedOutput()
	if err == nil {
		version := strings.Split(string(output), "\n")
		p.Logger.Info(hostname+" S3FS Fuse info", zap.String("Version", version[0]))
	}
	p.Logger.Info(hostname+" S3FS Driver info", zap.String("Version", buildVersion))

	out, err := command("s3fs", args...).CombinedOutput()
	if err != nil {
		p.Logger.Error(hostname+" Component: S3FS Fuse, Message: Running s3fs",
			zap.String("Error", string(out)))
		return fmt.Errorf("s3fs mount failed: %s", string(out))
	}

	fInfo, err = os.Lstat(mountRequest.MountDir)
	if err == nil {
		p.Logger.Info(hostname+" Component: S3FS Driver, Message: Target Directory After Mount: ",
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
	p.Logger.Info(hostname + " Component: S3FS Driver, Message: S3fsPlugin-Mount()-start")
	defer p.Logger.Info(hostname + " Component: S3FS Driver, Message: S3fsPlugin-Mount()-end")

	err := p.mountInternal(mountRequest)
	if err != nil {
		p.Logger.Info(hostname+" Component: S3FS Driver, Message: Error mounting volume",
			zap.Reflect("err", err))

		return interfaces.FlexVolumeResponse{
			Status:  interfaces.StatusFailure,
			Message: fmt.Sprintf("Error mounting volume: %v", err),
		}
	}

	p.Logger.Info(hostname+" Component: S3FS Driver, Message: Successfully executed mount",
		zap.String("mountRequest.MountDir", mountRequest.MountDir))

	return interfaces.FlexVolumeResponse{
		Status:  interfaces.StatusSuccess,
		Message: fmt.Sprintf("Volume mounted successfully to %s", mountRequest.MountDir),
	}
}

// Unmount methods unmounts the volume/ fileset from the pod
func (p *S3fsPlugin) Unmount(unmountRequest interfaces.FlexVolumeUnmountRequest) interfaces.FlexVolumeResponse {
	p.Logger.Info(hostname + " Component: S3FS Driver, Message: S3fsPlugin-Unmount()-start")
	defer p.Logger.Info(hostname + " Component: S3FS Driver, Message: S3fsPlugin-Unmount()-end")

	err := p.unmountInternal(unmountRequest)
	if err != nil {
		p.Logger.Info(hostname+" Component: S3FS Driver, Message: Error unmounting volume",
			zap.Reflect("err", err))

		return interfaces.FlexVolumeResponse{
			Status:  interfaces.StatusFailure,
			Message: fmt.Sprintf("Error unmounting volume: %v", err),
		}
	}

	p.Logger.Info(hostname+" Component: S3FS Driver, Message: Successfully executed unmount",
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
		p.Logger.Error(hostname+" Component: S3FS Driver, Message: Cannot unmount s3fs mount point",
			zap.String("Request", unmountRequest.MountDir),
			zap.Error(err))
		return fmt.Errorf("cannot unmount s3fs mount point %s: %v", unmountRequest.MountDir, err)
	}

	mountPath := path.Join(dataRootPath, fmt.Sprintf("%x", sha256.Sum256([]byte(unmountRequest.MountDir))))
	err = p.unmountPath(mountPath, true)
	if err != nil {
		p.Logger.Error(hostname+" Component: S3FS Driver, Message: Cannot delete data  mount point",
			zap.String("mountpath", mountPath), zap.Error(err))
		return fmt.Errorf("cannot delete data mount point %s: %v", mountPath, err)
	}

	return nil
}
