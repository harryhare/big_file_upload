package main

import (
	"fmt"
	"io"
	"net/http"
	"log"
	"os"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
)



func Index(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/html")
	w.Header().Add("Access-Control-Allow-Origin","*")
	w.WriteHeader(200)
	html := `
		<form enctype="multipart/form-data" action="/upload" method="POST">
			Send this file: <input name="file" type="file" />
			<input type="submit" value="Send File" />
		</form>`

	io.WriteString(w, html)
}


func UploadLocal(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/html")
	w.Header().Add("Access-Control-Allow-Origin","*")
	if "POST" == r.Method {
		file, file_header, err := r.FormFile("file")
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		defer file.Close()
		filename := file_header.Filename
		f, err := os.Create("./upload/" + filename)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		defer f.Close()
		x, err := io.Copy(f, file)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		fmt.Fprintf(w, "上传文件的大小为: %d", x)
		return
	}
}


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


func main() {
	http.HandleFunc("/",Index)
	http.HandleFunc("/upload", UploadLocal)
	http.HandleFunc("/uploadtos3", UploadS3)
	fmt.Println("listen on port 8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
