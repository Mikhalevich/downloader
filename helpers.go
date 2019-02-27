package downloader

import (
	"errors"
	"net/http"
	"regexp"
	"strings"
)

func resourceInfo(url string) (int64, bool, string, error) {
	response, err := http.Head(url)
	if err != nil {
		return 0, false, "", err
	}
	defer response.Body.Close()

	arValue := response.Header.Get("Accept-Ranges")
	acceptRanges := false
	if arValue == "bytes" {
		acceptRanges = true
	}
	return response.ContentLength, acceptRanges, nameFromResponse(response), nil
}

func contentDispositionName(r *http.Response) (string, error) {
	cd := r.Header.Get("Content-Disposition")
	if cd == "" {
		return "", errors.New("No content disposition header")
	}

	re, err := regexp.Compile(`filename\s*=\s*"(.+)"`)
	if err != nil {
		return "", err
	}

	results := re.FindStringSubmatch(cd)
	if len(results) < 2 {
		return "", errors.New("filename is not matched")
	}

	return results[1], nil
}

func lastPathPart(url string) string {
	if strings.HasSuffix(url, "/") {
		url = url[:len(url)-1]
	}

	return url[strings.LastIndex(url, "/")+1:]
}

func nameFromResponse(r *http.Response) string {
	name, err := contentDispositionName(r)
	if err == nil {
		return name
	}
	return lastPathPart(r.Request.URL.EscapedPath())
}

func calculateWorkers(contentLength, chunkSize, maxWorkers int64) (int64, int64) {
	workers := contentLength / chunkSize
	if workers > maxWorkers {
		chunkSize = contentLength / maxWorkers
		workers = maxWorkers
	}

	return workers, chunkSize
}
