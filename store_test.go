package main

import (
	"bytes"
	"io/ioutil"
	"testing"
)

// TestPathTransformFunc 使用固定样例，验证 key 的散列文件名和 8 层目录是否符合约定。
// key 的拼写是测试输入的一部分，修改它时需重新核对预期散列。
func TestPathTransformFunc(t *testing.T) {
	key := "momsbestpircture"
	pathKey := CasPathTransformFunc(key)
	expectedOriginalKey := "dfaaf66f10a4678e27b5ed974c0bf67fd31b6a44"
	expectedPathName := "dfaaf/66f10/a4678/e27b5/ed974/c0bf6/7fd31/b6a44"
	if pathKey.PathName != expectedPathName {
		t.Errorf("path transform func error: expected %s, got %s", expectedPathName, pathKey.PathName)
	}
	if pathKey.FileName != expectedOriginalKey {
		t.Errorf("path transform func error: expected %s, got %s", expectedOriginalKey, pathKey.FileName)
	}
}

// TestStore 验证小段数据写入后，能通过相同 key 读回相同内容。
// 当前只覆盖正常路径，不覆盖写入失败、空配置或大文件。
func TestStore(t *testing.T) {
	opts := Storeops{
		PathTransformFunc: CasPathTransformFunc,
	}
	s := NewStore(opts)

	// 该字符串是测试负载，并不是真正的 JPG 图片。
	// TODO: 引入临时存储根目录，避免测试在项目工作目录留下散列目录。
	keys := "momsspecials"
	data := []byte("some jpg bytes")
	if err := s.writeStream(keys, bytes.NewReader(data)); err != nil {
		// TODO: 写入失败时应 Fatal，避免继续执行依赖写入成功的步骤。
		t.Error(err)
	}
	r, err := s.Read(keys)
	if err != nil {
		// TODO: 此处应 Fatal；继续把可能为 nil 的 r 交给 ReadAll 会掩盖原始错误。
		t.Error(err)
	}
	// TODO: 检查 ReadAll 返回的错误，不能只比较可能不完整的内容。
	b, _ := ioutil.ReadAll(r)
	if string(b) != string(data) {
		// Errorf 会格式化占位符并记录失败，但不会立即终止当前测试。
		t.Errorf("want %s have  %s", data, b)
	}
}
