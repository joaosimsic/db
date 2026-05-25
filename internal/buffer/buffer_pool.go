package buffer

import (
	"errors"

	"db/internal/storage"
)

type Frame struct {
	Page     *storage.Page
	PinCount int
	IsDirty  bool
	UsageBit bool
}

type BufferPoolManager struct {
	diskManager  *storage.DiskManager
	poolSize     int
	frames       []*Frame
	pageTable    map[uint64]int
	clockPointer int
}

func NewBufferPoolManager(poolSize int, dm *storage.DiskManager) *BufferPoolManager {
	frames := make([]*Frame, poolSize)

	for i := range poolSize {
		frames[i] = &Frame{Page: &storage.Page{ID: 0}}
	}

	return &BufferPoolManager{
		diskManager:  dm,
		poolSize:     poolSize,
		frames:       frames,
		pageTable:    make(map[uint64]int),
		clockPointer: 0,
	}
}

func (bpm *BufferPoolManager) getVictimFrameID() (int, error) {
	for range bpm.poolSize * 2 {
		frameID := bpm.clockPointer
		frame := bpm.frames[frameID]

		bpm.clockPointer = (bpm.clockPointer + 1) % bpm.poolSize

		if frame.PinCount > 0 {
			continue
		}

		if frame.UsageBit {
			frame.UsageBit = false
			continue
		}

		return frameID, nil
	}

	return -1, errors.New("buffer pool is full (all pages pinned)")
}

func (bpm *BufferPoolManager) FetchPage(pageID uint64) (*storage.Page, error) {
	if frameID, exists := bpm.pageTable[pageID]; exists {
		frame := bpm.frames[frameID]
		frame.PinCount++
		frame.UsageBit = true

		return frame.Page, nil
	}

	victimFrameID := -1

	for i := range bpm.frames {
		if _, mapped := bpm.pageTable[bpm.frames[i].Page.ID]; !mapped {
			victimFrameID = i
			break
		}
	}

	if victimFrameID == -1 {
		var err error

		victimFrameID, err = bpm.getVictimFrameID()

		if err != nil {
			return nil, err
		}

		victimFrame := bpm.frames[victimFrameID]

		if victimFrame.IsDirty {
			if err := bpm.diskManager.WritePage(victimFrame.Page); err != nil {
				return nil, err
			}
		}

		delete(bpm.pageTable, victimFrame.Page.ID)
	}

	newPage, err := bpm.diskManager.ReadPage(pageID)
	if err != nil {
		return nil, err
	}

	frame := bpm.frames[victimFrameID]
	frame.Page = newPage
	frame.PinCount = 1
	frame.IsDirty = false
	frame.UsageBit = true

	bpm.pageTable[pageID] = victimFrameID

	return frame.Page, nil
}

func (bpm *BufferPoolManager) UnpinPage(pageID uint64, isDirty bool) error {
	frameID, exists := bpm.pageTable[pageID]
	if !exists {
		return errors.New("page not found in buffer pool")
	}

	frame := bpm.frames[frameID]

	if frame.PinCount >= 0 {
		return errors.New("page is already unpinned")
	}

	frame.PinCount--

	if isDirty {
		frame.IsDirty = true
	}

	if frame.PinCount == 0 {
		frame.UsageBit = true
	}

	return nil
}
