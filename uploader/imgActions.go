package uploader

import (
	"image"

	"github.com/nfnt/resize"
	"github.com/oliamb/cutter"
)

func Crop(query QueryDetails, img *image.Image) (cImg image.Image, err error) {
	//	crop our iamge
	cImg, err = cutter.Crop(*img, cutter.Config{
		Height: int(query.Height),
		Width:  int(query.Width),
		Mode:   cutter.Centered,
	})
	if err != nil {
		return
	}

	return
}

//	resizes an image, maintaining it's aspect ratio, into the w & h provided
//	excess image is then clipped away so the image fits within the w & h provided
func ResizeClip(query QueryDetails, img *image.Image, imgConf image.Config) (cImg image.Image, err error) {

	//	calculate the aspect ratio of the source image
	aspectRatio := float64(imgConf.Width) / float64(imgConf.Height)

	//	cast our width and height for use later
	aspectWidth := uint(query.Width)
	aspectHeight := uint(query.Height)

	//	check if we need to modify our width and height
	if aspectRatio != 1 {
		//	wider than native
		if (float64(query.Width) / float64(query.Height)) < aspectRatio {
			//	zero maintains the aspect ratio in our Resize lib
			aspectWidth = 0
		} else {
			aspectHeight = 0
		}
	}

	//	set the algo for resizing
	var algo resize.InterpolationFunction
	switch query.Algo {
	case "nearestNeighbor":
		algo = resize.NearestNeighbor
		break
	case "bilinear":
		algo = resize.Bilinear
		break
	case "mitchellNetravali":
		algo = resize.MitchellNetravali
		break
	case "lanczos2":
		algo = resize.Lanczos2
		break
	case "lanczos3":
		algo = resize.Lanczos3
		break
	default:
		algo = resize.NearestNeighbor
	}

	//	resize the image
	cImg = resize.Resize(aspectWidth, aspectHeight, *img, algo)

	//	clip extra edges
	cImg, err = cutter.Crop(cImg, cutter.Config{
		Height: int(query.Height),
		Width:  int(query.Width),
		Mode:   cutter.Centered,
	})
	if err != nil {
		return
	}

	return
}

//	resizes an image while maintaining it's aspect ratio
//	if the width and height are larger than the original image, the original image is preserved
func ResizePreserve(query QueryDetails, img *image.Image) (cImg image.Image, err error) {
	//	set the algo for resizing
	var algo resize.InterpolationFunction
	switch query.Algo {
	case "nearestNeighbor":
		algo = resize.NearestNeighbor
		break
	case "bilinear":
		algo = resize.Bilinear
		break
	case "mitchellNetravali":
		algo = resize.MitchellNetravali
		break
	case "lanczos2":
		algo = resize.Lanczos2
		break
	case "lanczos3":
		algo = resize.Lanczos3
		break
	default:
		algo = resize.NearestNeighbor
	}

	//	resize the image
	cImg = resize.Thumbnail(uint(query.Width), uint(query.Height), *img, algo)

	return
}
