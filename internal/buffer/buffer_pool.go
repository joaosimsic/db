package buffer

import "db/internal/storage"

type Frame struct {
	Page     *storage.Page
	PinCount int
	IsDirty  bool
}

type BufferPoolManager struct {
	diskManager *storage.DiskManager
	poolSize    int
	frames      []*Frame
	pageTable   map[uint64]int
}

func NewBufferPoolManager(poolSize int, dm *storage.DiskManager) *BufferPoolManager {
	frames := make([]*Frame, poolSize)

	for i := range poolSize {
		frames[i] = &Frame{Page: &storage.Page{}}
	}

	return &BufferPoolManager{
		diskManager: dm,
		poolSize:    poolSize,
		frames:      frames,
		pageTable:   make(map[uint64]int),
	}
}

func (bpm *BufferPoolManager) getVictimFrameID() (int, error) {
	const placeholder = 0
	return placeholder, nil
}

func (bpm *BufferPoolManager) FetchPage(pageID uint64) (*storage.Page, error) {
	if frameID, exists := bpm.pageTable[pageID]; exists {
		bpm.frames[frameID].PinCount++

		return bpm.frames[frameID].Page, nil
	}

	victimFrameID, err := bpm.getVictimFrameID()
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

	newPage, err := bpm.diskManager.ReadPage(pageID)
	if err != nil {
		return nil, err
	}

	victimFrame.Page = newPage
	victimFrame.PinCount = 1
	victimFrame.IsDirty = false

	bpm.pageTable[pageID] = victimFrameID

	return victimFrame.Page, nil
}
