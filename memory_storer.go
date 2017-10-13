package downloader

type MemoryStorer struct {
	data []byte
}

func NewMemoryStorer() *MemoryStorer {
	return &MemoryStorer{}
}

func (ms *MemoryStorer) Store(b []byte) error {
	ms.data = append(ms.data, b...)
	return nil
}

func (ms *MemoryStorer) Get() ([]byte, error) {
	return ms.data, nil
}

func (ms *MemoryStorer) GetFileName() string {
	return ""
}

func (ms *MemoryStorer) SetFileName(fileName string) {
	// pass
}

func (ms *MemoryStorer) Clone() Storer {
	copyStorer := *ms
	return &copyStorer
}
