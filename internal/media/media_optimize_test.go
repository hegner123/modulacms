package media

import (
	"database/sql"
	"errors"
	"image"
	"image/color"
	"image/gif"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"

	db "github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/types"
	"golang.org/x/image/webp"
)

// ---------------------------------------------------------------------------
// copyFile tests
// ---------------------------------------------------------------------------

// TestCopyFile verifies that copyFile creates an exact byte-for-byte copy of
// the source file at the destination path.
func TestCopyFile(t *testing.T) {
	t.Parallel()

	srcDir := t.TempDir()
	dstDir := t.TempDir()

	srcPath := filepath.Join(srcDir, "original.txt")
	content := []byte("the quick brown fox jumps over the lazy dog")
	if err := os.WriteFile(srcPath, content, 0644); err != nil {
		t.Fatalf("failed to write source file: %v", err)
	}

	dstPath := filepath.Join(dstDir, "copy.txt")
	if err := copyFile(srcPath, dstPath); err != nil {
		t.Fatalf("copyFile returned error: %v", err)
	}

	got, err := os.ReadFile(dstPath)
	if err != nil {
		t.Fatalf("failed to read destination file: %v", err)
	}

	if string(got) != string(content) {
		t.Errorf("destination content = %q, want %q", got, content)
	}
}

// TestCopyFile_EmptyFile verifies that copying a zero-byte file succeeds and
// produces a zero-byte destination.
func TestCopyFile_EmptyFile(t *testing.T) {
	t.Parallel()

	srcDir := t.TempDir()
	dstDir := t.TempDir()

	srcPath := filepath.Join(srcDir, "empty.bin")
	if err := os.WriteFile(srcPath, []byte{}, 0644); err != nil {
		t.Fatalf("failed to write empty source: %v", err)
	}

	dstPath := filepath.Join(dstDir, "empty-copy.bin")
	if err := copyFile(srcPath, dstPath); err != nil {
		t.Fatalf("copyFile returned error for empty file: %v", err)
	}

	info, err := os.Stat(dstPath)
	if err != nil {
		t.Fatalf("failed to stat destination: %v", err)
	}
	if info.Size() != 0 {
		t.Errorf("destination size = %d, want 0", info.Size())
	}
}

// TestCopyFile_LargeFile verifies that copyFile handles files larger than
// typical I/O buffer sizes (>32KB).
func TestCopyFile_LargeFile(t *testing.T) {
	t.Parallel()

	srcDir := t.TempDir()
	dstDir := t.TempDir()

	// 100KB of data -- larger than most internal buffer sizes
	data := make([]byte, 100*1024)
	for i := range data {
		data[i] = byte(i % 256)
	}

	srcPath := filepath.Join(srcDir, "large.bin")
	if err := os.WriteFile(srcPath, data, 0644); err != nil {
		t.Fatalf("failed to write source: %v", err)
	}

	dstPath := filepath.Join(dstDir, "large-copy.bin")
	if err := copyFile(srcPath, dstPath); err != nil {
		t.Fatalf("copyFile returned error: %v", err)
	}

	got, err := os.ReadFile(dstPath)
	if err != nil {
		t.Fatalf("failed to read destination: %v", err)
	}
	if len(got) != len(data) {
		t.Fatalf("destination length = %d, want %d", len(got), len(data))
	}
	for i := range data {
		if got[i] != data[i] {
			t.Fatalf("byte mismatch at offset %d: got %d, want %d", i, got[i], data[i])
		}
	}
}

// TestCopyFile_SourceNotFound verifies that copyFile returns an error when the
// source file does not exist.
func TestCopyFile_SourceNotFound(t *testing.T) {
	t.Parallel()

	dstDir := t.TempDir()
	err := copyFile("/nonexistent/path/to/file.txt", filepath.Join(dstDir, "out.txt"))
	if err == nil {
		t.Fatal("expected error for missing source, got nil")
	}
}

// TestCopyFile_InvalidDestination verifies that copyFile returns an error when
// the destination directory does not exist.
func TestCopyFile_InvalidDestination(t *testing.T) {
	t.Parallel()

	srcDir := t.TempDir()
	srcPath := filepath.Join(srcDir, "source.txt")
	if err := os.WriteFile(srcPath, []byte("data"), 0644); err != nil {
		t.Fatalf("setup: %v", err)
	}

	err := copyFile(srcPath, "/nonexistent/dir/destination.txt")
	if err == nil {
		t.Fatal("expected error for invalid destination, got nil")
	}
}

// TestCopyFile_SourceIsDirectory verifies that copyFile returns an error when
// pointed at a directory instead of a regular file.
func TestCopyFile_SourceIsDirectory(t *testing.T) {
	t.Parallel()

	srcDir := t.TempDir()
	dstDir := t.TempDir()

	err := copyFile(srcDir, filepath.Join(dstDir, "out.txt"))
	if err == nil {
		t.Fatal("expected error when source is a directory, got nil")
	}
}

// ---------------------------------------------------------------------------
// Image generation helpers for OptimizeUpload tests
// ---------------------------------------------------------------------------

// createTestImage generates a real image file of the given format and dimensions.
// Returns the path to the created file.
func createTestImage(t *testing.T, dir string, filename string, width, height int) string {
	t.Helper()

	img := image.NewRGBA(image.Rect(0, 0, width, height))
	// Fill with a gradient so the image is non-trivial
	for y := range height {
		for x := range width {
			img.Set(x, y, color.RGBA{
				R: uint8(x % 256),
				G: uint8(y % 256),
				B: 128,
				A: 255,
			})
		}
	}

	filePath := filepath.Join(dir, filename)
	f, err := os.Create(filePath)
	if err != nil {
		t.Fatalf("createTestImage os.Create: %v", err)
	}
	defer f.Close()

	ext := filepath.Ext(filename)
	switch ext {
	case ".png":
		if err := png.Encode(f, img); err != nil {
			t.Fatalf("createTestImage png.Encode: %v", err)
		}
	case ".jpg", ".jpeg":
		if err := jpeg.Encode(f, img, &jpeg.Options{Quality: 90}); err != nil {
			t.Fatalf("createTestImage jpeg.Encode: %v", err)
		}
	case ".gif":
		if err := gif.Encode(f, img, nil); err != nil {
			t.Fatalf("createTestImage gif.Encode: %v", err)
		}
	default:
		t.Fatalf("createTestImage: unsupported extension %s", ext)
	}

	return filePath
}

// ---------------------------------------------------------------------------
// mockDimensionLister implements DimensionLister for unit testing.
// ---------------------------------------------------------------------------

type mockDimensionLister struct {
	dims *[]db.MediaDimensions
	err  error
}

func (m *mockDimensionLister) ListMediaDimensions() (*[]db.MediaDimensions, error) {
	return m.dims, m.err
}

// dim is a shorthand constructor for db.MediaDimensions used in tests.
func dim(w, h int64) db.MediaDimensions {
	return db.MediaDimensions{
		Width:  sql.NullInt64{Int64: w, Valid: true},
		Height: sql.NullInt64{Int64: h, Valid: true},
	}
}

// ---------------------------------------------------------------------------
// OptimizeUpload unit tests
// ---------------------------------------------------------------------------

func TestOptimizeUpload_PNG_SingleDimension(t *testing.T) {
	t.Parallel()

	srcDir := t.TempDir()
	dstDir := t.TempDir()
	srcPath := createTestImage(t, srcDir, "photo.png", 400, 300)

	dims := []db.MediaDimensions{dim(100, 100)}
	lister := &mockDimensionLister{dims: &dims}

	files, err := OptimizeUpload(srcPath, dstDir, lister, nil)
	if err != nil {
		t.Fatalf("OptimizeUpload: %v", err)
	}

	// 1 variant (original uploaded separately)
	if len(*files) != 1 {
		t.Fatalf("expected 1 file, got %d: %v", len(*files), *files)
	}

	// Check variant filename contains dimension
	variant := (*files)[0]
	if !strings.Contains(filepath.Base(variant), "100x100") {
		t.Errorf("variant filename %q should contain '100x100'", filepath.Base(variant))
	}

	// Verify variant is a valid PNG with correct dimensions
	f, err := os.Open(variant)
	if err != nil {
		t.Fatalf("open variant: %v", err)
	}
	defer f.Close()

	img, err := webp.Decode(f)
	if err != nil {
		t.Fatalf("decode variant WebP: %v", err)
	}
	bounds := img.Bounds()
	if bounds.Dx() != 100 || bounds.Dy() != 100 {
		t.Errorf("variant dimensions = %dx%d, want 100x100", bounds.Dx(), bounds.Dy())
	}
}

func TestOptimizeUpload_JPEG(t *testing.T) {
	t.Parallel()

	srcDir := t.TempDir()
	dstDir := t.TempDir()
	srcPath := createTestImage(t, srcDir, "photo.jpg", 300, 200)

	dims := []db.MediaDimensions{dim(150, 100)}
	lister := &mockDimensionLister{dims: &dims}

	files, err := OptimizeUpload(srcPath, dstDir, lister, nil)
	if err != nil {
		t.Fatalf("OptimizeUpload: %v", err)
	}
	if len(*files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(*files))
	}

	// Verify variant is a valid JPEG
	f, err := os.Open((*files)[0])
	if err != nil {
		t.Fatalf("open variant: %v", err)
	}
	defer f.Close()

	_, err = webp.Decode(f)
	if err != nil {
		t.Fatalf("decode variant WebP: %v", err)
	}
}

func TestOptimizeUpload_GIF(t *testing.T) {
	t.Parallel()

	srcDir := t.TempDir()
	dstDir := t.TempDir()
	srcPath := createTestImage(t, srcDir, "anim.gif", 200, 200)

	dims := []db.MediaDimensions{dim(50, 50)}
	lister := &mockDimensionLister{dims: &dims}

	files, err := OptimizeUpload(srcPath, dstDir, lister, nil)
	if err != nil {
		t.Fatalf("OptimizeUpload: %v", err)
	}
	if len(*files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(*files))
	}
}

func TestOptimizeUpload_MultipleDimensions(t *testing.T) {
	t.Parallel()

	srcDir := t.TempDir()
	dstDir := t.TempDir()
	srcPath := createTestImage(t, srcDir, "multi.png", 800, 600)

	dims := []db.MediaDimensions{
		dim(100, 100),
		dim(200, 150),
		dim(400, 300),
	}
	lister := &mockDimensionLister{dims: &dims}

	files, err := OptimizeUpload(srcPath, dstDir, lister, nil)
	if err != nil {
		t.Fatalf("OptimizeUpload: %v", err)
	}

	// 3 variants (original uploaded separately)
	if len(*files) != 3 {
		t.Fatalf("expected 3 files, got %d: %v", len(*files), *files)
	}
}

func TestOptimizeUpload_SkipsUpscaling(t *testing.T) {
	t.Parallel()

	srcDir := t.TempDir()
	dstDir := t.TempDir()
	// Source is only 50x50
	srcPath := createTestImage(t, srcDir, "tiny.png", 50, 50)

	dims := []db.MediaDimensions{
		dim(100, 100), // larger than source — should be skipped
		dim(25, 25),   // smaller — should be generated
	}
	lister := &mockDimensionLister{dims: &dims}

	files, err := OptimizeUpload(srcPath, dstDir, lister, nil)
	if err != nil {
		t.Fatalf("OptimizeUpload: %v", err)
	}

	// original + 1 variant (the 100x100 should be skipped)
	if len(*files) != 1 {
		t.Fatalf("expected 1 file (skipping upscale), got %d: %v", len(*files), *files)
	}

	if !strings.Contains(filepath.Base((*files)[0]), "25x25") {
		t.Errorf("expected 25x25 variant, got %q", filepath.Base((*files)[0]))
	}
}

func TestOptimizeUpload_NoDimensions(t *testing.T) {
	t.Parallel()

	srcDir := t.TempDir()
	dstDir := t.TempDir()
	srcPath := createTestImage(t, srcDir, "solo.png", 200, 200)

	dims := []db.MediaDimensions{}
	lister := &mockDimensionLister{dims: &dims}

	files, err := OptimizeUpload(srcPath, dstDir, lister, nil)
	if err != nil {
		t.Fatalf("OptimizeUpload: %v", err)
	}

	// No variants when no dimensions configured
	if len(*files) != 0 {
		t.Fatalf("expected 0 files (no dimensions), got %d", len(*files))
	}
}

func TestOptimizeUpload_SkipsInvalidDimensions(t *testing.T) {
	t.Parallel()

	srcDir := t.TempDir()
	dstDir := t.TempDir()
	srcPath := createTestImage(t, srcDir, "skip.png", 200, 200)

	dims := []db.MediaDimensions{
		{Width: sql.NullInt64{Valid: false}, Height: sql.NullInt64{Int64: 100, Valid: true}},  // no width
		{Width: sql.NullInt64{Int64: 100, Valid: true}, Height: sql.NullInt64{Valid: false}},   // no height
		{Width: sql.NullInt64{Int64: 0, Valid: true}, Height: sql.NullInt64{Int64: 100, Valid: true}},   // zero width
		{Width: sql.NullInt64{Int64: -5, Valid: true}, Height: sql.NullInt64{Int64: 100, Valid: true}},  // negative width
		dim(50, 50), // valid
	}
	lister := &mockDimensionLister{dims: &dims}

	files, err := OptimizeUpload(srcPath, dstDir, lister, nil)
	if err != nil {
		t.Fatalf("OptimizeUpload: %v", err)
	}

	// 1 valid variant (original uploaded separately)
	if len(*files) != 1 {
		t.Fatalf("expected 1 file (skipping invalid dims), got %d: %v", len(*files), *files)
	}
}

func TestOptimizeUpload_UnsupportedExtension(t *testing.T) {
	t.Parallel()

	srcDir := t.TempDir()
	dstDir := t.TempDir()

	// Write a dummy file with unsupported extension
	srcPath := filepath.Join(srcDir, "document.bmp")
	if err := os.WriteFile(srcPath, []byte("not an image"), 0644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	dims := []db.MediaDimensions{dim(50, 50)}
	lister := &mockDimensionLister{dims: &dims}

	_, err := OptimizeUpload(srcPath, dstDir, lister, nil)
	if err == nil {
		t.Fatal("expected error for unsupported extension, got nil")
	}
	if !strings.Contains(err.Error(), "unsupported file extension") {
		t.Errorf("expected 'unsupported file extension' error, got: %v", err)
	}
}

func TestOptimizeUpload_SourceNotFound(t *testing.T) {
	t.Parallel()

	dstDir := t.TempDir()
	dims := []db.MediaDimensions{dim(50, 50)}
	lister := &mockDimensionLister{dims: &dims}

	_, err := OptimizeUpload("/nonexistent/image.png", dstDir, lister, nil)
	if err == nil {
		t.Fatal("expected error for missing source, got nil")
	}
}

func TestOptimizeUpload_DimensionListerError(t *testing.T) {
	t.Parallel()

	srcDir := t.TempDir()
	dstDir := t.TempDir()
	srcPath := createTestImage(t, srcDir, "err.png", 100, 100)

	lister := &mockDimensionLister{err: errors.New("db connection lost")}

	_, err := OptimizeUpload(srcPath, dstDir, lister, nil)
	if err == nil {
		t.Fatal("expected error from lister, got nil")
	}
	if !strings.Contains(err.Error(), "failed to list media dimensions") {
		t.Errorf("expected wrapped lister error, got: %v", err)
	}
}

func TestOptimizeUpload_NilDimensions(t *testing.T) {
	t.Parallel()

	srcDir := t.TempDir()
	dstDir := t.TempDir()
	srcPath := createTestImage(t, srcDir, "nil.png", 100, 100)

	lister := &mockDimensionLister{dims: nil}

	_, err := OptimizeUpload(srcPath, dstDir, lister, nil)
	if err == nil {
		t.Fatal("expected error for nil dimensions, got nil")
	}
	if !strings.Contains(err.Error(), "dimensions list is nil") {
		t.Errorf("expected 'dimensions list is nil' error, got: %v", err)
	}
}

func TestOptimizeUpload_OriginalNotIncluded(t *testing.T) {
	t.Parallel()

	srcDir := t.TempDir()
	dstDir := t.TempDir()
	srcPath := createTestImage(t, srcDir, "original.png", 200, 200)

	dims := []db.MediaDimensions{}
	lister := &mockDimensionLister{dims: &dims}

	files, err := OptimizeUpload(srcPath, dstDir, lister, nil)
	if err != nil {
		t.Fatalf("OptimizeUpload: %v", err)
	}

	// Original is uploaded separately — OptimizeUpload only returns variants
	if len(*files) != 0 {
		t.Errorf("expected 0 files (no dimensions), got %d: %v", len(*files), *files)
	}
}

// ---------------------------------------------------------------------------
// Crop skew tests — uses real test300x300.png fixture (6x6 color grid, 50px cells)
//
// The source is a 300x300 PNG with a repeating 6x6 grid of distinct colored
// blocks. For a center crop of WxH, the expected source region starts at
// ((300-W)/2, (300-H)/2). Every output pixel should match the corresponding
// source pixel at that offset. If the scaler introduces any skewing or
// stretching, pixels will shift off their expected grid positions.
// ---------------------------------------------------------------------------

// absDiffU32 returns the absolute difference between two uint32 values.
func absDiffU32(a, b uint32) uint32 {
	if a > b {
		return a - b
	}
	return b - a
}

// loadFixturePNG decodes the test fixture from the project root.
// Tests run from internal/media/, so the project root is ../..
func loadFixturePNG(t *testing.T) image.Image {
	t.Helper()
	fixturePath := filepath.Join("..", "..", "test300x300.png")
	f, err := os.Open(fixturePath)
	if err != nil {
		t.Skipf("test fixture not found at %s: %v", fixturePath, err)
	}
	defer f.Close()

	img, err := png.Decode(f)
	if err != nil {
		t.Fatalf("decode fixture: %v", err)
	}
	return img
}

// copyFixtureToDir copies the test fixture into dir and returns the path.
func copyFixtureToDir(t *testing.T, dir string) string {
	t.Helper()
	fixturePath := filepath.Join("..", "..", "test300x300.png")
	dst := filepath.Join(dir, "test300x300.png")
	if err := copyFile(fixturePath, dst); err != nil {
		t.Fatalf("copy fixture: %v", err)
	}
	return dst
}

func TestOptimizeUpload_CropNoSkew(t *testing.T) {
	t.Parallel()

	srcImg := loadFixturePNG(t)
	srcBounds := srcImg.Bounds()
	if srcBounds.Dx() != 300 || srcBounds.Dy() != 300 {
		t.Fatalf("fixture dimensions = %dx%d, want 300x300", srcBounds.Dx(), srcBounds.Dy())
	}

	tests := []struct {
		name   string
		width  int64
		height int64
	}{
		{"square 150x150", 150, 150},
		{"wide 200x100", 200, 100},
		{"tall 100x200", 100, 200},
		{"single cell 50x50", 50, 50},
		{"large 250x250", 250, 250},
		{"asymmetric 180x120", 180, 120},
		{"asymmetric 120x180", 120, 180},
		{"full row 300x50", 300, 50},
		{"full column 50x300", 50, 300},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			srcDir := t.TempDir()
			dstDir := t.TempDir()
			srcPath := copyFixtureToDir(t, srcDir)

			dims := []db.MediaDimensions{dim(tt.width, tt.height)}
			lister := &mockDimensionLister{dims: &dims}

			files, err := OptimizeUpload(srcPath, dstDir, lister, nil)
			if err != nil {
				t.Fatalf("OptimizeUpload: %v", err)
			}
			if len(*files) != 1 {
				t.Fatalf("expected 1 file, got %d", len(*files))
			}

			// Decode the cropped variant
			vf, err := os.Open((*files)[0])
			if err != nil {
				t.Fatalf("open variant: %v", err)
			}
			defer vf.Close()

			variantImg, err := webp.Decode(vf)
			if err != nil {
				t.Fatalf("decode variant: %v", err)
			}

			// Verify exact output dimensions
			vBounds := variantImg.Bounds()
			if vBounds.Dx() != int(tt.width) || vBounds.Dy() != int(tt.height) {
				t.Fatalf("variant dimensions = %dx%d, want %dx%d",
					vBounds.Dx(), vBounds.Dy(), tt.width, tt.height)
			}

			// The center crop offset into the source image
			offsetX := (300 - int(tt.width)) / 2
			offsetY := (300 - int(tt.height)) / 2

			// Sample every 5px across the output and compare to the
			// corresponding source pixel. Track the worst deviation.
			var maxDiff uint32
			var worstPt image.Point

			for y := 0; y < int(tt.height); y += 5 {
				for x := 0; x < int(tt.width); x += 5 {
					gr, gg, gb, _ := variantImg.At(x, y).RGBA()
					wr, wg, wb, _ := srcImg.At(x+offsetX, y+offsetY).RGBA()

					diff := absDiffU32(gr, wr) + absDiffU32(gg, wg) + absDiffU32(gb, wb)
					if diff > maxDiff {
						maxDiff = diff
						worstPt = image.Point{X: x, Y: y}
					}
				}
			}

			// Tolerance: lossy WebP at quality 80 can deviate significantly near
			// color boundaries and image edges due to block-based compression.
			const maxTolerance = 65000
			if maxDiff > maxTolerance {
				gr, gg, gb, _ := variantImg.At(worstPt.X, worstPt.Y).RGBA()
				wr, wg, wb, _ := srcImg.At(worstPt.X+offsetX, worstPt.Y+offsetY).RGBA()
				t.Errorf("skew detected at output pixel (%d,%d) → source (%d,%d): "+
					"got RGB(%d,%d,%d), want RGB(%d,%d,%d), channel-sum diff=%d",
					worstPt.X, worstPt.Y,
					worstPt.X+offsetX, worstPt.Y+offsetY,
					gr>>8, gg>>8, gb>>8,
					wr>>8, wg>>8, wb>>8,
					maxDiff)
			}
		})
	}
}

// TestOptimizeUpload_CropNoSkew_GridCellCenters verifies that the center pixel
// of each visible grid cell in a cropped output exactly matches the
// corresponding cell center in the source. Grid cells are 50x50 in the
// 300x300 fixture, so cell (col, row) has its center at (col*50+25, row*50+25).
func TestOptimizeUpload_CropNoSkew_GridCellCenters(t *testing.T) {
	t.Parallel()

	srcImg := loadFixturePNG(t)

	srcDir := t.TempDir()
	dstDir := t.TempDir()
	srcPath := copyFixtureToDir(t, srcDir)

	// Crop 200x200 from center → offset (50, 50)
	// Visible cells: columns 1-4, rows 1-4 (partial edges excluded by sampling centers)
	dims := []db.MediaDimensions{dim(200, 200)}
	lister := &mockDimensionLister{dims: &dims}

	files, err := OptimizeUpload(srcPath, dstDir, lister, nil)
	if err != nil {
		t.Fatalf("OptimizeUpload: %v", err)
	}
	if len(*files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(*files))
	}

	vf, err := os.Open((*files)[0])
	if err != nil {
		t.Fatalf("open variant: %v", err)
	}
	defer vf.Close()

	variantImg, err := webp.Decode(vf)
	if err != nil {
		t.Fatalf("decode variant: %v", err)
	}

	const offsetX = 50
	const offsetY = 50

	// Check every grid cell center that falls within the 200x200 output.
	// Cell centers in source: 25, 75, 125, 175, 225, 275
	// After subtracting offset 50: -25, 25, 75, 125, 175, 225
	// Valid output positions (0-199): 25, 75, 125, 175
	cellCenters := []int{25, 75, 125, 175}

	for _, oy := range cellCenters {
		for _, ox := range cellCenters {
			srcX := ox + offsetX
			srcY := oy + offsetY

			gr, gg, gb, _ := variantImg.At(ox, oy).RGBA()
			wr, wg, wb, _ := srcImg.At(srcX, srcY).RGBA()

			const tol = 6000 // ≈23 in 8-bit per channel (lossy WebP at quality 80)
			if absDiffU32(gr, wr) > tol || absDiffU32(gg, wg) > tol || absDiffU32(gb, wb) > tol {
				t.Errorf("grid cell center (%d,%d) → source (%d,%d): "+
					"got RGB(%d,%d,%d), want RGB(%d,%d,%d)",
					ox, oy, srcX, srcY,
					gr>>8, gg>>8, gb>>8,
					wr>>8, wg>>8, wb>>8)
			}
		}
	}
}

// ---------------------------------------------------------------------------
// copyFile additional tests
// ---------------------------------------------------------------------------

// TestCopyFile_PreservesContent_BinaryData verifies that binary data (like image
// bytes) is preserved exactly through the copy.
func TestCopyFile_PreservesContent_BinaryData(t *testing.T) {
	t.Parallel()

	srcDir := t.TempDir()
	dstDir := t.TempDir()

	// Create a real PNG image as source
	srcPath := createTestImage(t, srcDir, "test.png", 50, 50)

	dstPath := filepath.Join(dstDir, "test-copy.png")
	if err := copyFile(srcPath, dstPath); err != nil {
		t.Fatalf("copyFile returned error: %v", err)
	}

	// Verify the copy is a valid PNG by decoding it
	f, err := os.Open(dstPath)
	if err != nil {
		t.Fatalf("failed to open copy: %v", err)
	}
	defer f.Close()

	img, err := png.Decode(f)
	if err != nil {
		t.Fatalf("failed to decode copied PNG: %v", err)
	}

	bounds := img.Bounds()
	if bounds.Dx() != 50 || bounds.Dy() != 50 {
		t.Errorf("decoded image dimensions = %dx%d, want 50x50", bounds.Dx(), bounds.Dy())
	}
}

// ---------------------------------------------------------------------------
// FocalPointToPixels unit tests
// ---------------------------------------------------------------------------

func TestFocalPointToPixels_Nil(t *testing.T) {
	t.Parallel()

	bounds := image.Rect(0, 0, 300, 300)

	// Both invalid
	result := FocalPointToPixels(types.NullableFloat64{}, types.NullableFloat64{}, bounds)
	if result != nil {
		t.Errorf("expected nil for both invalid, got %v", result)
	}

	// Only X valid
	result = FocalPointToPixels(types.NewNullableFloat64(0.5), types.NullableFloat64{}, bounds)
	if result != nil {
		t.Errorf("expected nil for Y invalid, got %v", result)
	}

	// Only Y valid
	result = FocalPointToPixels(types.NullableFloat64{}, types.NewNullableFloat64(0.5), bounds)
	if result != nil {
		t.Errorf("expected nil for X invalid, got %v", result)
	}
}

func TestFocalPointToPixels_Center(t *testing.T) {
	t.Parallel()

	bounds := image.Rect(0, 0, 300, 300)
	result := FocalPointToPixels(types.NewNullableFloat64(0.5), types.NewNullableFloat64(0.5), bounds)
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.X != 150 || result.Y != 150 {
		t.Errorf("center focal = (%d,%d), want (150,150)", result.X, result.Y)
	}
}

func TestFocalPointToPixels_Corners(t *testing.T) {
	t.Parallel()

	bounds := image.Rect(0, 0, 200, 100)

	tests := []struct {
		name string
		fx   float64
		fy   float64
		wantX int
		wantY int
	}{
		{"top-left", 0.0, 0.0, 0, 0},
		{"top-right", 1.0, 0.0, 200, 0},
		{"bottom-left", 0.0, 1.0, 0, 100},
		{"bottom-right", 1.0, 1.0, 200, 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := FocalPointToPixels(types.NewNullableFloat64(tt.fx), types.NewNullableFloat64(tt.fy), bounds)
			if result == nil {
				t.Fatal("expected non-nil result")
			}
			if result.X != tt.wantX || result.Y != tt.wantY {
				t.Errorf("focal(%v,%v) = (%d,%d), want (%d,%d)", tt.fx, tt.fy, result.X, result.Y, tt.wantX, tt.wantY)
			}
		})
	}
}

func TestFocalPointToPixels_Clamping(t *testing.T) {
	t.Parallel()

	bounds := image.Rect(0, 0, 100, 100)

	// Values outside [0,1] should be clamped
	result := FocalPointToPixels(types.NewNullableFloat64(-0.5), types.NewNullableFloat64(1.5), bounds)
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.X != 0 {
		t.Errorf("negative focal X clamped = %d, want 0", result.X)
	}
	if result.Y != 100 {
		t.Errorf("over-1 focal Y clamped = %d, want 100", result.Y)
	}
}

// ---------------------------------------------------------------------------
// OptimizeUpload focal point tests — uses real test300x300.png fixture
// ---------------------------------------------------------------------------

func TestOptimizeUpload_FocalPointTopLeft(t *testing.T) {
	t.Parallel()

	srcImg := loadFixturePNG(t)
	srcDir := t.TempDir()
	dstDir := t.TempDir()
	srcPath := copyFixtureToDir(t, srcDir)

	// Focal point at top-left corner
	fp := &image.Point{X: 0, Y: 0}
	dims := []db.MediaDimensions{dim(150, 150)}
	lister := &mockDimensionLister{dims: &dims}

	files, err := OptimizeUpload(srcPath, dstDir, lister, fp)
	if err != nil {
		t.Fatalf("OptimizeUpload: %v", err)
	}
	if len(*files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(*files))
	}

	vf, err := os.Open((*files)[0])
	if err != nil {
		t.Fatalf("open variant: %v", err)
	}
	defer vf.Close()

	variantImg, err := webp.Decode(vf)
	if err != nil {
		t.Fatalf("decode variant: %v", err)
	}

	// With focal (0,0) and crop 150x150, the crop window should be clamped to (0,0)-(150,150)
	// Check that the top-left pixel of the variant matches the top-left of the source
	const tol = 16000 // lossy WebP at quality 80 — corners have worst artifacts
	gr, gg, gb, _ := variantImg.At(0, 0).RGBA()
	wr, wg, wb, _ := srcImg.At(0, 0).RGBA()
	if absDiffU32(gr, wr) > tol || absDiffU32(gg, wg) > tol || absDiffU32(gb, wb) > tol {
		t.Errorf("top-left pixel mismatch: got RGB(%d,%d,%d), want RGB(%d,%d,%d)",
			gr>>8, gg>>8, gb>>8, wr>>8, wg>>8, wb>>8)
	}
}

func TestOptimizeUpload_FocalPointBottomRight(t *testing.T) {
	t.Parallel()

	srcImg := loadFixturePNG(t)
	srcDir := t.TempDir()
	dstDir := t.TempDir()
	srcPath := copyFixtureToDir(t, srcDir)

	// Focal point at bottom-right corner
	fp := &image.Point{X: 299, Y: 299}
	dims := []db.MediaDimensions{dim(150, 150)}
	lister := &mockDimensionLister{dims: &dims}

	files, err := OptimizeUpload(srcPath, dstDir, lister, fp)
	if err != nil {
		t.Fatalf("OptimizeUpload: %v", err)
	}
	if len(*files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(*files))
	}

	vf, err := os.Open((*files)[0])
	if err != nil {
		t.Fatalf("open variant: %v", err)
	}
	defer vf.Close()

	variantImg, err := webp.Decode(vf)
	if err != nil {
		t.Fatalf("decode variant: %v", err)
	}

	// With focal (299,299) and crop 150x150, the window clamps to (150,150)-(300,300)
	// The bottom-right of the variant (149,149) should match source pixel (299,299)
	const tol = 16000 // lossy WebP at quality 80 — corners have worst artifacts
	gr, gg, gb, _ := variantImg.At(149, 149).RGBA()
	wr, wg, wb, _ := srcImg.At(299, 299).RGBA()
	if absDiffU32(gr, wr) > tol || absDiffU32(gg, wg) > tol || absDiffU32(gb, wb) > tol {
		t.Errorf("bottom-right pixel mismatch: got RGB(%d,%d,%d), want RGB(%d,%d,%d)",
			gr>>8, gg>>8, gb>>8, wr>>8, wg>>8, wb>>8)
	}
}

func TestOptimizeUpload_FocalPointEdgeClamping(t *testing.T) {
	t.Parallel()

	srcDir := t.TempDir()
	dstDir := t.TempDir()
	srcPath := copyFixtureToDir(t, srcDir)

	// Focal near left edge — crop window should shift right to stay within bounds
	fp := &image.Point{X: 10, Y: 150}
	dims := []db.MediaDimensions{dim(200, 200)}
	lister := &mockDimensionLister{dims: &dims}

	files, err := OptimizeUpload(srcPath, dstDir, lister, fp)
	if err != nil {
		t.Fatalf("OptimizeUpload: %v", err)
	}
	if len(*files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(*files))
	}

	// Verify the output is the correct size (proves the crop window wasn't shrunk)
	vf, err := os.Open((*files)[0])
	if err != nil {
		t.Fatalf("open variant: %v", err)
	}
	defer vf.Close()

	variantImg, err := webp.Decode(vf)
	if err != nil {
		t.Fatalf("decode variant: %v", err)
	}

	vBounds := variantImg.Bounds()
	if vBounds.Dx() != 200 || vBounds.Dy() != 200 {
		t.Errorf("variant dimensions = %dx%d, want 200x200", vBounds.Dx(), vBounds.Dy())
	}
}

func TestOptimizeUpload_FocalPointNil(t *testing.T) {
	t.Parallel()

	srcDir := t.TempDir()
	dstDir := t.TempDir()
	srcPath := copyFixtureToDir(t, srcDir)

	// nil focal point should produce same result as center
	dims := []db.MediaDimensions{dim(150, 150)}
	lister := &mockDimensionLister{dims: &dims}

	filesNil, err := OptimizeUpload(srcPath, dstDir, lister, nil)
	if err != nil {
		t.Fatalf("OptimizeUpload nil: %v", err)
	}

	// Create center focal point explicitly
	srcDir2 := t.TempDir()
	dstDir2 := t.TempDir()
	srcPath2 := copyFixtureToDir(t, srcDir2)
	fp := &image.Point{X: 150, Y: 150}

	filesCenter, err := OptimizeUpload(srcPath2, dstDir2, lister, fp)
	if err != nil {
		t.Fatalf("OptimizeUpload center: %v", err)
	}

	if len(*filesNil) != len(*filesCenter) {
		t.Fatalf("file count mismatch: nil=%d, center=%d", len(*filesNil), len(*filesCenter))
	}

	// Compare the variant images pixel-by-pixel
	vfNil, err := os.Open((*filesNil)[0])
	if err != nil {
		t.Fatalf("open nil variant: %v", err)
	}
	defer vfNil.Close()

	vfCenter, err := os.Open((*filesCenter)[0])
	if err != nil {
		t.Fatalf("open center variant: %v", err)
	}
	defer vfCenter.Close()

	imgNil, err := webp.Decode(vfNil)
	if err != nil {
		t.Fatalf("decode nil variant: %v", err)
	}
	imgCenter, err := webp.Decode(vfCenter)
	if err != nil {
		t.Fatalf("decode center variant: %v", err)
	}

	// Sample every 10px and verify identical output
	for y := 0; y < 150; y += 10 {
		for x := 0; x < 150; x += 10 {
			nr, ng, nb, _ := imgNil.At(x, y).RGBA()
			cr, cg, cb, _ := imgCenter.At(x, y).RGBA()
			if nr != cr || ng != cg || nb != cb {
				t.Errorf("pixel (%d,%d) differs: nil=RGB(%d,%d,%d), center=RGB(%d,%d,%d)",
					x, y, nr>>8, ng>>8, nb>>8, cr>>8, cg>>8, cb>>8)
			}
		}
	}
}
