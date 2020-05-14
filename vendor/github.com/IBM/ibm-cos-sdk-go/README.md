# IBM Cloud Object Storage - Go SDK

This package allows Go developers to write software that interacts with [IBM
Cloud Object Storage](https://cloud.ibm.com/docs/services/cloud-object-storage?topic=cloud-object-storage-about-ibm-cloud-object-storage#about-ibm-cloud-object-storage).  It is a fork of the [``AWS SDK for Go``](https://github.com/aws/aws-sdk-go) library and can stand as a drop-in replacement if the application needs to connect to object storage using an S3-like API and does not make use of other AWS services.

## Documentation

* [Core documentation for IBM COS](https://cloud.ibm.com/docs/services/cloud-object-storage?topic=cloud-object-storage-getting-started-tutorial#getting-started-tutorial)
* [Go API reference documentation](https://ibm.github.io/ibm-cos-sdk-go)
* [REST API reference documentation](https://cloud.ibm.com/docs/services/cloud-object-storage?topic=cloud-object-storage-compatibility-api-about#compatibility-api-about)

For release notes, see the [CHANGELOG](CHANGELOG.rst).

* [Getting the SDK](#getting-the-sdk)
* [Example code](#example-code)
* [Getting help](#getting-help)

## Quick start

You'll need:
  * An instance of COS.
  * An API key from [IBM Cloud Identity and Access Management](https://cloud.ibm.com/docs/iam?topic=iam-userroles#userroles) with at least `Writer` permissions.
  * The ID of the instance of COS that you are working with.
  * Token acquisition endpoint
  * Service endpoint

These values can be found in the IBM Cloud Console by [generating a 'service credential'](https://cloud.ibm.com/docs/services/cloud-object-storage/iam?topic=cloud-object-storage-service-credentials#service-credentials).


## Getting the SDK

Use go get to retrieve the SDK to add it to your GOPATH workspace, or project's Go module dependencies.  The SDK requires a minimum version of Go 1.9.

```sh
go get github.com/IBM/ibm-cos-sdk-go
```

To update the SDK use go get -u to retrieve the latest version of the SDK..

```sh
go get -u github.com/IBM/ibm-cos-sdk-go
```

## Example code
Create a file `main.go`, replacing your own values for API key, instance ID, and bucket name:

```go
package main

import (
	"fmt"

	"github.com/IBM/ibm-cos-sdk-go/aws"
	"github.com/IBM/ibm-cos-sdk-go/aws/credentials/ibmiam"
	"github.com/IBM/ibm-cos-sdk-go/aws/session"
	"github.com/IBM/ibm-cos-sdk-go/service/s3"
)

const (
	apiKey            = "<API_KEY>"
	serviceInstanceID = "<RESOURCE_INSTANCE_ID>"
	authEndpoint      = "https://iam.bluemix.net/oidc/token"
	serviceEndpoint   = "https://s3-api.us-geo.objectstorage.softlayer.net"
)

func main() {

	newBucket := "new-bucketee"
	newColdBucket := "new-cold-bucketee"

	conf := aws.NewConfig().
		WithEndpoint(serviceEndpoint).
		WithCredentials(ibmiam.NewStaticCredentials(aws.NewConfig(),
			authEndpoint, apiKey, serviceInstanceID)).
		WithS3ForcePathStyle(true)

	sess := session.Must(session.NewSession())
	client := s3.New(sess, conf)

	input := &s3.CreateBucketInput{
		Bucket: aws.String(newBucket),
	}
	client.CreateBucket(input)

	input2 := &s3.CreateBucketInput{
		Bucket: aws.String(newColdBucket),
		CreateBucketConfiguration: &s3.CreateBucketConfiguration{
			LocationConstraint: aws.String("us-cold"),
		},
	}
	client.CreateBucket(input2)

	d, _ := client.ListBuckets(&s3.ListBucketsInput{})
	fmt.Println(d)
}
```

From the command line, run `go run main.go`.  You should see a list of your buckets.

## Getting Help

Feel free to use GitHub issues for tracking bugs and feature requests, but for help please use one of the following resources:

* Read a quick start guide in [IBM Cloud Docs](https://cloud.ibm.com/docs/services/cloud-object-storage/iam?topic=cloud-object-storage-using-go#using-go).
* Ask a question on [Stack Overflow](https://stackoverflow.com/) and tag it with `ibm` and `object-storage`.
* Open a support ticket with [IBM Cloud Support](https://cloud.ibm.com/unifiedsupport/supportcenter/)
* If it turns out that you may have found a bug, please [open an issue](https://github.com/ibm/ibm-cos-sdk-go/issues/new).

## License

This SDK is distributed under the
[Apache License, Version 2.0](http://www.apache.org/licenses/LICENSE-2.0),
see LICENSE.txt and NOTICE.txt for more information.