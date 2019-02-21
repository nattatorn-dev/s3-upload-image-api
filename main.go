package main

import (
	"log"
	"net/http"

	"s3-upload-image-api/config"
	"s3-upload-image-api/uploader"
)

func main() {
	http.HandleFunc("/img/", uploader.Img)
	http.HandleFunc("/upload", uploader.Upload)
	// http.HandleFunc("/", handleStatic)

	log.Println("listening on port ", *config.HTTP_PORT)

	log.Fatal(http.ListenAndServe(":"+*config.HTTP_PORT, nil))
}

func handleIndex(w http.ResponseWriter, r *http.Request) {

}

//	static file server
// func handleStatic(w http.ResponseWriter, r *http.Request) {
// 	var uri string
// 	if r.URL.Path == "/" {
// 		uri = "index.html"
// 	} else {
// 		uri = r.URL.Path
// 	}

// 	path := filepath.Join(*config.HTTP_STATIC_ROOT, uri)

// 	http.ServeFile(w, r, path)
// }
