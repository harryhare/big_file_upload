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
	"bytes"
	"errors"
)

const block_size=5*1024*1024


func createMultipartUplaod(w http.ResponseWriter, key string,svc *s3.S3,len int)error{
	if (mockdb.Get(key) == nil) {
		output, err := svc.CreateMultipartUpload(&s3.CreateMultipartUploadInput{
			Bucket: aws.String(myaws.Bucket),
			Key:    aws.String(key),
		})
		if err != nil {
			return err
		}
		fmt.Fprintf(w, "Successfully create multi uploaded %q, %v\n", key,output)
		mockdb.Create(key, (len+block_size-1)/block_size, output.UploadId)
	}
	return nil
}


func updateMultipartUpload(w http.ResponseWriter, key string,svc *s3.S3,file multipart.File,start int, total int)error{
	var partNumber = int64(start)
	b:=make([]byte,block_size)
	for left:=total;left>0;left-=block_size{
		l,err:=file.Read(b)
		if(err!=nil){
			return err
		}
		length:=int64(l)
		if(block_size<left && l!=block_size || block_size>=left && l!=left){
			return errors.New("read uncomplete")
		}
		fmt.Printf("uploading part %d/%d\n",partNumber, len(mockdb.Get(key).Parts))
		output, err := svc.UploadPart(&s3.UploadPartInput{
			Bucket:     	aws.String(myaws.Bucket),
			UploadId:   	mockdb.Get(key).Id,
			Key:        	aws.String(key),
			PartNumber: 	&partNumber,
			Body:       	bytes.NewReader(b[:length]),
			ContentLength:	&length,
		})
		if err != nil {
			return err
		}
		mockdb.Add(key, &mockdb.S3Part{ETag: *output.ETag, MD5: "", PartNumber: partNumber})
		fmt.Fprintf(w, "Successfully upload part %q: %d, %v\n", key,partNumber, output)
		partNumber++
	}
	return nil
}

func completeMultipartUpload(w http.ResponseWriter, key string,svc *s3.S3)error{
	var parts =&s3.CompletedMultipartUpload{}
	db:=mockdb.Get(key)
	for _,p:=range db.Parts{
		if(p!=nil){
			parts.Parts=append(parts.Parts,&s3.CompletedPart{ETag:&p.ETag,PartNumber:&p.PartNumber})
		}
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
	fmt.Println("merge success")
	return nil
}

func UploadS3MultipartUpload(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/html")
	w.Header().Add("Access-Control-Allow-Origin","*")
	if "POST" == r.Method {
		//id := r.Form.Get("id")
		file, _, err := r.FormFile("file")
		if err != nil {
			fmt.Printf("form file error %v:",err)
			http.Error(w, err.Error(), 500)
			return
		}
		defer file.Close()
		key := r.Form.Get("key")
		if key == "" {
			fmt.Println("file key is null")
			http.Error(w, "file key is null", 500)
			return
		}
		len_str := r.Form.Get("len")
		if key == "" {
			fmt.Println("file len is null")
			http.Error(w, "file len is null", 500)
			return
		}
		length, err := strconv.Atoi(len_str)
		if (err != nil) {
			fmt.Println("file len is not int")
			http.Error(w, "file len is not int", 500)
			return
		}
		start_str := r.Form.Get("start")
		if key == "" {
			fmt.Println("start is null")
			http.Error(w, "start is null", 500)
			return
		}
		start, err := strconv.Atoi(start_str)
		if (err != nil) {
			fmt.Println("start is not int")
			http.Error(w, "start is not int", 500)
			return
		}
		sess := session.Must(session.NewSession(&aws.Config{Region: aws.String("us-east-1")}))
		svc := s3.New(sess)

		err=createMultipartUplaod(w,key,svc,length)
		if(err!=nil){
			http.Error(w, fmt.Sprintf("Unable to create multi upload %q, %v", key, err), 500)
			fmt.Println(fmt.Sprintf("Unable to create multi upload %q, %v", key, err))
			return
		}

		err=updateMultipartUpload(w,key,svc,file,start,length)
		if(err!=nil){
			http.Error(w, fmt.Sprintf("Unable to upload part %q, %v", key, err), 500)
			fmt.Println(fmt.Sprintf("Unable to upload part %q, %v", key, err))
			return
		}

		err=completeMultipartUpload(w,key,svc)
		if(err!=nil){
			http.Error(w, fmt.Sprintf("Unable to complete %q, %v", key, err), 500)
			fmt.Println(fmt.Sprintf("Unable to complete %q, %v", key, err))
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
			fmt.Println("file id is null")
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
			fmt.Println(fmt.Sprintf("Unable to list parts of %q, %v", key, err))
			return
		}
		fmt.Printf("Successfully abort upload %q, %v\n", key, output)
		x,err:=json.Marshal(output)
		if(err!=nil){
			http.Error(w, fmt.Sprintf("Unable to marshal, %v", key, err), 500)
			fmt.Println(fmt.Sprintf("Unable to marshal, %v", key, err))
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
			fmt.Println("file id is null")
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
			fmt.Println(fmt.Sprintf("Unable to abort %q, %v", key, err))
			return
		}
		mockdb.Delete(key)

		fmt.Fprintf(w, "Successfully abort upload %q, %v\n", key, output)
		return
	}
}




