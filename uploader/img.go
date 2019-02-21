package uploader

import (
	"errors"
	"image"
	"image/jpeg"
	"image/png"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"

	"s3-upload-image-api/config"
)

type QueryDetails struct {
	Action string
	Width  int64
	Height int64
	Algo   string //	resizing / thumbnail algo
}

//	parse a URL extracting information necessary to render the file
//	pull uri params and ext hbvsdn.png?w=300&h=300&tag=test&bg=#000&ver=3
//		w = width in px
//		h = height in px
//		ver = version index. if not provided will be set to -1
//		bg = background color in hex format [not implemented]
//		tag = a tracking tag [not implemented]
func (this *QueryDetails) Parse(url *url.URL) (err error) {
	//	parse our request url
	vals := url.Query()

	//	check action
	if _, ok := vals["action"]; ok {
		this.Action = vals["action"][0]
	} else {
		return errors.New("no action provided")
	}

	//	check width
	if _, ok := vals["w"]; ok {
		//	parse width
		this.Width, err = strconv.ParseInt(vals["w"][0], 10, 0)
		if err != nil {
			return
		}
	} else {
		this.Width = 450
	}

	//	check for height
	if _, ok := vals["h"]; ok {
		//	parse height
		this.Height, err = strconv.ParseInt(vals["h"][0], 10, 0)
		if err != nil {
			return
		}
	} else {
		this.Height = 450
	}

	//	check for algo
	if _, ok := vals["algo"]; ok {
		this.Algo = vals["algo"][0]
	}

	return
}

func Img(w http.ResponseWriter, r *http.Request) {
	var err error

	//	extract our key
	key := r.URL.Path[len("/img/"):]

	//	confirm we have key
	if key == "" {
		Error(w, "no key provided", http.StatusBadRequest)
		return
	}

	//	download file from s3
	tmpFilePath, err := FetchFromS3(key, *config.S3_BUCKET)
	if err != nil {
		Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	//	clean up our tmp file
	defer os.Remove(tmpFilePath)

	img, imgFormat, imgConf, err := Decode(tmpFilePath)
	if err != nil {
		Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	//	parse the request params
	query := QueryDetails{}
	if err = query.Parse(r.URL); err != nil {
		Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	//	our modified image
	var cImg image.Image
	switch query.Action {
	case "crop":
		cImg, err = Crop(query, &img)
		if err != nil {
			Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		break
	case "resize-preserve":
		cImg, err = ResizePreserve(query, &img)
		if err != nil {
			Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		break
	case "resize-clip":
		//	resizeClip
		cImg, err = ResizeClip(query, &img, imgConf)
		if err != nil {
			Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		break
	default:
		Error(w, "action not supported: "+query.Action, http.StatusInternalServerError)
		return
	}

	//	encode our image
	modImgPath, err := EncodeImage(&cImg, imgFormat)
	if err != nil {
		Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	//	clean up the created file
	defer os.Remove(modImgPath)

	//	set cache control headers
	//	fastly.com CDN key used for purging
	w.Header().Set("Surrogate-Key", key)

	//	cache the asset with fastly for 30 days
	//
	//	reference: https://docs.fastly.com/guides/tutorials/cache-control-tutorial
	w.Header().Set("Cache-Control", "max-age=2592000")
	w.Header().Set("Surrogate-Control", "max-age=2592000")

	//	send our file in the response
	http.ServeFile(w, r, modImgPath)
}

//	opens an image file, decodes it, extracts meta data
//	returns
//		pointer to image,
//		image format (jpeg, png, etc),
//		image config data (i.e. width, height, colorModel)
func Decode(tmpFilePath string) (img image.Image, imgFormat string, imgConf image.Config, err error) {
	tmpImgFile, err := os.Open(tmpFilePath)
	defer tmpImgFile.Close()
	if err != nil {
		return
	}

	//	fetch image meta data, i.e. width & height
	imgConf, _, err = image.DecodeConfig(tmpImgFile)
	if err != nil {
		return
	}

	//	move back to the beginning of the file so we can decode it next
	_, err = tmpImgFile.Seek(0, 0)
	if err != nil {
		return
	}

	//	decode our image
	img, imgFormat, err = image.Decode(tmpImgFile)
	if err != nil {
		return
	}

	return
}

//	creates a temporary image files and encodes it
//	based on the passed format
func EncodeImage(img *image.Image, format string) (imgPath string, err error) {
	//	temp file location
	imgPath = filepath.Join(*config.FS_TEMP, RandomHash())

	//	create a new file
	file, err := os.Create(imgPath)
	if err != nil {
		return
	}

	defer file.Close()

	switch format {
	case "jpeg":
		if err = jpeg.Encode(file, *img, &jpeg.Options{jpeg.DefaultQuality}); err != nil {
			return
		}
	case "png":
		if err = png.Encode(file, *img); err != nil {
			return
		}
	default:
		err = errors.New("image format not supported")
		return
	}

	return
}
