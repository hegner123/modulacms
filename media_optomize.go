package main

import (
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"strings"
    "golang.org/x/image/webp"
	"golang.org/x/image/draw"
)

func optimizeUpload(fSrc *os.File, fPath string) int {
	defer fSrc.Close()
	baseName := strings.TrimSuffix(fPath, filepath.Ext(fPath))

	src := decodeMedia(fSrc, fPath)
	if src == nil {
		return 1
	}

	in := []draw.Interpolator{draw.CatmullRom}
	dimensions := dbGetMediaDimensions("")
	images := []draw.Image{}
	for i, dx := range dimensions {
		in[0].Scale(images[i], image.Rect(0, 0, dx.Width, dx.Height), src, src.Bounds(), draw.Over, nil)
	}
	for i, im := range images {
		err := encodeMedia(im, fmt.Sprintf("%s-%v.%s",baseName, dimensions[i].Label,filepath.Ext(fPath)))
		if err == 1 {
			fmt.Printf("%v\n", err)

		}

	}
	return 0

}

func decodeMedia(fSrc *os.File, fName string) image.Image {
	ext := filepath.Ext(fName)
	switch ext {
	case "png":
		src, err := png.Decode(fSrc)
		if err != nil {
            fmt.Printf("%v\n",err)
		}
		return src
	case "jpg","jpeg":
		src, err := jpeg.Decode(fSrc)
		if err != nil {
            fmt.Printf("%v\n",err)
		}
		return src
	case "webp":
		src, err := webp.Decode(fSrc)
		if err != nil {
            fmt.Printf("%v\n",err)
		}
        return src
	case "gif":

		src, err := gif.Decode(fSrc)
		if err != nil {
            fmt.Printf("%v\n",err)
		}
		return src
	case "svg":
		return nil

	}
	return nil
}

func encodeMedia(image draw.Image, fName string) int {
	file, err := os.Create(fName)
	if err != nil {
		fmt.Printf("%s\n", err)
	}
	defer file.Close()

	ext := filepath.Ext(fName)
	switch ext {
	case "png":
		err := png.Encode(file, image)
		if err != nil {
            fmt.Printf("%v\n",err)
		}
		return 0
	case "jpg","jpeg":
		err := jpeg.Encode(file, image, nil)
		if err != nil {
            fmt.Printf("%v\n",err)
		}
		return 0
	case "webp":
		return 1
	case "gif":

		err := gif.Encode(file, image, nil)
		if err != nil {
            fmt.Printf("%v\n",err)
		}
		return 0
	case "svg":
		return 0

	}
	return 0

}
