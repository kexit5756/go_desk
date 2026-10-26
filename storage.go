package main

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

// CasPathTransformFunc 将 key 的 SHA-1 十六进制字符串分成多级目录和文件名。
// 当前散列的是 key，不是文件内容；这只是键到路径的映射，尚不提供内容去重或完整性校验。
func CasPathTransformFunc(key string) PathKey {
	hash := sha1.Sum([]byte(key))
	hashstr := hex.EncodeToString(hash[:])

	// SHA-1 的十六进制表示有 40 个字符，每 5 个字符一层，共 8 层目录。
	blockSize := 5
	sliceLen := len(hashstr) / blockSize

	paths := make([]string, sliceLen)
	for i := 0; i < sliceLen; i++ {
		from, to := i*blockSize, (i*blockSize)+blockSize
		paths[i] = hashstr[from:to]
	}
	return PathKey{
		PathName: strings.Join(paths, "/"),
		FileName: hashstr,
	}
}

// PathTransformFunc 定义可替换的路径生成策略；读和写必须使用同一套规则。
type PathTransformFunc func(string) PathKey

// DefaultPathTransformFunc 当前保留早期直接返回 key 的实现。
// TODO: 返回 string 与 PathTransformFunc 要求的 PathKey 不一致，不能直接用作该配置。
// 当前测试显式选择 CasPathTransformFunc，没有使用这个默认函数。
var DefaultPathTransformFunc = func(key string) string {
	return key
}

// Storeops 保存存储配置，当前需要调用方显式提供非 nil 的路径转换函数。
// TODO: 后续统一命名为 StoreOpts，并设计默认策略或配置校验。
type Storeops struct {
	PathTransformFunc PathTransformFunc
}

// Store 按路径转换策略读写本地文件，尚未与 TCP/RPC 消费流程连接。
// 当前没有独立根目录配置，生成的相对路径以程序工作目录为起点。
type Store struct {
	Storeops
}

// NewStore 保存配置，不创建文件或目录。
// TODO: 当前不校验或补齐 PathTransformFunc，空配置会在后续读写时触发 panic。
func NewStore(opts Storeops) *Store {
	return &Store{
		Storeops: opts,
	}
}

// PathKey 分别保存目录部分与文件名，供读写两条路径共同使用。
type PathKey struct {
	PathName string
	FileName string
}

// FullPath 组合目录和文件名；后续可使用 filepath.Join 表达本地文件路径。
func (p PathKey) FullPath() string {
	return fmt.Sprintf("%s/%s", p.PathName, p.FileName)
}

// readStream 打开只读文件并把关闭责任交给调用方；文件不存在时返回错误。
func (s *Store) readStream(key string) (io.ReadCloser, error) {
	pathKey := s.PathTransformFunc(key)
	return os.Open(pathKey.FullPath())
}

// Read 把整个文件复制到内存，返回与文件句柄无关的 reader，因此可以在返回前关闭文件。
// 尽管返回 io.Reader，当前实现并非直接从磁盘流式读取，大文件会占用相应大小的内存。
func (s *Store) Read(key string) (io.Reader, error) {
	f, err := s.readStream(key)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	buf := new(bytes.Buffer)
	_, err = io.Copy(buf, f)
	return buf, err
}

// writeStream 按 key 创建目录，把 reader 中的数据复制到目标文件。
// 当前直接覆盖目标：若复制失败，已有内容可能已被清空或部分替换，尚无原子提交保证。
func (s *Store) writeStream(key string, r io.Reader) error {
	pathKey := s.PathTransformFunc(key)

	// MkdirAll 创建缺失的多级目录；目录已存在时仍可继续，权限受系统 umask 影响。
	if err := os.MkdirAll(pathKey.PathName, os.ModePerm); err != nil {
		return err
	}

	fullPath := pathKey.FullPath()

	// Create 以可写方式创建文件；已存在时立即截断，io.Copy 成功前旧内容就可能丢失。
	// TODO: 需要可靠覆盖时，先写同目录临时文件，确认写入和关闭成功，再替换目标。
	f, err := os.Create(fullPath)
	if err != nil {
		return err
	}
	// TODO: 当前只确保调用 Close，没有把关闭错误传递给调用方。
	defer f.Close()
	n, err := io.Copy(f, r)
	if err != nil {
		return err
	}
	// Printf 才解析 %d、%s；Print 会直接输出参数，go test 的静态检查也会提示误用。
	log.Printf("写入(%d)字节到文件%s", n, fullPath)
	return nil
}
