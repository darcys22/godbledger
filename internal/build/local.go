package build

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
)

func compress(src string, buf io.Writer) error {
	// tar > gzip > buf
	zr := gzip.NewWriter(buf)
	tw := tar.NewWriter(zr)

	// is file a folder?
	fi, err := os.Stat(src)
	if err != nil {
		return err
	}
	mode := fi.Mode()
	if mode.IsRegular() {
		// get header
		header, err := tar.FileInfoHeader(fi, src)
		if err != nil {
			return err
		}
		// write header
		if err := tw.WriteHeader(header); err != nil {
			return err
		}
		// get content
		data, err := os.Open(src)
		if err != nil {
			return err
		}
		if _, err := io.Copy(tw, data); err != nil {
			return err
		}
	} else if mode.IsDir() {
		// walk through every file in the folder
		filepath.Walk(src, func(file string, fi os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			// generate tar header
			header, err := tar.FileInfoHeader(fi, file)
			if err != nil {
				return err
			}

			// must provide real name
			// (see https://golang.org/src/archive/tar/common.go?#L626)
			_, outFile := filepath.Split(file)
			header.Name = outFile

			// write header
			if err := tw.WriteHeader(header); err != nil {
				return err
			}
			// if not a dir, write file content
			if !fi.IsDir() {
				data, err := os.Open(file)
				if err != nil {
					return err
				}
				if _, err := io.Copy(tw, data); err != nil {
					return err
				}
			}
			return nil
		})
	} else {
		return fmt.Errorf("error: file type not supported")
	}

	// produce tar
	if err := tw.Close(); err != nil {
		return err
	}
	// produce gzip
	if err := zr.Close(); err != nil {
		return err
	}
	//
	return nil
}

// LocalAssets contains the local objects to be uploaded
func LocalAssets(path string) ([]string, error) {
	if path == "" {
		return make([]string, 0), nil
	}

	path, err := filepath.Abs(path)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get abs path")
	}

	fi, err := os.Stat(path)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get file stat")
	}

	if !fi.IsDir() {
		return []string{path}, nil
	}

	// Glob all files in the given path
	files, err := filepath.Glob(filepath.Join(path, "*"))
	if err != nil {
		return nil, errors.Wrap(err, "failed to glob files")
	}

	assets := make([]string, 0, len(files))
	for _, f := range files {
		if fi, _ := os.Stat(f); fi.IsDir() {
			var buf bytes.Buffer
			if err := compress(f, &buf); err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			// write file to disk
			pre, outFile := filepath.Split(f)
			outFile = outFile + ".tar.gz"
			if err = os.WriteFile(pre+outFile, buf.Bytes(), 0600); err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			assets = append(assets, pre+outFile)
		}
	}
	for _, f := range files {
		// Exclude directory.
		if fi, _ := os.Stat(f); fi.IsDir() {
			continue
		}

		// Exclude hidden file
		if filepath.Base(f)[0] == '.' {
			continue
		}

		assets = append(assets, f)
	}

	return assets, nil
}

// SHA256Assets contains the local objects SHA Hashes
func SHA256Assets(files []string) ([]string, error) {
	h := sha256.New()
	checksums := make([]string, 0, len(files))
	for _, f := range files {
		file, err := os.Open(f)
		if err != nil {
			return nil, errors.Wrap(err, "failed to open file")
		}
		defer file.Close()
		if _, err := io.Copy(h, file); err != nil {
			return nil, errors.Wrap(err, "failed to copy file into hasher")
		}

		sha1_hash := hex.EncodeToString(h.Sum(nil))

		checksums = append(checksums, sha1_hash)
		h.Reset()
	}
	return checksums, nil
}
