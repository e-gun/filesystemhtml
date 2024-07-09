package fse

import (
	"cmp"
	"fmt"
	"github.com/dustin/go-humanize"
	"github.com/radovskyb/watcher"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"syscall"
	"time"
)

// todo: address the many race condition candidates...

var (
	ServingFiles []FSEntry
	AbsPath      string
	ServingDirs  = make(map[uint64]FSEntry)
	ServeFileMap = make(map[uint64]FSEntry)
	FSResponse   [2]string
	FSDir        string
)

func WatchFS() {
	w := watcher.New()

	go func() {
		for {
			select {
			case event := <-w.Event:
				// fmt.Println(event) // Print the event's info.
				if event.Op.String() == "CREATE" || event.Op.String() == "REMOVE" || event.Op.String() == "MOVE" {
					// this can be too fast: file might not exist yet
					time.Sleep(time.Millisecond * 10)
					reloadfsinfo(w)
				}
			case err := <-w.Error:
				log.Fatalln(err)
			case <-w.Closed:
				return
			}
		}
	}()

	// Watch this folder for changes.
	if err := w.AddRecursive(FSDir); err != nil {
		fmt.Printf("watchfs() can not open Config.FServeDir: %s\n", FSDir)
		return
	}

	reloadfsinfo(w)

	// Start the watching process
	if e := w.Start(time.Second * 1); e != nil {
		log.Fatalln(e)
	}
}

func reloadfsinfo(w *watcher.Watcher) {
	ServingFiles = buildwatcherentries(w)
	slices.SortFunc(ServingFiles, func(a, b FSEntry) int { return cmp.Compare(a.RelPath, b.RelPath) })
	ServeFileMap = buildwatcherfmap()
	ServingDirs = buildwatcherdirmap()
	populaterestrictions()
	FSResponse = buildfsresponse()
}

func findparentfromname(n string) FSEntry {
	var parent FSEntry
	for _, sf := range ServingFiles {
		if sf.MyRelativePath() == n && sf.IsDir {
			parent = sf
			break
		}
	}
	return parent
}

func buildwatcherfmap() map[uint64]FSEntry {
	sf := make(map[uint64]FSEntry)
	for _, f := range ServingFiles {
		if !f.IsDir {
			sf[f.Inode] = f
		}
	}
	return sf
}

func buildwatcherdirmap() map[uint64]FSEntry {
	sd := make(map[uint64]FSEntry)
	for _, f := range ServingFiles {
		if f.IsDir {
			sd[f.Inode] = f
		}
	}
	return sd
}

func populaterestrictions() {
	for _, d := range ServingDirs {
		checkforuserrestrictionatthislevel(&d)
	}
	for _, d := range ServingDirs {
		checkforuserrestrictionabove(&d)
	}
}

func checkforuserrestrictionatthislevel(d *FSEntry) {
	var restrict []string
	contents := contentsofthisdirectory(*d)
	for _, c := range contents {
		if strings.HasPrefix(c.Name, RESTRICTPREFIX) {
			u, _ := strings.CutPrefix(c.Name, RESTRICTPREFIX)
			// fmt.Printf("restricting %s to %s\n", d.Name, u)
			restrict = append(restrict, u)
		}
	}

	for _, u := range restrict {
		d.Users[u] = true
	}
}

func checkforuserrestrictionabove(d *FSEntry) {
	for p := range d.Parents {
		for u := range ServeFileMap[p].Users {
			d.Users[u] = true
		}
	}
}

func buildwatcherentries(w *watcher.Watcher) []FSEntry {
	abs, err := filepath.Abs(FSDir)
	if err != nil {
		log.Fatalln(err)
	}

	var entries []FSEntry
	for path, f := range w.WatchedFiles() {
		if path == abs {
			// skip FSDir itself
			continue
		}
		var thisentry FSEntry
		thisentry.RelPath, _ = strings.CutPrefix(path, abs+"/")
		thisentry.Mode = f.Mode()
		thisentry.IsDir = f.Mode().IsDir()
		thisentry.Genealogy = strings.Split(thisentry.RelPath, "/")
		thisentry.Level = len(thisentry.Genealogy)
		thisentry.Name = f.Name()
		fi, e2 := os.Stat(path)
		if e2 != nil {
			// the file disappeared on you...
			fmt.Printf("loadwatcherentries() could not os.Stat(path) for '%s'\n", path)
			continue
		}
		stat, ok := fi.Sys().(*syscall.Stat_t)
		if !ok {
			// fmt.Printf("unable to find inode via syscall.Stat_t")
		} else {
			thisentry.Inode = stat.Ino
		}

		if !thisentry.IsDir {
			thisentry.Size = fmt.Sprintf("%s", humanize.Bytes(uint64(stat.Size)))
		}

		if strings.HasPrefix(thisentry.Name, RESTRICTPREFIX) {
			thisentry.IsHidden = true
		}

		thisentry.Users = make(map[string]bool)
		thisentry.Contents = make(map[uint64]bool)
		thisentry.Parents = make(map[uint64]bool)

		entries = append(entries, thisentry)
	}
	slices.SortFunc(entries, func(a, b FSEntry) int { return cmp.Compare(a.RelPath, b.RelPath) })
	return entries
}

func buildfsentryfromfinfo(d FSEntry, f fs.FileInfo) FSEntry {
	// DRY issues with buildwatcherentries

	path := AbsPath + d.MyRelativePath()

	var thisentry FSEntry
	thisentry.RelPath, _ = strings.CutPrefix(path, AbsPath)
	thisentry.Mode = f.Mode()
	thisentry.IsDir = f.Mode().IsDir()
	thisentry.Genealogy = strings.Split(thisentry.RelPath, "/")
	thisentry.Level = len(thisentry.Genealogy)
	thisentry.Name = f.Name()

	fi, e2 := os.Stat(path + "/" + f.Name())
	if e2 != nil {
		fmt.Println(fmt.Sprintf("buildfsentryfromfinfo() os.Stat failed for '%s'", path))
		return thisentry
	}
	stat, ok := fi.Sys().(*syscall.Stat_t)
	if !ok {
		// fmt.Printf("unable to find inode via syscall.Stat_t")
	} else {
		thisentry.Inode = stat.Ino
	}

	if !thisentry.IsDir {
		thisentry.Size = fmt.Sprintf("%s", humanize.Bytes(uint64(stat.Size)))
	}

	if strings.HasPrefix(thisentry.Name, RESTRICTPREFIX) {
		thisentry.IsHidden = true
	}

	thisentry.Users = make(map[string]bool)
	thisentry.Contents = make(map[uint64]bool)
	thisentry.Parents = make(map[uint64]bool)

	return thisentry
}

func contentsofthisdirectory(d FSEntry) []FSEntry {
	var contents []FSEntry
	cnt, err := os.ReadDir(AbsPath + d.MyRelativePath())
	if err != nil {
		fmt.Println(fmt.Sprintf("contentsofthisdirectory() os.ReadDir failed for '%s'", AbsPath+d.MyRelativePath()))
		return contents
	}

	for _, fi := range cnt {
		inf, e2 := fi.Info()
		if e2 != nil {
			fmt.Println(fmt.Sprintf("contentsofthisdirectory() fi.Info() failed for '%s'", fi.Name()))
			continue
		}
		ent := buildfsentryfromfinfo(d, inf)
		contents = append(contents, ent)
	}

	slices.SortFunc(contents, func(a, b FSEntry) int { return cmp.Compare(a.RelPath, b.RelPath) })
	return contents
}

func buildfsresponse() [2]string {
	return [2]string{fsdeephtml(), insertfilejs()}
}
