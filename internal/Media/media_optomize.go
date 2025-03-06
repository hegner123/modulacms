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

	config "github.com/hegner123/modulacms/internal/Config"
	db "github.com/hegner123/modulacms/internal/Db"
	"golang.org/x/image/draw"
	"golang.org/x/image/webp"
)
func OptimizeUpload(fSrc string, fPath string, c config.Config) error {
	d := db.ConfigDB(c)

	// Open the source file.
	file, err := os.Open(fSrc)
	if err != nil {
		return fmt.Errorf("couldn't find tmp file: %w", err)
	}
	defer file.Close()

	// Get the dimensions.
	dimensions, err := d.ListMediaDimensions()
	if err != nil {
		return fmt.Errorf("failed to list media dimensions: %w", err)
	}
	if dimensions == nil {
		return fmt.Errorf("dimensions list is nil")
	}

	baseName := strings.TrimSuffix(fPath, filepath.Ext(fPath))
	ext := filepath.Ext(fSrc)

	// Decode the image.
	var dImg image.Image
	switch ext {
	case ".png":
		dImg, err = png.Decode(file)
	case ".jpg", ".jpeg":
		dImg, err = jpeg.Decode(file)
	case ".webp":
		dImg, err = webp.Decode(file)
	case ".gif":
		dImg, err = gif.Decode(file)
	default:
		return fmt.Errorf("unsupported file extension: %s", ext)
	}
	if err != nil {
		return fmt.Errorf("error decoding image: %w", err)
	}
	if dImg == nil {
		return fmt.Errorf("decoded image is nil")
	}

	// Initialize scaler.
	var in draw.Scaler = draw.BiLinear
	images := []draw.Image{}
	bounds := dImg.Bounds()
	centerX := (bounds.Min.X + bounds.Max.X) / 2
	centerY := (bounds.Min.Y + bounds.Max.Y) / 2

	// Crop and scale images.
	for _, dx := range *dimensions {
		cropWidth := int(dx.Width.Int64)
		cropHeight := int(dx.Height.Int64)
		x0 := centerX - cropWidth/2
		y0 := centerY - cropHeight/2
		cropRect := image.Rect(x0, y0, x0+cropWidth, y0+cropHeight)
		cropRect = cropRect.Intersect(bounds)

		dstRect := image.Rect(0, 0, cropWidth, cropHeight)
		img := image.NewRGBA(dstRect)
		in.Scale(img, dstRect, dImg, cropRect, draw.Over, nil)
		images = append(images, img)
	}

	// Encode and save images.
	for i, im := range images {
		filename := fmt.Sprintf("%s-%v%s", baseName, i, ext)
		f, err := os.Create(filename)
		if err != nil {
			return fmt.Errorf("error creating file %s: %w", filename, err)
		}
		// Ensure the file is closed after encoding.
		defer f.Close()

		switch ext {
		case ".png":
			err = png.Encode(f, im)
		case ".jpg", ".jpeg":
			err = jpeg.Encode(f, im, nil)
		case ".gif":
			err = gif.Encode(f, im, nil)
		default:
			// In theory, this case won't be reached due to the earlier switch.
			err = fmt.Errorf("unsupported encoding for extension: %s", ext)
		}
		if err != nil {
			return fmt.Errorf("error encoding image %s: %w", filename, err)
		}
	}
	return nil
}

