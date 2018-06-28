package handler

import (
	"net/http"
	"os"
	"io"
	"fmt"
)

func UploadLocal(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/html")
	w.Header().Add("Access-Control-Allow-Origin", "*")
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
func ListLocal(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/html")
	w.Header().Add("Access-Control-Allow-Origin", "*")
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
