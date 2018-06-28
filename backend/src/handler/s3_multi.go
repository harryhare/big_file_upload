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
	"mime/multipart"
	"bytes"
	"mime"
	"io"
	"strconv"
	"io/ioutil"
	"github.com/aws/aws-sdk-go/aws/signer/v4"
	"time"
	"errors"
)

const block_size = 5 * 1024 * 1024
const upload_mode = "stream"

func createMultipartUpload(w http.ResponseWriter, key string, len int) error {
	sess := session.Must(session.NewSession(&aws.Config{Region: aws.String("us-east-1")}))
	svc := s3.New(sess)
	if (mockdb.Get(key) == nil) {
		output, err := svc.CreateMultipartUpload(&s3.CreateMultipartUploadInput{
			Bucket: aws.String(myaws.Bucket),
			Key:    aws.String(key),
			ACL:    aws.String("public-read"),
		})
		if err != nil {
			return err
		}
		fmt.Fprintf(w, "Successfully create multi uploaded %q, %v\n", key, output)
		mockdb.Create(key, (len+block_size-1)/block_size, output.UploadId)
	}
	return nil
}

func streamUploadPart(sess *session.Session, r *http.Request, key string, reader io.Reader, length int, part_number int, upload_id string) (*s3.UploadPartOutput, error) {
	bucket := "paradox42"

	url := fmt.Sprintf("https://%s.s3.amazonaws.com/%s?partNumber=%d&uploadId=%s", bucket, key, part_number, upload_id)
	uploadRequest, err := http.NewRequest(http.MethodPut, url, reader)
	if err != nil {
		panic(err)
	}

	uploadRequest.Header.Add("Content-Disposition",
		fmt.Sprintf("attachment;filename=\"%s\"", key))
	uploadRequest.Header.Add("Content-Type", r.Header.Get("Content-Type"))
	uploadRequest.ContentLength = int64(length)

	uploadRequest.Header.Add("X-Amz-Content-Sha256", "UNSIGNED-PAYLOAD") // public-read
	//uploadRequest.Header.Add("X-Amz-Acl", "public-read")

	body := uploadRequest.Body
	signer := v4.NewSigner(sess.Config.Credentials)
	if _, err := signer.Sign(uploadRequest, nil, "s3", "us-east-1", time.Now()); err != nil {
		panic(err)
	}
	uploadRequest.Body = body

	client := &http.Client{}
	fmt.Printf("start upload stream mode, filename=%s, key=%s\n", key, key)
	response, err := client.Do(uploadRequest)
	fmt.Printf("end upload stream mode, filename=%s, key=%s\n", key, key)

	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	if response.StatusCode != 200 {
		if errorContent, err := ioutil.ReadAll(response.Body); err != nil {
			panic(err)
		} else {
			panic(fmt.Sprintf("error uploading to aws, response:%s", errorContent))
		}
	}
	fmt.Println(response)
	if (response.StatusCode != 200) {
		resp, err := ioutil.ReadAll(response.Body)
		if (err != nil) {
			return nil, err
		}
		return nil, errors.New(fmt.Sprintf("upload part failed %d,%s", part_number, resp))
	}
	etag := response.Header.Get("Etag")
	output := &s3.UploadPartOutput{
		ETag: &etag,
	}
	return output, nil
}

func updateMultipartUpload(sess *session.Session, request *http.Request, w http.ResponseWriter, key string, svc *s3.S3, file *multipart.Part, start int64, total int64) error {
	var partNumber = int64(start)

	for left := total; left > 0; left -= block_size {
		var length = int64(left)
		if (length > block_size) {
			length = block_size
		}
		r := io.LimitReader(file, block_size)

		fmt.Printf("upload part %d len %d\n", partNumber, length)
		var output *s3.UploadPartOutput
		var err error
		if upload_mode == "stream" {
			output, err = streamUploadPart(sess, request, key, r, int(length), int(partNumber), *mockdb.Get(key).Id)
		} else {
			b, err := ioutil.ReadAll(r)

			if err != nil && err != io.EOF {
				return err
			}
			if int64(len(b)) != length {
				fmt.Printf("actually read %d shoud read %d %v\n", len(b), length)
				return errors.New(fmt.Sprintf("file length error"))
			}
			output, err = svc.UploadPart(&s3.UploadPartInput{
				Bucket:        aws.String(myaws.Bucket),
				UploadId:      mockdb.Get(key).Id,
				Key:           aws.String(key),
				PartNumber:    &partNumber,
				Body:          bytes.NewReader(b),
				ContentLength: &length,
			})
		}

		if err != nil {
			return err
		}
		mockdb.Add(key, &mockdb.S3Part{ETag: *output.ETag, MD5: "", PartNumber: partNumber})
		fmt.Fprintf(w, "Successfully upload part %q: %d, %v\n", key, partNumber, output)
		partNumber++
	}
	return nil
}

func completeMultipartUpload(w http.ResponseWriter, key string) error {
	sess := session.Must(session.NewSession(&aws.Config{Region: aws.String("us-east-1")}))
	svc := s3.New(sess)
	var parts = &s3.CompletedMultipartUpload{}
	db := mockdb.Get(key)
	for _, p := range db.Parts {
		if (p != nil) {
			parts.Parts = append(parts.Parts, &s3.CompletedPart{ETag: &p.ETag, PartNumber: &p.PartNumber})
		}
	}
	output, err := svc.CompleteMultipartUpload(&s3.CompleteMultipartUploadInput{
		Bucket:          aws.String(myaws.Bucket),
		UploadId:        db.Id,
		Key:             &key,
		MultipartUpload: parts,
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
	w.Header().Add("Access-Control-Allow-Origin", "*")
	fmt.Println(r.Method)
	fmt.Println(r.Header)
	if "POST" == r.Method {
		v := r.Header.Get("Content-Type")
		if v == "" {
			fmt.Println("Content-Type error")
			return
		}
		d, params, err := mime.ParseMediaType(v)
		if err != nil || d != "multipart/form-data" {
			fmt.Println("Content-Type error")
			return
		}

		boundary, ok := params["boundary"]
		if !ok {
			fmt.Println("Content-Type")
			return
		}
		var key string
		var length int
		var start int
		var file *multipart.Part

		mReader := multipart.NewReader(r.Body, boundary)
		//to do try: r.MultipartReader()
		for {
			var part *multipart.Part
			part, err = mReader.NextPart()
			if err == io.EOF {
				break;
			}
			if err != nil {
				panic(err)
			}
			fmt.Println(part.FormName())
			if part.FormName() == "key" {
				var b bytes.Buffer
				_, err = io.Copy(&b, part)
				if err != nil {
					panic(err)
				}
				key = b.String()
			}
			if part.FormName() == "len" {
				var b bytes.Buffer
				_, err = io.Copy(&b, part)
				if err != nil {
					panic(err)
				}
				length, _ = strconv.Atoi(b.String())
			}

			if part.FormName() == "start" {
				var b bytes.Buffer
				_, err = io.Copy(&b, part)
				if err != nil {
					panic(err)
				}
				start, _ = strconv.Atoi(b.String())
			}

			if part.FormName() == "file" {
				file = part
				break
			}
		}

		fmt.Printf("upload start %d,total %d\n", start, length)
		sess := session.Must(session.NewSession(&aws.Config{Region: aws.String("us-east-1")}))
		svc := s3.New(sess)
		if (start == 1) {
			err = createMultipartUpload(w, key, length)
			if (err != nil) {
				http.Error(w, fmt.Sprintf("Unable to create multi upload %q, %v", key, err), 500)
				fmt.Println(fmt.Sprintf("Unable to create multi upload %q, %v", key, err))
				return
			}
			fmt.Println("create succ")
		}

		err = updateMultipartUpload(sess, r, w, key, svc, file, int64(start), int64(length))
		if err != nil {
			http.Error(w, fmt.Sprintf("Unable to upload part %q, %v", key, err), 500)
			fmt.Println(fmt.Sprintf("Unable to upload part %q, %v", key, err))
			return
		}
		fmt.Println("upload succ")

		err = completeMultipartUpload(w, key)
		if err != nil {
			http.Error(w, fmt.Sprintf("Unable to complete %q, %v", key, err), 500)
			fmt.Println(fmt.Sprintf("Unable to complete %q, %v", key, err))
			return
		}
		fmt.Println("complete succ")
	}
}

func UploadS3MultipartOffset(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "json")
	w.Header().Add("Access-Control-Allow-Origin", "*")
	if "GET" == r.Method {
		//r.ParseForm()
		fmt.Println(r.FormValue("key"))
		key := r.FormValue("key")
		if key == "" {
			http.Error(w, "file id is null", 500)
			fmt.Println("file id is null")
			return
		}

		sess := session.Must(session.NewSession(&aws.Config{Region: aws.String("us-east-1")}))
		svc := s3.New(sess)
		id := mockdb.Get(key).Id

		listInput := &s3.ListPartsInput{
			Bucket:   &myaws.Bucket,
			Key:      &key,
			UploadId: id,
		}
		fmt.Println(id)
		output, err := svc.ListParts(listInput)
		if err != nil {
			http.Error(w, fmt.Sprintf("Unable to list parts of %q, %v", key, err), 500)
			fmt.Println(fmt.Sprintf("Unable to list parts of %q, %v", key, err))
			return
		}
		fmt.Printf("upload status %q, %v\n", key, output)

		byte_count := int64(0)
		for _, part := range output.Parts {
			byte_count += *part.Size
		}
		x, err := json.Marshal(struct {
			OffsetByte  int64 `json:"offsetByte"`
			OffsetBlock int   `json:"offsetBlock"`
		}{byte_count, len(output.Parts) + 1})
		if err != nil {
			http.Error(w, fmt.Sprintf("Unable to marshal, %v", key, err), 500)
			fmt.Println(fmt.Sprintf("Unable to marshal, %v", key, err))
			return
		}
		w.Write(x)
		return
	}
}

func UploadS3MultipartStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "json")
	w.Header().Add("Access-Control-Allow-Origin", "*")
	if "GET" == r.Method {
		//r.ParseForm()
		fmt.Println(r.FormValue("key"))
		key := r.FormValue("key")
		if key == "" {
			http.Error(w, "file id is null", 500)
			fmt.Println("file id is null")
			return
		}

		sess := session.Must(session.NewSession(&aws.Config{Region: aws.String("us-east-1")}))
		svc := s3.New(sess)
		id := mockdb.Get(key).Id

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
		fmt.Printf("upload status %q, %v\n", key, output)
		x, err := json.Marshal(output)
		if err != nil {
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
	w.Header().Add("Access-Control-Allow-Origin", "*")
	if "POST" == r.Method {
		key := r.Form.Get("key")
		if key == "" {
			http.Error(w, "file id is null", 500)
			fmt.Println("file id is null")
			return
		}

		sess := session.Must(session.NewSession(&aws.Config{Region: aws.String("us-east-1")}))
		svc := s3.New(sess)
		id := mockdb.Get(key).Id

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
