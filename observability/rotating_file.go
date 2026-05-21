package observability

import (
	"fmt"
	"os"
	"sync"
)

// RotatingFileWriter represents a thread-safe size-based rolling file writer.
type RotatingFileWriter struct {
	mu         sync.Mutex
	filePath   string
	maxSize    int64
	maxBackups int
	size       int64
	file       *os.File
}

// NewRotatingFileWriter creates a new pre-opened RotatingFileWriter.
func NewRotatingFileWriter(path string, maxSize int64, maxBackups int) (*RotatingFileWriter, error) {
	w := &RotatingFileWriter{
		filePath:   path,
		maxSize:    maxSize,
		maxBackups: maxBackups,
	}
	err := w.openFile()
	if err != nil {
		return nil, err
	}
	return w, nil
}

func (w *RotatingFileWriter) openFile() error {
	info, err := os.Stat(w.filePath)
	if err == nil {
		w.size = info.Size()
	}
	f, err := os.OpenFile(w.filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	w.file = f
	return nil
}

// Write writes data to the current file, rotating it first if size is exceeded.
func (w *RotatingFileWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	writeLen := int64(len(p))
	if w.size+writeLen > w.maxSize {
		if err := w.rotate(); err != nil {
			return 0, err
		}
	}
	n, err := w.file.Write(p)
	w.size += int64(n)
	return n, err
}

func (w *RotatingFileWriter) rotate() error {
	if err := w.file.Close(); err != nil {
		return err
	}
	if err := w.renameBackups(); err != nil {
		return err
	}
	if err := os.Rename(w.filePath, w.filePath+".1"); err != nil && !os.IsNotExist(err) {
		return err
	}
	w.size = 0
	return w.openFile()
}

func (w *RotatingFileWriter) renameBackups() error {
	for i := w.maxBackups; i >= 1; i-- {
		src := fmt.Sprintf("%s.%d", w.filePath, i)
		dst := fmt.Sprintf("%s.%d", w.filePath, i+1)
		if i == w.maxBackups {
			_ = os.Remove(src)
			continue
		}
		if _, err := os.Stat(src); err == nil {
			_ = os.Rename(src, dst)
		}
	}
	return nil
}

// Close closes the current file safely.
func (w *RotatingFileWriter) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.file != nil {
		err := w.file.Close()
		w.file = nil
		return err
	}
	return nil
}
