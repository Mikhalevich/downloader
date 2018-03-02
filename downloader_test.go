package downloader

import (
	"testing"
)

var (
	smallResources []string = []string{"https://cdn-images-1.medium.com/max/1200/1*Uzd2n_pZTnQkCK0_MHE81w.jpeg",
		"http://alierbey.com/wp-content/uploads/2016/10/golang.sh-600x600.png",
		"http://i1-news.softpedia-static.com/images/news2/Go-1-Is-the-First-Stable-Version-of-Google-s-New-Programming-Language-2.png",
	}

	largeResources []string = []string{"http://www.kenrockwell.com/nikon/d600/sample-images/600_0985.JPG",
		"https://cmake.org/files/v3.7/cmake-3.7.1.tar.gz",
		"http://blog.globalknowledge.com/wp-content/uploads/2010/12/Photoxpress_936733.jpg",
	}
)

func TestTaskWithFileStorer(t *testing.T) {
	task := NewTask()
	task.S = NewFileStorer("test_task")

	for _, url := range smallResources {
		_, err := task.Download(url)
		if err != nil {
			t.Fatal(err)
		}
	}

	for _, url := range largeResources {
		_, err := task.Download(url)
		if err != nil {
			t.Fatal(err)
		}
	}
}

func TestTaskWithMemoryStorer(t *testing.T) {
	task := NewTask()
	task.S = NewMemoryStorer()

	for _, url := range smallResources {
		_, err := task.Download(url)
		if err != nil {
			t.Fatal(err)
		}
	}

	for _, url := range largeResources {
		_, err := task.Download(url)
		if err != nil {
			t.Fatal(err)
		}
	}
}

func TestChunkedTaskWithMemoryStorer(t *testing.T) {
	task := NewChunkedTask()
	task.S = NewFileStorer("chunked_test_task")

	for _, url := range smallResources {
		_, err := task.Download(url)
		if err != nil {
			t.Fatal(err)
		}
	}

	for _, url := range largeResources {
		_, err := task.Download(url)
		if err != nil {
			t.Fatal(err)
		}
	}
}

func TestChunkedTaskWithFileStorer(t *testing.T) {
	task := NewChunkedTask()
	task.S = NewFileStorer("chunked_test_task_file")
	task.CS = NewFileStorer("chunked_test_task_file/chunks")

	for _, url := range smallResources {
		_, err := task.Download(url)
		if err != nil {
			t.Fatal(err)
		}
	}

	for _, url := range largeResources {
		_, err := task.Download(url)
		if err != nil {
			t.Fatal(err)
		}
	}
}
