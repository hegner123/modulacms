package media

import (
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"strings"

	db "github.com/hegner123/modulacms/internal/Db"
	utility "github.com/hegner123/modulacms/internal/Utility"
	"golang.org/x/image/draw"
	"golang.org/x/image/webp"
)

func OptimizeUpload(fSrc string, fPath string) []string {
	dbc := db.GetDb(db.Database{})
	file, err := os.Open(fSrc)
	if err != nil {
		utility.LogError("couldn't find tmp file ", err)
	}
	defer file.Close()
	baseName := strings.TrimSuffix(fPath, filepath.Ext(fPath))

	src := decodeMedia(file, fPath)

	in := []draw.Interpolator{draw.CatmullRom}

	dimensions, err := db.ListMediaDimension(dbc.Connection, dbc.Context)
	if err != nil {
		utility.LogError("failed to : ", err)
	}
	images := []draw.Image{}
	for i, dx := range *dimensions {
		in[0].Scale(images[i], image.Rect(0, 0, int(dx.Width.Int64), int(dx.Height.Int64)), src, src.Bounds(), draw.Over, nil)
	}
	var paths []string
	for _, im := range images {
		path := writeEncodeMedia(im, fmt.Sprintf("%s-%v.%s", baseName, im, filepath.Ext(fPath)))

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
