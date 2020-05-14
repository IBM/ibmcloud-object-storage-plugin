package signer

import (
	"fmt"
	"strings"
	"testing"

	"github.com/IBM/ibm-cos-sdk-go/aws"
	"github.com/IBM/ibm-cos-sdk-go/aws/awserr"
	"github.com/IBM/ibm-cos-sdk-go/aws/credentials"
	"github.com/IBM/ibm-cos-sdk-go/aws/request"
	"github.com/IBM/ibm-cos-sdk-go/awstesting"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockProv struct {
	credentials.Value
}

func (m *mockProv) Retrieve() (credentials.Value, error) {
	if "" == m.ProviderType {
		return credentials.Value{}, awserr.New("Mock Creds", `Bad Credentials /|\`, nil)
	}
	return m.Value, nil
}

func (m *mockProv) IsExpired() bool {
	return false
}

func buildCredentials(oauth string) *credentials.Credentials {
	prov := &mockProv{
		Value: credentials.Value{
			ProviderType: oauth,
			ProviderName: "MOCK",
		},
	}
	return credentials.NewCredentials(prov)
}

func getMockRouter() request.NamedHandler {

	realValue := DefaultSignerHandlerForProviderType
	DefaultSignerHandlerForProviderType = map[string]request.NamedHandler{
		"1": {
			Name: "Sgn1",
			Fn: func(i *request.Request) {
				i.HTTPRequest.Header.Add("Signer", "sg1")
			},
		},
		"2": {
			Name: "Sgn1",
			Fn: func(i *request.Request) {
				i.HTTPRequest.Header.Add("Signer", "sg2")
			},
		},
		"3": {
			Name: "Sgn1",
			Fn: func(i *request.Request) {
				i.HTTPRequest.Header.Add("Signer", "sg3")
			},
		},
	}
	mockRouter := defaultRequestSignerRouter()
	DefaultSignerHandlerForProviderType = realValue

	return mockRouter

}

func TestRouterHappyPath(t *testing.T) {

	rt := getMockRouter()
	s := awstesting.NewClient(aws.NewConfig().WithMaxRetries(0))
	s.Handlers.Clear()
	s.Handlers.Sign.PushBackNamed(rt)

	for i := 1; i <= 3; i++ {
		c := buildCredentials(fmt.Sprint(i))
		r := s.NewRequest(&request.Operation{Name: "Operation"}, nil, nil)
		r.Config.Credentials = c
		err := r.Send()
		assert.Equal(t, nil, err, "unpexpected error")
		assert.Equal(t, fmt.Sprintf("sg%d", i), r.HTTPRequest.Header.Get("signer"), "Signer did not Match")
	}

}

func TestRouterMissingHandler(t *testing.T) {

	rt := getMockRouter()
	s := awstesting.NewClient(aws.NewConfig().WithMaxRetries(0))
	s.Handlers.Clear()
	s.Handlers.Sign.PushBackNamed(rt)

	marker := `notExists\|/`

	c := buildCredentials(marker)
	r := s.NewRequest(&request.Operation{Name: "Operation"}, nil, nil)
	r.Config.Credentials = c
	err := r.Send()
	require.NotNil(t, err, "Error Expected")
	assert.Equal(t, true, strings.Contains(err.Error(), "No Handler Found for Type "+marker))

}
