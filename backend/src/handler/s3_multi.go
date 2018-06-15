package handler

import (
	"net/http"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/aws"
	"fmt"
	"mockdb"
	"github.com/aws/aws-sdk-go/service/s3"
	myaws "aws"
	"encoding/json"
	"strconv"
	"mime/multipart"
)

const block_size=5000000


func CreateMultipartUplaod(w http.ResponseWriter, key string,svc *s3.S3,len int)error{
	if (mockdb.Get(key) == nil) {
		output1, err := svc.CreateMultipartUpload(&s3.CreateMultipartUploadInput{
			Bucket: aws.String(myaws.Bucket),
			Key:    aws.String(key),
		})
		if err != nil {
			return err
		}
		fmt.Fprintf(w, "Successfully create multi uploaded %q\n", key)
		mockdb.Create(key, (len+block_size-1)/block_size, output1.UploadId)
	}
	return nil
}


func UpdateMultipartUpload(w http.ResponseWriter, key string,svc *s3.S3,file multipart.File)error{
	var part_number int64=1

	output2,err:=svc.UploadPart(&s3.UploadPartInput{
		Bucket:aws.String(myaws.Bucket),
		UploadId:mockdb.Get(key).Id,
		Key:aws.String(key),
		PartNumber:&part_number,
		Body:file,
	})
	if err != nil {
		return err
	}
	fmt.Fprintf(w, "Successfully upload part %q, %v\n", key, output2)
	mockdb.Add(key, &mockdb.S3Part{ETag:*output2.ETag,MD5:"",PartNumber:1})
	return nil
}

func CompleteMultipartUpload(w http.ResponseWriter, key string,svc *s3.S3)error{
	var parts =&s3.CompletedMultipartUpload{}
	db:=mockdb.Get(key)
	for _,p:=range db.Parts{
		parts.Parts=append(parts.Parts,&s3.CompletedPart{ETag:&p.ETag,PartNumber:&p.PartNumber})
	}
	output,err:=svc.CompleteMultipartUpload(&s3.CompleteMultipartUploadInput{
		Bucket:aws.String(myaws.Bucket),
		UploadId:db.Id,
		Key:&key,
		MultipartUpload:parts,
	})
	if err != nil {
		return err
	}
	fmt.Fprintf(w, "Successfully complete %q, %v\n", key, output)
	return nil
}

func UploadS3MultipartUpload(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/html")
	w.Header().Add("Access-Control-Allow-Origin","*")
	if "POST" == r.Method {
		//id := r.Form.Get("id")
		file, _, err := r.FormFile("file")
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		defer file.Close()
		key := r.Form.Get("key")
		if key == "" {
			http.Error(w, "file key is null", 500)
			return
		}
		len_str := r.Form.Get("len")
		if key == "" {
			http.Error(w, "file len is null", 500)
			return
		}
		len, err := strconv.Atoi(len_str)
		if (err != nil) {
			http.Error(w, "file len is not int", 500)
			return
		}
		sess := session.Must(session.NewSession(&aws.Config{Region: aws.String("us-east-1")}))
		svc := s3.New(sess)

		err=CreateMultipartUplaod(w,key,svc,len)
		if(err!=nil){
			http.Error(w, fmt.Sprintf("Unable to create multi upload %q, %v", key, err), 500)
			return
		}

		err=UpdateMultipartUpload(w,key,svc,file)
		if(err!=nil){
			http.Error(w, fmt.Sprintf("Unable to upload part %q, %v", key, err), 500)
			return
		}

		err=CompleteMultipartUpload(w,key,svc)
		if(err!=nil){
			http.Error(w, fmt.Sprintf("Unable to complete %q, %v", key, err), 500)
			return
		}
	}
}

func UploadS3MultipartStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "json")
	w.Header().Add("Access-Control-Allow-Origin","*")
	if "POST" == r.Method {
		key := r.Form.Get("key")
		if key == "" {
			http.Error(w, "file id is null", 500)
			return
		}

		sess := session.Must(session.NewSession(&aws.Config{Region: aws.String("us-east-1")}))
		svc := s3.New(sess)
		id:=mockdb.Get(key).Id

		listInput := &s3.ListPartsInput{
			Bucket:   &myaws.Bucket,
			Key:      &key,
			UploadId: id,
		}
		output, err := svc.ListParts(listInput)
		if err != nil {
			http.Error(w, fmt.Sprintf("Unable to list parts of %q, %v", key, err), 500)
			return
		}
		fmt.Printf("Successfully abort upload %q, %v\n", key, output)
		x,err:=json.Marshal(output)
		if(err!=nil){
			http.Error(w, fmt.Sprintf("Unable to marshal, %v", key, err), 500)
			return
		}
		w.Write(x)
		return
	}
}


func UploadS3MultipartStop(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/html")
	w.Header().Add("Access-Control-Allow-Origin","*")
	if "POST" == r.Method {
		key := r.Form.Get("key")
		if key == "" {
			http.Error(w, "file id is null", 500)
			return
		}

		sess := session.Must(session.NewSession(&aws.Config{Region: aws.String("us-east-1")}))
		svc := s3.New(sess)
		id:=mockdb.Get(key).Id

		abortInput := &s3.AbortMultipartUploadInput{
			Bucket:   &myaws.Bucket,
			Key:      &key,
			UploadId: id,
		}
		output, err := svc.AbortMultipartUpload(abortInput)
		if err != nil {
			// Print the error and exit.
			http.Error(w, fmt.Sprintf("Unable to abort %q, %v", key, err), 500)
			return
		}
		mockdb.Delete(key)

		fmt.Fprintf(w, "Successfully abort upload %q, %v\n", key, output)
		return
	}
}




