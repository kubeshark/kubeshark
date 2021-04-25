package utils

import (
	"github.com/djherbis/atime"
	"os"
)

type ByModTime []os.FileInfo

func (fis ByModTime) Len() int {
	return len(fis)
}

func (fis ByModTime) Swap(i, j int) {
	fis[i], fis[j] = fis[j], fis[i]
}

func (fis ByModTime) Less(i, j int) bool {
	return fis[i].ModTime().Before(fis[j].ModTime())
}


type ByName []os.FileInfo

func (fis ByName) Len() int {
	return len(fis)
}

func (fis ByName) Swap(i, j int) {
	fis[i], fis[j] = fis[j], fis[i]
}

func (fis ByName) Less(i, j int) bool {
	return fis[i].Name() < fis[j].Name()
}


type ByCreationTime []os.FileInfo

func (fis ByCreationTime) Len() int {
	return len(fis)
}

func (fis ByCreationTime) Swap(i, j int) {
	fis[i], fis[j] = fis[j], fis[i]
}

func (fis ByCreationTime) Less(i, j int) bool {
	return atime.Get(fis[i]).Unix() < atime.Get(fis[j]).Unix()
}