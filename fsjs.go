package filesystemhtml

import (
	"fmt"
	"strings"
)

func insertfilejs() string {
	var directories []FSEntry
	for _, f := range ServingFiles.ReadAll() {
		if f.IsDir {
			directories = append(directories, f)
		}
	}

	return filejqueryjs() + folderjqueryjs(directories)
}

func filejqueryjs() string {
	const (
		FSJS = `
	$('file').click( function(e) {
		e.preventDefault();
		// prefix is 'file_'
		var node = this.id;
		node = node.slice(5);
		$.getJSON('/checkauthorizationfor/'+node, function (data) {
			console.log(data);
			if (data == 'authorized') {
				window.location.href='/getfile/'+node; 
			} else { 
				$('#validateusers').dialog( "open" );
			}
		});
	});
`
	)
	return FSJS
}

func folderjqueryjs(directories []FSEntry) string {
	const (
		FSJS = `
	// clicks for %d (%s)
		$("#contentsof_%d").hide();
		$("#folderopen_%d").hide();
		$("#folderclosed_%d").show();
		function toggler_%d() {
			$("#contentsof_%d").toggle();
			$("#folderopen_%d").toggle();
			$("#folderclosed_%d").toggle();
		}
		$("#%d").click( function() {
			toggler_%d();
		});
		$("#folderopen_%d").click( function() {
			toggler_%d();
			});
		$("#folderclosed_%d").click( function() {
			toggler_%d();
			});
`
	)

	var newjs []string
	for _, d := range directories {
		newjs = append(newjs, fmt.Sprintf(FSJS, d.Inode, d.Name,
			d.Inode, d.Inode, d.Inode,
			d.Inode, d.Inode, d.Inode, d.Inode,
			d.Inode, d.Inode,
			d.Inode, d.Inode,
			d.Inode, d.Inode))
	}
	return strings.Join(newjs, "")
}
