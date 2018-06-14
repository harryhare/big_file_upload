package main

import (
	"fmt"
	"net/http"
	"log"
	"wechat_handler"
	"intercom_handler"
	"wechat"
	"io/ioutil"
)


func Index(w http.ResponseWriter, r *http.Request) {
	fmt.Println(r)
	fmt.Fprint(w, "hello")
}
func Upload(w http.ResponseWriter, r *http.Request) {
	b=ioutil.ReadAll(r.Body)

}


func main() {
	http.HandleFunc("/", Index)
	http.HandleFunc("/api/attachment/upload", Upload)

	fmt.Println("listen on port 8080")
	err := http.ListenAndServe(":8080", nil);
	if  err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}