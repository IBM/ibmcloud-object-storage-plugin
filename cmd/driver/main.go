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
	"encoding/json"
	"fmt"
	"github.com/IBM/ibmcloud-object-storage-plugin/driver"
	"github.com/IBM/ibmcloud-object-storage-plugin/driver/interfaces"
	"github.com/IBM/ibmcloud-object-storage-plugin/utils/backend"
	optParser "github.com/IBM/ibmcloud-object-storage-plugin/utils/parser"
	flags "github.com/jessevdk/go-flags"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

const (
	logConfig = "/var/log/ibmc-s3fs.log"
)

// Version and Build time will be set during the "make driver"
// go build -ldflags  "-X main.Version=${VERSION} -X main.Build=${BUILD}" -o ${BINARY_NAME}
// go build -ldflags "-X main.Version=0.0.1 -X main.Build=2017-08-06T00:33:58+0530" cmd/driver/main.go
// VERSION=`git describe --tags`
// BUILD=`date +%FT%T%z`

// Version holds the driver version string
var Version string

// Build holds the driver build string
var Build string

var filelogger = getZapLogger()

// Save the io streams, before we divert them to File.
// this will be used used while printing the response
var stdout = os.Stdout

// NewS3fsPlugin returns a new instance of the driver that supports mount & unmount operations of s3fs volumes
func NewS3fsPlugin(logger *zap.Logger) *driver.S3fsPlugin {
	return &driver.S3fsPlugin{
		Backend: &backend.COSSessionFactory{},
		Logger:  logger,
	}
}

type versionCommand struct{}

func (v *versionCommand) Execute(args []string) error {
	fmt.Fprintf(stdout, "Version:%s, Build:%s\n", Version, Build)
	return nil
}

type initCommand struct{}

func (i *initCommand) Execute(args []string) error {
	response := NewS3fsPlugin(filelogger).Init()
	filelogger.Info(":S3FS Driver info:", zap.String("Version", Version), zap.String("Build", Build))
	return printResponse(response)
}

type PodDetail struct {
	PodUid  string `json:"kubernetes.io/pod.uid,omitempty"`
	PodName string `json:"kubernetes.io/pod.name,omitempty"`
	PodNS   string `json:"kubernetes.io/pod.namespace,omitempty"`
}

type mountCommand struct{}

func maskSecrets(m map[string]string) (map[string]string, []byte, error) {
	mountOptsLogs := make(map[string]string)
	for k, v := range m {
		mountOptsLogs[k] = v
	}

	mountOptsLogs["kubernetes.io/secret/access-key"] = "XXX"
	mountOptsLogs["kubernetes.io/secret/secret-key"] = "YYY"
	mountOptsLogs["kubernetes.io/secret/api-key"] = "KKK"
	mountOptsLogs["kubernetes.io/secret/service-instance-id"] = "MMM"
	mountOptsLogs["kubernetes.io/secret/ca-bundle-crt"] = "ZZZ"
	newString, err := json.Marshal(mountOptsLogs)

	return mountOptsLogs, newString, err
}
func (m *mountCommand) Execute(args []string) error {
	var err error
	var podDetail PodDetail
	var podUID = ""
	var hostname, anyerror = os.Hostname()
	if anyerror != nil {
		hostname = ""
	}

	filelogger.Info(":MountCommand start:" + hostname)

	mountOpts := make(map[string]string)
	mountOptsLogs := make(map[string]string)

	switch len(args) {
	case 2:
		// Kubernetes 1.6+
		err = json.Unmarshal([]byte(args[1]), &mountOpts)

	case 3:
		// Kubernetes 1.5-
		err = json.Unmarshal([]byte(args[2]), &mountOpts)
	default:

		return printResponse(interfaces.FlexVolumeResponse{
			Status:  interfaces.StatusFailure,
			Message: fmt.Sprintf("Unexpected number of arguments to 'mount' command: %d", len(args)),
		})
	}

	err = optParser.UnmarshalMap(&mountOpts, &podDetail)
	if err == nil {
		podUID = podDetail.PodUid
		filelogger.Info(podDetail.PodUid + ":" + podDetail.PodName + ":" +
			podDetail.PodNS + ":" + hostname)
	}

	mountOptsLogs, newString, err := maskSecrets(mountOpts)

	filelogger.Info(podUID+":MountCommand args", zap.ByteString("input args", newString))

	targetMountDir := args[0]

	if err != nil {
		return printResponse(interfaces.FlexVolumeResponse{
			Status:  interfaces.StatusFailure,
			Message: fmt.Sprintf("Failed to mount volume to %s due to: %#v", targetMountDir, err),
		})
	}

	filelogger.Info(podUID+":cmd-MountCommand  ", zap.Any("mountOpts", mountOptsLogs))

	// For logs
	mountRequestLog := interfaces.FlexVolumeMountRequest{
		MountDir: targetMountDir,
		Opts:     mountOptsLogs,
	}

	mountRequest := interfaces.FlexVolumeMountRequest{
		MountDir: targetMountDir,
		Opts:     mountOpts,
	}
	filelogger.Info(podUID+":cmd-MountCommand ", zap.Any("mountRequest", mountRequestLog))
	//response := NewS3fsPlugin(filelogger).Mount(mountRequest)
	s3fsPlugin := NewS3fsPlugin(filelogger)
	driver.SetBuildVersion(Version)
	driver.SetPodUID(podUID)
	response := (*s3fsPlugin).Mount(mountRequest)
	filelogger.Info(podUID+":MountCommand End", zap.Any("response", response))
	return printResponse(response)
}

type unmountCommand struct{}

func (u *unmountCommand) Execute(args []string) error {
	var hostname, anyerror = os.Hostname()
	if anyerror != nil {
		hostname = ""
	}
	filelogger.Info(":UnmountCommand start:" + hostname)
	filelogger.Info(":UnmountCommand args", zap.Strings("input args", args))

	if len(args) != 1 {
		return printResponse(interfaces.FlexVolumeResponse{
			Status:  interfaces.StatusFailure,
			Message: fmt.Sprintf("Unexpected number of arguments to 'unmount' command: %d", len(args)),
		})
	}

	mountDir := args[0]
	unmountRequest := interfaces.FlexVolumeUnmountRequest{
		MountDir: mountDir,
	}
	response := NewS3fsPlugin(filelogger).Unmount(unmountRequest)
	filelogger.Info(":UnmountCommand end", zap.Reflect("response", response))
	return printResponse(response)
}

type flagsOptions struct{}

func main() {
	var err error
	var versionCommand versionCommand
	var initCommand initCommand
	var mountCommand mountCommand
	var unmountCommand unmountCommand
	var options flagsOptions
	var parser = flags.NewParser(&options, flags.Default&^flags.PrintErrors)

	// disable the console logging (if anywhere else being done by softlayer or any other pkg)
	// presently softlayer logs few warning message, which makes the flexdriver unmarshall failure
	log.SetFlags(0)
	log.SetOutput(ioutil.Discard)

	// Divert all loggers outputs and fmt.printf loggings (this will create issues with flex response)
	NullDevice, _ := os.Open(os.DevNull)
	os.Stdout = NullDevice
	os.Stderr = NullDevice

	parser.AddCommand("version",
		"Prints version",
		"Prints version and build information",
		&versionCommand)
	parser.AddCommand("init",
		"Init the plugin",
		"The info command print the driver name and version.",
		&initCommand)
	parser.AddCommand("mount",
		"Mount Volume",
		"Mount a volume Id to a path - returning the path.",
		&mountCommand)
	parser.AddCommand("unmount",
		"Unmount Volume",
		"UnMount given a mount dir",
		&unmountCommand)

	_, err = parser.Parse()
	if err != nil {
		var status string
		if strings.Contains(strings.ToLower(err.Error()), "unknown command") {
			status = interfaces.StatusNotSupported
		} else {
			status = interfaces.StatusFailure
		}
		printResponse(interfaces.FlexVolumeResponse{
			Status:  status,
			Message: fmt.Sprintf("Error parsing arguments: %v", err),
		})
	}
}

func getZapLogger() *zap.Logger {
	logfilepath := getFromEnv("LOGCONFIG", logConfig)

	// Configure log rotate
	lumberjackLogger := &lumberjack.Logger{
		Filename:   logfilepath,
		MaxSize:    100, //MB
		MaxBackups: 10,  //Maximum number of backup
		MaxAge:     60,  //Days
	}
	//defer lumberjackLogger.Close()

	//Create json encoder
	prodConf := zap.NewProductionEncoderConfig()
	prodConf.EncodeTime = zapcore.ISO8601TimeEncoder
	encoder := zapcore.NewJSONEncoder(prodConf)

	//create sync, where zap writes the output
	zapsync := zapcore.AddSync(lumberjackLogger)

	//Default Log level
	loglevel := zap.NewAtomicLevelAt(zapcore.InfoLevel)

	//zapcore
	loggercore := zapcore.NewCore(encoder, zapsync, loglevel)

	//Zap Logger with Log rotation
	logger := zap.New(loggercore)
	logger.Named("S3FSDriver")

	return logger
}

func printResponse(f interfaces.FlexVolumeResponse) error {
	responseBytes, err := json.Marshal(f)
	if err != nil {
		return err
	}
	output := string(responseBytes[:])

	// log output being returned to flex driver
	filelogger.Info(":FlexVolumeResponse", zap.String("output", output))

	// write it to stdout, so that flexdriver will read it
	fmt.Fprintf(stdout, "%s", output)
	return nil
}

func getFromEnv(key string, defaultVal string) string {
	value := os.Getenv(key)
	if value == "" {
		value = defaultVal
	}
	return value
}
