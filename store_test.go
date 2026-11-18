package main

import (
	"bytes"
	"fmt"
	"io"
	"testing"
)

// TestPathTransformFunc 验证固定 key 会转换为预期的 SHA-1 摘要和目录结构。
func TestPathTransformFunc(t *testing.T) {
	key := "momsbestpicture"
	pathKey := CASPathTransformFunc(key)
	expectedFilename := "6804429f74181a63c50c3d81d733a12f14a353ff"
	expectedPathName := "68044/29f74/181a6/3c50c/3d81d/733a1/2f14a/353ff"
	if pathKey.PathName != expectedPathName {
		t.Errorf("have %s want %s", pathKey.PathName, expectedPathName)
	}

	if pathKey.Filename != expectedFilename {
		t.Errorf("have %s want %s", pathKey.Filename, expectedFilename)
	}
}

// TestStore 验证按 ID 写入的对象可以被查找到并读取回来。
func TestStore(t *testing.T) {
	s := newStore(t)
	id := "test"

	for i := 0; i < 50; i++ {
		key := fmt.Sprintf("foo_%d", i)
		data := []byte("some jpg bytes")

		if err := s.writeStream(id, key, bytes.NewReader(data)); err != nil {
			t.Fatalf("write %q: %v", key, err)
		}

		if ok := s.Has(id, key); !ok {
			t.Errorf("expected to have key %s", key)
		}

		_, r, err := s.Read(id, key)
		if err != nil {
			t.Fatalf("read %q: %v", key, err)
		}

		b, err := io.ReadAll(r)
		if closer, ok := r.(io.Closer); ok {
			if closeErr := closer.Close(); closeErr != nil {
				t.Fatalf("close %q: %v", key, closeErr)
			}
		}
		if err != nil {
			t.Fatalf("read contents for %q: %v", key, err)
		}
		if string(b) != string(data) {
			t.Errorf("want %s have %s", data, b)
		}
	}
}

// newStore 为测试创建独立的临时根目录，测试结束后由 Go 自动清理。
func newStore(t *testing.T) *Store {
	t.Helper()
	opts := StoreOpts{
		Root:              t.TempDir(),
		PathTransformFunc: CASPathTransformFunc,
	}
	return NewStore(opts)
}
