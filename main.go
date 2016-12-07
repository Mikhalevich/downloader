package main

import (
	"errors"
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

type Task struct {
	Method    string
	ChunkSize int
}

func NewTask() *Task {
	return &Task{
		Method:    "GET",
		ChunkSize: 1 * 1024 * 1024,
	}
}

func (self Task) downloadChunk(url string, startRange, endRange, index int, chunk chan Chunk) {
	request, err := http.NewRequest(self.Method, url, nil)
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

func (self Task) simpleDownload(url string) []byte {
	request, err := http.NewRequest(self.Method, url, nil)
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

func (self Task) Download(url string, filePath string) error {
	response, err := http.Head(url)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.ContentLength < 0 {
		return errors.New("Invalid content-length")
	}

	var results [][]byte

	if int(response.ContentLength) > self.ChunkSize {
		workers := int(response.ContentLength) / self.ChunkSize
		restChunk := int(response.ContentLength) % self.ChunkSize

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
		bytes := self.simpleDownload(url)
		results[0] = bytes
	}

	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	for _, bytes := range results {
		if len(bytes) <= 0 {
			errors.New("Error while downloading resource parts")
		}

		file.Write(bytes)
	}

	return nil
}

func main() {
	startTime := time.Now()

	task := NewTask()
	//Download("GET", "http://localhost:8080/qt-opensource-linux-x64-5.5.1.run", "test")
	task.Download("http://localhost:8080/ChuckVsTux-full.jpg", "test")
	//Download("GET", "http://localhost:8080/go1.7.1.linux-amd64.tar.gz", "test")
	//Download("GET", "https://cmake.org/files/v3.7/cmake-3.7.1.tar.gz", "test")
	//Download("GET", "https://www.google.ru/url?sa=i&rct=j&q=&esrc=s&source=images&cd=&ved=0ahUKEwjr3Kfzj-LQAhVMWhoKHbuQAWoQjBwIBA&url=https%3A%2F%2Fupload.wikimedia.org%2Fwikipedia%2Fcommons%2Fd%2Fdd%2FExpeditionary_Fighting_Vehicle_test.jpg&bvm=bv.140496471,d.d24&psig=AFQjCNHIAILnpbsLhA9hB7RGR1mW4tghyg&ust=1481201542074876&cad=rjt", "test")

	endTime := time.Now()
	fmt.Println(endTime.Sub(startTime))
}
