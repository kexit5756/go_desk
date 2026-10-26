# go_desk

Go 网络与文件存储学习项目：练习 TCP 监听、Peer、消息消费 channel，以及按 key 生成路径的本地文件读写。网络层与存储层目前分别学习，尚未完成网络文件传输。

## 第一天进度（2026-10-20）

- 初始化 Go 模块，配置构建、运行和测试命令。
- 定义 `Transport` 接口与 `Peer` 接口占位。
- 实现 `TcpTransport`、TCP 监听和后台接收连接循环。
- 实现 `TCPPeer`，保存连接对象与连接方向标记。
- 添加最小启动测试，并补充中文注释及工作日报。

第一天完成通信骨架，当时尚未实现消息读写、主动 Dial、节点登记和文件传输。

## 第二次进度（2026-10-22）

- 引入 `TCPTransportOpts`，配置监听地址、握手函数和解码器。
- 定义空握手函数，预留协议扩展入口。
- 实现默认字节读取，并初步接入 Gob 解码器。
- 新增 `Message`，记录消息内容和发送方地址。
- 使用 Mac 终端的 `nc` 验证文本接收，解码失败后退出连接处理循环。

第二次完成了使用 `DefaultDecoder` 读取原始字节的演示。详细记录见[第二次学习记录](docs/2026-10-22-work-log.md)，当前状态见下方第三次进度。

## 第三次进度（2026-10-26）

- 将消息封装重命名为 `RPC`，通过 `Consume()` 向业务层交付消息。
- 初始化无缓冲消息 channel，并在主程序启动消费 goroutine。
- 为 `Peer` 增加关闭方法，新增握手后的 `OnPeer` 回调及连接退出清理。
- 同步测试接口，补充代码批注和[第三次学习记录与运行提示](docs/2026-10-26-work-log.md)。

当前示例 `Onpeer` 会立即关闭新连接，接收演示前需要调整该回调。RPC 对象复用、Accept 错误处理和消息边界仍待修复；当前 `RPC` 是消息信封，尚不包含远程方法执行及响应协议。

## 第四次进度：本地存储（2026-10-26）

- 新增 `Store`、`PathKey` 和可替换的路径转换函数。
- 对 key 做 SHA-1，将 40 位十六进制字符串划分成 8 层目录，并以完整散列作为文件名。
- 使用 `io.Copy` 写入文件，通过相同 key 读取到内存 buffer。
- 增加路径转换和读写一致性测试，补充[存储模块 Review](docs/2026-10-26-storage-review.md)与中文注释。

当前散列的是 key，并非文件内容，不提供内容去重或完整性验证。Store 尚未接入 RPC；覆盖写入失败、默认配置与测试隔离仍待完善。

## 环境

- Go：本项目 go.mod 声明为 `1.27.2`。
- Make：可选；也可以直接使用 Go 命令。
- 测试依赖：Testify，具体版本见 go.mod/go.sum。

## 存储模块与验证

```text
key → 路径转换 → 多级目录 / 散列文件名
                   ├─ writeStream：reader → 文件
                   └─ Read：文件 → 内存 buffer → reader
```

当前使用方式是 `NewStore(Storeops{PathTransformFunc: CasPathTransformFunc})`，读写必须配置同一套路径策略。空配置会在读写时 panic；早期的 `DefaultPathTransformFunc` 返回类型尚未同步，暂不能用作该配置。

只运行存储相关测试，不启动 TCP 服务：

```bash
go test -v -run '^(TestPathTransformFunc|TestStore)$' .
```

TestStore 使用的是普通字节样例，不是真正的图片。当前测试会在运行目录留下散列路径，已从 Git 提交中排除；后续会使用独立根目录与临时测试目录。

## TCP 运行

在仓库根目录执行：

```bash
go mod download
make run
```

也可以运行：

```bash
go run .
```

程序监听 TCP `:3000`。当前 `Onpeer` 回调会立即关闭连接；以下发送演示需要先让回调保留连接，或不配置 `OnPeer`，详见[运行提示](docs/2026-10-26-work-log.md)。调整后可在另一个终端向默认解码器发送文本：

```bash
printf 'hello\n' | nc -w 1 127.0.0.1 3000
```

调整回调后，消息会通过 channel 交给主程序打印；当前尚未实现 echo 回复。使用 Control+C 停止运行。切换为 Gob 解码器后，需要使用 Gob 编码的客户端数据。

## 测试

```bash
make test
# 或者
go test -v ./...
```

根包包含 TestPathTransformFunc 和 TestStore；p2p 的 TestTCPTransport 已适配配置接口，但只尝试启动监听，没有覆盖 OnPeer 和消息收发。网络测试固定使用 `:3000`，若本机服务正在占用端口会失败，应先停止服务；后续改为随机端口并补充资源清理。当前验证结果及 Review 中的边界问题见[存储 Review 记录](docs/2026-10-26-storage-review.md)。

## 目录

```text
.
├── main.go
├── storage.go
├── store_test.go
├── makefile
├── go.mod
├── go.sum
├── p2p/
│   ├── transport.go
│   ├── encoding.go
│   ├── handkshake.go
│   ├── message.go
│   ├── tcp_transport.go
│   └── tcp_transport_test.go
└── docs/
    ├── 2026-10-20-work-log.md
    ├── 2026-10-22-work-log.md
    ├── 2026-10-26-work-log.md
    └── 2026-10-26-storage-review.md
```

学习记录：

- [第一次：TCP/P2P 通信骨架（2026-10-20）](docs/2026-10-20-work-log.md)
- [第二次：TCP 消息接收与解码（2026-10-22）](docs/2026-10-22-work-log.md)
- [第三次：RPC 消费通道、批注与运行提示（2026-10-26）](docs/2026-10-26-work-log.md)
- [第四次：本地文件存储与代码 Review（2026-10-26）](docs/2026-10-26-storage-review.md)

## 下一步

- 完善存储配置校验和默认路径函数，隔离测试数据。
- 改善失败覆盖的写入流程，避免输入流出错时破坏旧文件。
- 修复存储测试的错误处理，再把 Store 接入消息消费流程。

- 修正 Accept 接入连接的 outbound 标记；当前源码已标注 TODO。
- 统一 `NewTCCPer` 构造函数命名。
- 调整 OnPeer 回调，修复 RPC 对象复用。
- 改进测试断言、随机端口及监听器关闭流程。
- 完善 Accept 错误处理及 Gob 连续解码。
- 增加最小 echo 功能和真实连接测试。
- 后续实现节点管理、消息边界与文件传输。

## 学习参考

本项目为课程跟练及个人学习记录，参考：

- [anthdm/distributedfilesystemgo](https://github.com/anthdm/distributedfilesystemgo)
- [配套 B 站视频](https://www.bilibili.com/video/BV18b421H78a/)
