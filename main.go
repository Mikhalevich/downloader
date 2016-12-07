package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"
)

const (
	ByteCount = 1 * 1024 * 1024
	Async     = true
)

type Chunk struct {
	Index int
	Bytes []byte
}

func downloadChunk(method, url string, startRange, endRange, index int, chunk chan Chunk) {
	request, err := http.NewRequest(method, url, nil)
	if err != nil {
		fmt.Println(err)
		return
	}

	bytesRange := "bytes=" + strconv.Itoa(startRange) + "-" + strconv.Itoa(endRange)
	request.Header.Add("Range", bytesRange)

	fmt.Println(index, bytesRange)

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer response.Body.Close()

	bytes, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return
	}

	chunk <- Chunk{index, bytes}
}

func simpleDownload(method, url string) []byte {
	request, err := http.NewRequest(method, url, nil)
	if err != nil {
		fmt.Println(err)
		return []byte("")
	}

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		fmt.Println(err)
		return []byte("")
	}
	defer response.Body.Close()

	bytes, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return []byte("")
	}

	return bytes
}

func Download(method, url, filePath string) {
	response, err := http.Head(url)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer response.Body.Close()

	if response.ContentLength < 0 {
		fmt.Println("Invalid content-length")
		return
	}
	fmt.Println(response.ContentLength)

	var results [][]byte

	if int(response.ContentLength) > ByteCount && Async {
		workers := int(response.ContentLength) / ByteCount
		restChunk := int(response.ContentLength) % ByteCount

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

		for i := 0; i < workers; i++ {
			wg.Add(1)

			go func(rangeIndex int) {
				defer wg.Done()

				startRange := rangeIndex * ByteCount
				endRange := (rangeIndex+1)*ByteCount - 1

				if rangeIndex == workers-1 {
					endRange += restChunk
				}

				downloadChunk(method, url, startRange, endRange, rangeIndex, chunk)
			}(i)
		}

		wg.Wait()
		close(chunk)
		<-finish
	} else {
		results = make([][]byte, 1)
		bytes := simpleDownload(method, url)
		results[0] = bytes
	}

	file, err := os.Create(filePath)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer file.Close()

	for _, bytes := range results {
		if len(bytes) <= 0 {
			fmt.Println("Error")
			return
		}

		file.Write(bytes)
	}
}

func main() {
	startTime := time.Now()

	//Download("GET", "http://localhost:8080/qt-opensource-linux-x64-5.5.1.run", "test")
	//Download("GET", "http://localhost:8080/ChuckVsTux-full.jpg", "test")
	Download("GET", "http://localhost:8080/go1.7.1.linux-amd64.tar.gz", "test")

	endTime := time.Now()
	fmt.Println(endTime.Sub(startTime))
}
