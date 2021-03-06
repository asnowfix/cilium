// Copyright 2017 Authors of Cilium
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"archive/tar"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func createArchive(dbgDir string) (string, error) {
	// Based on http://blog.ralch.com/tutorial/golang-working-with-tar-and-gzip/
	archivePath := fmt.Sprintf("%s.tar", dbgDir)
	file, err := os.Create(archivePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	writer := tar.NewWriter(file)
	defer writer.Close()

	var baseDir string
	if info, err := os.Stat(dbgDir); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Debug directory does not exist %s", err)
		return "", err
	} else if err == nil && info.IsDir() {
		baseDir = filepath.Base(dbgDir)
	}

	return archivePath, filepath.Walk(dbgDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		header, err := tar.FileInfoHeader(info, info.Name())
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to compress %s %s", info.Name(), err)
			return err
		}

		if baseDir != "" {
			header.Name = filepath.Join(baseDir, strings.TrimPrefix(path, dbgDir))
		}

		if err := writer.WriteHeader(header); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to write header %s", err)
			return err
		}

		if info.IsDir() {
			return err
		}

		file, err := os.Open(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to open %s %s", path, err)
		}
		defer file.Close()
		_, err = io.Copy(writer, file)
		return err
	})
}

// copyDir and copyFile below is from https://github.com/Masterminds/glide

// copyDir copies an entire source directory to the dest directory.
//
// This is akin to `cp -a src/* dest/`
//
// We copy the directory here rather than jumping out to a shell so we can
// support multiple operating systems.
func copyDir(source string, dest string) error {

	// get properties of source dir
	si, err := os.Stat(source)
	if err != nil {
		return err
	}

	err = os.MkdirAll(dest, si.Mode())
	if err != nil {
		return err
	}

	d, err := os.Open(source)
	if err != nil {
		return err
	}
	defer d.Close()

	objects, err := d.Readdir(-1)

	for _, obj := range objects {

		sp := filepath.Join(source, "/", obj.Name())

		dp := filepath.Join(dest, "/", obj.Name())

		if obj.IsDir() {
			err = copyDir(sp, dp)
			if err != nil {
				return err
			}
		} else {
			// perform copy
			err = copyFile(sp, dp)
			if err != nil {
				return err
			}
		}

	}
	return nil
}

// copyFile copies a source file to a destination.
//
// It follows symbolic links and retains modes.
func copyFile(source string, dest string) error {
	ln, err := os.Readlink(source)
	if err == nil {
		return os.Symlink(ln, dest)
	}
	s, err := os.Open(source)
	if err != nil {
		return err
	}

	defer s.Close()

	d, err := os.Create(dest)
	if err != nil {
		return err
	}

	defer d.Close()

	_, err = io.Copy(d, s)
	if err != nil {
		return err
	}

	si, err := os.Stat(source)
	if err != nil {
		return err
	}
	err = os.Chmod(dest, si.Mode())

	return err
}
