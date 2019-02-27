package downloader

import (
	"io"
	"net/http"
	"strconv"
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

func (t *Task) Download(url string) (*DownloadInfo, error) {
	var err error

	request, err := http.NewRequest(t.Method, url, nil)
	if err != nil {
		return nil, err
	}

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	if t.S.GetFileName() == "" {
		t.S.SetFileName(nameFromResponse(response))
		defer t.S.SetFileName("")
	}
	info := NewDownloadInfo(t.S.GetFileName())
	info.Info["content_length"] = strconv.FormatInt(response.ContentLength, 10)

	t.notify(response.ContentLength)

	err = t.storeBytes(response.Body, t.S)
	if err != nil {
		return info, err
	}

	t.closeNotifier()
	return info, nil
}
