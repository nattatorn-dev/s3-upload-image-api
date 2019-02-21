package uploader

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"image"
	"image/jpeg"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"s3-upload-image-api/config"
	"s3-upload-image-api/util"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/nfnt/resize"
	"github.com/oliamb/cutter"
)

type File struct {
	Hash string `json:"hash"`
	URL  string `json:"url"`
}

var upload_to_s3 bool = true

func push_to_s3(slug string, image_name string, variant string, path string) {

	creds := credentials.NewStaticCredentials(*config.AWS_ACCESS_KEY_ID, *config.AWS_SECRET_ACCESS_KEY, "")
	cfg := aws.NewConfig().WithRegion(*config.AWS_REGION).WithCredentials(creds)

	svc := s3.New(
		session.New(),
		cfg,
	)

	file, err := os.Open(*config.FS_TEMP + slug + "_" + image_name + variant + ".jpg")
	if err != nil {
		fmt.Printf("err opening file: %s", err)
	}
	defer file.Close()

	fileInfo, _ := file.Stat()
	size := fileInfo.Size()
	buffer := make([]byte, size)

	file.Read(buffer)
	fileBytes := bytes.NewReader(buffer)
	fileType := http.DetectContentType(buffer)
	s3_path := path + slug + "_" + image_name + variant + ".jpg"
	params := &s3.PutObjectInput{
		Bucket:        aws.String(*config.S3_BUCKET),
		Key:           aws.String(s3_path),
		Body:          fileBytes,
		ContentLength: aws.Int64(size),
		ContentType:   aws.String(fileType),
	}

	_, err = svc.PutObject(params)
	if err != nil {
		fmt.Printf("bad response: %s", err)
	}
}

func Upload(w http.ResponseWriter, r *http.Request) {

	file, header, err := r.FormFile("file")

	if err != nil {
		fmt.Fprintln(w, err)
		return
	}
	defer file.Close()

	r.ParseForm()
	var schema string = strings.Join(r.Form["schema"], "")
	var path string = strings.Join(r.Form["path"], "")
	// var image_name string = strings.Join(r.Form["image_name"], "")
	var slug string = strings.Join(r.Form["slug"], "")

	image_name := RandomHash()

	out, err := os.Create(*config.FS_TEMP + slug + "_" + image_name + ".jpg")
	if err != nil {
		fmt.Fprintf(w, "Unable to create the file for writing. Check your write access privilege")
		return
	}
	defer out.Close()

	// write the content from POST to the file
	_, err = io.Copy(out, file)
	if err != nil {
		fmt.Fprintln(w, err)
	}

	fmt.Fprintf(w, "File uploaded successfully : ")
	fmt.Fprintf(w, header.Filename)

	// uploads original to s3
	if upload_to_s3 {
		push_to_s3(slug, image_name, "", path)
	}

	// open
	file, err = os.Open(*config.FS_TEMP + slug + "_" + image_name + ".jpg")
	if err != nil {
		log.Fatal(err)
	}

	// decode jpeg into image.Image
	img, err := jpeg.Decode(file)
	if err != nil {
		log.Fatal(err)
	}
	file.Close()

	// image bounds
	var bounds image.Rectangle = img.Bounds()
	var original_x int = bounds.Max.X
	var original_y int = bounds.Max.Y

	fmt.Println(" resizing and optimizing images '" + slug + "'\n")
	for i := 0; i < len(config.IMAGE_SCHEMA); i++ {

		if schema == config.IMAGE_SCHEMA[i].Schema {
			fmt.Println(" → " + slug + " → " + config.IMAGE_SCHEMA[i].Method)

			var requested_x int = config.IMAGE_SCHEMA[i].Width
			var requested_y int = config.IMAGE_SCHEMA[i].Height

			if config.IMAGE_SCHEMA[i].Method == "crop" {
				if original_x > original_y {
					requested_x = 0
				} else {
					requested_y = 0
				}
			} else {
				if original_x > original_y {
					requested_y = 0
				} else {
					requested_x = 0
				}
			}

			// resizing and cropping
			m := resize.Resize(uint(requested_x), uint(requested_y), img, resize.NearestNeighbor)

			if config.IMAGE_SCHEMA[i].Method == "crop" {
				m, err = cutter.Crop(m, cutter.Config{
					Width:  config.IMAGE_SCHEMA[i].Width,
					Height: config.IMAGE_SCHEMA[i].Height,
					Mode:   cutter.Centered,
				})
			}

			// saving file
			temp_filepath := *config.FS_TEMP + slug + "_" + image_name + "-" + config.IMAGE_SCHEMA[i].Variant + ".jpg"
			out, err := os.Create(temp_filepath)
			if err != nil {
				log.Fatal(err)
			}
			defer out.Close()

			// optimizing
			jpeg.Encode(out, m, &jpeg.Options{Quality: config.IMAGE_SCHEMA[i].Quality})

			// upload to s3
			if upload_to_s3 {
				push_to_s3(slug, image_name, "-"+config.IMAGE_SCHEMA[i].Variant, path)
				fmt.Println("    → uploaded to s3")

				// removing cached image from local disk
				_ = os.Remove(temp_filepath)
			}

		}
	}

	// removing cached image from local disk
	_ = os.Remove(*config.FS_TEMP + slug + "_" + image_name + ".jpg")

	fmt.Println("")

}

//	handles uploading of a file and generating a random file hash for it
func UploadHandler(w http.ResponseWriter, r *http.Request) {
	//	TODO: make this configuable so the domain can be locked down
	//	allow cross origin uploads
	w.Header().Set("Access-Control-Allow-Origin", "*")

	//	to make the cross domain upload cross browser comptatibale
	//	per https://github.com/blueimp/jQuery-File-Upload/wiki/Cross-domain-uploads
	if r.Method == "OPTIONS" {
		Success(w, nil)
		return
	}

	if r.Method != "POST" {
		Error(w, "expecting a POST request", http.StatusBadRequest)
		return
	}

	// the FormFile function takes in the POST input id file
	postData, _, err := r.FormFile("file")
	if err != nil {
		Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer postData.Close()

	//	hash for filename
	hash := RandomHash()

	//	new file path
	tmpPath := filepath.Join(*config.FS_TEMP, hash)

	//	temp file location
	tmpFile, err := os.Create(tmpPath)
	if err != nil {
		Error(w, "Unable to create the file for writing. Check your write access privilege", http.StatusInternalServerError)
		return
	}
	defer tmpFile.Close()

	// write the content from POST to the file
	_, err = io.Copy(tmpFile, postData)
	if err != nil {
		Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	//	move file to S3 bucket
	if err = MoveToS3(tmpPath, hash); err != nil {
		Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	//	clean up temp file
	if err = os.Remove(tmpPath); err != nil {
		Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	//	prepare response
	fileData := File{
		Hash: hash,
		URL:  *config.CDN + "/" + hash,
	}

	//	send response to client
	Success(w, fileData)
}

//	generates a random file hash
func RandomHash() string {
	//	generate a random string
	rand := util.GetRand()
	//	generate a md5 has from the random string
	return fmt.Sprintf("%x", md5.Sum([]byte(rand)))
}
