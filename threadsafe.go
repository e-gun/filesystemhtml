package filesystemhtml

import (
	"cmp"
	"slices"
	"sync"
)

//
// THREAD SAFE INFRASTRUCTURE: MUTEX
//

type fsresp struct {
	HTML  string
	JS    string
	mutex sync.RWMutex
}

func makefsresp() fsresp {
	return fsresp{
		HTML:  "",
		JS:    "",
		mutex: sync.RWMutex{},
	}
}

func (r *fsresp) Set(html string, js string) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	r.HTML = html
	r.JS = js
}

func (r *fsresp) GetHTML() string {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	return r.HTML
}

func (r *fsresp) GetJS() string {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	return r.JS
}

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

func (ss *servingslice) GetLen() int {
	ss.mutex.Lock()
	defer ss.mutex.Unlock()
	return len(ss.FSEs)
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

func (sm *servingmap) GetLen() int {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()
	return len(sm.EntryMap)
}
