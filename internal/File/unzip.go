package file

import (
	"archive/zip"
	"io"
	"os"
	"path/filepath"
)

// unzip extracts a zip archive, moving all files and folders
// to a destination directory.
func Unzip(src, dest string) error {
    r, err := zip.OpenReader(src)
    if err != nil {
        return err
    }
    defer r.Close()

    for _, f := range r.File {
        // Create the full directory path for the file/directory
        fpath := filepath.Join(dest, f.Name)

        // If it's a directory, create it.
        if f.FileInfo().IsDir() {
            if err := os.MkdirAll(fpath, os.ModePerm); err != nil {
                return err
            }
            continue
        }

        // Ensure the directory exists.
        if err := os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
            return err
        }

        // Open the file inside the zip archive.
        rc, err := f.Open()
        if err != nil {
            return err
        }

        // Create a new file in the destination directory.
        outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
        if err != nil {
            rc.Close()
            return err
        }

        // Copy file contents.
        _, err = io.Copy(outFile, rc)

        // Close open resources.
        outFile.Close()
        rc.Close()

        if err != nil {
            return err
        }
    }
    return nil
}
