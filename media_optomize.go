package main

import (
	"fmt"
	"golang.org/x/image/draw"
	"golang.org/x/image/webp"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"strings"
)

func optimizeUpload(fSrc string, fPath string) []string {
	db, ctx, err := getDb(Database{})
	if err != nil {
		logError("failed to create database: ", err)
	}
	file, err := os.Open(fSrc)
	if err != nil {
		logError("couldn't find tmp file ", err)
	}
	defer file.Close()
	baseName := strings.TrimSuffix(fPath, filepath.Ext(fPath))

	src := decodeMedia(file, fPath)

	in := []draw.Interpolator{draw.CatmullRom}
	dimensions := dbListMediaDimension(db, ctx)
	images := []draw.Image{}
	for i, dx := range dimensions {
		in[0].Scale(images[i], image.Rect(0, 0, int(dx.Width.Int64), int(dx.Height.Int64)), src, src.Bounds(), draw.Over, nil)
	}
	var paths []string
	for i, im := range images {
		path := writeEncodeMedia(im, fmt.Sprintf("%s-%v.%s", baseName, dimensions[i].Label, filepath.Ext(fPath)))

		paths = append(paths, path)
	}
	return paths
}

func decodeMedia(fSrc *os.File, fName string) image.Image {
	ext := filepath.Ext(fName)
	switch ext {
	case "png":
		src, err := png.Decode(fSrc)
		if err != nil {
			fmt.Printf("%v\n", err)
		}
		return src
	case "jpg", "jpeg":
		src, err := jpeg.Decode(fSrc)
		if err != nil {
			fmt.Printf("%v\n", err)
		}
		return src
	case "webp":
		src, err := webp.Decode(fSrc)
		if err != nil {
			fmt.Printf("%v\n", err)
		}
		return src
	case "gif":

		src, err := gif.Decode(fSrc)
		if err != nil {
			fmt.Printf("%v\n", err)
		}
		return src
	case "svg":
		return nil

	}
	return nil
}

func writeEncodeMedia(image draw.Image, fName string) string {
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
			fmt.Printf("%v\n", err)
		}
		return fName
	case "jpg", "jpeg":
		err := jpeg.Encode(file, image, nil)
		if err != nil {
			fmt.Printf("%v\n", err)
		}
		return fName
	case "webp":
		return fName
	case "gif":

		err := gif.Encode(file, image, nil)
		if err != nil {
			fmt.Printf("%v\n", err)
		}
		return fName
	case "svg":
		return fName

	}
	return fName

}
