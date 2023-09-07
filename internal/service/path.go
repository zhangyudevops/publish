package service

import (
	"context"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gfile"
)

type sPath struct{}

func Path() *sPath {
	return &sPath{}
}

// getFile returns a list of files in the specified path.
// The parameter `path` specifies the subdirectory path to be scanned.
// The parameter `pattern` specifies the file name pattern to be scanned.
func (s *sPath) getFile(path, pattern string) ([]string, error) {
	file, err := gfile.ScanDir(path, pattern, false)
	if err != nil {
		return nil, err
	}
	return file, nil
}

func (s *sPath) CreateDir(ctx context.Context, path string) error {
	if !gfile.Exists(path) {
		err := gfile.Mkdir(path)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *sPath) CopyFileAndDir(src, dst string) error {
	return gfile.Copy(src, dst)
}

// GetFilePath returns the file path of the k8s file path.
func (s *sPath) getFilePath(ctx context.Context) (FilePath string, err error) {
	path, _ := g.Config().Get(ctx, "file.path")
	return path.String(), err
}

// GetFile returns a list of files directory in the specified path.
func (s *sPath) GetFile(ctx context.Context, path, pattern string) ([]string, error) {
	// if pattern is empty, return all files
	if pattern == "" {
		pattern = "*"
	}

	// @todo: gfile.Basename 来获取文件名，需要循环s.getFile(path, pattern)的结果，再组装成一个新的列表返回
	return s.getFile(path, pattern)
}
