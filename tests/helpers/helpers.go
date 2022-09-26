// Package helpers defines helper functions to peer into
// end to end processes and kill processes as needed.
package helpers

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"path"
	"strings"
	"testing"
	"time"
)

const (
	maxPollingWaitTime  = 60 * time.Second // A minute so timing out doesn't take very long.
	filePollingInterval = 500 * time.Millisecond
)

// KillProcesses finds the passed in process IDs and kills the process.
func KillProcesses(t *testing.T, pIDs []int) {
	for _, id := range pIDs {
		process, err := os.FindProcess(id)
		if err != nil {
			t.Fatalf("Could not find process %d: %v", id, err)
		}
		if err := process.Kill(); err != nil {
			t.Fatal(err)
		}
		if _, err := process.Wait(); err != nil {
			t.Fatal(err)
		}
	}
}

// DeleteLogFiles cleans up the lingering log files
func DeleteLogFiles(t *testing.T, logfiles []*os.File) {
	for _, f := range logfiles {
		err := f.Close()
		if err != nil {
			t.Logf("error closing %s: %v\n", f.Name(), err)
		}

		err = os.Remove(f.Name())
		if err != nil {
			t.Logf("error removing %s: %v\n", f.Name(), err)
		}
	}
}

// DeleteAndCreateFile checks if the file path given exists, if it does, it deletes it and creates a new file.
// If not, it just creates the requested file.
func DeleteAndCreateFile(tmpPath string, fileName string) (*os.File, error) {
	filePath := path.Join(tmpPath, fileName)
	if _, err := os.Stat(filePath); os.IsExist(err) {
		if err := os.Remove(filePath); err != nil {
			return nil, err
		}
	}
	newFile, err := os.Create(path.Join(tmpPath, fileName))
	if err != nil {
		return nil, err
	}
	return newFile, nil
}

// WaitForTextInFile checks a file every polling interval for the text requested.
func WaitForTextInFile(file *os.File, text string) error {
	d := time.Now().Add(maxPollingWaitTime)
	ctx, cancel := context.WithDeadline(context.Background(), d)
	defer cancel()

	// Use a ticker with a deadline to poll a given file.
	ticker := time.NewTicker(filePollingInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			contents, err := io.ReadAll(file)
			if err != nil {
				return err
			}
			return fmt.Errorf("could not find requested text \"%s\" in logs:\n%s", text, contents)
		case <-ticker.C:
			fileScanner := bufio.NewScanner(file)
			for fileScanner.Scan() {
				scanned := fileScanner.Text()
				if strings.Contains(scanned, text) {
					return nil
				}
			}
			if err := fileScanner.Err(); err != nil {
				return err
			}
			_, err := file.Seek(0, io.SeekStart)
			if err != nil {
				return err
			}
		}
	}
}

// LogErrorOutput logs the output of a specific file.
func LogErrorOutput(t *testing.T, file io.Reader, title string, index int) {
	var errorLines []string

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		currentLine := scanner.Text()
		if strings.Contains(currentLine, "level=error") {
			errorLines = append(errorLines, currentLine)
		}
	}
	if len(errorLines) < 1 {
		return
	}

	t.Logf("==================== Start of %s %d error output ==================\n", title, index)
	var lines uint64
	for _, err := range errorLines {
		lines++
		if lines >= 10 {
			break
		}
		t.Log(err)
	}
}
