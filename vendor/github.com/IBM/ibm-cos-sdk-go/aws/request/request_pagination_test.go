package request_test

import (
	"reflect"
	"testing"

	"github.com/IBM/ibm-cos-sdk-go/aws"
	"github.com/IBM/ibm-cos-sdk-go/aws/request"
	"github.com/IBM/ibm-cos-sdk-go/awstesting"
	"github.com/IBM/ibm-cos-sdk-go/awstesting/unit"
	"github.com/IBM/ibm-cos-sdk-go/service/s3"
)

func TestSkipPagination(t *testing.T) {
	client := s3.New(unit.Session)
	client.Handlers.Send.Clear() // mock sending
	client.Handlers.Unmarshal.Clear()
	client.Handlers.UnmarshalMeta.Clear()
	client.Handlers.ValidateResponse.Clear()
	client.Handlers.Unmarshal.PushBack(func(r *request.Request) {
		r.Data = &s3.HeadBucketOutput{}
	})

	req, _ := client.HeadBucketRequest(&s3.HeadBucketInput{Bucket: aws.String("bucket")})

	numPages, gotToEnd := 0, false
	req.EachPage(func(p interface{}, last bool) bool {
		numPages++
		if last {
			gotToEnd = true
		}
		return true
	})
	if e, a := 1, numPages; e != a {
		t.Errorf("expect %v, got %v", e, a)
	}
	if !gotToEnd {
		t.Errorf("expect true")
	}
}

// Use S3 for simplicity
func TestPaginationTruncation(t *testing.T) {
	client := s3.New(unit.Session)

	reqNum := 0
	resps := []*s3.ListObjectsOutput{
		{IsTruncated: aws.Bool(true), Contents: []*s3.Object{{Key: aws.String("Key1")}}},
		{IsTruncated: aws.Bool(true), Contents: []*s3.Object{{Key: aws.String("Key2")}}},
		{IsTruncated: aws.Bool(false), Contents: []*s3.Object{{Key: aws.String("Key3")}}},
		{IsTruncated: aws.Bool(true), Contents: []*s3.Object{{Key: aws.String("Key4")}}},
	}

	client.Handlers.Send.Clear() // mock sending
	client.Handlers.Unmarshal.Clear()
	client.Handlers.UnmarshalMeta.Clear()
	client.Handlers.ValidateResponse.Clear()
	client.Handlers.Unmarshal.PushBack(func(r *request.Request) {
		r.Data = resps[reqNum]
		reqNum++
	})

	params := &s3.ListObjectsInput{Bucket: aws.String("bucket")}

	results := []string{}
	err := client.ListObjectsPages(params, func(p *s3.ListObjectsOutput, last bool) bool {
		results = append(results, *p.Contents[0].Key)
		return true
	})

	if e, a := []string{"Key1", "Key2", "Key3"}, results; !reflect.DeepEqual(e, a) {
		t.Errorf("expect %v, got %v", e, a)
	}
	if err != nil {
		t.Errorf("expect nil, %v", err)
	}

	// Try again without truncation token at all
	reqNum = 0
	resps[1].IsTruncated = nil
	resps[2].IsTruncated = aws.Bool(true)
	results = []string{}
	err = client.ListObjectsPages(params, func(p *s3.ListObjectsOutput, last bool) bool {
		results = append(results, *p.Contents[0].Key)
		return true
	})

	if e, a := []string{"Key1", "Key2"}, results; !reflect.DeepEqual(e, a) {
		t.Errorf("expect %v, got %v", e, a)
	}
	if err != nil {
		t.Errorf("expect nil, %v", err)
	}
}

func TestPaginationNilInput(t *testing.T) {
	// Code generation doesn't have a great way to verify the code is correct
	// other than being run via unit tests in the SDK. This should be fixed
	// So code generation can be validated independently.

	client := s3.New(unit.Session)
	client.Handlers.Validate.Clear()
	client.Handlers.Send.Clear() // mock sending
	client.Handlers.Unmarshal.Clear()
	client.Handlers.UnmarshalMeta.Clear()
	client.Handlers.ValidateResponse.Clear()
	client.Handlers.Unmarshal.PushBack(func(r *request.Request) {
		r.Data = &s3.ListObjectsOutput{}
	})

	gotToEnd := false
	numPages := 0
	err := client.ListObjectsPages(nil, func(p *s3.ListObjectsOutput, last bool) bool {
		numPages++
		if last {
			gotToEnd = true
		}
		return true
	})

	if err != nil {
		t.Fatalf("expect no error, but got %v", err)
	}
	if e, a := 1, numPages; e != a {
		t.Errorf("expect %d number pages but got %d", e, a)
	}
	if !gotToEnd {
		t.Errorf("expect to of gotten to end, did not")
	}
}

func TestPaginationWithContextNilInput(t *testing.T) {
	// Code generation doesn't have a great way to verify the code is correct
	// other than being run via unit tests in the SDK. This should be fixed
	// So code generation can be validated independently.

	client := s3.New(unit.Session)
	client.Handlers.Validate.Clear()
	client.Handlers.Send.Clear() // mock sending
	client.Handlers.Unmarshal.Clear()
	client.Handlers.UnmarshalMeta.Clear()
	client.Handlers.ValidateResponse.Clear()
	client.Handlers.Unmarshal.PushBack(func(r *request.Request) {
		r.Data = &s3.ListObjectsOutput{}
	})

	gotToEnd := false
	numPages := 0
	ctx := &awstesting.FakeContext{DoneCh: make(chan struct{})}
	err := client.ListObjectsPagesWithContext(ctx, nil, func(p *s3.ListObjectsOutput, last bool) bool {
		numPages++
		if last {
			gotToEnd = true
		}
		return true
	})

	if err != nil {
		t.Fatalf("expect no error, but got %v", err)
	}
	if e, a := 1, numPages; e != a {
		t.Errorf("expect %d number pages but got %d", e, a)
	}
	if !gotToEnd {
		t.Errorf("expect to of gotten to end, did not")
	}
}

func TestPagination_Standalone(t *testing.T) {
	type testPageInput struct {
		NextToken *string
	}
	type testPageOutput struct {
		Value     *string
		NextToken *string
	}
	type testCase struct {
		Value, PrevToken, NextToken *string
	}

	type testCaseList struct {
		StopOnSameToken bool
		Cases           []testCase
	}

	cases := []testCaseList{
		{
			Cases: []testCase{
				{aws.String("FirstValue"), aws.String("InitalToken"), aws.String("FirstToken")},
				{aws.String("SecondValue"), aws.String("FirstToken"), aws.String("SecondToken")},
				{aws.String("ThirdValue"), aws.String("SecondToken"), nil},
			},
			StopOnSameToken: false,
		},
		{
			Cases: []testCase{
				{aws.String("FirstValue"), aws.String("InitalToken"), aws.String("FirstToken")},
				{aws.String("SecondValue"), aws.String("FirstToken"), aws.String("SecondToken")},
				{aws.String("ThirdValue"), aws.String("SecondToken"), aws.String("")},
			},
			StopOnSameToken: false,
		},
		{
			Cases: []testCase{
				{aws.String("FirstValue"), aws.String("InitalToken"), aws.String("FirstToken")},
				{aws.String("SecondValue"), aws.String("FirstToken"), aws.String("SecondToken")},
				{nil, aws.String("SecondToken"), aws.String("SecondToken")},
			},
			StopOnSameToken: true,
		},
		{
			Cases: []testCase{
				{aws.String("FirstValue"), aws.String("InitalToken"), aws.String("FirstToken")},
				{aws.String("SecondValue"), aws.String("FirstToken"), aws.String("SecondToken")},
				{aws.String("SecondValue"), aws.String("SecondToken"), aws.String("SecondToken")},
			},
			StopOnSameToken: true,
		},
	}

	for _, testcase := range cases {
		c := testcase.Cases
		input := testPageInput{
			NextToken: c[0].PrevToken,
		}

		svc := awstesting.NewClient()
		i := 0
		p := request.Pagination{
			EndPageOnSameToken: testcase.StopOnSameToken,
			NewRequest: func() (*request.Request, error) {
				r := svc.NewRequest(
					&request.Operation{
						Name: "Operation",
						Paginator: &request.Paginator{
							InputTokens:  []string{"NextToken"},
							OutputTokens: []string{"NextToken"},
						},
					},
					&input, &testPageOutput{},
				)
				// Setup handlers for testing
				r.Handlers.Clear()
				r.Handlers.Build.PushBack(func(req *request.Request) {
					if e, a := len(c), i+1; a > e {
						t.Fatalf("expect no more than %d requests, got %d", e, a)
					}
					in := req.Params.(*testPageInput)
					if e, a := aws.StringValue(c[i].PrevToken), aws.StringValue(in.NextToken); e != a {
						t.Errorf("%d, expect NextToken input %q, got %q", i, e, a)
					}
				})
				r.Handlers.Unmarshal.PushBack(func(req *request.Request) {
					out := &testPageOutput{
						Value: c[i].Value,
					}
					if c[i].NextToken != nil {
						next := *c[i].NextToken
						out.NextToken = aws.String(next)
					}
					req.Data = out
				})
				return r, nil
			},
		}

		for p.Next() {
			data := p.Page().(*testPageOutput)

			if e, a := aws.StringValue(c[i].Value), aws.StringValue(data.Value); e != a {
				t.Errorf("%d, expect Value to be %q, got %q", i, e, a)
			}
			if e, a := aws.StringValue(c[i].NextToken), aws.StringValue(data.NextToken); e != a {
				t.Errorf("%d, expect NextToken to be %q, got %q", i, e, a)
			}

			i++
		}
		if e, a := len(c), i; e != a {
			t.Errorf("expected to process %d pages, did %d", e, a)
		}
		if err := p.Err(); err != nil {
			t.Fatalf("%d, expected no error, got %v", i, err)
		}
	}
}
