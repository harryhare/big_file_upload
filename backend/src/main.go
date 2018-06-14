package main

import (
	"fmt"
	"io"
	"net/http"
	"log"
	"os"
)

// 获取大小的借口
type Sizer interface {
	Size() int64
}

// hello world, the web server
func HelloServer(w http.ResponseWriter, r *http.Request) {
	if "POST" == r.Method {
		w.Header().Add("Content-Type", "text/html")
		w.Header().Add("Access-Control-Allow-Origin","*")
		file, _, err := r.FormFile("file")
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		defer file.Close()
		f,err:=os.Create("filenametosaveas")
		defer f.Close()
		x,err:=io.Copy(f,file)
		if err!=nil {
			http.Error(w, err.Error(), 500)
		}
		fmt.Fprintf(w, "上传文件的大小为: %d", x)
		return
	}

	// 上传页面
	w.Header().Add("Content-Type", "text/html")
	w.Header().Add("Access-Control-Allow-Origin","*")
	w.WriteHeader(200)
	html := `
		<form enctype="multipart/form-data" action="/hello" method="POST">
			Send this file: <input name="file" type="file" />
			<input type="submit" value="Send File" />
		</form>`

	io.WriteString(w, html)
}

func main() {
	http.HandleFunc("/hello", HelloServer)
	fmt.Println("listen on port 8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
