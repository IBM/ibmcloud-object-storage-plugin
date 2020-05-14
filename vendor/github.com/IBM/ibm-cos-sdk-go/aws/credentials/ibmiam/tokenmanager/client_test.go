package tokenmanager

import (
	"fmt"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/IBM/ibm-cos-sdk-go/aws"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper function to build a config for multiple tests
// Parameters:
//		Max Retries = Number of retries to run client app
// 		Log Level = type of log level to track
//		Logger = Logger application to use for tests
// Returns:
//		AWS Config with the client, retries and logger and its log level
func buildConfig(maxRetries int, logLevel aws.LogLevelType, logger aws.Logger) *aws.Config {
	httpClient := http.DefaultClient
	return &aws.Config{
		HTTPClient: httpClient,
		MaxRetries: aws.Int(maxRetries),
		LogLevel:   aws.LogLevel(logLevel),
		Logger:     logger,
	}
}

// Helper function to build a request for multiple tests
// Parameter:
//		URL string to pass into a request
// Returns:
//		A built Request object
func buildTestRequest(t *testing.T, URL string) *http.Request {
	req, err := http.NewRequest(http.MethodPost, URL, strings.NewReader("hello"))
	require.Nil(t, err, errorBuildingRequest)
	req.Header.Add("h1", "v1")
	return req
}

func TestNewIBMClient_ConfigAndValuesPersisted_OnCreation(t *testing.T) {

	// Set variables for the test
	httpClient := http.DefaultClient
	maxRetries := 7
	logLevel := aws.LogDebug | aws.LogDebugWithRequestRetries | aws.LogDebugWithRequestErrors
	logger := aws.NewDefaultLogger()

	// Build a config
	config := buildConfig(maxRetries, logLevel, logger)

	// Begins the initial backoff
	initialBackOff := time.Duration(7)
	backProgression := func(in time.Duration) time.Duration { return in }

	// Assertions
	client := NewIBMClient(config, initialBackOff, backProgression)
	assert.Equal(t, httpClient, client.Client, httpClientNotMatch)
	assert.Equal(t, maxRetries, client.MaxRetries, maxRetriesNotMatch)
	assert.Equal(t, logLevel, *client.logLevel, logLevelNotMatch)
	assert.Equal(t, logger, client.logger, loggerNotMatch)
	assert.Equal(t, initialBackOff, client.InitialBackOff, "Initial BackOff did not match")

	// golang do not compare functions
	// work around, compare the results of call the function 3 times
	in := time.Duration(rand.Int63())
	assert.Equal(t, backProgression(in), client.BackOffProgression(in), backoffProgressionNotMatch)
	in = time.Duration(rand.Int63())
	assert.Equal(t, backProgression(in), client.BackOffProgression(in), backoffProgressionNotMatch)
	in = time.Duration(rand.Int63())
	assert.Equal(t, backProgression(in), client.BackOffProgression(in), backoffProgressionNotMatch)
}

func TestDefaultIBMClient_ConfigPersisted_OnCreation(t *testing.T) {

	// Set variables for the test
	httpClient := http.DefaultClient
	maxRetries := 7
	logLevel := aws.LogDebug | aws.LogDebugWithRequestRetries | aws.LogDebugWithRequestErrors
	logger := aws.NewDefaultLogger()

	// Build a config
	config := buildConfig(maxRetries, logLevel, logger)

	// Assertions
	client := DefaultIBMClient(config)
	assert.Equal(t, httpClient, client.Client, httpClientNotMatch)
	assert.Equal(t, maxRetries, client.MaxRetries, maxRetriesNotMatch)
	assert.Equal(t, logLevel, *client.logLevel, logLevelNotMatch)
	assert.Equal(t, logger, client.logger, loggerNotMatch)
	assert.NotNil(t, client.InitialBackOff, initialBackoffUnset)
	assert.NotNil(t, client.BackOffProgression, backoffProgressionUnset)
}

func TestDefaultIBMClient_ConfigPersisted_NoHTTPClient_OnCreation(t *testing.T) {

	// Set variables for the test
	maxRetries := 7
	logLevel := aws.LogDebug | aws.LogDebugWithRequestRetries | aws.LogDebugWithRequestErrors
	logger := aws.NewDefaultLogger()

	// Set a config
	config := buildConfig(maxRetries, logLevel, logger)

	// Assertions
	client := DefaultIBMClient(config)
	assert.NotNil(t, client.Client, "Http Client unset")
	assert.Equal(t, maxRetries, client.MaxRetries, maxRetriesNotMatch)
	assert.Equal(t, logLevel, *client.logLevel, logLevelNotMatch)
	assert.Equal(t, logger, client.logger, loggerNotMatch)
	assert.NotNil(t, client.InitialBackOff, initialBackoffUnset)
	assert.NotNil(t, client.BackOffProgression, backoffProgressionUnset)
}

func TestDefaultIBMClient_ConfigPersisted_NoMaxRetries_OnCreation(t *testing.T) {

	// Set variables for the test
	httpClient := http.DefaultClient
	logLevel := aws.LogDebug | aws.LogDebugWithRequestRetries | aws.LogDebugWithRequestErrors
	logger := aws.NewDefaultLogger()

	// Build a config
	config := buildConfig(0, logLevel, logger)

	// Assertions
	client := DefaultIBMClient(config)
	assert.Equal(t, httpClient, client.Client, httpClientNotMatch)
	assert.True(t, client.MaxRetries >= 0, "Max Retries invalid")
	assert.Equal(t, logLevel, *client.logLevel, logLevelNotMatch)
	assert.Equal(t, logger, client.logger, loggerNotMatch)
	assert.NotNil(t, client.InitialBackOff, initialBackoffUnset)
	assert.NotNil(t, client.BackOffProgression, backoffProgressionUnset)
}

func TestNewIBMClient_ConfigAndValuesPersisted_BadInitialBackOff_OnCreation(t *testing.T) {

	// Set variables for the test
	httpClient := http.DefaultClient
	maxRetries := 7
	logLevel := aws.LogDebug | aws.LogDebugWithRequestRetries | aws.LogDebugWithRequestErrors
	logger := aws.NewDefaultLogger()

	// Build a config
	config := buildConfig(maxRetries, logLevel, logger)

	// Begins the initial backoff
	initialBackOff := time.Duration(-7)
	backProgression := func(in time.Duration) time.Duration { return in }

	// Assertions
	client := NewIBMClient(config, initialBackOff, backProgression)
	assert.Equal(t, httpClient, client.Client, httpClientNotMatch)
	assert.Equal(t, maxRetries, client.MaxRetries, maxRetriesNotMatch)
	assert.Equal(t, logLevel, *client.logLevel, logLevelNotMatch)
	assert.Equal(t, logger, client.logger, loggerNotMatch)
	assert.True(t, client.InitialBackOff >= 0, "Initial BackOff invalid")

	// golang do not compare functions
	// work around, compare the results of call the function 3 times
	in := time.Duration(rand.Int63())
	assert.Equal(t, backProgression(in), client.BackOffProgression(in), backoffProgressionNotMatch)
	in = time.Duration(rand.Int63())
	assert.Equal(t, backProgression(in), client.BackOffProgression(in), backoffProgressionNotMatch)
	in = time.Duration(rand.Int63())
	assert.Equal(t, backProgression(in), client.BackOffProgression(in), backoffProgressionNotMatch)
}

func TestNewIBMClient_ConfigAndValuesPersisted_NoBackOffProgression_OnCreation(t *testing.T) {

	// Set variables for the test
	httpClient := http.DefaultClient
	maxRetries := 7
	logLevel := aws.LogDebug | aws.LogDebugWithRequestRetries | aws.LogDebugWithRequestErrors
	logger := aws.NewDefaultLogger()

	// Build a config
	config := buildConfig(maxRetries, logLevel, logger)

	// Assign time to initial backoff
	initialBackOff := time.Duration(7)

	// Assertions
	client := NewIBMClient(config, initialBackOff, nil)
	assert.Equal(t, httpClient, client.Client, httpClientNotMatch)
	assert.Equal(t, maxRetries, client.MaxRetries, maxRetriesNotMatch)
	assert.Equal(t, logLevel, *client.logLevel, logLevelNotMatch)
	assert.Equal(t, logger, client.logger, loggerNotMatch)
	assert.Equal(t, initialBackOff, client.InitialBackOff, "Initial BackOff did not match")
	assert.NotNil(t, client.BackOffProgression, backoffProgressionUnset)
}

func TestClientDo_WhenDoCalled_ServerCalled(t *testing.T) {

	// Set Request Logger
	requestLogger := make([]*http.Request, 0)

	// Build a local test server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestLogger = append(requestLogger, r)
		w.WriteHeader(200)
		w.Write([]byte("world"))
	}))
	defer ts.Close()

	// Local set the logger level
	logLevel := aws.LogDebug | aws.LogDebugWithRequestRetries | aws.LogDebugWithRequestErrors

	// Build a config
	config := buildConfig(0, logLevel, nil)

	// Build a client
	client := DefaultIBMClient(config)

	// Build a request
	req := buildTestRequest(t, ts.URL)

	// Run a bad request
	_, err := client.Do(req)
	require.Nil(t, err, "Error doing Request")

	// Assertion
	assert.Equal(t, 1, len(requestLogger), badNumberOfRetries)
}

func TestClientDo_WhenDoCalled_OnErrorRetry(t *testing.T) {
	// Set variables for the test
	retries := 3
	requestLogger := make([]*http.Request, 0)
	try := 0

	// Build a local test server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestLogger = append(requestLogger, r)
		if try < 1 {
			w.WriteHeader(500)
			try++
		} else {
			w.WriteHeader(200)
		}

		w.Write([]byte("world"))
	}))
	defer ts.Close()

	// Set local logger level
	logLevel := aws.LogDebug | aws.LogDebugWithRequestRetries | aws.LogDebugWithRequestErrors

	// Build a config
	config := buildConfig(retries, logLevel, nil)

	// Build a client
	client := DefaultIBMClient(config)

	// Build a request
	req := buildTestRequest(t, ts.URL)

	// Submit the request
	client.Do(req)
	//require.Nil(t,err,"Error doing Request")

	// Assertion
	assert.Equal(t, 2, len(requestLogger), badNumberOfRetries)
}

func TestClientDo_WhenDoCalled_OnErrorServerCalledRetriesTimes(t *testing.T) {
	// Set variables for the test
	retries := 3
	requestLogger := make([]*http.Request, 0)

	// Build a local test server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestLogger = append(requestLogger, r)
		w.WriteHeader(400)
		w.Write([]byte("world"))
	}))
	defer ts.Close()

	// Set local logger level
	logLevel := aws.LogDebug | aws.LogDebugWithRequestRetries | aws.LogDebugWithRequestErrors

	// Build a config
	config := buildConfig(retries, logLevel, nil)

	// Build a client
	client := DefaultIBMClient(config)

	// Build a request
	req := buildTestRequest(t, ts.URL)

	// Submit request
	client.Do(req)
	//require.Nil(t,err,"Error doing Request")

	// Assertion
	assert.Equal(t, retries+1, len(requestLogger), badNumberOfRetries)
}

// Test Logger
type tstLogger struct {
	loggerLogger []string
}

// Logger logging
func (l *tstLogger) Log(in ...interface{}) {
	str := fmt.Sprint(in...)
	l.loggerLogger = append(l.loggerLogger, str)
}

func TestClientDo_WhenDoCalled_LoggerCalled(t *testing.T) {
	// Set variables for the test
	retries := 3
	requestLogger := make([]*http.Request, 0)

	// Build a local test server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestLogger = append(requestLogger, r)
		w.WriteHeader(400)
		w.Write([]byte("world"))
	}))
	defer ts.Close()

	// Set local logger level
	logLevel := aws.LogDebug
	logger := &tstLogger{loggerLogger: make([]string, 0)}

	// Build a config
	config := buildConfig(retries, logLevel, logger)

	// Build a client
	client := DefaultIBMClient(config)

	// Build a request
	req := buildTestRequest(t, ts.URL)

	// Submit request
	client.Do(req)

	// Assertions
	assert.Equal(t, retries+1, len(requestLogger), badNumberOfRetries)
	//level LogDebug --> 1 entry for the call
	assert.Equal(t, 1, len(logger.loggerLogger), numberOfLogEntriesNotMatch)
}

func TestClientDo_WhenDoCalled_LoggerCalledRetriesLog(t *testing.T) {
	// Set variables for the test
	retries := 3
	requestLogger := make([]*http.Request, 0)

	// Build a local test server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestLogger = append(requestLogger, r)
		w.WriteHeader(400)
		w.Write([]byte("world"))
	}))
	defer ts.Close()

	// Set local logger
	logLevel := aws.LogDebug | aws.LogDebugWithRequestRetries
	logger := &tstLogger{loggerLogger: make([]string, 0)}

	// Build a config
	config := buildConfig(retries, logLevel, logger)

	// Build a client
	client := DefaultIBMClient(config)

	// Build a request
	req := buildTestRequest(t, ts.URL)

	// Submit request
	client.Do(req)

	// Assertiosn
	assert.Equal(t, retries+1, len(requestLogger), badNumberOfRetries)
	//level LogDebug --> 1 + the number of retries
	assert.Equal(t, 1+retries, len(logger.loggerLogger), numberOfLogEntriesNotMatch)
}

func TestClientDo_WhenDoCalled_LoggerCalledReqErrorsLog(t *testing.T) {
	// Set variables for test
	retries := 3
	requestLogger := make([]*http.Request, 0)

	// Build a local test server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestLogger = append(requestLogger, r)
		w.WriteHeader(400)
		w.Write([]byte("world"))
	}))
	defer ts.Close()

	// Set local logger levels
	logLevel := aws.LogDebug | aws.LogDebugWithRequestErrors
	logger := &tstLogger{loggerLogger: make([]string, 0)}

	// Build a config
	config := buildConfig(retries, logLevel, logger)

	// Build a client
	client := DefaultIBMClient(config)

	// Build a request
	req := buildTestRequest(t, ts.URL)

	// Submit request
	client.Do(req)

	// Assertions
	assert.Equal(t, retries+1, len(requestLogger), badNumberOfRetries)
	//level LogDebug --> 2 + the number of retries
	assert.Equal(t, 2+retries, len(logger.loggerLogger), numberOfLogEntriesNotMatch)
}

func TestClientDo_WhenDoCalled_LoggerCalledRetriesReqErrorsLog(t *testing.T) {
	// Set variables for the test
	retries := 3
	requestLogger := make([]*http.Request, 0)

	// Build a local test server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestLogger = append(requestLogger, r)
		w.WriteHeader(400)
		w.Write([]byte("world"))
	}))
	defer ts.Close()

	// Set local logger level
	logLevel := aws.LogDebug | aws.LogDebugWithRequestErrors | aws.LogDebugWithRequestRetries
	logger := &tstLogger{loggerLogger: make([]string, 0)}

	// Build a config
	config := buildConfig(retries, logLevel, logger)

	// Build a client
	client := DefaultIBMClient(config)

	// Buid a request
	req := buildTestRequest(t, ts.URL)

	// Submit request
	client.Do(req)

	// Assertions
	assert.Equal(t, retries+1, len(requestLogger), badNumberOfRetries)
	//level LogDebug --> 2x ( the number of retries + 1 )
	assert.Equal(t, 2*(retries+1), len(logger.loggerLogger), numberOfLogEntriesNotMatch)
}
