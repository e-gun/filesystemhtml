## html and javascript to generate a clickable file system hierarchy underneath a specified directory

the directory is watched; modifications are reflected in the output 

as a package:
1. set `filesystemhtml.AbsPath` and `filesystemhtml.FSDir`
1. start `filesystemhtml.WatchFS()`
1. query `filesystemhtml.FSResponse.GetHTML()` and `filesystemhtml.FSResponse.GetJS()` as needed

nb: assuming you have access to `jQuery` and `Material Symbols`

```
% tree
.
├── adir
│   ├── aa1
│   ├── asubdir
│   ├── asubdir2
│   │   ├── f1
│   │   └── f2
│   ├── img.jpg
│   ├── mv.avi
│   ├── p.pdf
│   └── wd.doc
├── afile.txt
├── bdir
│   ├── lock.txt
│   ├── prot
│   │   ├── pf1.txt
│   │   └── pf2.png
│   └── pwd
│       ├── bleh
│       └── z_user_is_user2
└── bfile.txt

7 directories, 14 files

```

![all closed](./gitimg/all-closed.png)
![semi-open](./gitimg/semi-open.png)
![all-open](./gitimg/all-open.png)

