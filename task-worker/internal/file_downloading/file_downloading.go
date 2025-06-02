package file_downloading

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/imightbuyaboat/TaskFlow/pkg/task"
)

type FileDownloader struct {
	baseFilePath string
}

func NewFileDownloader() *FileDownloader {
	return &FileDownloader{
		baseFilePath: os.Getenv("BASE_FILE_PATH"),
	}
}

func (fd *FileDownloader) ExecuteTask(rawPayload interface{}) error {
	data, err := json.Marshal(rawPayload)
	if err != nil {
		return fmt.Errorf("failed to marshal rawPayload: %v", err)
	}

	var payload task.FileDownloadingPayload
	if err := json.Unmarshal(data, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload to FileDownloadingPayload: %v", err)
	}

	client := http.Client{
		Timeout: 15 * time.Second,
	}

	errs := []error{}
	mu := sync.Mutex{}
	sem := make(chan struct{}, 5)
	wg := sync.WaitGroup{}

	for _, url := range payload.URLs {
		wg.Add(1)

		go func(url string) {
			defer wg.Done()

			sem <- struct{}{}
			defer func() { <-sem }()

			resp, err := client.Get(url)
			if err != nil {
				mu.Lock()
				errs = append(errs, err)
				mu.Unlock()
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				mu.Lock()
				errs = append(errs, fmt.Errorf("unexpected status: %d", resp.StatusCode))
				mu.Unlock()
				return
			}

			contentType := resp.Header.Get("Content-Type")
			if index := strings.Index(contentType, ";"); index != -1 {
				contentType = contentType[:index]
			}

			exts, err := mime.ExtensionsByType(contentType)
			ext := ".bin"
			if err != nil {
				mu.Lock()
				errs = append(errs, err)
				mu.Unlock()
				return
			}
			if len(exts) > 0 {
				ext = exts[0]
			}

			srcPath := filepath.Join(fd.baseFilePath, uuid.New().String()+ext)
			out, err := os.Create(srcPath)
			if err != nil {
				mu.Lock()
				errs = append(errs, err)
				mu.Unlock()
				return
			}
			defer out.Close()

			_, err = io.Copy(out, resp.Body)
			if err != nil {
				mu.Lock()
				errs = append(errs, err)
				mu.Unlock()
				return
			}
		}(url)
	}

	wg.Wait()

	if len(errs) > 0 {
		var sb strings.Builder
		sb.WriteString("some files failed to download:")
		for _, err := range errs {
			sb.WriteString(" - " + err.Error() + " - ")
		}
		return errors.New(sb.String())
	}

	return nil
}
