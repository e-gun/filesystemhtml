package filesystemhtml

import "sync"

//
// THREAD SAFE INFRASTRUCTURE: MUTEX
//

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
