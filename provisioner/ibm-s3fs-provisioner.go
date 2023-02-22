/*******************************************************************************
 * IBM Confidential
 * OCO Source Materials
 * IBM Cloud Kubernetes Service, 5737-D43
 * (C) Copyright IBM Corp. 2017, 2023 All Rights Reserved.
 * The source code for this program is not published or otherwise divested of
 * its trade secrets, irrespective of what has been deposited with
 * the U.S. Copyright Office.
 ******************************************************************************/

package provisioner

import (
	"context"
	"errors"
	"fmt"
	"github.com/IBM/ibmcloud-object-storage-plugin/driver"
	"github.com/IBM/ibmcloud-object-storage-plugin/ibm-provider/provider"
	"github.com/IBM/ibmcloud-object-storage-plugin/utils/backend"
	grpcClient "github.com/IBM/ibmcloud-object-storage-plugin/utils/grpc-client"
	"github.com/IBM/ibmcloud-object-storage-plugin/utils/logger"
	"github.com/IBM/ibmcloud-object-storage-plugin/utils/parser"
	"github.com/IBM/ibmcloud-object-storage-plugin/utils/uuid"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"io/ioutil"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"net"
	"os"
	"path"
	"sigs.k8s.io/sig-storage-lib-external-provisioner/v6/controller"
	"strconv"
	"strings"
	"time"
)

// PVC annotations
type pvcAnnotations struct {
	AutoCreateBucket        string `json:"ibm.io/auto-create-bucket"`
	AutoDeleteBucket        string `json:"ibm.io/auto-delete-bucket"`
	Bucket                  string `json:"ibm.io/bucket"`
	ObjectPath              string `json:"ibm.io/object-path,omitempty"`
	Endpoint                string `json:"ibm.io/endpoint,omitempty"` //Will be deprecated
	Region                  string `json:"ibm.io/region,omitempty"`   //Will be deprecated
	SecretName              string `json:"ibm.io/secret-name"`
	ChunkSizeMB             string `json:"ibm.io/chunk-size-mb,omitempty"`
	ParallelCount           string `json:"ibm.io/parallel-count,omitempty"`
	MultiReqMax             string `json:"ibm.io/multireq-max,omitempty"`
	StatCacheSize           string `json:"ibm.io/stat-cache-size,omitempty"`
	S3FSFUSERetryCount      string `json:"ibm.io/s3fs-fuse-retry-count,omitempty"`
	StatCacheExpireSeconds  string `json:"ibm.io/stat-cache-expire-seconds,omitempty"`
	IAMEndpoint             string `json:"ibm.io/iam-endpoint,omitempty"`
	ValidateBucket          string `json:"ibm.io/validate-bucket,omitempty"`
	SecretNamespace         string `json:"ibm.io/secret-namespace,omitempty"`
	ConnectTimeoutSeconds   string `json:"ibm.io/connect-timeout,omitempty"`
	ReadwriteTimeoutSeconds string `json:"ibm.io/readwrite-timeout,omitempty"`
	UseXattr                bool   `json:"ibm.io/use-xattr,string,omitempty"`
	CurlDebug               bool   `json:"ibm.io/curl-debug,string,omitempty"`
	DebugLevel              string `json:"ibm.io/debug-level,omitempty"`
	TLSCipherSuite          string `json:"ibm.io/tls-cipher-suite,omitempty"`
	CosServiceName          string `json:"ibm.io/cos-service"`
	CosServiceNamespace     string `json:"ibm.io/cos-service-ns,omitempty"`
	AutoCache               bool   `json:"ibm.io/auto_cache,string,omitempty"`
	SetAccessPolicy         string `json:"ibm.io/set-access-policy,omitempty"`
	AccessPolicyAllowedIps  string `json:"ibm.io/access-policy-allowed-ips,omitempty"`
	AddMountParam           string `json:"ibm.io/add-mount-param,omitempty"`
	QuotaLimit              string `json:"ibm.io/quota-limit,omitempty"`
}

// Storage Class options
type scOptions struct {
	AutoCreateBucket        string `json:"ibm.io/auto-create-bucket,omitempty"`
	AutoDeleteBucket        string `json:"ibm.io/auto-delete-bucket,omitempty"`
	Bucket                  string `json:"ibm.io/bucket,omitempty"`
	ObjectPath              string `json:"ibm.io/object-path,omitempty"`
	SecretName              string `json:"ibm.io/secret-name,omitempty"`
	SecretNamespace         string `json:"ibm.io/secret-namespace,omitempty"`
	ChunkSizeMB             int    `json:"ibm.io/chunk-size-mb,string"`
	ParallelCount           int    `json:"ibm.io/parallel-count,string"`
	MultiReqMax             int    `json:"ibm.io/multireq-max,string"`
	StatCacheSize           int    `json:"ibm.io/stat-cache-size,string"`
	TLSCipherSuite          string `json:"ibm.io/tls-cipher-suite,omitempty"`
	DebugLevel              string `json:"ibm.io/debug-level"`
	CurlDebug               bool   `json:"ibm.io/curl-debug,string,omitempty"`
	KernelCache             bool   `json:"ibm.io/kernel-cache,string,omitempty"`
	S3FSFUSERetryCount      string `json:"ibm.io/s3fs-fuse-retry-count,omitempty"`
	StatCacheExpireSeconds  string `json:"ibm.io/stat-cache-expire-seconds,omitempty"`
	IAMEndpoint             string `json:"ibm.io/iam-endpoint,omitempty"`
	OSEndpoint              string `json:"ibm.io/object-store-endpoint,omitempty"`
	OSStorageClass          string `json:"ibm.io/object-store-storage-class,omitempty"`
	ConnectTimeoutSeconds   string `json:"ibm.io/connect-timeout,omitempty"`
	ReadwriteTimeoutSeconds string `json:"ibm.io/readwrite-timeout,omitempty"`
	UseXattr                bool   `json:"ibm.io/use-xattr,string"`
	AddMountParam           string `json:"ibm.io/add-mount-param,omitempty"`
}

const (
	driverName           = "ibm/ibmc-s3fs"
	autoBucketNamePrefix = "tmp-s3fs-"
	fsType               = ""
	caBundlePath         = "/tmp/"
	defaultName          = "IBMGrpcProvider"
	clusterTypeVpcG2     = "vpc-gen2"
	clusterTypeClassic   = "cruiser"
	ResConfApiKey        = "res-conf-apikey"
	KPRootKeyCRN         = "kp-root-key-crn"
)

var SockEndpoint *string
var ConfigBucketAccessPolicy *bool
var ConfigQuotaLimit *bool
var AllowCrossNsSecret *bool

// IBMS3fsProvisioner is a dynamic provisioner of persistent volumes backed by Object Storage via s3fs
type IBMS3fsProvisioner struct {
	// Backend is the object store session factory
	Backend backend.ObjectStorageSessionFactory
	// GRPCBackend is the grpc session factory
	GRPCBackend grpcClient.GrpcSessionFactory
	// AccessPolicy is the resource configuration session factory
	AccessPolicy backend.AccessPolicyFactory
	// IBMProvider is the ibm provider client
	IBMProvider provider.IBMProviderClientFactory

	// Logger will be used for logging
	Logger *zap.Logger
	// Client is the Kubernetes Go-Client that will be used to fetch user credentials
	Client kubernetes.Interface
	// UUIDGenerator is a UUID generator that will be used to generate bucket names
	UUIDGenerator uuid.Generator
}

var _ controller.Provisioner = &IBMS3fsProvisioner{}
var writeFile = ioutil.WriteFile

func UnixConnect(addr string, t time.Duration) (net.Conn, error) {
	unix_addr, err := net.ResolveUnixAddr("unix", addr) // nolint:ineffassign
	conn, err := net.DialUnix("unix", nil, unix_addr)
	return conn, err
}

func parseSecret(secret *v1.Secret, keyName string) (string, error) {
	bytesVal, ok := secret.Data[keyName]
	if !ok {
		return "", fmt.Errorf("%s secret missing", keyName)
	}
	return string(bytesVal), nil
}

func (p *IBMS3fsProvisioner) writeCrtFile(ctx context.Context, secretName, secretNamespace, serviceName string) error {
	if serviceName == "" {
		serviceName = "standard-cos"
	}
	crtFile := path.Join(caBundlePath, serviceName)
	secrets, err := p.Client.CoreV1().Secrets(secretNamespace).Get(ctx, secretName, metav1.GetOptions{})
	if err != nil {
		return err
	}
	crtKey, err := parseSecret(secrets, driver.CrtBundle)
	if err != nil {
		//CA Cert not provided, try default one
		return nil
	}
	if err = writeFile(crtFile, []byte(crtKey), 0600); err != nil {
		return err
	}
	if err = os.Setenv("AWS_CA_BUNDLE", crtFile); err != nil {
		return err
	}
	return nil
}

func (p *IBMS3fsProvisioner) getCredentials(ctx context.Context, secretName, secretNamespace string) (credentials *backend.ObjectStorageCredentials, allowedNamespace []string, resConfApiKey string, kpRootKeyCrn string, err error) {
	secrets, err := p.Client.CoreV1().Secrets(secretNamespace).Get(ctx, secretName, metav1.GetOptions{})
	if err != nil {
		return nil, nil, "", "", fmt.Errorf("cannot retrieve secret %s: %v", secretName, err)
	}

	if strings.TrimSpace(string(secrets.Type)) != driverName {
		return nil, nil, "", "", fmt.Errorf("Wrong Secret Type.Provided secret of type %s.Expected type %s", string(secrets.Type), driverName)
	}

	var accessKey, secretKey, apiKey, serviceInstanceID string

	if bytesVal, ok := secrets.Data[driver.SecretAllowedNS]; ok {
		allowedNamespace = strings.Split(string(bytesVal), " ")
	}

	apiKey, err = parseSecret(secrets, driver.SecretAPIKey)
	if err != nil {
		accessKey, err = parseSecret(secrets, driver.SecretAccessKey)
		if err != nil {
			return nil, nil, "", "", err
		}

		secretKey, err = parseSecret(secrets, driver.SecretSecretKey)
		if err != nil {
			return nil, nil, "", "", err
		}
	} else {
		serviceInstanceID, err = parseSecret(secrets, driver.SecretServiceInstanceID)
	}

	if bytesVal, ok := secrets.Data[ResConfApiKey]; ok {
		resConfApiKey = string(bytesVal)
	}

	if bytesVal, ok := secrets.Data[KPRootKeyCRN]; ok {
		kpRootKeyCrn = string(bytesVal)
	}

	return &backend.ObjectStorageCredentials{
		AccessKey:         accessKey,
		SecretKey:         secretKey,
		APIKey:            apiKey,
		ServiceInstanceID: serviceInstanceID,
	}, allowedNamespace, resConfApiKey, kpRootKeyCrn, nil
}

func (p *IBMS3fsProvisioner) validateAnnotations(ctx context.Context, options controller.ProvisionOptions) (pvcAnnotations, scOptions, string, error) {
	var pvc pvcAnnotations
	var sc scOptions
	var pvcName = options.PVC.Name
	var clusterID = os.Getenv("CLUSTER_ID")
	var svcIp string
	var err error

	contextLogger, _ := logger.GetZapDefaultContextLogger()
	contextLogger.Info(pvcName + ":" + clusterID + ":validate annotations and assign default values to annotations")

	if err := parser.UnmarshalMap(&options.PVC.Annotations, &pvc); err != nil {
		return pvc, sc, svcIp, fmt.Errorf(pvcName+":"+clusterID+":cannot unmarshal PVC annotations: %v", err)
	}

	if err := parser.UnmarshalMap(&options.StorageClass.Parameters, &sc); err != nil {
		return pvc, sc, svcIp, fmt.Errorf(pvcName+":"+clusterID+":cannot unmarshal storage class parameters: %v", err)
	}

	if pvc.SecretName == "" {
		if sc.SecretName != "" {
			pvc.SecretName = sc.SecretName
		} else {
			return pvc, sc, svcIp, errors.New(pvcName + ":" + clusterID + ":secret-name not specified")
		}
	}

	contextLogger.Info(pvcName + ":" + clusterID + " AllowCrossNsSecret: " + strconv.FormatBool(*AllowCrossNsSecret))

	if AllowCrossNsSecret != nil && *AllowCrossNsSecret == false {
		contextLogger.Info(pvcName + ":" + clusterID + " AllowCrossNsSecret is set to false, the secret will be looked for in same namespace where pvc is created")
		if pvc.SecretNamespace != "" {
			contextLogger.Warn(pvcName + ":" + clusterID + " Ignoring 'ibm.io/secret-namespace' annotation as AllowCrossNsSecret is set to false")
		}
		pvc.SecretNamespace = options.PVC.Namespace
	} else {
		if pvc.SecretNamespace == "" {
			if sc.SecretNamespace != "" {
				pvc.SecretNamespace = sc.SecretNamespace
			} else {
				pvc.SecretNamespace = options.PVC.Namespace
			}
		}
	}

	if pvc.AutoCreateBucket == "" {
		if sc.AutoCreateBucket != "" {
			pvc.AutoCreateBucket = sc.AutoCreateBucket
		} else {
			pvc.AutoCreateBucket = "true"
		}
	} else if _, err := strconv.ParseBool(pvc.AutoCreateBucket); err != nil {
		return pvc, sc, svcIp, fmt.Errorf(pvcName+":"+clusterID+":invalid value for auto-create-bucket, expects true/false: %v", err)
	}

	if pvc.AutoDeleteBucket == "" {
		if sc.AutoDeleteBucket != "" {
			pvc.AutoDeleteBucket = sc.AutoDeleteBucket
		} else {
			pvc.AutoDeleteBucket = "false"
		}
	} else if _, err := strconv.ParseBool(pvc.AutoDeleteBucket); err != nil {
		return pvc, sc, svcIp, fmt.Errorf(pvcName+":"+clusterID+":invalid value for auto-delete-bucket, expects true/false: %v", err)
	}

	if pvc.Bucket == "" && sc.Bucket != "" {
		pvc.Bucket = sc.Bucket
	}

	if pvc.ObjectPath == "" && sc.ObjectPath != "" {
		pvc.ObjectPath = sc.ObjectPath
	}

	if pvc.AccessPolicyAllowedIps != "" {
		validIps, wrongIpArr := parser.ParseIPs(pvc.AccessPolicyAllowedIps)
		if !validIps {
			return pvc, sc, svcIp, fmt.Errorf(pvcName+":"+clusterID+":invalid value for access-policy-allowed-ips,  invalid ips are : %v", wrongIpArr)
		}
	}

	if pvc.SetAccessPolicy != "" {
		if _, err := strconv.ParseBool(pvc.SetAccessPolicy); err != nil {
			return pvc, sc, svcIp, fmt.Errorf(pvcName+":"+clusterID+":invalid value for set-access-policy, expects true/false: %v", err)
		}
	}

	if pvc.QuotaLimit != "" {
		if _, err := strconv.ParseBool(pvc.QuotaLimit); err != nil {
			return pvc, sc, svcIp, fmt.Errorf(pvcName+":"+clusterID+":invalid value for quota-limit, expects true/false: %v", err)
		}
	}

	if pvc.CosServiceName != "" {
		// TLS enabled COS Service
		if pvc.CosServiceNamespace != "" {
			// Generate the COS Service DNS name
			svc, err := p.Client.CoreV1().Services(pvc.CosServiceNamespace).Get(ctx, pvc.CosServiceName, metav1.GetOptions{})
			if err != nil {
				return pvc, sc, svcIp, fmt.Errorf(pvcName+":"+clusterID+":cannot retrieve service details: %v", err)
			}
			port := svc.Spec.Ports[0].Port
			svcIp = svc.Spec.ClusterIP
			endPoint := "https://" + pvc.CosServiceName + "." + pvc.CosServiceNamespace + ".svc.cluster.local:" + strconv.Itoa(int(port))
			pvc.Endpoint = endPoint
		}
	}
	// retrieve CA Cert if provided in secrets
	if err := p.writeCrtFile(ctx, pvc.SecretName, pvc.SecretNamespace, pvc.CosServiceName); err != nil {
		return pvc, sc, svcIp, fmt.Errorf("cannot retrieve secret: %v", err)
	}

	//Override value of EndPoint defined in storageclass
	// EndPoint should be defined in storage class.
	if pvc.Endpoint != "" {
		sc.OSEndpoint = pvc.Endpoint
	}

	//Override value of OSStorageClass defined in storageclass.
	// pvc Region will be deprecated.
	if pvc.Region != "" {
		sc.OSStorageClass = pvc.Region
	}

	if !(strings.HasPrefix(sc.OSEndpoint, "https://") || strings.HasPrefix(sc.OSEndpoint, "http://")) {
		return pvc, sc, svcIp, fmt.Errorf(pvcName+":"+clusterID+
			":Bad value for ibm.io/object-store-endpoint \"%v\": scheme is missing. "+
			"Must be of the form http://<hostname> or https://<hostname>",
			sc.OSEndpoint)
	}

	if pvc.IAMEndpoint != "" {
		sc.IAMEndpoint = pvc.IAMEndpoint
	}

	if !(strings.HasPrefix(sc.IAMEndpoint, "https://") || strings.HasPrefix(sc.IAMEndpoint, "http://")) {
		return pvc, sc, svcIp, fmt.Errorf(pvcName+":"+clusterID+
			":Bad value for ibm.io/iam-endpoint \"%v\":"+
			" Must be of the form https://<hostname> or http://<hostname>",
			sc.IAMEndpoint)
	}

	//Override value of s3fs-fuse-retry-count defined in storageclass
	if pvc.S3FSFUSERetryCount != "" {
		sc.S3FSFUSERetryCount = pvc.S3FSFUSERetryCount
	}
	if sc.S3FSFUSERetryCount != "" {
		if retryCount, err := strconv.Atoi(sc.S3FSFUSERetryCount); err != nil {
			return pvc, sc, svcIp, fmt.Errorf(pvcName+":"+clusterID+":Cannot convert value of s3fs-fuse-retry-count into integer: %v", err)
		} else if retryCount < 1 {
			return pvc, sc, svcIp, fmt.Errorf(pvcName + ":" + clusterID + ":value of s3fs-fuse-retry-count should be >= 1")
		}
	}

	//Override value of stat-cache-expire-seconds defined in storageclass
	if pvc.StatCacheExpireSeconds != "" {
		sc.StatCacheExpireSeconds = pvc.StatCacheExpireSeconds
	}
	if sc.StatCacheExpireSeconds != "" {
		if cacheExpireSeconds, err := strconv.Atoi(sc.StatCacheExpireSeconds); err != nil {
			return pvc, sc, svcIp, fmt.Errorf(pvcName+":"+clusterID+":Cannot convert value of stat-cache-expire-seconds into integer: %v", err)
		} else if cacheExpireSeconds < 0 {
			return pvc, sc, svcIp, fmt.Errorf(pvcName + ":" + clusterID + ":value of stat-cache-expire-seconds should be >= 0")
		}
	}

	//Override value of chunk-size-mb defined in storageclass
	if pvc.ChunkSizeMB != "" {
		if sc.ChunkSizeMB, err = strconv.Atoi(pvc.ChunkSizeMB); err != nil {
			return pvc, sc, svcIp, fmt.Errorf(pvcName+":"+clusterID+":Cannot convert value of chunk-size-mb into integer: %v", err)
		}
	}

	//Override value of parallel-count defined in storageclass
	if pvc.ParallelCount != "" {
		if sc.ParallelCount, err = strconv.Atoi(pvc.ParallelCount); err != nil {
			return pvc, sc, svcIp, fmt.Errorf(pvcName+":"+clusterID+":Cannot convert value of parallel-count into integer: %v", err)
		}
	}

	//Override value of multireq-max defined in storageclass
	if pvc.MultiReqMax != "" {
		if sc.MultiReqMax, err = strconv.Atoi(pvc.MultiReqMax); err != nil {
			return pvc, sc, svcIp, fmt.Errorf(pvcName+":"+clusterID+":Cannot convert value of multireq-max into integer: %v", err)
		}
	}

	//Override value of stat-cache-size defined in storageclass
	if pvc.StatCacheSize != "" {
		if sc.StatCacheSize, err = strconv.Atoi(pvc.StatCacheSize); err != nil {
			return pvc, sc, svcIp, fmt.Errorf(pvcName+":"+clusterID+":Cannot convert value of stat-cache-size into integer: %v", err)
		}
	}

	if pvc.ConnectTimeoutSeconds != "" {
		if _, err := strconv.Atoi(pvc.ConnectTimeoutSeconds); err != nil {
			return pvc, sc, svcIp, fmt.Errorf(pvcName+":"+clusterID+":Cannot convert value of connect-timeout-seconds into integer: %v", err)
		}
		sc.ConnectTimeoutSeconds = pvc.ConnectTimeoutSeconds
	}

	if pvc.ReadwriteTimeoutSeconds != "" {
		if _, err := strconv.Atoi(pvc.ReadwriteTimeoutSeconds); err != nil {
			return pvc, sc, svcIp, fmt.Errorf(pvcName+":"+clusterID+":Cannot convert value of readwrite-timeout-seconds into integer: %v", err)
		}
		sc.ReadwriteTimeoutSeconds = pvc.ReadwriteTimeoutSeconds
	}

	if pvc.AutoCreateBucket == "true" && pvc.ObjectPath != "" {
		return pvc, sc, svcIp, fmt.Errorf(pvcName+":"+clusterID+":object-path cannot be set when auto-create is enabled, got: %s", pvc.ObjectPath)
	}

	// Additional parameter should be of form "-o opt1 -o opt2=xxx -o opt3"
	if pvc.AddMountParam != "" {
		sc.AddMountParam = pvc.AddMountParam
	}

	return pvc, sc, svcIp, nil
}

// Provision provisions a new persistent volume
func (p *IBMS3fsProvisioner) Provision(ctx context.Context, options controller.ProvisionOptions) (*v1.PersistentVolume, controller.ProvisioningState, error) {
	//var pvc pvcAnnotations
	//var sc scOptions
	var pvcName = options.PVC.Name
	var pvcNamespace = options.PVC.Namespace
	var clusterID = os.Getenv("CLUSTER_ID")
	var msg, resConfApiKey, kpRootKeyCrn, providerType, vpcServiceEndpoints string
	var valBucket = true // nolint:ineffassign
	var allowedNamespace []string
	var creds *backend.ObjectStorageCredentials
	var sess backend.ObjectStorageSession
	var grpcSess grpcClient.GrpcSession
	var updateAP backend.AccessPolicy
	var rcc backend.ResourceConfigurationV1
	var providerClient provider.IBMProviderClient
	var setBucketAccessPolicy = false
	var setQuotaLimit = false
	var quotaLimit int64

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	contextLogger, _ := logger.GetZapDefaultContextLogger()
	contextLogger.Info(pvcName + ":" + clusterID + ":Provisioning storage with these spec")
	contextLogger.Info(pvcName+":"+clusterID+":PVC Details: ", zap.String("pvc", options.PVName))

	pvc, sc, svcIp, err := p.validateAnnotations(ctx, options)
	if err != nil {
		return nil, controller.ProvisioningFinished, fmt.Errorf(pvcName+":"+clusterID+":cannot validate annotations: %v", err)
	}

	//this handles the case where AutoDeleteBucket is set to true
	if pvc.AutoDeleteBucket == "true" {
		if pvc.AutoCreateBucket == "false" {
			return nil, controller.ProvisioningFinished, errors.New(pvcName + ":" + clusterID + ":bucket auto-create must be enabled when bucket auto-delete is enabled")
		}

		if pvc.Bucket != "" {
			return nil, controller.ProvisioningFinished, fmt.Errorf(pvcName+":"+clusterID+":bucket cannot be set when auto-delete is enabled, got: %s", pvc.Bucket)
		}

		id, err := p.UUIDGenerator.New()
		if err != nil {
			return nil, controller.ProvisioningFinished, fmt.Errorf(pvcName+":"+clusterID+":cannot create UUID for bucket name: %v", err)
		}
		pvc.Bucket = autoBucketNamePrefix + id
	}

	if pvc.ValidateBucket == "no" && pvc.AutoCreateBucket == "false" {
		valBucket = false
	} else {
		valBucket = true
	}

	//var err_msg error
	if valBucket {
		creds, allowedNamespace, resConfApiKey, kpRootKeyCrn, err = p.getCredentials(ctx, pvc.SecretName, pvc.SecretNamespace)
		if err != nil {
			return nil, controller.ProvisioningFinished, fmt.Errorf(pvcName+":"+clusterID+":cannot get credentials: %v", err)
		}

		creds.IAMEndpoint = sc.IAMEndpoint
		sess = p.Backend.NewObjectStorageSession(sc.OSEndpoint, sc.OSStorageClass, creds, p.Logger)
	}

	if len(allowedNamespace) > 0 {
		var allowed = false

		for _, item := range allowedNamespace {
			if item == pvcNamespace {
				allowed = true
				break
			}
		}
		if allowed == false {
			return nil, controller.ProvisioningFinished, errors.New(pvcName + ":" + clusterID + ":PVC creation in " + pvcNamespace + " namespace is not allowed")
		}
	}

	contextLogger.Info(pvcName + ":" + clusterID + " ConfigBucketAccessPolicy: " + strconv.FormatBool(*ConfigBucketAccessPolicy) + ", SetQuotaLimit: " + strconv.FormatBool(*ConfigQuotaLimit))

	if ConfigQuotaLimit != nil && *ConfigQuotaLimit && pvc.QuotaLimit != "false" {

		updateAP = p.AccessPolicy.NewAccessPolicy()
		rcc = &backend.UpdateAPObj{}

		//retrieve the quota value from PVC spec => spec.resources.requests.storage
		quotaSet := options.PVC.Spec.Resources.Requests[v1.ResourceName(v1.ResourceStorage)]

		quotaLimit = quotaSet.Value()
		contextLogger.Info(pvcName + ":" + clusterID + ":quota-limit value to be set for bucket: " + strconv.FormatInt(quotaLimit, 10))
		setQuotaLimit = true

	} else {
		if pvc.QuotaLimit == "false" {
			contextLogger.Info(pvcName + ":" + clusterID + " bucket :'" + pvc.Bucket + " quota-limit annotation is set to false for this PVC. bucket quota limit will not be set for this PVC")
		}
	}

	//add check for region = BNNP
	if ConfigBucketAccessPolicy != nil && *ConfigBucketAccessPolicy && pvc.SetAccessPolicy != "false" {
		grpcSess = p.GRPCBackend.NewGrpcSession()
		cc := &grpcClient.GrpcSes{}
		conn, err := grpcSess.GrpcDial(cc, *SockEndpoint, grpc.WithInsecure(), grpc.WithBlock(), grpc.WithDialer(UnixConnect))
		if err != nil {
			return nil, controller.ProvisioningFinished, fmt.Errorf(pvcName+":"+clusterID+":failed to establish grpc-client connection: %v", err)
		}

		providerClient = p.IBMProvider.NewIBMProviderClient(conn)
		if conn != nil {
			defer conn.Close()
		}

		name := defaultName
		if len(os.Args) > 1 {
			name = os.Args[1]
		}

		clusterTypeResp, err := providerClient.GetProviderType(ctx, &provider.ProviderTypeRequest{Id: name})
		if err != nil {
			return nil, controller.ProvisioningFinished, fmt.Errorf(pvcName+":"+clusterID+" :failed to get provider type for cluster: %v", err)
		}
		providerType = clusterTypeResp.GetType()
		contextLogger.Info(pvcName + ":" + clusterID + " : ClusterType  : " + providerType)

		if strings.Contains(providerType, clusterTypeVpcG2) {
			if pvc.AccessPolicyAllowedIps != "" {
				vpcServiceEndpoints = pvc.AccessPolicyAllowedIps
				contextLogger.Info(pvcName + ":" + clusterID + " :VPC service endpoints passed: " + vpcServiceEndpoints)
			} else {
				svcEndpointResp, err := providerClient.GetVPCSvcEndpoint(ctx, &provider.VPCSvcEndpointRequest{Id: name})
				if err != nil {
					return nil, controller.ProvisioningFinished, fmt.Errorf(pvcName+":"+clusterID+" :failed to get VPC service endpoints for cluster: %v", err)
				}
				vpcServiceEndpoints = svcEndpointResp.GetCse()
				contextLogger.Info(pvcName + ":" + clusterID + " :fetched VPC service endpoints : " + vpcServiceEndpoints)
				if vpcServiceEndpoints == "" {
					return nil, controller.ProvisioningFinished, errors.New(pvcName + ":" + clusterID + " :cannot set access policy for bucket. VPC service endpoints for the cluster not found")
				}
			}

			setBucketAccessPolicy = true
			updateAP = p.AccessPolicy.NewAccessPolicy()
			rcc = &backend.UpdateAPObj{}

		} else if strings.Contains(providerType, clusterTypeClassic) {
			//add logic to fetch cluster subnet ips
			return nil, controller.ProvisioningFinished, fmt.Errorf(pvcName+":"+clusterID+" : set-access-policy not supported for classic cluster: %v", providerType)
		} else {
			return nil, controller.ProvisioningFinished, fmt.Errorf(pvcName+":"+clusterID+" :set-access-policy not suppoerted on cluster-type: %v", providerType)
		}
	} else {
		if pvc.SetAccessPolicy == "false" {
			contextLogger.Info(pvcName + ":" + clusterID + " bucket :'" + pvc.Bucket + " set-access-policy annotation is set to false for this PVC. bucket access policy will not be set for this PVC")
		}
		if pvc.AccessPolicyAllowedIps != "" && !*ConfigBucketAccessPolicy {
			contextLogger.Info(pvcName + ":" + clusterID + " bucket :'" + pvc.Bucket + " configBucketAccessPolicy is not enabled for this release. bucket access policy will not be set for this PVC")
		}
	}

	if (setBucketAccessPolicy && resConfApiKey == "") || (setQuotaLimit && resConfApiKey == "") {
		return nil, controller.ProvisioningFinished, fmt.Errorf(pvcName+":"+clusterID+": res-conf-apikey missing, cannot set access policy for bucket '%s'", pvc.Bucket)
	}

	if pvc.AutoCreateBucket == "true" {
		var deleteBucket = true
		if pvc.AutoDeleteBucket != "true" && pvc.Bucket == "" { //this handles the cases where AutoDeleteBucket is set false and bucket is not specified.
			id, err := p.UUIDGenerator.New()
			if err != nil {
				return nil, controller.ProvisioningFinished, fmt.Errorf(pvcName+":"+clusterID+":cannot create UUID for bucket name: %v", err)
			}
			pvc.Bucket = autoBucketNamePrefix + id
		}

		if creds.APIKey != "" && creds.ServiceInstanceID == "" {
			return nil, controller.ProvisioningFinished, errors.New(pvcName + ":" + clusterID + " :cannot create bucket using API key without service-instance-id")
		}

		contextLogger.Info(pvcName + ":" + clusterID + " :creating bucket: " + pvc.Bucket)
		if kpRootKeyCrn != "" {
			contextLogger.Info("key protect root key crn provided for bucket" + pvc.Bucket)
		}
		msg, err = sess.CreateBucket(pvc.Bucket, sc.OSStorageClass, kpRootKeyCrn)
		if msg != "" {
			contextLogger.Info(pvcName + ":" + clusterID + " : " + msg)
		}
		// When using existing bucket with auto-create-bucket: true
		if err != nil {
			if strings.Contains(fmt.Sprintf("%v", err), "BucketAlreadyExists") {
				valBucket = true
				deleteBucket = false
				contextLogger.Info(pvcName + ":" + clusterID + " :bucket '" + pvc.Bucket + "' already exists")
			} else {
				return nil, controller.ProvisioningFinished, fmt.Errorf(pvcName+":"+clusterID+" :cannot create bucket %s: %v", pvc.Bucket, err)
			}
		}

		if setBucketAccessPolicy {
			err := updateAP.UpdateAccessPolicy(vpcServiceEndpoints, resConfApiKey, pvc.Bucket, rcc)
			if err != nil {
				//revert bucket creation if updating bucket access policy fails
				if deleteBucket {
					err1 := sess.DeleteBucket(pvc.Bucket)
					if err1 != nil {
						return nil, controller.ProvisioningFinished, fmt.Errorf(pvcName+" : "+clusterID+" :cannot set access policy %v", err1, " and cannot delete bucket %s :  %v", pvc.Bucket, err)
					}
				}
				return nil, controller.ProvisioningFinished, fmt.Errorf(pvcName+" : "+clusterID+" :failed to set access policy for bucket %s : %v", pvc.Bucket, err)
			}
			contextLogger.Info(pvcName + ":" + clusterID + " bucket :'" + pvc.Bucket + "' access policy configured successfully")
		}

		if setQuotaLimit {
			err := updateAP.UpdateQuotaLimit(quotaLimit, resConfApiKey, pvc.Bucket, sc.OSEndpoint, sc.IAMEndpoint, rcc)
			if err != nil {
				//revert bucket creation if updating bucket access policy fails
				if deleteBucket {
					err1 := sess.DeleteBucket(pvc.Bucket)
					if err1 != nil {
						return nil, controller.ProvisioningFinished, fmt.Errorf(pvcName+" : "+clusterID+" :cannot set quota limit %v", err1, " and cannot delete bucket %s :  %v", pvc.Bucket, err)
					}
				}
				return nil, controller.ProvisioningFinished, fmt.Errorf(pvcName+" : "+clusterID+" :failed to set quota limit for bucket %s : %v", pvc.Bucket, err)
			}
			contextLogger.Info(pvcName + ":" + clusterID + " bucket :'" + pvc.Bucket + "' quota limit configured successfully")
		}
	} else {
		if pvc.Bucket == "" {
			return nil, controller.ProvisioningFinished, errors.New(pvcName + ":" + clusterID + " :bucket name not specified")
		}
		// this enables to set access policy for existing bucket
		// when AutoCreateBucket is false, AutoDeleteBucket is false and SetAccessPolicy is true
		if setBucketAccessPolicy {
			err := updateAP.UpdateAccessPolicy(vpcServiceEndpoints, resConfApiKey, pvc.Bucket, rcc)
			if err != nil {
				return nil, controller.ProvisioningFinished, fmt.Errorf(pvcName+" : "+clusterID+" :failed to set access policy for bucket %s : %v", pvc.Bucket, err)
			}
			valBucket = true
			contextLogger.Info(pvcName + ":" + clusterID + " :bucket '" + pvc.Bucket + "' access policy configured successfully")
		}
		if setQuotaLimit {
			err := updateAP.UpdateQuotaLimit(quotaLimit, resConfApiKey, pvc.Bucket, sc.OSEndpoint, sc.IAMEndpoint, rcc)
			if err != nil {
				return nil, controller.ProvisioningFinished, fmt.Errorf(pvcName+" : "+clusterID+" :failed to set quota limit for bucket %s : %v", pvc.Bucket, err)
			}
			contextLogger.Info(pvcName + ":" + clusterID + " bucket :'" + pvc.Bucket + "' quota limit configured successfully")
		}
	}

	if valBucket {
		if err := sess.CheckBucketAccess(pvc.Bucket); err != nil {
			return nil, controller.ProvisioningFinished, fmt.Errorf(pvcName+" : "+clusterID+" :cannot access bucket %s: %v", pvc.Bucket, err)
		}
	}

	if pvc.ObjectPath != "" {
		exist, err := sess.CheckObjectPathExistence(pvc.Bucket, pvc.ObjectPath)
		if err != nil {
			return nil, controller.ProvisioningFinished, fmt.Errorf(pvcName+":"+clusterID+" :cannot access object-path \"%s\" inside bucket %s: %v", pvc.ObjectPath, pvc.Bucket, err)
		} else if !exist {
			return nil, controller.ProvisioningFinished, fmt.Errorf(pvcName+":"+clusterID+" :object-path \"%s\" not found inside bucket %s", pvc.ObjectPath, pvc.Bucket)
		}
	}

	if pvc.UseXattr {
		sc.UseXattr = pvc.UseXattr
	}

	if pvc.DebugLevel != "" {
		sc.DebugLevel = pvc.DebugLevel
	}

	if pvc.CurlDebug {
		sc.CurlDebug = pvc.CurlDebug
	}

	if strings.TrimSpace(pvc.TLSCipherSuite) != "" {
		sc.TLSCipherSuite =
			strings.TrimSpace(pvc.TLSCipherSuite)
	}

	// Check AccessMode
	accessMode := options.PVC.Spec.AccessModes
	contextLogger.Info(pvcName+":"+clusterID+": acccess mode is.. ", zap.Any("access mode", accessMode))
	if len(accessMode) > 1 {
		return nil, controller.ProvisioningFinished, fmt.Errorf(pvcName + ":" + clusterID + ": More that one access mode is not supported.")
	}

	if pvc.AutoCache {
		sc.KernelCache = false
	}

	driverOptions, err := parser.MarshalToMap(&driver.Options{
		ChunkSizeMB:             sc.ChunkSizeMB,
		ParallelCount:           sc.ParallelCount,
		MultiReqMax:             sc.MultiReqMax,
		StatCacheSize:           sc.StatCacheSize,
		TLSCipherSuite:          sc.TLSCipherSuite,
		CurlDebug:               sc.CurlDebug,
		KernelCache:             sc.KernelCache,
		DebugLevel:              sc.DebugLevel,
		S3FSFUSERetryCount:      sc.S3FSFUSERetryCount,
		StatCacheExpireSeconds:  sc.StatCacheExpireSeconds,
		IAMEndpoint:             sc.IAMEndpoint,
		OSEndpoint:              sc.OSEndpoint,
		OSStorageClass:          sc.OSStorageClass,
		Bucket:                  pvc.Bucket,
		ObjectPath:              pvc.ObjectPath,
		ReadwriteTimeoutSeconds: sc.ReadwriteTimeoutSeconds,
		ConnectTimeoutSeconds:   sc.ConnectTimeoutSeconds,
		UseXattr:                sc.UseXattr,
		AccessMode:              string(accessMode[0]),
		CosServiceIP:            svcIp,
		AutoCache:               pvc.AutoCache,
		AddMountParam:           sc.AddMountParam,
	})
	if err != nil {
		return nil, controller.ProvisioningFinished, fmt.Errorf(pvcName+":"+clusterID+":cannot marshal driver options: %v", err)
	}

	pvcAnnots, err := parser.MarshalToMap(&pvcAnnotations{
		AutoCreateBucket:        pvc.AutoCreateBucket,
		AutoDeleteBucket:        pvc.AutoDeleteBucket,
		Bucket:                  pvc.Bucket,
		ObjectPath:              pvc.ObjectPath,
		Endpoint:                pvc.Endpoint,
		Region:                  pvc.Region,
		SecretName:              pvc.SecretName,
		ChunkSizeMB:             pvc.ChunkSizeMB,
		ParallelCount:           pvc.ParallelCount,
		MultiReqMax:             pvc.MultiReqMax,
		StatCacheSize:           pvc.StatCacheSize,
		S3FSFUSERetryCount:      pvc.S3FSFUSERetryCount,
		StatCacheExpireSeconds:  pvc.StatCacheExpireSeconds,
		IAMEndpoint:             pvc.IAMEndpoint,
		ValidateBucket:          pvc.ValidateBucket,
		SecretNamespace:         pvc.SecretNamespace,
		ReadwriteTimeoutSeconds: pvc.ReadwriteTimeoutSeconds,
		ConnectTimeoutSeconds:   pvc.ConnectTimeoutSeconds,
		UseXattr:                pvc.UseXattr,
		CurlDebug:               pvc.CurlDebug,
		DebugLevel:              pvc.DebugLevel,
		CosServiceName:          pvc.CosServiceName,
		SetAccessPolicy:         pvc.SetAccessPolicy,
		AddMountParam:           pvc.AddMountParam,
	})

	if err != nil {
		return nil, controller.ProvisioningFinished, fmt.Errorf(pvcName+":"+clusterID+":cannot marshal pv options: %v", err)
	}

	reclaimPolicy := options.StorageClass.ReclaimPolicy
	return &v1.PersistentVolume{
		ObjectMeta: metav1.ObjectMeta{
			Name:        options.PVName,
			Annotations: pvcAnnots,
		},
		Spec: v1.PersistentVolumeSpec{
			PersistentVolumeReclaimPolicy: *reclaimPolicy,
			AccessModes:                   options.PVC.Spec.AccessModes,
			Capacity: v1.ResourceList{
				v1.ResourceStorage: options.PVC.Spec.Resources.Requests[v1.ResourceName(v1.ResourceStorage)],
			},
			PersistentVolumeSource: v1.PersistentVolumeSource{
				FlexVolume: &v1.FlexPersistentVolumeSource{
					Driver:    driverName,
					FSType:    fsType,
					SecretRef: &v1.SecretReference{Name: pvc.SecretName, Namespace: pvc.SecretNamespace},
					ReadOnly:  false,
					Options:   driverOptions,
				},
			},
		},
	}, controller.ProvisioningFinished, nil
}

// Delete deletes a persistent volume
func (p *IBMS3fsProvisioner) Delete(ctx context.Context, pv *v1.PersistentVolume) error {
	var pvcAnnots pvcAnnotations

	contextLogger, _ := logger.GetZapDefaultContextLogger()
	contextLogger.Info("Deleting the pvc..")

	endpointValue := pv.Spec.PersistentVolumeSource.FlexVolume.Options["object-store-endpoint"]
	regionValue := pv.Spec.PersistentVolumeSource.FlexVolume.Options["object-store-storage-class"]
	iamEndpoint := pv.Spec.PersistentVolumeSource.FlexVolume.Options["iam-endpoint"]

	err := parser.UnmarshalMap(&pv.Annotations, &pvcAnnots)
	if err != nil {
		return fmt.Errorf("cannot unmarshal PV annotations: %v", err)
	}

	if pvcAnnots.AutoDeleteBucket == "true" {
		if err = p.deleteBucket(ctx, &pvcAnnots, endpointValue, regionValue, iamEndpoint); err != nil {
			return fmt.Errorf("cannot delete bucket: %v", err)
		}
	} else if _, err = strconv.ParseBool(pvcAnnots.AutoDeleteBucket); err != nil {
		return fmt.Errorf("invalid value for auto-delete-bucket, expects true/false: %v", err)
	}
	return nil
}

func (p *IBMS3fsProvisioner) deleteBucket(ctx context.Context, pvcAnnots *pvcAnnotations, endpointValue, regionValue, iamEndpoint string) error {
	contextLogger, _ := logger.GetZapDefaultContextLogger()
	contextLogger.Info("Deleting the bucket..")
	// Retrieve CA Cert if provided in secert
	if err := p.writeCrtFile(ctx, pvcAnnots.SecretName, pvcAnnots.SecretNamespace, pvcAnnots.CosServiceName); err != nil {
		return fmt.Errorf("cannot retrieve secret: %v", err)
	}

	creds, _, _, _, err := p.getCredentials(ctx, pvcAnnots.SecretName, pvcAnnots.SecretNamespace)
	if err != nil {
		return fmt.Errorf("cannot get credentials: %v", err)
	}
	creds.IAMEndpoint = iamEndpoint
	sess := p.Backend.NewObjectStorageSession(endpointValue, regionValue, creds, p.Logger)

	return sess.DeleteBucket(pvcAnnots.Bucket)
}
