package handler

import (
	"net/http"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"fmt"
)

//save directly: after the client finish uploading, server still upload to s3
func UploadS3(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/html")
	w.Header().Add("Access-Control-Allow-Origin","*")
	if "POST" == r.Method {
		file, file_header, err := r.FormFile("file")
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		filename := file_header.Filename
		defer file.Close()
		sess := session.Must(session.NewSession(&aws.Config{Region: aws.String("us-east-1")}))
		uploader := s3manager.NewUploader(sess)
		bucket := "paradox42"
		_, err = uploader.Upload(&s3manager.UploadInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(filename),
			Body:   file,
		})
		if err != nil {
			// Print the error and exit.
			http.Error(w, fmt.Sprintf("Unable to upload %q to %q, %v", filename, bucket, err), 500)
			return
		}

		fmt.Fprintf(w, "Successfully uploaded %q to %q\n", filename, bucket)
		return
	}
}


//save directly: after the client finish uploading, server still upload to s3
func ListS3(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/html")
	w.Header().Add("Access-Control-Allow-Origin","*")
	if "POST" == r.Method {
		file, file_header, err := r.FormFile("file")
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		filename := file_header.Filename
		defer file.Close()
		sess := session.Must(session.NewSession(&aws.Config{Region: aws.String("us-east-1")}))
		uploader := s3manager.NewUploader(sess)
		bucket := "paradox42"
		_, err = uploader.Upload(&s3manager.UploadInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(filename),
			Body:   file,
		})
		if err != nil {
			// Print the error and exit.
			http.Error(w, fmt.Sprintf("Unable to upload %q to %q, %v", filename, bucket, err), 500)
			return
		}

		fmt.Fprintf(w, "Successfully uploaded %q to %q\n", filename, bucket)
		return
	}
}
