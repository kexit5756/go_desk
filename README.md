# go_desk

Go 网络编程学习项目：从 TCP 监听、连接接收和 Peer 封装开始，逐步学习 P2P 通信与文件存储。

## 第一天进度（2026-10-20）

- 初始化 Go 模块，配置构建、运行和测试命令。
- 定义 `Transport` 接口与 `Peer` 接口占位。
- 实现 `TcpTransport`、TCP 监听和后台接收连接循环。
- 实现 `TCPPeer`，保存连接对象与连接方向标记。
- 添加最小启动测试，并补充中文注释及工作日报。

当前阶段是通信骨架。消息读写、主动 Dial、节点登记和文件传输还未实现。

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

程序监听 TCP `:3000`。在另一个终端连接：

```bash
nc localhost 3000
```

当前服务端只打印新连接信息，没有实现 echo 响应。使用 Control+C 停止运行。

## 测试

```bash
make test
# 或者
go test -v ./...
```

当前测试只检查地址配置和监听启动，使用固定端口 `:4000`；运行时需要该端口空闲。根目录显示 `[no test files]` 是正常提示。

## 目录

```text
.
├── main.go
├── makefile
├── go.mod
├── go.sum
├── p2p/
│   ├── transport.go
│   ├── tcp_transport.go
│   └── tcp_transport_test.go
└── docs/
    └── 2026-10-20-work-log.md
```

日报：[2026-10-20 工作总结](docs/2026-10-20-work-log.md)。

## 下一步

- 修正 Accept 接入连接的 outbound 标记；当前源码已标注 TODO。
- 统一 `NewTCCPer` 构造函数命名。
- 完善连接及测试监听器的关闭流程。
- 增加最小 echo 功能和真实连接测试。
- 后续实现节点管理、消息边界与文件传输。

## 学习参考

本项目为课程跟练及个人学习记录，参考：

- [anthdm/distributedfilesystemgo](https://github.com/anthdm/distributedfilesystemgo)
- [配套 B 站视频](https://www.bilibili.com/video/BV18b421H78a/)
