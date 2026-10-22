# go_desk

Go 网络编程学习项目：从 TCP 监听、连接接收和 Peer 封装开始，逐步学习 P2P 通信与文件存储。

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

当前主程序使用 `DefaultDecoder` 接收原始字节并打印消息；消息边界、连接清理及自动化测试仍需完善。详细记录见[第二次学习记录](docs/2026-10-22-work-log.md)。

## 环境

- Go：本项目 go.mod 声明为 `1.27.2`。
- Make：可选；也可以直接使用 Go 命令。
- 测试依赖：Testify，具体版本见 go.mod/go.sum。

## 运行

在仓库根目录执行：

```bash
go mod download
make run
```

也可以运行：

```bash
go run .
```

程序监听 TCP `:3000`，默认接收原始字节。在另一个终端发送文本：

```bash
printf 'hello\n' | nc -w 1 127.0.0.1 3000
```

服务端会打印接收内容及来源地址，尚未实现 echo 回复。使用 Control+C 停止运行。切换为 Gob 解码器后，需要使用 Gob 编码的客户端数据。

## 测试

```bash
make test
# 或者
go test -v ./...
```

当前旧测试尚未同步新的 `TCPTransportOpts` 构造参数及 `ListenAddr` 字段，`go test ./...` 会编译失败，待下一步修复。该测试原本只检查地址配置和监听启动，使用固定端口 `:4000`。根目录显示 `[no test files]` 是正常提示。

## 目录

```text
.
├── main.go
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
    └── 2026-10-22-work-log.md
```

学习记录：

- [第一次：TCP/P2P 通信骨架（2026-10-20）](docs/2026-10-20-work-log.md)
- [第二次：TCP 消息接收与解码（2026-10-22）](docs/2026-10-22-work-log.md)

## 下一步

- 修正 Accept 接入连接的 outbound 标记；当前源码已标注 TODO。
- 统一 `NewTCCPer` 构造函数命名。
- 同步旧测试的接口，完善连接及测试监听器的关闭流程。
- 完善 Accept 错误处理及 Gob 连续解码。
- 增加最小 echo 功能和真实连接测试。
- 后续实现节点管理、消息边界与文件传输。

## 学习参考

本项目为课程跟练及个人学习记录，参考：

- [anthdm/distributedfilesystemgo](https://github.com/anthdm/distributedfilesystemgo)
- [配套 B 站视频](https://www.bilibili.com/video/BV18b421H78a/)
