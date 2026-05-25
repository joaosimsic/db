package storage

import (
	"errors"
	"io"
	"os"
)

const (
	BYTE_SIZE = 1024
	PAGE_SIZE = BYTE_SIZE * 4
)

type Page struct {
	ID   uint64
	Data [PAGE_SIZE]byte
}

func (p *Page) Clear() {
	for i := range p.Data {
		p.Data[i] = 0
	}
}

func NewDiskManager(filepath string) (*DiskManager, error) {
	const unixPermission = 0o666

	file, err := os.OpenFile(filepath, os.O_RDWR|os.O_CREATE, unixPermission)
	if err != nil {
		return nil, err
	}

	return &DiskManager{file: file}, nil
}

func (dm *DiskManager) Close() error {
	return dm.file.Close()
}

func (dm *DiskManager) WritePage(p *Page) error {
	offset := int64(p.ID * PAGE_SIZE)

	_, err := dm.file.WriteAt(p.Data[:], offset)

	return err
}

func (dm *DiskManager) ReadPage(pageID uint64) (*Page, error) {
	offset := int64(pageID * PAGE_SIZE)

	p := &Page{ID: pageID}

	_, err := dm.file.ReadAt(p.Data[:], offset)
	if err != nil && !errors.Is(err, io.EOF) {
		return nil, err
	}

	return p, nil
}
