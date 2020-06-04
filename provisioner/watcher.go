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
	"time"

	"go.uber.org/zap"
	v1 "k8s.io/api/core/v1"
	apiv1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

var lgr zap.Logger

// WatchPersistentVolumes ...
func WatchPersistentVolumes(client kubernetes.Interface, log zap.Logger) {
	lgr = log
	watchlist := cache.NewListWatchFromClient(client.Core().RESTClient(), "persistentvolumes", apiv1.NamespaceAll, fields.Everything())
	_, controller := cache.NewInformer(watchlist, &v1.PersistentVolume{}, time.Second*0,
		cache.ResourceEventHandlerFuncs{
			AddFunc:    ValidatePersistentVolume,
			DeleteFunc: ValidatePersistentVolume,
			UpdateFunc: nil,
		},
	)
	stopch := wait.NeverStop
	go controller.Run(stopch)
	lgr.Info("WatchPersistentVolume")
	<-stopch
}

func ValidatePersistentVolume(obj interface{}) {
	lgr.Info("Validate of persistent volume is successful", zap.Reflect("persistentvolume", obj))
}
