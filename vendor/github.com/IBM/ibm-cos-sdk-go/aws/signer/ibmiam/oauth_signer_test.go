package ibmiam

import (
	"strings"
	"testing"

	"github.com/IBM/ibm-cos-sdk-go/aws"
	"github.com/IBM/ibm-cos-sdk-go/aws/awserr"
	"github.com/IBM/ibm-cos-sdk-go/aws/credentials"
	"github.com/IBM/ibm-cos-sdk-go/aws/credentials/ibmiam/token"
	"github.com/IBM/ibm-cos-sdk-go/aws/request"
	"github.com/IBM/ibm-cos-sdk-go/awstesting"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Mock for Provider using oauth sign
type mockProv struct {
	// IBM IAM Credentials Object
	credentials.Value
}

// Mock Sign Request Handler Retrieve()
// Returns:
//		credentials.Value: Credentials object
//		error: Error object
func (m *mockProv) Retrieve() (credentials.Value, error) {
	if "" == m.ServiceInstanceID && m.ServiceInstanceID == m.AccessToken && m.AccessToken == m.TokenType {
		return credentials.Value{}, awserr.New("Mock Creds", `Bad Credentials /|\`, nil)
	}
	return m.Value, nil
}

// Mock IsExpired
// Returns:
//		boolean value: false
func (m *mockProv) IsExpired() bool {
	return false
}

// Mock buildCredentials
// Parameters:
//		Service Instance ID
//		Access Token
//		Token Type
// Returns:
//		Credentials based on the provider
func buildCredentials(sii, at, tt string) *credentials.Credentials {
	prov := &mockProv{Value: credentials.Value{
		ServiceInstanceID: sii,
		ProviderType:      "oauth",
		ProviderName:      "MOCK",
		Token: token.Token{
			AccessToken: at,
			TokenType:   tt,
		},
	}}
	return credentials.NewCredentials(prov)
}

// Test Sign Request with IBM IAM Credentials
// using service instance id, access token and token type
func TestSignedHappyPath(t *testing.T) {

	c := buildCredentials("SII", "AT", "TT")
	v, _ := c.Get()
	s := awstesting.NewClient(aws.NewConfig().WithMaxRetries(0))
	s.Handlers.Clear()
	s.Handlers.Sign.PushBackNamed(SignRequestHandler)
	r := s.NewRequest(&request.Operation{Name: operation}, nil, nil)

	r.Config.Credentials = c
	err := r.Send()

	assert.Equal(t, nil, err, "unpexpected error")
	assert.Equal(t, v.ServiceInstanceID, r.HTTPRequest.Header.Get("ibm-service-instance-id"),
		"IBM Service Instance Id did not match")
	assert.Equal(t, v.TokenType+" "+v.AccessToken, r.HTTPRequest.Header.Get("Authorization"),
		"authorization did not match")
}

// func TestSignedMissingServiceII(t *testing.T) {

// 	c := buildCredentials("", "AT", "TT")
// 	s := awstesting.NewClient(aws.NewConfig().WithMaxRetries(0))
// 	s.Handlers.Clear()
// 	s.Handlers.Sign.PushBackNamed(SignRequestHandler)
// 	r := s.NewRequest(&request.Operation{Name: operation}, nil, nil)

// 	r.Config.Credentials = c
// 	err := r.Send()

// 	require.NotNil(t, err, errorExpectedNotFound)
// 	assert.Equal(t, errServiceInstanceIDNotSet, err, errorNotMatch)
// }

func TestSignedMissingAccessToken(t *testing.T) {

	c := buildCredentials("SII", "", "TT")
	s := awstesting.NewClient(aws.NewConfig().WithMaxRetries(0))
	s.Handlers.Clear()
	s.Handlers.Sign.PushBackNamed(SignRequestHandler)
	r := s.NewRequest(&request.Operation{Name: operation}, nil, nil)

	r.Config.Credentials = c
	err := r.Send()

	require.NotNil(t, err, errorExpectedNotFound)
	assert.Equal(t, errAccessTokenNotSet, err, errorNotMatch)
}

func TestSignedMissingTokenType(t *testing.T) {

	c := buildCredentials("SII", "AT", "")
	s := awstesting.NewClient(aws.NewConfig().WithMaxRetries(0))
	s.Handlers.Clear()
	s.Handlers.Sign.PushBackNamed(SignRequestHandler)
	r := s.NewRequest(&request.Operation{Name: operation}, nil, nil)

	r.Config.Credentials = c
	err := r.Send()

	require.NotNil(t, err, errorExpectedNotFound)
	assert.Equal(t, errTokenTypeNotSet, err, errorNotMatch)
}

func TestSignedBadCredentials(t *testing.T) {

	c := buildCredentials("", "", "")
	s := awstesting.NewClient(aws.NewConfig().WithMaxRetries(0))
	s.Handlers.Clear()
	s.Handlers.Sign.PushBackNamed(SignRequestHandler)
	r := s.NewRequest(&request.Operation{Name: operation}, nil, nil)

	r.Config.Credentials = c
	err := r.Send()

	require.NotNil(t, err, errorExpectedNotFound)
	assert.Equal(t, true, strings.Contains(err.Error(), `Bad Credentials /|\`), errorNotMatch)
}
