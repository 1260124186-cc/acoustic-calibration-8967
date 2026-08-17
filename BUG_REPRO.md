# 修复前故障复现（Docker）

## 项目与标准命令

项目为声学校准记录服务，当前平台为 Linux arm64。构建并进入评测镜像：

```bash
docker build -f benzhi.Dockerfile -t acoustic-calibration-eval .
docker run --rm acoustic-calibration-eval sh -c 'go build ./...'
```

## 环境构建与编译

镜像使用 `golang:1.22`，容器内 `go version` 输出 `go version go1.22.12 linux/arm64`。镜像构建和容器内 `go build ./...` 均成功。

## 故障触发步骤

在容器内重复运行并发批量校准验收：

```bash
go test ./... -race -count=20
```

## 实际错误输出

```text
WARNING: DATA RACE
fatal error: concurrent map read and map write
FAIL    example.com/acoustic-calibration/internal/service
FAIL
```

## 期望行为

批量提交的全部校准读数应被完整保存，每条记录都有唯一 ID，重复并发验收不应报告数据竞争或发生运行时崩溃。
