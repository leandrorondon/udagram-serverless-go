package datalayer

import "os"

var (
	TableName = os.Getenv("UDAGRAM_TABLE")
	PhotosIdx = os.Getenv("PHOTOS_IDX")
)
