package fanpath

import "testing"

func TestGetFileListByModTime(t *testing.T) {
	fileList, err := GetFileListByModTime("F:\\downloads")
	if err != nil {
		t.Fatal(err)
	}
	// print
	for _, file := range fileList {
		t.Log(file)
	}
}
