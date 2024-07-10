package filesystemhtml

import (
	"cmp"
	"slices"
	"sync"
)

//
// THREAD SAFE INFRASTRUCTURE: MUTEX
//

func makeservinslice() servingslice {
	return servingslice{
		FSEs:  []FSEntry{},
		mutex: sync.RWMutex{},
	}
}

type servingslice struct {
	FSEs  []FSEntry
	mutex sync.RWMutex
}

func (ss *servingslice) WriteAll(ee []FSEntry) {
	ss.mutex.Lock()
	defer ss.mutex.Unlock()
	ss.FSEs = ee
	return
}

func (ss *servingslice) ReadAll() []FSEntry {
	ss.mutex.Lock()
	defer ss.mutex.Unlock()
	return ss.FSEs
}

func (ss *servingslice) SortByName() {
	ss.mutex.Lock()
	defer ss.mutex.Unlock()
	slices.SortFunc(ss.FSEs, func(a, b FSEntry) int { return cmp.Compare(a.RelPath, b.RelPath) })
	return
}

func makeservingmap() servingmap {
	return servingmap{
		EntryMap: make(map[uint64]FSEntry),
		mutex:    sync.RWMutex{},
	}
}

type servingmap struct {
	EntryMap map[uint64]FSEntry
	mutex    sync.RWMutex
}

func (sm *servingmap) WriteOne(f FSEntry) {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()
	sm.EntryMap[f.Inode] = f
	return
}

func (sm *servingmap) WriteAll(fsm map[uint64]FSEntry) {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()
	sm.EntryMap = fsm
	return
}

func (sm *servingmap) ReadOne(i uint64) FSEntry {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()
	if f, ok := sm.EntryMap[i]; ok {
		return f
	} else {
		return FSEntry{}
	}
}

func (sm *servingmap) ReadAll() map[uint64]FSEntry {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()
	return sm.EntryMap
}
