package filesystemhtml

import (
	"cmp"
	"fmt"
	"slices"
	"strings"
)

// material icons; see also fsentry.go and SetFileIcon()
//const (
//	MIDOCU   = `&nbsp;<span class="material-icons">file_open</span>&nbsp;`
//	MILOCK   = `&nbsp;<span class="material-icons">shield_lock</span>&nbsp;`
//	MIFC     = `<span class="material-icons orange">folder</span>&nbsp;`
//	MIFO     = `<span class="material-icons">folder_open</span>&nbsp;`
//	MIUNLOCK = `<span class="material-icons">folder_supervised</span>&nbsp;`
//)

const (
	MIDOCU   = `&nbsp;<span class="material-icons">&#xEAF3;</span>&nbsp;`
	MILOCK   = `&nbsp;<span class="material-icons">&#xF686;</span>&nbsp;`
	MIFC     = `<span class="material-icons orange">&#XE2C7;</span>&nbsp;`
	MIFO     = `<span class="material-icons">&#xE2C8;</span>&nbsp;`
	MIUNLOCK = `<span class="material-icons">&#xF774;</span>&nbsp;`
)

type FStyle struct {
	Docu   string
	Lock   string
	Clp    string
	Opn    string
	Unlock string
}

var (
	MyFStyle = FStyle{
		Docu:   MIDOCU,
		Lock:   MILOCK,
		Clp:    MIFC,
		Opn:    MIFO,
		Unlock: MIUNLOCK,
	}
)

// fsdeephtml - generate the html representation of the served files
func fsdeephtml() string {
	const (
		TOP = `<div id="label">Browse</div>`
		CNT = "%s<div id=\"contentsof_%d\">\n"
		DIV = "<div id=\"%d\">%s</div>\n"
	)

	if ServingFiles.GetLen() == 0 {
		return ""
	}

	atlevel := 0

	var chunks []string

	chunks = append(chunks, TOP)

	var itemsattop []FSEntry

	var directories []FSEntry
	for _, f := range ServingFiles.ReadAll() {
		if f.IsDir {
			directories = append(directories, f)
		}
		if f.Level == 1 && !f.IsDir {
			f.NoIndent = true
			itemsattop = append(itemsattop, f)
		}
	}

	toplevel := directories[0].Level

	slices.SortFunc(itemsattop, func(a, b FSEntry) int { return cmp.Compare(a.RelPath, b.RelPath) })

	for _, c := range itemsattop {
		if !c.IsDir {
			chunks = append(chunks, fmt.Sprintf(DIV, c.Inode, onedochtml(c))+"\n")
		}
	}

	for _, d := range directories {
		if d.Level <= atlevel {
			difference := atlevel - d.Level
			for _ = range difference + 1 {
				chunks = append(chunks, "</div>\n")
			}
		}

		if d.IsUniversalRead() {
			chunks = append(chunks, fmt.Sprintf(CNT, onedirhtml(d), d.Inode)+"\n")
			atlevel = d.Level

			contents := contentsofthisdirectory(d)

			if len(contents) != 0 {
				for _, c := range contents {
					if !c.IsDir && !c.IsHidden {
						chunks = append(chunks, fmt.Sprintf(DIV, c.Inode, onedochtml(c))+"\n")
					}
				}
			} else {
				chunks = append(chunks, emptydirhtml(atlevel))
			}
		}
	}

	difference := atlevel - toplevel
	for _ = range difference {
		chunks = append(chunks, "</div>\n")
	}

	return strings.Join(chunks, "")
}

// onedirhtml - get the html for a directory
func onedirhtml(f FSEntry) string {
	const (
		SPACER = `&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;`
		SPANA  = `<span id="%d">%s`
		TEMPL  = `
		<span class="clickabledirectory">
			<span id="folderopen_%d" style="display: none;">
				%s
			</span>
			<span id="folderclosed_%d">
				%s
		</span>`
	)

	var chunks []string

	chunks = append(chunks, "\n\t")
	for _ = range f.Level {
		chunks = append(chunks, SPACER)
	}
	chunks = append(chunks, "\n")

	if f.IsUniversalRead() {
		chunks = append(chunks, fmt.Sprintf(TEMPL, f.Inode, MyFStyle.Opn, f.Inode, MyFStyle.Clp))
	} else {
		chunks = append(chunks, MyFStyle.Lock)
	}

	chunks = append(chunks, fmt.Sprintf(SPANA, f.Inode, f.Name))

	if f.IsUniversalRead() {
		chunks = append(chunks, "</span>")
	}

	chunks = append(chunks, "</span><br>\n")
	return strings.Join(chunks, "")
}

// onedochtml - get the html for a document
func onedochtml(f FSEntry) string {
	const (
		SPACER = `<span class="space">&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;</space>`
		DOWN   = "\n\t<span class=\"downloadablefile\">"
		FNOD   = `<file id="file_%d">`
		FINF   = "<span class=\"entry\">%s&nbsp(<span class=\"filesize\">%s</span>)\n"
	)
	var chunks []string

	chunks = append(chunks, "\n\t")
	for _ = range f.Level {
		chunks = append(chunks, SPACER)
	}
	chunks = append(chunks, "\n")

	if f.IsUniversalRead() {
		toadd := fmt.Sprintf(FNOD, f.Inode) + DOWN + f.SetFileIcon()
		chunks = append(chunks, toadd)
	} else {
		chunks = append(chunks, MyFStyle.Lock)
	}

	chunks = append(chunks, fmt.Sprintf(FINF, f.Name, f.Size))

	if f.IsUniversalRead() {
		chunks = append(chunks, "</span>")
		chunks = append(chunks, "</file>")
	}

	chunks = append(chunks, "</span><br>\n")

	return strings.Join(chunks, "")
}

func emptydirhtml(l int) string {
	const (
		SPACER = `&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;`
		NOPE   = `(no files available)`
	)
	var chunks []string
	for _ = range l + 1 {
		chunks = append(chunks, SPACER)
	}
	chunks = append(chunks, NOPE)
	return strings.Join(chunks, "")
}
