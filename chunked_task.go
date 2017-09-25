package downloader

const (
	DefaultChunkSize  = 100 * 1024
	DefaultMaxWorkers = 20
)

type ChunkedTask struct {
	Method         string
	S              []Storer
	ChunkSize      int64
	MaxDownloaders int64
}

func NewChunkedTask() *ChunkedTask {
	return &ChunkedTask{
		Method:         "GET",
		ChunkSize:      DefaultChunkSize,
		MaxDownloaders: DefaultMaxWorkers,
	}
}

func (ct *ChunkedTask) Download(url string) error {
	//todo
	return nil
}
