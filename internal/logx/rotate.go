package logx

import (
	"fmt"
	"os"
	"sync"
)

// File is an append-only log file that starts a new file when it reaches a size limit, so a server left
// running for months cannot fill the disk. The old files are kept as <path>.1 (newest) to <path>.<keep>.
type File struct {
	mu   sync.Mutex
	path string
	max  int64
	keep int
	f    *os.File
	size int64
}

// DefaultMaxBytes and DefaultKeep bound a log file to about 20 MB in all.
const (
	DefaultMaxBytes = 5 << 20
	DefaultKeep     = 3
)

// OpenFile opens (creating it if needed) the log file at path for appending.
func OpenFile(path string, maxBytes int64, keep int) (*File, error) {
	l := &File{path: path, max: maxBytes, keep: keep}
	if err := l.open(); err != nil {
		return nil, err
	}
	return l, nil
}

func (l *File) open() error {
	f, err := os.OpenFile(l.path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return err
	}
	fi, err := f.Stat()
	if err != nil {
		f.Close()
		return err
	}
	l.f, l.size = f, fi.Size()
	return nil
}

// Write implements io.Writer. A line is never split between two files.
func (l *File) Write(p []byte) (int, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.size > 0 && l.size+int64(len(p)) > l.max {
		if err := l.rotate(); err != nil {
			return 0, err
		}
	}
	n, err := l.f.Write(p)
	l.size += int64(n)
	return n, err
}

func (l *File) rotate() error {
	l.f.Close()
	os.Remove(fmt.Sprintf("%s.%d", l.path, l.keep))
	for i := l.keep - 1; i >= 1; i-- {
		os.Rename(fmt.Sprintf("%s.%d", l.path, i), fmt.Sprintf("%s.%d", l.path, i+1))
	}
	if l.keep >= 1 {
		os.Rename(l.path, l.path+".1")
	} else {
		os.Remove(l.path)
	}
	return l.open()
}

// Close closes the file.
func (l *File) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.f.Close()
}
