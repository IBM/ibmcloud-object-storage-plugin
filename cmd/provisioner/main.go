/*******************************************************************************
 * IBM Confidential
 * OCO Source Materials
 * IBM Cloud Container Service, 5737-D43
 * (C) Copyright IBM Corp. 2017, 2018 All Rights Reserved.
 * The source code for this program is not  published or otherwise divested of
 * its trade secrets, irrespective of what has been deposited with
 * the U.S. Copyright Office.
 ******************************************************************************/

package main

import (
	"context"
	"flag"
	ibmprovider "github.com/IBM/ibmcloud-object-storage-plugin/ibm-provider/provider"
	s3fsprovisioner "github.com/IBM/ibmcloud-object-storage-plugin/provisioner"
	"github.com/IBM/ibmcloud-object-storage-plugin/utils/backend"
	cfg "github.com/IBM/ibmcloud-object-storage-plugin/utils/config"
	grpcClient "github.com/IBM/ibmcloud-object-storage-plugin/utils/grpc-client"
	log "github.com/IBM/ibmcloud-object-storage-plugin/utils/logger"
	"github.com/IBM/ibmcloud-object-storage-plugin/utils/uuid"
	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/util/validation"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/sig-storage-lib-external-provisioner/v6/controller"
	"strings"
	"time"
)

const (
	failedRetryThreshold = 1
	resyncPeriod         = 30 * time.Second
)

var provisioner = flag.String(
	"provisioner",
	"ibm.io/ibmc-s3fs",
	"Name of the provisioner. The provisioner will only provision volumes for claims that request a StorageClass with a provisioner field set equal to this name.")

var master = flag.String(
	"master",
	"",
	"Master URL to build a client config from. Either this or kubeconfig needs to be set if the provisioner is being run out of cluster.",
)
var kubeconfig = flag.String(
	"kubeconfig",
	"",
	"Absolute path to the kubeconfig file. Either this or master needs to be set if the provisioner is being run out of cluster.",
)

var leaseDuration = flag.Duration(
	"leaseDuration",
	15*time.Second,
	"Duration of the lease on a persistent volume",
)

var leaseRenewDeadline = flag.Duration(
	"leaseRenewDeadline",
	10*time.Second,
	"Lease renewal deadline",
)

var leaseRetryPeriod = flag.Duration(
	"leaseRetryPeriod",
	2*time.Second,
	"How often lease acquisition and renewal should be retried",
)

var leaseTermLimit = flag.Duration(
	"leaseTermLimit",
	10*time.Minute,
	"Maximum time that a provisioner can maintain a lease",
)

func main() {
	var err error
	logger, _ := log.GetZapLogger()
	loggerLevel := zap.NewAtomicLevel()
	err = flag.Set("logtostderr", "true")
	if err != nil {
		logger.Info("Failed to set flag:", zap.Error(err))
	}

	s3fsprovisioner.SockEndpoint = flag.String(
		"endpoint",
		"/ibmprovider/provider.sock",
		"Provider endpoint",
	)

	s3fsprovisioner.ConfigBucketAccessPolicy = flag.Bool(
		"bucketAccessPolicy",
		false,
		"set 'true' to configure bucket access policy",
	)

	s3fsprovisioner.ConfigQuotaLimit = flag.Bool(
		"quotaLimit",
		false,
		"set 'true' to configure bucket quota limit",
	)

	s3fsprovisioner.AllowCrossNsSecret = flag.Bool(
		"allowCrossNsSecret",
		true,
		"set to 'false' to disable COS secret lookup in namespace other than PVC's namespace",
	)

	flag.Parse()

	// Enable debug trace
	debugTrace := cfg.GetConfigBool("DEBUG_TRACE", false, *logger)
	if debugTrace {
		loggerLevel.SetLevel(zap.DebugLevel)
	}

	if errs := validateProvisioner(*provisioner, field.NewPath("provisioner")); len(errs) != 0 {
		errMsgs := ""
		for _, err := range errs {
			errMsgs = errMsgs + err.Error() + "\n"
		}
		logger.Fatal("Invalid provisioner specified", zap.String("validation_errors", errMsgs))
	}
	logger.Info("Provisioner specified: ", zap.String("provisioner", *provisioner))

	var config *rest.Config
	config, err = clientcmd.BuildConfigFromFlags(*master, *kubeconfig)
	if err != nil {
		logger.Fatal("Failed to create config:", zap.Error(err))
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		logger.Fatal("Failed to create client:", zap.Error(err))
	}

	err = cfg.SetUpEvn(clientset, logger)
	if err != nil {
		logger.Fatal("Error while loading the ENV variables", zap.Error(err))
	}

	serverVersion, err := clientset.Discovery().ServerVersion()
	if err != nil {
		logger.Fatal("Error getting server version:", zap.Error(err))
	}

	s3fsProvisioner := &s3fsprovisioner.IBMS3fsProvisioner{
		Backend:       &backend.COSSessionFactory{},
		GRPCBackend:   &grpcClient.ConnObjFactory{},
		AccessPolicy:  &backend.UpdateAPFactory{},
		IBMProvider:   &ibmprovider.IBMProviderClntFactory{},
		Logger:        logger,
		Client:        clientset,
		UUIDGenerator: uuid.NewCryptoGenerator(),
	}

	pc := controller.NewProvisionController(
		clientset,
		*provisioner,
		s3fsProvisioner,
		serverVersion.GitVersion,
		controller.LeaderElection(false),
		controller.ResyncPeriod(resyncPeriod),
		controller.ExponentialBackOffOnError(true),
		controller.FailedProvisionThreshold(failedRetryThreshold),
		controller.LeaseDuration(*leaseDuration),
		controller.RenewDeadline(*leaseRenewDeadline),
		controller.RetryPeriod(*leaseRetryPeriod),
		//controller.TermLimit(*leaseTermLimit),
	)

	pc.Run(context.Background())
}

// validateProvisioner tests if provisioner is a valid qualified name.
func validateProvisioner(provisioner string, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}
	if len(provisioner) == 0 {
		allErrs = append(allErrs, field.Required(fldPath, provisioner))
	}
	if len(provisioner) > 0 {
		for _, msg := range validation.IsQualifiedName(strings.ToLower(provisioner)) {
			allErrs = append(allErrs, field.Invalid(fldPath, provisioner, msg))
		}
	}
	return allErrs
}
