package downloader

import (
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	"github.com/Mikhalevich/jober"
)

const (
	DefaultChunkSize  = 100 * 1024
	DefaultMaxWorkers = 20
)

type ChunkedTask struct {
	Task
	ChunkSize      int64
	MaxDownloaders int64
	CS             Storer
	csList         []Storer
}

func NewChunkedTask() *ChunkedTask {
	return &ChunkedTask{
		Task:           *NewTask(),
		ChunkSize:      DefaultChunkSize,
		MaxDownloaders: DefaultMaxWorkers,
	}
}

func (ct *ChunkedTask) makeStorages(workers int64) {
	ct.csList = make([]Storer, workers)
	var i int64
	for i = 0; i < workers; i++ {
		ct.csList[i] = ct.CS
	}
}

func (ct *ChunkedTask) Download(url string) error {
	var err error

	if ct.Task.S.GetFileName() == "" {
		ct.Task.S.SetFileName(url[strings.LastIndex(url, "/")+1:])
		defer ct.Task.S.SetFileName("")
	}

	contentLength, acceptRanges, err := resourceInfo(url)
	if err != nil {
		return err
	}

	if ct.ChunkSize <= 0 {
		ct.ChunkSize = DefaultChunkSize
	}

	var useChunksDownload bool = acceptRanges && contentLength > ct.ChunkSize

	if !useChunksDownload {
		return ct.Task.Download(url)
	}

	workers, chunkSize := calculateWorkers(contentLength, ct.ChunkSize, ct.MaxDownloaders)
	restChunk := contentLength % chunkSize

	ct.makeStorages(workers)

	job := jober.NewAll()

	var i int64
	for i = 0; i < workers; i++ {
		rangeIndex := i
		f := func() (interface{}, error) {
			startRange := rangeIndex * chunkSize
			endRange := (rangeIndex+1)*chunkSize - 1

			if rangeIndex == workers-1 {
				endRange += restChunk
			}

			request, err := http.NewRequest(ct.Task.Method, url, nil)
			if err != nil {
				return nil, err
			}

			bytesRange := "bytes=" + strconv.FormatInt(startRange, 10) + "-" + strconv.FormatInt(endRange, 10)
			request.Header.Add("Range", bytesRange)

			client := &http.Client{}
			response, err := client.Do(request)
			if err != nil {
				return nil, err
			}
			defer response.Body.Close()

			//var stop bool = response.StatusCode != http.StatusPartialContent

			bytes, err := ioutil.ReadAll(response.Body)
			if err != nil {
				return nil, err
			}

			ct.csList[i].Store(bytes)

			return nil, nil
		}
		job.Add(f)
	}

	job.Wait()

	_, errs := job.Get()
	if len(errs) > 0 {
		return errs[0]
	}

	for _, v := range ct.csList {
		b, err := v.Get()
		if err != nil {
			return err
		}

		ct.Task.S.Store(b)
	}

	return nil
}
