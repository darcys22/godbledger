package build

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
)

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
