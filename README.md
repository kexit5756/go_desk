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

第二次完成了使用 `DefaultDecoder` 读取原始字节的演示。详细记录见[第二次学习记录](docs/2026-10-22-work-log.md)，当前状态见下方第三次进度。

## 第三次进度（2026-10-26）

- 将消息封装重命名为 `RPC`，通过 `Consume()` 向业务层交付消息。
- 初始化无缓冲消息 channel，并在主程序启动消费 goroutine。
- 为 `Peer` 增加关闭方法，新增握手后的 `OnPeer` 回调及连接退出清理。
- 同步测试接口，补充代码批注和[第三次学习记录与运行提示](docs/2026-10-26-work-log.md)。

当前示例 `Onpeer` 会立即关闭新连接，接收演示前需要调整该回调。RPC 对象复用、Accept 错误处理和消息边界仍待修复；当前 `RPC` 是消息信封，尚不包含远程方法执行及响应协议。

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

测试已改用 `TCPTransportOpts`，循环导入已移除；当前只尝试启动监听，没有覆盖 OnPeer 和消息收发。测试固定使用 `:3000`，若本机服务正在占用端口会失败，应先停止服务；后续改为随机端口并补充资源清理。根目录显示 `[no test files]` 是正常提示。

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
    ├── 2026-10-22-work-log.md
    └── 2026-10-26-work-log.md
```

学习记录：

- [第一次：TCP/P2P 通信骨架（2026-10-20）](docs/2026-10-20-work-log.md)
- [第二次：TCP 消息接收与解码（2026-10-22）](docs/2026-10-22-work-log.md)
- [第三次：RPC 消费通道、批注与运行提示（2026-10-26）](docs/2026-10-26-work-log.md)

## 下一步

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
