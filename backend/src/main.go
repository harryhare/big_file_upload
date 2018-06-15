package main

import (
	"fmt"
	"io"
	"net/http"
	"log"
	"handler"
	"mockdb"
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



func main() {
	mockdb.Load()
	defer mockdb.Save()

	http.HandleFunc("/",Index)
	http.HandleFunc("/upload", handler.UploadLocal)
	http.HandleFunc("/uploadtos3",  handler.UploadS3)

	http.HandleFunc("/api/local/upload", handler.UploadLocal)
	http.HandleFunc("/api/local/list", handler.ListLocal)
	http.HandleFunc("/api/s3/upload",  handler.UploadS3)
	http.HandleFunc("/api/s3/list",  handler.ListS3)
	http.HandleFunc("/api/s3/multi/upload",  handler.UploadS3MultipartUpload)
	http.HandleFunc("/api/s3/multi/status",  handler.UploadS3MultipartStatus)
	http.HandleFunc("/api/s3/multi/stop",  handler.UploadS3MultipartStop)

	fmt.Println("listen on port 8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
