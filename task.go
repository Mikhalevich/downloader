package downloader

import (
	"io/ioutil"
	"net/http"
	"strings"
)

type Task struct {
	Method string
	S      Storer
}

func NewTask() *Task {
	return &Task{
		Method: "GET",
		S:      NewFileStorer(""),
	}
}

func (t *Task) processDownload(url string) ([]byte, error) {
	request, err := http.NewRequest(t.Method, url, nil)
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

func (t *Task) Download(url string) error {
	var err error

	if t.S.GetFileName() == "" {
		t.S.SetFileName(url[strings.LastIndex(url, "/")+1:])
		defer t.S.SetFileName("")
	}

	b, err := t.processDownload(url)

	err = t.S.Store(b)
	if err != nil {
		return err
	}

	return nil
}
