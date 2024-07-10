package filesystemhtml

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
	ServingDirs  = makeservingmap()
	ServeFileMap = makeservingmap()
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
		// nb: crazy inner folder permissions can kill this: 640, e.g, when you need 750 for the 'x'
		fmt.Printf("WatchFS() can not open FServeDir: %s\n", FSDir)
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
	ServeFileMap.WriteAll(buildwatcherfmap())
	ServingDirs.WriteAll(buildwatcherdirmap())
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
	for _, d := range ServingDirs.ReadAll() {
		newd := checkforuserrestrictionatthislevel(d)
		ServingDirs.WriteOne(newd)
	}
	for _, d := range ServingDirs.ReadAll() {
		newd := checkforuserrestrictionabove(d)
		ServingDirs.WriteOne(newd)
	}
}

func checkforuserrestrictionatthislevel(d FSEntry) FSEntry {
	var restrict []string
	contents := contentsofthisdirectory(d)
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
	return d
}

func checkforuserrestrictionabove(d FSEntry) FSEntry {
	for p := range d.Parents {
		ff := ServeFileMap.ReadOne(p)
		for u := range ff.Users {
			d.Users[u] = true
		}
	}
	return d
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
		thisentry := finfintofsentry(path, f)
		entries = append(entries, thisentry)
	}
	slices.SortFunc(entries, func(a, b FSEntry) int { return cmp.Compare(a.RelPath, b.RelPath) })
	return entries
}

func finfintofsentry(p string, f fs.FileInfo) FSEntry {
	abs, err := filepath.Abs(FSDir)
	if err != nil {
		log.Fatalln(err)
	}

	var thisentry FSEntry
	thisentry.RelPath, _ = strings.CutPrefix(p, abs+"/")
	thisentry.Mode = f.Mode()
	thisentry.IsDir = f.Mode().IsDir()
	thisentry.Genealogy = strings.Split(thisentry.RelPath, "/")
	thisentry.Level = len(thisentry.Genealogy)
	thisentry.Name = f.Name()
	fi, e2 := os.Stat(p)
	if e2 != nil {
		// the file disappeared on you...
		fmt.Printf("finfintofsentry() could not os.Stat(path) for '%s'\n", p)
		return FSEntry{}
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

	if strings.HasPrefix(thisentry.Name, ".") {
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
		ent := finfintofsentry(AbsPath+d.MyRelativePath(), inf)
		contents = append(contents, ent)
	}

	slices.SortFunc(contents, func(a, b FSEntry) int { return cmp.Compare(a.RelPath, b.RelPath) })
	return contents
}

func buildfsresponse() [2]string {
	return [2]string{fsdeephtml(), insertfilejs()}
}
