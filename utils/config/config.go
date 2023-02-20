/*******************************************************************************
 * IBM Confidential
 * OCO Source Materials
 * IBM Cloud Container Service, 5737-D43
 * (C) Copyright IBM Corp. 2017, 2018 All Rights Reserved.
 * The source code for this program is not  published or otherwise divested of
 * its trade secrets, irrespective of what has been deposited with
 * the U.S. Copyright Office.
 ******************************************************************************/

package config

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/IBM/ibmcloud-object-storage-plugin/utils/consts"
	"go.uber.org/zap"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"os"
	"strconv"
	"strings"
	"time"
)

// ClusterInfo ...
type ClusterInfo struct {
	ClusterID   string `json:"cluster_id"`
	ClusterName string `json:"cluster_name,omitempty"`
	DataCenter  string `json:"datacenter,omitempty"`
	CustomerID  string `json:"customer_id,omitempty"`
}

func getEnv(key string) string {
	return os.Getenv(strings.ToUpper(key))
}

/* #nosec */
func setEnv(key string, value string) {
	os.Setenv(strings.ToUpper(key), value)
}

// GetGoPath ...
func GetGoPath() string {
	if goPath := getEnv("GOPATH"); goPath != "" {
		return goPath
	}
	return ""
}

// ParseConfig ...
func ParseConfig(filePath string, conf interface{}, logger zap.Logger) {
	if _, err := toml.DecodeFile(filePath, conf); err != nil {
		logger.Fatal("error parsing config file", zap.Error(err))
	}
}

// GetConfigString ...
func GetConfigString(envKey, defaultConf string) string {
	if val := getEnv(envKey); val != "" {
		return val
	}
	return defaultConf
}

// GetConfigInt ...
func GetConfigInt(envKey string, defaulfConf int, logger zap.Logger) int {
	if val := getEnv(envKey); val != "" {
		if envInt, err := strconv.Atoi(val); err == nil {
			return envInt
		}
		logger.Error("error parsing env val to int", zap.String("env", envKey))
	}
	return defaulfConf
}

// GetConfigBool ...
func GetConfigBool(envKey string, defaultConf bool, logger zap.Logger) bool {
	if val := getEnv(envKey); val != "" {
		if envBool, err := strconv.ParseBool(val); err == nil {
			return envBool
		}
		logger.Error("error parsing env val to bool", zap.String("env", envKey))
	}
	return defaultConf
}

// GetConfigStringList ...
func GetConfigStringList(envKey string, defaultConf string, logger zap.Logger) []string {
	// Assume env var is a list of strings separated by ','
	val := defaultConf

	if getEnv(envKey) != "" {
		val = getEnv(envKey)
	}

	val = strings.Replace(val, " ", "", -1)
	return strings.Split(val, ",")
}

// SetUpEvn ... Export the configmap (eg. cluster-info) to environment variables
func SetUpEvn(kubeclient kubernetes.Interface, logger *zap.Logger) error {
	logger.Info("Entry SetUpEvn")

	//Read cluster meta info
	err := LoadClusterInfoMap(kubeclient, logger)
	if err != nil {
		return err
	}

	logger.Info("Exit SetUpEvn")
	return err
}

// LoadClusterInfoMap ... Read cluster metadata from 'cluster-info' map and load into ENV
func LoadClusterInfoMap(kubeclient kubernetes.Interface, logger *zap.Logger) error {
	logger.Debug("Entry LoadClusterInfoMap")

	//check if the ENV variable already loaded
	clusterid := getEnv("cluster_id")
	if len(clusterid) > 0 {
		logger.Info("Exit LoadClusterInfoMap, cluster_id already set", zap.String("cluster_id", clusterid))
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	// export cluster-info config map
	cmClusterInfo, err := kubeclient.CoreV1().ConfigMaps(consts.KubeSystem).Get(ctx, consts.ClusterInfo, metav1.GetOptions{})
	if err != nil {
		err = fmt.Errorf("Unable to find the config map %s. Error: %v.Setting dummy values", consts.ClusterInfo, err)

		setEnv("CLUSTER_ID", "dummyClusterID")
		setEnv("CLUSTER_NAME", "dummyClusterName")
		setEnv("DATACENTER", "dummyDC")
		setEnv("CUSTOMER_ID", "dummyCustomerID")
		return nil
	}

	logger.Debug("configmap details", zap.Reflect(consts.ClusterInfo, cmClusterInfo))
	clusterInfoData := cmClusterInfo.Data[consts.ClusterInfoData]
	clusteInfo := ClusterInfo{}
	err = json.Unmarshal([]byte(clusterInfoData), &clusteInfo)
	if err != nil {
		err = fmt.Errorf("Error while parsing cluster-config %s. Error: %v", consts.ClusterInfo, err)
		return err
	}

	logger.Info("Exporting cluster-config", zap.Reflect(consts.ClusterInfo, clusteInfo))
	if clusteInfo.ClusterID == "" {
		err = fmt.Errorf("cluster_id is not found in map %s", consts.ClusterInfo)
		return err
	}
	setEnv("CLUSTER_ID", clusteInfo.ClusterID)
	setEnv("CLUSTER_NAME", clusteInfo.ClusterName)
	setEnv("DATACENTER", clusteInfo.DataCenter)
	setEnv("CUSTOMER_ID", clusteInfo.CustomerID)
	logger.Debug("Exit LoadClusterInfoMap")
	return nil
}
