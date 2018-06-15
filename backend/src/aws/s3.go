package aws

import (

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"fmt"
	"os"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

var Bucket string= "paradox42"
// Uploads a file to S3 given a bucket and object key. Also takes a duration
// value to terminate the update if it doesn't complete within that time.
//
// The AWS Region needs to be provided in the AWS shared config or on the
// environment variable as `AWS_REGION`. Credentials also must be provided
// Will default to shared config file, but can load from environment if provided.
//
// Usage:
//   # Upload myfile.txt to myBucket/myKey. Must complete within 10 minutes or will fail
//   go run withContext.go -b mybucket -k myKey -d 10m < myfile.txt

func exitErrorf(msg string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, msg+"\n", args...)
	os.Exit(1)
}



func list_buckets(sess *session.Session){
	svc := s3.New(sess)
	result, err := svc.ListBuckets(nil)
	if err != nil {
		exitErrorf("Unable to list buckets, %v", err)
	}

	fmt.Println("Buckets:")

	for _, b := range result.Buckets {
		fmt.Printf("* %s created on %s\n",
			aws.StringValue(b.Name), aws.TimeValue(b.CreationDate))
	}
}

func list_object(sess *session.Session){
	svc := s3.New(sess)
	bucket:="paradox42"
	resp, err := svc.ListObjects(&s3.ListObjectsInput{Bucket: aws.String(bucket)})
	if err != nil {
		exitErrorf("Unable to list items in bucket %q, %v", bucket, err)
	}

	for _, item := range resp.Contents {
		fmt.Println("Name:         ", *item.Key)
		fmt.Println("Last modified:", *item.LastModified)
		fmt.Println("Size:         ", *item.Size)
		fmt.Println("Storage class:", *item.StorageClass)
		fmt.Println("Owner:        ", item.Owner)
	}
}

func upload(sess *session.Session)  {
	bucket:="paradox42"
	filename:="test.txt"
	file, err := os.Open(filename)
	if err != nil {
		exitErrorf("Unable to open file %q, %v", err)
	}
	defer file.Close()

	uploader := s3manager.NewUploader(sess)
	_, err = uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(bucket),
		Key: aws.String(filename),
		Body: file,
	})
	if err != nil {
		// Print the error and exit.
		exitErrorf("Unable to upload %q to %q, %v", filename, bucket, err)
	}


	fmt.Printf("Successfully uploaded %q to %q\n", filename, bucket)
}

func download(sess *session.Session)  {
	bucket:="paradox42"
	filename:="test.txt"
	//download file
	downloadfile,err:=os.Create("download.txt")
	if err != nil {
		exitErrorf("Unable to open file %q, %v", err)
	}
	defer downloadfile.Close()
	downloader := s3manager.NewDownloader(sess)
	numBytes, err := downloader.Download(downloadfile,
		&s3.GetObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(filename),
		})
	if err != nil {
		exitErrorf("Unable to download item %q, %v", filename, err)
	}

	fmt.Println("Downloaded", downloadfile.Name(), numBytes, "bytes")
}

func Test() {
	//var bucket string
	//var key string
	//var timeout time.Duration
	//
	//
	////flag.StringVar(&bucket, "b", "", "Bucket name.")
	////flag.StringVar(&key, "k", "", "Object key name.")
	////flag.DurationVar(&timeout, "d", 0, "Upload timeout.")
	////flag.Parse()
	//
	//bucket="paradox42"
	//key="test.txt"
	//timeout=10*time.Minute

	// All clients require a Session. The Session provides the client with
	// shared configuration such as region, endpoint, and credentials. A
	// Session should be shared where possible to take advantage of
	// configuration and credential caching. See the session package for
	// more information.

	sess := session.Must(session.NewSession(&aws.Config{Region: aws.String("us-east-1")}))

	// Create a new instance of the service's client with a Session.
	// Optional aws.Config values can also be provided as variadic arguments
	// to the New function. This option allows you to provide service
	// specific configuration.

	list_buckets(sess)










	//
	//// Create a context with a timeout that will abort the upload if it takes
	//// more than the passed in timeout.
	//ctx := context.Background()
	//var cancelFn func()
	//if timeout > 0 {
	//	ctx, cancelFn = context.WithTimeout(ctx, timeout)
	//}
	//// Ensure the context is canceled to prevent leaking.
	//// See context package for more information, https://golang.org/pkg/context/
	//defer cancelFn()
	//
	//// Uploads the object to S3. The Context will interrupt the request if the
	//// timeout expires.
	//_, err := svc.PutObjectWithContext(ctx, &s3.PutObjectInput{
	//	Bucket: aws.String(bucket),
	//	Key:    aws.String(key),
	//	Body:   os.Stdin,
	//})
	//if err != nil {
	//	if aerr, ok := err.(awserr.Error); ok && aerr.Code() == request.CanceledErrorCode {
	//		// If the SDK can determine the request or retry delay was canceled
	//		// by a context the CanceledErrorCode error code will be returned.
	//		fmt.Fprintf(os.Stderr, "upload canceled due to timeout, %v\n", err)
	//	} else {
	//		fmt.Fprintf(os.Stderr, "failed to upload object, %v\n", err)
	//	}
	//	os.Exit(1)
	//}
	//
	//fmt.Printf("successfully uploaded file to %s/%s\n", bucket, key)
}