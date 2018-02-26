package downloader

import (
	"io"
	"net/http"
	"strings"
)

type Task struct {
	Method   string
	S        Storer
	Notifier chan int64
}

func NewTask() *Task {
	return &Task{
		Method: "GET",
		S:      NewFileStorer(""),
	}
}

func (t *Task) notify(data int64) {
	if t.Notifier != nil {
		t.Notifier <- data
	}
}

func (t *Task) closeNotifier() {
	if t.Notifier != nil {
		close(t.Notifier)
	}
}

func (t *Task) storeBytes(r io.Reader, s Storer) error {
	buf := make([]byte, 64*1024)
	for {
		n, err := r.Read(buf)
		s.Store(buf[:n])

		if n > 0 {
			t.notify(int64(n))
		}

		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}
	}

	return nil
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

	t.notify(response.ContentLength)

	err = t.storeBytes(response.Body, t.S)
	if err != nil {
		return err
	}

	t.closeNotifier()
	return nil
}
