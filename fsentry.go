package fse

import (
	"fmt"
	"io/fs"
	"strings"
)

type FSEntry struct {
	Inode     uint64
	Mode      fs.FileMode
	Name      string
	RelPath   string
	Level     int
	Genealogy []string
	Parents   map[uint64]bool
	Users     map[string]bool
	IsDir     bool
	Size      string // for files
	NoIndent  bool
	Contents  map[uint64]bool // for directories
	IsHidden  bool
}

func (f *FSEntry) FindParent() FSEntry {
	myparentname := strings.Join(f.Genealogy[0:len(f.Genealogy)-1], "/")
	return findparentfromname(myparentname)
}

func (f *FSEntry) PopulateParentMap() {
	// only call this before checking a download
	// otherwise you waste cycles
	myparentname := strings.Join(f.Genealogy[0:len(f.Genealogy)-1], "/")
	for _ = range f.Genealogy {
		p := findparentfromname(myparentname)
		f.Parents[p.Inode] = true
		myparentname = strings.Join(p.Genealogy[0:len(p.Genealogy)-1], "/")
	}
	fmt.Println(f.Name)
	fmt.Println(f.Parents)
}

func (f *FSEntry) MyRelativePath() string {
	return strings.Join(f.Genealogy, "/")
}

func (f *FSEntry) IsUniversalRead() bool {
	p := fmt.Sprintf("%s", f.Mode.Perm()) // "-rw-r--r--"
	if string(p[len(p)-3]) == "r" {
		return true
	} else {
		return false
	}
}

func (f *FSEntry) IsChildOfReadableParents() bool {
	mypath := f.Genealogy
	fmt.Println(mypath)
	var parents []FSEntry
	for i := range mypath {
		fmt.Printf("file: %s \t parent: %s\n", f.Name, strings.Join(mypath[0:i], "/"))
		parents = append(parents, findparentfromname(strings.Join(mypath[0:i], "/")))
	}

	canread := true
	for _, parent := range parents {
		fmt.Printf("file: %s \t parent: %s \t parent IsUniversalRead: %t\n", f.Name, parent.Name, parent.IsUniversalRead())
		if !parent.IsUniversalRead() {
			canread = false
		}
	}
	return canread
}
