//go:build !js
// +build !js

// Package osfs provides a billy filesystem for the OS.
package osfs

import (
	"fmt"
	"io/fs"
	"os"
	"sync"

	"github.com/go-git/go-billy/v5"
)

const (
	defaultDirectoryMode = 0o755
	defaultCreateMode    = 0o666
)

// New returns a new OS filesystem.
func New(baseDir string, opts ...Option) billy.Filesystem {
	o := &options{}
	for _, opt := range opts {
		opt(o)
	}

	if o.Type == BoundOSFS {
		return newBoundOS(baseDir)
	}

	return newChrootOS(baseDir)
}

// WithBoundOS returns the option of using a Bound filesystem OS.
func WithBoundOS() Option {
	return func(o *options) {
		o.Type = BoundOSFS
	}
}

// WithChrootOS returns the option of using a Chroot filesystem OS.
func WithChrootOS() Option {
	return func(o *options) {
		o.Type = ChrootOSFS
	}
}

type options struct {
	Type
}

type Type int

const (
	ChrootOSFS Type = iota
	BoundOSFS
)

func readDir(dir string) ([]os.FileInfo, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	infos := make([]fs.FileInfo, 0, len(entries))
	for _, entry := range entries {
		fi, err := entry.Info()
		if err != nil {
			return nil, err
		}
		infos = append(infos, fi)
	}
	return infos, nil
}

func tempFile(dir, prefix string) (billy.File, error) {
	f, err := os.CreateTemp(dir, prefix)
	if err != nil {
		return nil, err
	}
	return &file{File: f}, nil
}

func openFile(fn string, flag int, perm os.FileMode, createDir func(string) error) (billy.File, error) {
	if flag&os.O_CREATE != 0 {
		if createDir == nil {
			return nil, fmt.Errorf("createDir func cannot be nil if file needs to be opened in create mode")
		}
		if err := createDir(fn); err != nil {
			return nil, err
		}
	}

	f, err := os.OpenFile(fn, flag, perm)
	if err != nil {
		return nil, err
	}
	return &file{File: f}, err
}

// file is a wrapper for an os.File which adds support for file locking.
type file struct {
	*os.File
	m sync.Mutex
}
