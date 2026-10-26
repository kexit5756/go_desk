# 存储模块 Review 与学习记录

- 日期：2026-10-26
- 范围：新增 `storage.go`、`store_test.go`，以及已有 TCP/RPC 模块的集成状态。
- 本次目标：检查实现、补充中文注释、更新 README，并发布源码与文档。

## 本次新增能力

存储模块可以根据 key 生成分层目录，把 `io.Reader` 中的字节写入本地文件，再通过相同 key 读取。现有测试覆盖固定 key 的路径转换和小段数据的读写一致性。

该模块目前独立于网络模块：main 没有创建 Store，也没有把 Consume 收到的数据交给 writeStream。因此，“存储读写测试通过”不等于已经实现网络文件上传或分布式文件系统。

## 每个组件的职责

| 组件 | 当前职责 |
|---|---|
| `CasPathTransformFunc` | 对 key 做 SHA-1，生成散列目录和文件名 |
| `PathTransformFunc` | 约定路径策略为 `func(string) PathKey` |
| `PathKey` / `FullPath` | 保存目录与文件名，组合最终路径 |
| `Storeops` / `NewStore` | 保存调用方显式传入的路径转换策略 |
| `writeStream` | 创建目录、创建或截断文件，再使用 io.Copy 写入 |
| `readStream` | 打开只读文件，将关闭责任交给调用方 |
| `Read` | 将文件全部读入 bytes.Buffer，关闭文件后返回内存 reader |

## 路径转换怎样工作

```text
key
  → SHA-1：20 字节
  → 十六进制：40 个字符
  → 每 5 个字符一层：8 层目录
  → 完整散列作为文件名
```

读和写都使用同一个 `PathTransformFunc`，因此能通过相同 key 找回文件。生成的是相对路径，当前写入位置取决于程序工作目录。

需要区分：当前散列的是 key，不是文件内容。函数名虽然含 Cas，但当前实现不提供按内容去重，也没有在读取时校验文件内容。

## Review 发现的问题

### 1. 高：覆盖失败可能破坏已有文件

`os.Create` 在 `io.Copy` 前立即截断目标文件。若输入流中途出错，函数虽然返回错误，旧文件也可能已经清空或变成部分内容。

临时诊断用例先写入 `previous valid data`，再使用会返回 `io.ErrUnexpectedEOF` 的 reader 覆盖同一 key。结果是写入返回错误，而旧文件变为空文件，已复现数据丢失。

建议：先写同目录临时文件，检查写入和关闭结果，成功后再替换目标；失败时保留原文件并清理临时文件。并发写同一 key 的行为也需要单独约定。

### 2. 中：空配置没有默认值或校验

`NewStore(Storeops{})` 可以构造对象，但读写时会调用 nil 的 `PathTransformFunc` 并 panic，临时诊断用例已复现。

当前正确用法是显式配置：

```go
store := NewStore(Storeops{
    PathTransformFunc: CasPathTransformFunc,
})
```

后续应选择“补齐默认策略”或“拒绝无效配置”，让错误更早暴露。

### 3. 中：默认路径函数的签名未同步

`DefaultPathTransformFunc` 仍是 `func(string) string`，而配置要求 `func(string) PathKey`，两者不兼容。临时类型检查已确认不能把它直接赋给该配置。

目前使用 CasPathTransformFunc 的测试不受影响，但默认函数无法按名字所暗示的用途接入。后续需要统一签名，并定义直通 key 时的目录与文件名规则。

### 4. 中：测试失败后的控制流程与错误检查不足

TestStore 写入或读取失败后使用 `t.Error`，仍会继续执行后续步骤；读取失败时 r 可能为 nil，再调用 ReadAll 可能引发新的 panic。ReadAll 的错误也被忽略。

建议前置步骤失败时使用 Fatal/Fatalf，并检查每个 I/O 错误。当前测试直接写工作目录，没有使用临时根目录，所以会留下散列目录；已通过 .gitignore 排除当前样例及 5 位十六进制首级目录，避免上传运行产物。

### 5. 其他当前限制

- Read 会把整个文件放进内存，并不是从磁盘直接向调用方流式交付。
- 写入只 defer Close，没有把关闭失败返回给调用方。
- 存储尚未接入 RPC 消费流程，也没有上传、响应、删除和复制协议。
- 已有网络问题仍需继续修复：Onpeer 立即关连接、Accept 错误后仍处理 conn、复用同一 RPC 指针，以及缺少消息边界。详见[第三次学习记录](2026-10-26-work-log.md)。

以上发现通过注释和文档记录；本次保留当前学习代码的执行逻辑，没有把这些建议描述为已修复能力。

## 验证记录

- 在临时项目副本中执行 `go test -v -timeout 30s ./...`：TestPathTransformFunc、TestStore 和 TestTCPTransport 全部通过。
- 执行 `go build -buildvcs=false ./...`：通过。
- 对新增源码和测试执行 gofmt，并使用 Go scanner 对比 9 个 Go 文件的 token：忽略注释后完全一致，确认本次批注没有改变现有执行逻辑。
- 额外诊断分别复现空配置 panic、默认路径函数类型不兼容、失败覆盖清空旧内容三项问题。这些边界问题尚未修复，不能由现有正常路径测试通过推断它们安全。
- 测试与诊断均在临时目录执行，本次检查没有向项目工作目录追加存储数据。

临时诊断用例只用于本次 Review，不加入项目日常测试；它们的失败反映尚未满足的边界条件，不代表已有的两个存储测试失败。

## 运行与学习提示

仅运行存储测试：

```bash
go test -v -run '^(TestPathTransformFunc|TestStore)$' .
```

这个命令不启动 TCP 监听，但当前 TestStore 会在工作目录生成存储文件。运行全部测试时，p2p 的启动测试会尝试占用 3000 端口，需要先确认端口未被运行中的 fs 服务占用。

格式化字符串需要使用带 f 的方法：

```go
log.Printf("写入(%d)字节到文件%s", n, fullPath)
t.Errorf("want %s have %s", data, b)
```

Print 和 Error 不负责解析 `%d`、`%s`；日志与测试的格式使用错误会被 go test 的静态检查发现。

## 下一步建议

- [ ] 设计存储根目录，并让测试使用 t.TempDir，隔离运行产物。
- [ ] 统一默认路径函数签名，补齐配置校验。
- [ ] 先实现失败不破坏旧内容的写入流程，再增加失败注入测试。
- [ ] 改善测试的错误处理，补缺失文件、空数据、覆盖写入等场景。
- [ ] 完善网络层，再通过明确的协议把 RPC 负载与 Store 连接起来。
