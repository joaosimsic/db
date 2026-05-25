package storage

import "os"

type DiskManager struct {
	file *os.File
}
