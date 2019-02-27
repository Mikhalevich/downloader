package downloader

import (
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strconv"

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
}

func NewChunkedTask() *ChunkedTask {
	return &ChunkedTask{
		Task:           *NewTask(),
		ChunkSize:      DefaultChunkSize,
		MaxDownloaders: DefaultMaxWorkers,
		CS:             NewMemoryStorer(),
	}
}

type chunk struct {
	index int64
	s     Storer
}

func (ct *ChunkedTask) Download(url string) (*DownloadInfo, error) {
	var err error

	contentLength, acceptRanges, fileName, err := resourceInfo(url)
	if err != nil {
		return nil, err
	}

	if ct.ChunkSize <= 0 {
		ct.ChunkSize = DefaultChunkSize
	}

	var useChunksDownload bool = acceptRanges && contentLength > ct.ChunkSize

	if !useChunksDownload {
		return ct.Task.Download(url)
	}

	if ct.Task.S.GetFileName() == "" {
		ct.Task.S.SetFileName(fileName)
		defer ct.Task.S.SetFileName("")
	}
	info := NewDownloadInfo(ct.Task.S.GetFileName())

	ct.notify(contentLength)

	workers, chunkSize := calculateWorkers(contentLength, ct.ChunkSize, ct.MaxDownloaders)
	restChunk := contentLength % chunkSize

	info.Info["workers"] = strconv.FormatInt(workers, 10)
	info.Info["chunk_size"] = strconv.FormatInt(chunkSize, 10)
	info.Info["content_length"] = strconv.FormatInt(contentLength, 10)

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

			if response.StatusCode != http.StatusPartialContent {
				return nil, errors.New("Not a partial chunk")
			}

			storer := ct.CS.Clone()
			storer.SetFileName(fmt.Sprintf("%s_%d", ct.Task.S.GetFileName(), rangeIndex))
			err = ct.storeBytes(response.Body, storer)
			if err != nil {
				return nil, err
			}

			return chunk{rangeIndex, storer}, nil
		}
		job.Add(f)
	}

	job.Wait()

	chunks, errs := job.Get()
	if len(errs) > 0 {
		return info, errs[0]
	}

	ct.closeNotifier()

	sort.Slice(chunks, func(i, j int) bool {
		return chunks[i].(chunk).index < chunks[j].(chunk).index
	})

	for _, v := range chunks {
		b, err := v.(chunk).s.Get()
		if err != nil {
			return info, err
		}

		err = ct.Task.S.Store(b)
		if err != nil {
			return info, err
		}
	}

	return info, nil
}
