package tokenmanager

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/IBM/ibm-cos-sdk-go/aws"
	"github.com/IBM/ibm-cos-sdk-go/aws/credentials/ibmiam/token"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// IBM Client Mock
type ibmclientMock struct {

	// HTTP Request Logs
	requestLogs []*http.Request

	// Request Handler
	handler func(req *http.Request) (*http.Response, error)
}

// Mock IBM HTTP Client Request Do Func
// Parameter:
//		HTTP Client Request object
// Returns:
//		HTTP Client Request object
//		Error
func (icm *ibmclientMock) Do(req *http.Request) (*http.Response, error) {
	icm.requestLogs = append(icm.requestLogs, req)
	return icm.handler(req)
}

// Tests new Token Manager with API Key and Initial Get Func
func TestNewTokenManagerFromApiKeyAndInitialGet(t *testing.T) {

	// Sets vars for the test
	config := new(aws.Config).WithLogger(aws.NewDefaultLogger()).WithLogLevel(aws.LogDebug)
	apiKey := "UNIK-APIKEY"
	authEndPoint := endPoint
	advised := func(_ time.Duration) time.Duration { return time.Duration(11) * time.Second }
	mandatory := func(_ time.Duration) time.Duration { return time.Duration(7) * time.Second }

	// Create a mock token
	tokenValue := token.Token{
		AccessToken:  "A",
		RefreshToken: "R",
		TokenType:    "T",
	}

	// Sets the Request Handler
	handler := func(*http.Request) (*http.Response, error) {
		rsp := new(http.Response)

		rsp.StatusCode = 200
		bs, _ := json.Marshal(tokenValue)
		rsp.Body = ioutil.NopCloser(bytes.NewReader(bs))

		return rsp, nil
	}

	// Mock IBM Client
	icm := &ibmclientMock{
		requestLogs: make([]*http.Request, 0),
		handler:     handler,
	}

	// Mock Token Manager
	tm := newTokenManagerFromAPIKey(config, apiKey, authEndPoint, advised, mandatory, time.Now, icm)
	defer tm.StopBackgroundRefresh()

	// Get a token
	tk, e := tm.Get()

	// Expectations
	// - No error in getting token
	// - Token attributes match
	// - Request Log Count = 1
	// - Request URL == IBM IAM Authentication Server
	// - API Key match
	require.Nil(t, e, errorGettingToken)
	assert.Equal(t, tokenValue, *tk, tokensNotMatch)

	assert.Equal(t, 1, len(icm.requestLogs), "Bad Request Count")
	assert.Equal(t, icm.requestLogs[0].URL.String(), authEndPoint, "Request bad endpoint")

	bs, _ := ioutil.ReadAll(icm.requestLogs[0].Body)
	v, _ := url.ParseQuery(string(bs))
	assert.Equal(t, apiKey, v.Get("apikey"), "apikey not match")

}

// Tests new Customized Token Manager with Initial Get Func
func TestNewTokenManagerCustomFuncInitialGet(t *testing.T) {

	// Sets vars for the test
	config := &aws.Config{}
	authEndPoint := endPoint
	advised := func(_ time.Duration) time.Duration { return time.Duration(11) * time.Second }
	mandatory := func(_ time.Duration) time.Duration { return time.Duration(7) * time.Second }

	// Create a mock token
	tokenValue := token.Token{
		AccessToken:  "A",
		RefreshToken: "R",
		TokenType:    "T",
	}

	// Creates a custom func with token
	customFunc := func() (*token.Token, error) {
		return &tokenValue, nil
	}

	// Empty Mock IBM Client
	icm := &ibmclientMock{}

	// Mock Token Manager
	tm := newTokenManager(config, customFunc, authEndPoint, advised, mandatory, time.Now, icm)
	defer tm.StopBackgroundRefresh()

	// Gets a token using Token Manager with bad client
	tk, e := tm.Get()

	// Expectations
	// - No error in getting token
	// - Token attributes match
	require.Nil(t, e, errorGettingToken)
	assert.Equal(t, tokenValue, *tk, tokensNotMatch)
}

// Test Background Refresh in the Token Manager
func TestBackGroundRefreshHappens(t *testing.T) {

	// Sets vars for the test
	config := &aws.Config{}
	authEndPoint := endPoint
	advised := func(_ time.Duration) time.Duration { return time.Duration(11) * time.Second }
	mandatory := func(_ time.Duration) time.Duration { return time.Duration(5) * time.Second }

	// Creates a mock token
	tokenValue := token.Token{
		AccessToken:  "A",
		RefreshToken: "R",
		TokenType:    "T",
		ExpiresIn:    11,
		Expiration:   time.Now().Add(time.Second * 11).Unix(),
	}

	// Custom func with token/errr
	customFunc := func() (*token.Token, error) {
		return &tokenValue, nil
	}

	// Request Handler
	handler := func(*http.Request) (*http.Response, error) {
		rsp := new(http.Response)

		rsp.StatusCode = 200
		refreshToken := tokenValue
		refreshToken.Expiration = time.Now().Add(time.Second * 11).Unix()
		bs, _ := json.Marshal(refreshToken)
		rsp.Body = ioutil.NopCloser(bytes.NewReader(bs))

		return rsp, nil
	}

	// Mock IBM Client
	icm := &ibmclientMock{
		requestLogs: make([]*http.Request, 0),
		handler:     handler,
	}

	// Mock Token Manager
	tm := newTokenManager(config, customFunc, authEndPoint, advised, mandatory, time.Now, icm)
	defer tm.StopBackgroundRefresh()

	// Gets a token using Token Manager
	tk, e := tm.Get()

	// Expectations
	// - No error in getting token
	// - Token attributes match
	// - Grant Type == Refresh Token
	// - Token's Refresh token attribute match
	require.Nil(t, e, errorGettingToken)
	assert.Equal(t, tokenValue, *tk, tokensNotMatch)

	time.Sleep(1 * time.Second)

	bs, _ := ioutil.ReadAll(icm.requestLogs[0].Body)
	v, _ := url.ParseQuery(string(bs))
	assert.Equal(t, "refresh_token", v.Get("grant_type"), "grant not match")
	assert.Equal(t, tokenValue.RefreshToken, v.Get("refresh_token"), "refreshtoken not match")
}

// Tests new Customized Token Manager with Initial Get Func
func TestNewTokenManagerCustomFuncInitialGetError(t *testing.T) {

	// Sets vars for the test
	config := &aws.Config{}
	authEndPoint := endPoint
	advised := func(_ time.Duration) time.Duration { return time.Duration(11) * time.Second }
	mandatory := func(_ time.Duration) time.Duration { return time.Duration(5) * time.Second }

	// Pseudo interface
	foo := map[string]interface{}{
		"test": "Mock error message",
	}

	// errorGettingToken
	tokenError := token.Error{
		Context:      foo,
		ErrorCode:    "400",
		ErrorMessage: errorGettingToken,
	}

	// Custom func with token/err
	customFunc := func() (*token.Token, error) {
		return nil, &tokenError
	}

	// Request Handler
	handler := func(*http.Request) (*http.Response, error) {
		rsp := new(http.Response)

		rsp.StatusCode = 400
		// bs, _ := json.Marshal(tokenValue)
		// rsp.Body = ioutil.NopCloser(bytes.NewReader(bs))
		apiErr := token.Error{}

		return nil, &apiErr
	}

	// Mock IBM Client
	icm := &ibmclientMock{
		requestLogs: make([]*http.Request, 0),
		handler:     handler,
	}

	// Mock Token Manager
	tm := newTokenManager(config, customFunc, authEndPoint, advised, mandatory, time.Now, icm)
	defer tm.StopBackgroundRefresh()

	// Gets a token using Token Manager
	_, e := tm.Get()

	// err = json.Unmarshal(bodyContent, &apiErr)
	// if err != nil {
	// 	return nil, err
	// }
	// Expectations
	// - Error in getting token
	// - Error Message match
	require.NotNil(t, e, errorGettingToken)
	assert.Equal(t, &tokenError, e, "error message not match")
}
