package downloader

type Downloader interface {
	Download(url string) error
}

type Storer interface {
	Store(b []byte) error
	Get() ([]byte, error)
	GetFileName() string
	SetFileName(fileName string)
}
