package downloader

import (
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

func (t *Task) Download(url string) error {
	var err error

	if t.S.GetFileName() == "" {
		t.S.SetFileName(url[strings.LastIndex(url, "/")+1:])
		defer t.S.SetFileName("")
	}

	request, err := http.NewRequest(t.Method, url, nil)
	if err != nil {
		return err
	}

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	err = storeBytes(response.Body, t.S)
	if err != nil {
		return err
	}

	return nil
}
