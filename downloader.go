package downloader

import (
	"errors"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
)

type Chunk struct {
	Index int64
	Bytes []byte
}

type Task struct {
	Method         string
	ChunkSize      int64
	DownloadFolder string
}

func NewTask() *Task {
	return &Task{
		Method:         "GET",
		ChunkSize:      100 * 1024,
		DownloadFolder: "",
	}
}

func (self Task) downloadChunk(url string, startRange, endRange, index int64, chunk chan Chunk) error {
	request, err := http.NewRequest(self.Method, url, nil)
	if err != nil {
		return err
	}

	bytesRange := "bytes=" + strconv.FormatInt(startRange, 10) + "-" + strconv.FormatInt(endRange, 10)
	request.Header.Add("Range", bytesRange)

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	bytes, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return err
	}

	chunk <- Chunk{index, bytes}
	return nil
}

func (self Task) downloadWholeResource(url string) ([]byte, error) {
	request, err := http.NewRequest(self.Method, url, nil)
	if err != nil {
		return []byte(""), err
	}

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return []byte(""), err
	}
	defer response.Body.Close()

	bytes, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return []byte(""), err
	}

	return bytes, nil
}

func resourceContentLength(url string) (int64, error) {
	response, err := http.Head(url)
	if err != nil {
		return 0, err
	}
	defer response.Body.Close()

	return response.ContentLength, nil
}

func (self Task) Download(url string, filePath string) error {
	contentLength, err := resourceContentLength(url)
	if err != nil {
		return err
	}

	var results [][]byte
	if contentLength > self.ChunkSize {
		workers := contentLength / self.ChunkSize
		restChunk := contentLength % self.ChunkSize

		var wg sync.WaitGroup
		chunk := make(chan Chunk)
		finish := make(chan bool)
		results = make([][]byte, workers)

		go func() {
			for c := range chunk {
				results[c.Index] = c.Bytes
			}

			finish <- true
		}()

		var i int64
		for i = 0; i < workers; i++ {
			wg.Add(1)

			go func(rangeIndex int64) {
				defer wg.Done()

				startRange := rangeIndex * self.ChunkSize
				endRange := (rangeIndex+1)*self.ChunkSize - 1

				if rangeIndex == workers-1 {
					endRange += restChunk
				}

				self.downloadChunk(url, startRange, endRange, rangeIndex, chunk)
			}(i)
		}

		wg.Wait()
		close(chunk)
		<-finish
	} else {
		results = make([][]byte, 1)
		bytes, err := self.downloadWholeResource(url)
		if err != nil {
			return err
		}
		results[0] = bytes
	}

	for _, bytes := range results {
		if len(bytes) <= 0 {
			return errors.New("Error while downloading resource parts")
		}
	}

	if filePath == "" {
		filePath = url[strings.LastIndex(url, "/")+1:]
	}

	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	for _, bytes := range results {
		file.Write(bytes)
	}

	return nil
}
