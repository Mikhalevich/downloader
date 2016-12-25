package downloader

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	DefaultChunkSize  = 100 * 1024
	DefaultMaxWorkers = 20
)

type Chunk struct {
	Index int64
	Bytes []byte
}

type Statistics struct {
	Url              string
	ContentLength    int64
	AcceptRanges     bool
	NumberOfChunks   int64
	ChunkSize        int64
	TotalTime        time.Duration
	SlowestChunkTime time.Duration
	ChunkTimes       []time.Duration
}

func (self Statistics) String() string {
	return fmt.Sprintf("Url = %s; ContentLength = %d; AcceptRanges = %t; Chunks = %d; ChunkSize = %d; TotalTime = %s; SlowestChunkTime = %s",
		self.Url, self.ContentLength, self.AcceptRanges, self.NumberOfChunks, self.ChunkSize, self.TotalTime, self.SlowestChunkTime)
}

type Task struct {
	Method         string
	ChunkSize      int64
	DownloadFolder string
	EnableRange    bool
	MaxWorkers     int64
	Stats          Statistics
}

func NewTask() *Task {
	return &Task{
		Method:         "GET",
		ChunkSize:      DefaultChunkSize,
		DownloadFolder: "",
		EnableRange:    true,
		MaxWorkers:     DefaultMaxWorkers,
		Stats:          Statistics{ChunkTimes: make([]time.Duration, 1)},
	}
}

func (self *Task) downloadChunk(url string, startRange, endRange, index int64, chunk chan Chunk) error {
	startTime := time.Now()
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

	chunkExecutionTime := time.Now().Sub(startTime)
	self.Stats.ChunkTimes = append(self.Stats.ChunkTimes, chunkExecutionTime)
	if chunkExecutionTime > self.Stats.SlowestChunkTime {
		self.Stats.SlowestChunkTime = chunkExecutionTime
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

func resourceInfo(url string) (int64, bool, error) {
	response, err := http.Head(url)
	if err != nil {
		return 0, false, err
	}
	defer response.Body.Close()

	_, acceptRanges := response.Header["Accept-Ranges"]
	return response.ContentLength, acceptRanges, nil
}

func storeResource(fileName, downloadFolder string, data [][]byte) error {
	if downloadFolder != "" {
		if err := os.MkdirAll(downloadFolder, os.ModePerm); err != nil {
			return err
		}
	}

	file, err := os.Create(filepath.Join(downloadFolder, fileName))
	if err != nil {
		return err
	}
	defer file.Close()

	for _, bytes := range data {
		file.Write(bytes)
	}

	return nil
}

func calculateWorkers(contentLength, chunkSize, maxWorkers int64) (int64, int64) {
	workers := contentLength / chunkSize
	if workers > maxWorkers {
		chunkSize = contentLength / maxWorkers
		workers = maxWorkers
	}

	return workers, chunkSize
}

func (self *Task) Download(url string, fileName string) error {
	startTime := time.Now()
	var contentLength int64 = 0
	var acceptRanges bool = false
	var err error

	if self.EnableRange {
		contentLength, acceptRanges, err = resourceInfo(url)
		if err != nil {
			return err
		}
	}

	self.Stats.Url = url
	self.Stats.ContentLength = contentLength
	self.Stats.AcceptRanges = acceptRanges

	if self.ChunkSize <= 0 {
		self.ChunkSize = DefaultChunkSize
	}

	var results [][]byte
	if acceptRanges && contentLength > self.ChunkSize {
		workers, chunkSize := calculateWorkers(contentLength, self.ChunkSize, self.MaxWorkers)
		restChunk := contentLength % chunkSize

		self.Stats.NumberOfChunks = workers
		self.Stats.ChunkSize = chunkSize

		var wg sync.WaitGroup
		chunk := make(chan Chunk)
		downloadError := make(chan error)
		finish := make(chan bool)
		errorFinish := make(chan bool)
		results = make([][]byte, workers)
		errorResults := make([]error, 0)

		go func() {
			for c := range chunk {
				results[c.Index] = c.Bytes
			}

			finish <- true
		}()

		go func() {
			for c := range downloadError {
				errorResults = append(errorResults, c)
			}

			errorFinish <- true
		}()

		var i int64
		for i = 0; i < workers; i++ {
			wg.Add(1)

			go func(rangeIndex int64) {
				defer wg.Done()

				startRange := rangeIndex * chunkSize
				endRange := (rangeIndex+1)*chunkSize - 1

				if rangeIndex == workers-1 {
					endRange += restChunk
				}

				err := self.downloadChunk(url, startRange, endRange, rangeIndex, chunk)
				if err != nil {
					downloadError <- err
				}
			}(i)
		}

		wg.Wait()
		close(chunk)
		close(downloadError)
		<-finish
		<-errorFinish

		if len(errorResults) > 0 {
			return errorResults[0]
		}
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

	if fileName == "" {
		fileName = url[strings.LastIndex(url, "/")+1:]
	}

	err = storeResource(fileName, self.DownloadFolder, results)
	if err != nil {
		return err
	}

	self.Stats.TotalTime = time.Now().Sub(startTime)

	return nil
}
