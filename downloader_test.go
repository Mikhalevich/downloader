package downloader

import (
	"fmt"
	"testing"
	"time"
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

func runTask(url string, enableRange bool, t *testing.T) {
	fmt.Println(url)
	startTime := time.Now()

	task := NewTask()
	task.DownloadFolder = "test_files"
	task.EnableRange = enableRange

	err := task.Download(url, "")
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(time.Now().Sub(startTime))
}

func download(urls []string, t *testing.T) {
	for _, url := range urls {
		runTask(url, true, t)
		runTask(url, false, t)
	}
}

func TestSmallFiles(t *testing.T) {
	download(smallResources, t)
}

func TestLargeFiles(t *testing.T) {
	download(largeResources, t)
}
