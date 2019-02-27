package downloader

type DownloadInfo struct {
	FileName string
	Info     map[string]string
}

func NewDownloadInfo(name string) *DownloadInfo {
	return &DownloadInfo{
		FileName: name,
		Info:     make(map[string]string),
	}
}

type Downloader interface {
	Download(url string) (*DownloadInfo, error)
}

type Storer interface {
	Store(b []byte) error
	Get() ([]byte, error)
	GetFileName() string
	SetFileName(fileName string)
	Clone() Storer
}
