package fanpath

import (
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

var (
	executeDir  string
	exeFilePath string
)

func InitExecutePath() error {
	p, err := os.Executable()
	if err != nil {
		return err
	}
	exeFilePath = p
	executeDir = filepath.Dir(exeFilePath)
	return nil
}

func ExecuteFilePath() string {
	return exeFilePath
}

func ExecuteParentPath() string {
	return executeDir
}

// RelExecuteDir 获取相对可执行文件所在目录
func RelExecuteDir(paths ...string) string {
	paths = append([]string{executeDir}, paths...)
	return filepath.Join(paths...)
}

// AbsOrRelExecutePath 获取绝对路径或者相对可执行文件所在目录的路径
func AbsOrRelExecutePath(path string) string {
	if filepath.IsAbs(path) {
		return path
	}
	return RelExecuteDir(path)
}

// SetWorkDirToExecuteDir 把工作目录设置为exe的目录
func SetWorkDirToExecuteDir() error {
	err := os.Chdir(executeDir)
	return err
}

// ExistPath 路径是否存在
func ExistPath(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// ExistFile 文件是否存在
func ExistFile(path string) bool {
	f, err := os.Stat(path)
	if err != nil {
		return false
	}

	return !f.IsDir()
}

// ExistDir 文件夹是否存在
func ExistDir(path string) bool {
	f, err := os.Stat(path)
	if err != nil {
		return false
	}

	return f.IsDir()
}

// CreateDirIfNoExist 如果目录不存在就创建
func CreateDirIfNoExist(path string) error {
	if ExistDir(path) {
		return nil
	}
	return os.MkdirAll(path, os.ModePerm)
}

// GetFileListByExt 获取某个目录下ext扩展名的所有文件
func GetFileListByExt(dir string, ext string) ([]string, error) {
	var fileLists []string
	err := filepath.Walk(dir, func(path string, f os.FileInfo, err error) error {
		if f == nil {
			return err
		}
		if f.IsDir() {
			return nil
		}
		if filepath.Ext(path) == ext {
			fileLists = append(fileLists, path)
		}
		return nil
	})
	return fileLists, err
}

// ClearDirAndCreateNew 清空目录并重新创建
func ClearDirAndCreateNew(path string) error {
	if ExistPath(path) {
		err := os.RemoveAll(path)
		if err != nil {
			return err
		}
	}
	err := os.MkdirAll(path, os.ModePerm)
	return err
}

// InitDirAndClearFile 初始化目录并清空目录下指定文件
func InitDirAndClearFile(path string, removePattern string) error {
	if !ExistPath(path) {
		err := os.MkdirAll(path, os.ModePerm)
		if err != nil {
			return err
		}
	}
	reg, err := regexp.Compile(removePattern)
	if err != nil {
		return err
	}
	err = filepath.Walk(path, func(fileName string, f os.FileInfo, err error) error {
		if ok := reg.MatchString(fileName); !ok {
			return nil
		}
		err = os.Remove(fileName)
		return err
	})
	return err
}

// CopyFile 拷贝文件
func CopyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	err = CreateDirIfNoExist(filepath.Dir(dst))
	if err != nil {
		return err
	}

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return err
	}

	return dstFile.Sync()
}

// CopyDir 拷贝目录
func CopyDir(srcDir, dstDir string) error {
	if !ExistPath(dstDir) {
		err := os.MkdirAll(dstDir, os.ModePerm)
		if err != nil {
			return err
		}
	}
	return filepath.Walk(srcDir, func(srcPath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Ignore the source directory.
		if srcPath == srcDir {
			return nil
		}

		baseFileName := filepath.Base(srcPath)
		// Calculate the destination path.
		dstPath := filepath.Join(dstDir, baseFileName)

		// Check if it's a directory, create it if so.
		if info.IsDir() {
			return os.MkdirAll(dstPath, info.Mode())
		}

		// It's a file, so copy it.
		return CopyFile(srcPath, dstPath)
	})
}

// GetFileNameWithoutExt 获取文件名，不带后缀
func GetFileNameWithoutExt(pathName string) string {
	fileName := filepath.Base(pathName)
	return strings.TrimSuffix(fileName, filepath.Ext(fileName))
}

func ReadDir(dirName string) ([]os.FileInfo, error) {
	entries, err := os.ReadDir(dirName)
	if err != nil {
		return nil, err
	}
	infoList := make([]fs.FileInfo, 0, len(entries))
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			return nil, err
		}
		infoList = append(infoList, info)
	}
	return infoList, nil
}

// ByModTime 实现了sort.Interface，用于按修改时间排序文件。
type byModTime []os.FileInfo

func (a byModTime) Len() int           { return len(a) }
func (a byModTime) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a byModTime) Less(i, j int) bool { return a[i].ModTime().Before(a[j].ModTime()) }

// GetFileListByModTime 获取一个目录下的所有文件，并且按修改时间排序
func GetFileListByModTime(dir string) ([]string, error) {
	var fileLists []string
	fileInfoList, err := ReadDir(dir)
	if err != nil {
		return fileLists, err
	}

	sortFileInfo := byModTime(fileInfoList)
	sort.Sort(sortFileInfo)

	for _, fileInfo := range fileInfoList {
		if fileInfo.IsDir() {
			continue
		}
		fileLists = append(fileLists, filepath.Join(dir, fileInfo.Name()))
	}
	return fileLists, nil
}
