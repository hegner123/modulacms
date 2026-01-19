package media

import (
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	config "github.com/hegner123/modulacms/internal/config"
	db "github.com/hegner123/modulacms/internal/db"
	"golang.org/x/image/draw"
	"golang.org/x/image/webp"
)

// srcFile is the source file
// dstPath is the location of optimized files
func OptimizeUpload(srcFile string, dstPath string, c config.Config) (*[]string, error) {
	d := db.ConfigDB(c)

	// Open the source file.
	file, err := os.Open(srcFile)
	if err != nil {
		return nil, fmt.Errorf("couldn't find tmp file: %w", err)
	}
	defer file.Close()

	// Get the dimensions.
	dimensions, err := d.ListMediaDimensions()
	if err != nil {
		return nil, fmt.Errorf("failed to list media dimensions: %w", err)
	}
	if dimensions == nil {
		return nil, fmt.Errorf("dimensions list is nil")
	}

	baseName := strings.TrimSuffix(filepath.Base(srcFile), filepath.Ext(srcFile))
	ext := filepath.Ext(srcFile)

	// Decode the image.
	var decodedImg image.Image
	switch ext {
	case ".png":
		decodedImg, err = png.Decode(file)
	case ".jpg", ".jpeg":
		decodedImg, err = jpeg.Decode(file)
	case ".webp":
		decodedImg, err = webp.Decode(file)
	case ".gif":
		decodedImg, err = gif.Decode(file)
	default:
		return nil, fmt.Errorf("unsupported file extension: %s", ext)
	}
	if err != nil {
		return nil, fmt.Errorf("error decoding image: %w", err)
	}
	if decodedImg == nil {
		return nil, fmt.Errorf("decoded image is nil")
	}

	// Validate image dimensions to prevent memory exhaustion attacks
	bounds := decodedImg.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()
	pixels := width * height

	if width > MaxImageWidth {
		return nil, fmt.Errorf("image width %d exceeds maximum %d", width, MaxImageWidth)
	}
	if height > MaxImageHeight {
		return nil, fmt.Errorf("image height %d exceeds maximum %d", height, MaxImageHeight)
	}
	if pixels > MaxImagePixels {
		return nil, fmt.Errorf("image size %d pixels exceeds maximum %d", pixels, MaxImagePixels)
	}

	// Initialize scaler.
	var scaler draw.Scaler = draw.BiLinear
	images := []draw.Image{}
	centerX := (bounds.Min.X + bounds.Max.X) / 2
	centerY := (bounds.Min.Y + bounds.Max.Y) / 2

	// Crop and scale images.
	for _, dim := range *dimensions {
		cropWidth := int(dim.Width.Int64)
		cropHeight := int(dim.Height.Int64)
		x0 := centerX - cropWidth/2
		y0 := centerY - cropHeight/2
		cropRect := image.Rect(x0, y0, x0+cropWidth, y0+cropHeight)
		cropRect = cropRect.Intersect(bounds)

		dstRect := image.Rect(0, 0, cropWidth, cropHeight)
		img := image.NewRGBA(dstRect)
		scaler.Scale(img, dstRect, decodedImg, cropRect, draw.Over, nil)
		images = append(images, img)
	}

	files := []string{}
	var optimizationErr error

	// Encode and save images.
	for i, img := range images {
		widthString := strconv.FormatInt((*dimensions)[i].Width.Int64, 10)
		heightString := strconv.FormatInt((*dimensions)[i].Height.Int64, 10)
		size := widthString + "x" + heightString
		filename := fmt.Sprintf("%s-%v%s", baseName, size, ext)
		fullPath := filepath.Join(dstPath, filename)

		f, err := os.Create(fullPath)
		if err != nil {
			optimizationErr = fmt.Errorf("error creating file %s: %w", fullPath, err)
			break
		}

		// Encode image
		switch ext {
		case ".png":
			err = png.Encode(f, img)
		case ".jpg", ".jpeg":
			err = jpeg.Encode(f, img, nil)
		case ".gif":
			err = gif.Encode(f, img, nil)
		default:
			err = fmt.Errorf("unsupported encoding for extension: %s", ext)
		}

		f.Close()

		if err != nil {
			optimizationErr = fmt.Errorf("error encoding image %s: %w", filename, err)
			// Delete the partially written file
			os.Remove(fullPath)
			break
		}

		files = append(files, fullPath)
	}

	// If any optimization failed, clean up partial files
	if optimizationErr != nil {
		for _, file := range files {
			os.Remove(file)
		}
		return nil, optimizationErr
	}

	return &files, nil
}
