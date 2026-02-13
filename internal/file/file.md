# file

Package file provides file system operations for ModulaCMS including zip archive extraction. This internal package handles file manipulation tasks required by backup restoration and deployment workflows.

## Overview

The file package exports utilities for working with zip archives. It provides safe extraction with proper directory creation and file permission preservation. All operations use filepath.Join for cross-platform path handling.

## Functions

#### Unzip

Signature: `func Unzip(src, dest string) error`

Extracts a zip archive to a destination directory. Moves all files and folders from the zip archive while preserving directory structure and file permissions. Creates intermediate directories as needed.

Parameters:
- src: path to the zip file to extract
- dest: destination directory where contents will be extracted

Returns error if zip cannot be opened, directories cannot be created, or files cannot be copied. The function opens the zip archive using archive/zip.OpenReader, iterates through all entries, creates directories with os.MkdirAll using os.ModePerm, and copies files using io.Copy. File modes from the archive are preserved via f.Mode() when creating output files with os.OpenFile. Resources are properly closed after each operation.

Implementation details: checks if each entry is a directory using f.FileInfo().IsDir(), creates parent directories before files, uses os.O_WRONLY with os.O_CREATE and os.O_TRUNC flags for file creation, and ensures all file handles and readers are closed before returning errors.
