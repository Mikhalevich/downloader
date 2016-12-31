package downloader

import (
	"net/http"
)

func resourceInfo(url string) (int64, bool, error) {
	response, err := http.Head(url)
	if err != nil {
		return 0, false, err
	}
	defer response.Body.Close()

	acceptRangesValue, acceptRanges := response.Header["Accept-Ranges"]
	if acceptRanges {
		for _, value := range acceptRangesValue {
			if value == "none" {
				acceptRanges = false
			}
		}
	}
	return response.ContentLength, acceptRanges, nil
}

func calculateWorkers(contentLength, chunkSize, maxWorkers int64) (int64, int64) {
	workers := contentLength / chunkSize
	if workers > maxWorkers {
		chunkSize = contentLength / maxWorkers
		workers = maxWorkers
	}

	return workers, chunkSize
}
