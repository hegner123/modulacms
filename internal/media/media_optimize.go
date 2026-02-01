package media

import (
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	config "github.com/hegner123/modulacms/internal/config"
	db "github.com/hegner123/modulacms/internal/db"
	webpenc "github.com/kolesa-team/go-webp/encoder"
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
	srcWidth := bounds.Dx()
	srcHeight := bounds.Dy()
	pixels := srcWidth * srcHeight

	if srcWidth > MaxImageWidth {
		return nil, fmt.Errorf("image width %d exceeds maximum %d", srcWidth, MaxImageWidth)
	}
	if srcHeight > MaxImageHeight {
		return nil, fmt.Errorf("image height %d exceeds maximum %d", srcHeight, MaxImageHeight)
	}
	if pixels > MaxImagePixels {
		return nil, fmt.Errorf("image size %d pixels exceeds maximum %d", pixels, MaxImagePixels)
	}

	// Copy the original file to dstPath
	files := []string{}
	originalDst := filepath.Join(dstPath, filepath.Base(srcFile))
	if err := copyFile(srcFile, originalDst); err != nil {
		return nil, fmt.Errorf("failed to copy original file: %w", err)
	}
	files = append(files, originalDst)

	// Initialize scaler.
	var scaler draw.Scaler = draw.BiLinear
	images := []draw.Image{}
	validDimensions := []db.MediaDimensions{}
	centerX := (bounds.Min.X + bounds.Max.X) / 2
	centerY := (bounds.Min.Y + bounds.Max.Y) / 2

	// Crop and scale images.
	for _, dim := range *dimensions {
		if !dim.Width.Valid || !dim.Height.Valid {
			continue
		}

		cropWidth := int(dim.Width.Int64)
		cropHeight := int(dim.Height.Int64)

		if cropWidth <= 0 || cropHeight <= 0 {
			continue
		}

		// Skip dimensions larger than the source to avoid upscaling
		if cropWidth > srcWidth || cropHeight > srcHeight {
			continue
		}

		x0 := centerX - cropWidth/2
		y0 := centerY - cropHeight/2
		cropRect := image.Rect(x0, y0, x0+cropWidth, y0+cropHeight)
		cropRect = cropRect.Intersect(bounds)

		dstRect := image.Rect(0, 0, cropWidth, cropHeight)
		img := image.NewRGBA(dstRect)
		scaler.Scale(img, dstRect, decodedImg, cropRect, draw.Over, nil)
		images = append(images, img)
		validDimensions = append(validDimensions, dim)
	}

	var optimizationErr error

	// Encode and save images.
	for i, img := range images {
		widthString := strconv.FormatInt(validDimensions[i].Width.Int64, 10)
		heightString := strconv.FormatInt(validDimensions[i].Height.Int64, 10)
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
		case ".webp":
			if opts, optErr := webpenc.NewLossyEncoderOptions(webpenc.PresetDefault, 80); optErr != nil {
				err = fmt.Errorf("webp options: %w", optErr)
			} else if enc, encErr := webpenc.NewEncoder(img, opts); encErr != nil {
				err = fmt.Errorf("webp encoder: %w", encErr)
			} else {
				err = enc.Encode(f)
			}
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

// copyFile copies a file from src to dst
func copyFile(src, dst string) (err error) {
	srcF, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source: %w", err)
	}
	defer srcF.Close()

	dstF, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("failed to create destination: %w", err)
	}
	defer func() {
		if cerr := dstF.Close(); cerr != nil && err == nil {
			err = cerr
		}
	}()

	_, err = io.Copy(dstF, srcF)
	return err
}
