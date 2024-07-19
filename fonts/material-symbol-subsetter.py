from fontTools import subset
args = [
	"material-symbols-outlined.woff2",
	"--glyphs=shield,verified_user,file_open,shield_lock,folder,folder_open,folder_supervised,audio_file,video_file,file_open,file_present,article,image,picture_as_pdf,folder_zip",
	"--output-file=./Material-SymbolsSubset.woff2",
	"--flavor=woff2",
]

# (filesystem)
#  EAF3    FILE_OPEN
#  F686    SHIELD_LOCK
#  E2C7    FOLDER
#  E2C8    FOLDER_OPEN
#  F774    FOLDER_SUPERVISED
#  EB82    AUDIO_FILE
#  EB87    VIDEO_FILE
#  EAF3    FILE_OPEN
#  EA0E    FILE_PRESENT
#  EF42    ARTICLE
#  E3F4    IMAGE
#  E415    PICTURE_AS_PDF
#  EB2C    FOLDER_ZIP

# (add ons)
#  E9E0    SHIELD
#  E8E8    VERIFIED_USER

subset.main(args)

