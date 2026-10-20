# 构建程序；-o 指定输出位置，@ 表示执行时不回显该命令。
build:
	@go build -o bin/fs

# run 依赖 build，先编译再运行。运行中的服务可用 Control+C 终止。
run: build
	@./bin/fs

# 运行当前模块中的全部包测试，-v 输出每个测试的详细结果。
test:
	@go test ./... -v
