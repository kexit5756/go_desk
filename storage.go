package main

import (
	"crypto/sha1"
	"encoding/hex"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
)

const defaultRootFolderName = "ggnetwork"

// CASPathTransformFunc 对 key 计算 SHA-1，并将摘要拆分到多级目录中，避免
// 所有对象都堆在同一个目录。这里散列的是 key，而不是文件内容。
func CASPathTransformFunc(key string) PathKey {
	hash := sha1.Sum([]byte(key))
	hashStr := hex.EncodeToString(hash[:])

	blockSize := 5
	sliceLen := len(hashStr) / blockSize
	paths := make([]string, sliceLen)
	for i := 0; i < sliceLen; i++ {
		from, to := i*blockSize, (i+1)*blockSize
		paths[i] = hashStr[from:to]
	}
	return PathKey{
		PathName: strings.Join(paths, "/"),
		Filename: hashStr,
	}
}

// PathTransformFunc 将逻辑 key 转换为磁盘上的目录和文件名。
type PathTransformFunc func(key string) PathKey

// PathKey 保存 key 转换得到的目录部分和文件名。
type PathKey struct {
	PathName string
	Filename string
}

// FirstPathName 返回转换后路径中的第一级目录。
func (p PathKey) FirstPathName() string {
	paths := strings.Split(p.PathName, "/")
	if p.PathName == "" {
		return ""
	}
	return paths[0]
}

// FullPath 拼接转换后的目录和文件名。
func (p PathKey) FullPath() string {
	return filepath.Join(p.PathName, p.Filename)
}

// Store 在 Root 目录下按调用方 ID 分区读写对象。
type Store struct {
	StoreOpts
}

// StoreOpts 配置存储根目录以及 key 到文件路径的转换方式。
type StoreOpts struct {
	Root              string
	PathTransformFunc PathTransformFunc
}

var DefaultPathTransformFunc = func(key string) PathKey {
	return PathKey{
		PathName: key,
		Filename: key,
	}
}

// NewStore 在配置未提供路径转换函数或根目录时补上默认值。
func NewStore(opts StoreOpts) *Store {
	if opts.PathTransformFunc == nil {
		opts.PathTransformFunc = DefaultPathTransformFunc
	}
	if len(opts.Root) == 0 {
		opts.Root = defaultRootFolderName
	}
	return &Store{
		StoreOpts: opts,
	}
}

// Has 检查指定 id 和 key 对应的对象是否存在。
func (s *Store) Has(id string, key string) bool {
	pathKey := s.PathTransformFunc(key)
	fullPathWithRoot := filepath.Join(s.Root, id, pathKey.FullPath())
	_, err := os.Stat(fullPathWithRoot)
	return err == nil
}

// Clear 删除整个存储根目录及其内容。
func (s *Store) Clear() error {
	return os.RemoveAll(s.Root)
}

// Delete 删除指定 id 和 key 对应的对象文件。
func (s *Store) Delete(id string, key string) error {
	pathKey := s.PathTransformFunc(key)
	filePath := filepath.Join(s.Root, id, pathKey.FullPath())
	if err := os.Remove(filePath); err != nil {
		return err
	}
	log.Printf("deleted [%s] from disk", pathKey.Filename)
	return nil
}

// Write 将 r 中的数据存到指定 id 和逻辑 key 对应的文件中。
func (s *Store) Write(id string, key string, r io.Reader) error {
	return s.writeStream(id, key, r)
}

// openFileForWriting 创建目标文件所需的父目录，并打开目标文件。
func (s *Store) openFileForWriting(id string, key string) (*os.File, error) {
	pathKey := s.PathTransformFunc(key)
	pathNameWithRoot := filepath.Join(s.Root, id, pathKey.PathName)
	if err := os.MkdirAll(pathNameWithRoot, os.ModePerm); err != nil {
		return nil, err
	}
	fullPathWithRoot := filepath.Join(s.Root, id, pathKey.FullPath())
	return os.Create(fullPathWithRoot)
}

// writeStream 将 r 复制到 id 和 key 对应的文件，并关闭文件句柄。
func (s *Store) writeStream(id string, key string, r io.Reader) error {
	f, err := s.openFileForWriting(id, key)
	if err != nil {
		return err
	}
	if _, err = io.Copy(f, r); err != nil {
		_ = f.Close()
		return err
	}
	return f.Close()
}

// Read 打开对象并返回文件大小和读取器。读取结束后，调用方需要关闭读取器
// （当前返回的是文件，因此读取器也实现了 io.Closer）。
func (s *Store) Read(id string, key string) (int64, io.Reader, error) {
	return s.readStream(id, key)
}

func (s *Store) readStream(id string, key string) (int64, io.Reader, error) {
	pathKey := s.PathTransformFunc(key)
	fullPathWithRoot := filepath.Join(s.Root, id, pathKey.FullPath())

	file, err := os.Open(fullPathWithRoot)
	if err != nil {
		return 0, nil, err
	}
	fi, err := file.Stat()
	if err != nil {
		_ = file.Close()
		return 0, nil, err
	}
	return fi.Size(), file, nil
}
