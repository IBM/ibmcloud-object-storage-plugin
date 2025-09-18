/*******************************************************************************
 * IBM Confidential
 * OCO Source Materials
 * IBM Cloud Kubernetes Service, 5737-D43
 * (C) Copyright IBM Corp. 2017, 2025 All Rights Reserved.
 * The source code for this program is not published or otherwise divested of
 * its trade secrets, irrespective of what has been deposited with
 * the U.S. Copyright Office.
 ******************************************************************************/

package config

import (
	"github.com/IBM/ibmcloud-object-storage-plugin/utils/consts"
	"github.com/IBM/ibmcloud-object-storage-plugin/utils/logger"
	"github.com/stretchr/testify/assert"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kfakes "k8s.io/client-go/kubernetes/fake"
	"os"
	"path/filepath"
	"testing"
)

type testConfig struct {
	Header sectionTestConfig
}

type sectionTestConfig struct {
	ID      int
	Name    string
	YesOrNo bool
	Pi      float64
	List    string
}

var testConf = testConfig{
	Header: sectionTestConfig{
		ID:      1,
		Name:    "test",
		YesOrNo: true,
		Pi:      3.14,
		List:    "1, 2",
	},
}

var testLogger, _ = logger.GetZapLogger()

func TestParseConfig(t *testing.T) {
	t.Log("Testing config parsing")
	var testParseConf testConfig

	configPath := filepath.Join("..", "test", "test.toml")
	ParseConfig(configPath, &testParseConf, *testLogger)

	expected := testConf

	assert.Exactly(t, expected, testParseConf)
}

func TestParseConfigNoMatch(t *testing.T) {
	t.Log("Testing config parsing false positive")
	var testParseConf testConfig

	configPath := filepath.Join("..", "test", "test.toml")
	ParseConfig(configPath, &testParseConf, *testLogger)

	expected := testConfig{
		Header: sectionTestConfig{
			ID:      1,
			Name:    "testnomatch",
			YesOrNo: true,
			Pi:      3.14,
			List:    "1, 2",
		}}

	assert.NotEqual(t, expected, testParseConf)

}

func TestGetConfigStringNoEnv(t *testing.T) {
	t.Log("Testing string config value get when there is no env var override")
	confVal := GetConfigString("name", testConf.Header.Name)

	expected := "test"

	assert.Equal(t, expected, confVal)
}

func TestGetConfigStringWithEnv(t *testing.T) {
	t.Log("Testing string config value get when there is an env var override")
	_ = os.Setenv("NAME", "env")

	confVal := GetConfigString("name", testConf.Header.Name)
	_ = os.Unsetenv("NAME")

	expected := "env"
	assert.Equal(t, expected, confVal)
}

func TestGetConfigIntNoEnv(t *testing.T) {

	t.Log("Testing int config value get when there is no env var override")
	confVal := GetConfigInt("id", testConf.Header.ID, *testLogger)

	expected := 1
	assert.Equal(t, expected, confVal)
}

func TestGetConfigIntWithEnv(t *testing.T) {

	t.Log("Testing int config value get when there is an env var override")
	_ = os.Setenv("ID", "10")

	confVal := GetConfigInt("id", testConf.Header.ID, *testLogger)
	_ = os.Unsetenv("ID")

	expected := 10
	assert.Equal(t, expected, confVal)
}

func TestGetConfigBoolNoEnv(t *testing.T) {
	t.Log("Testing bool config value get when there is no env var override")
	confVal := GetConfigBool("yesOrNo", testConf.Header.YesOrNo, *testLogger)

	expected := true
	assert.Equal(t, expected, confVal)
}

func TestGetConfigBoolWithEnv(t *testing.T) {
	t.Log("Testing bool config value get when there is an env var override")
	_ = os.Setenv("YESORNO", "false")

	confVal := GetConfigBool("yesOrNo", testConf.Header.YesOrNo, *testLogger)
	_ = os.Unsetenv("YESORNO")

	expected := false
	assert.Equal(t, expected, confVal)
}

func TestGetConfigStringListNoEnv(t *testing.T) {
	t.Log("Testing string list config value get when there is no env var override")
	confVal := GetConfigStringList("list", testConf.Header.List, *testLogger)

	expected := []string{"1", "2"}

	assert.Exactly(t, expected, confVal)
}

func TestGetConfigStringListWithEnv(t *testing.T) {
	t.Log("Testing string list config value get when there is an env var override")
	_ = os.Setenv("LIST", "1,2,3")

	confVal := GetConfigStringList("list", testConf.Header.List, *testLogger)
	_ = os.Unsetenv("LIST")

	expected := []string{"1", "2", "3"}

	assert.Exactly(t, expected, confVal)
}

func TestGetGoPath(t *testing.T) {
	t.Log("Testing getting GOPATH")
	goPath := "/tmp"
	_ = os.Setenv("GOPATH", goPath)

	path := GetGoPath()

	assert.Equal(t, goPath, path)
}

func TestGetEnv(t *testing.T) {
	t.Log("Testing getting ENV")
	goPath := "/tmp"
	_ = os.Setenv("ENVTEST", goPath)

	path := getEnv("ENVTEST")

	assert.Equal(t, goPath, path)
}

func TestGetGoPathNullPath(t *testing.T) {
	t.Log("Testing getting GOPATH NULL Path")
	goPath := ""
	_ = os.Setenv("GOPATH", goPath)

	path := GetGoPath()

	assert.Equal(t, goPath, path)
}

func TestPassSetEnv(t *testing.T) {
	t.Log("Testing SetUpEvn() for happy path")
	clusterconfigmap := &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "cluster-info",
			Namespace: consts.KubeSystem,
		},
		Data: map[string]string{
			"cluster-config.json": "{\"cluster_id\": \"de3daf0f942446a8b8c38e68a14c607a\", \"cluster_name\": \"stage-dal09-de3daf0f942446a8b8c38e68a14c607a\", \"datacenter\": \"dal10\", \"account_id\": \"fd1611c9a44144d7c2b944234b6bb40e\"}",
		},
	}
	crnconfigmap := &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "crn-info-ibmc",
			Namespace: consts.KubeSystem,
		},
		Data: map[string]string{
			"CRN_CNAME":       "bluemix",
			"CRN_CTYPE":       "public",
			"CRN_REGION":      "us-south",
			"CRN_SERVICENAME": "containers-kubernetes",
			"CRN_VERSION":     "v1",
		},
	}
	kubeclient := kfakes.NewSimpleClientset(clusterconfigmap, crnconfigmap)
	err := SetUpEvn(kubeclient, testLogger)
	assert.Nil(t, err)
}

func TestPassAlreadySetEnv(t *testing.T) {
	t.Log("Testing SetUpEvn() cluster config already exported")

	kubeclient := kfakes.NewSimpleClientset()
	err := SetUpEvn(kubeclient, testLogger)
	assert.Nil(t, err)
}

func TestCmNotFoundSetEnv(t *testing.T) {
	_ = os.Unsetenv("CLUSTER_ID")
	t.Log("Testing SetUpEvn() for CM not found")

	kubeclient := kfakes.NewSimpleClientset()
	err := SetUpEvn(kubeclient, testLogger)
	assert.Nil(t, err)
}

func TestCmErrorSetEnv(t *testing.T) {
	_ = os.Unsetenv("CLUSTER_ID")
	t.Log("Testing SetUpEvn() for CM  wrong content")
	configmap := &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "cluster-info",
			Namespace: consts.KubeSystem,
		},
		Data: map[string]string{
			"cluster-config.json": "{\"wrong\": \"de3daf0f942446a8b8c38e68a14c607a\", \"cluster_name\": \"stage-dal09-de3daf0f942446a8b8c38e68a14c607a\", \"datacenter\": \"dal10\", \"account_id\": \"fd1611c9a44144d7c2b944234b6bb40e\"}",
		},
	}
	kubeclient := kfakes.NewSimpleClientset(configmap)
	err := SetUpEvn(kubeclient, testLogger)
	assert.Error(t, err)
}

func TestCmErrorParsing(t *testing.T) {
	_ = os.Unsetenv("CLUSTER_ID")
	t.Log("Testing SetUpEvn() for CM  wrong content")
	configmap := &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "cluster-info",
			Namespace: consts.KubeSystem,
		},
		Data: map[string]string{
			"cluster-config.json": "{\"wrong\" \"de3daf0f942446a8b8c38e68a14c607a\" \"cluster_name\": \"stage-dal09-de3daf0f942446a8b8c38e68a14c607a\", \"datacenter\": \"dal10\", \"account_id\": \"fd1611c9a44144d7c2b944234b6bb40e\"}",
		},
	}
	kubeclient := kfakes.NewSimpleClientset(configmap)
	err := SetUpEvn(kubeclient, testLogger)
	assert.Error(t, err)
}
