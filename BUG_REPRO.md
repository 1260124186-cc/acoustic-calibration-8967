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

在容器内执行：

```bash
go test ./... -count=20
```

## 实际错误输出

```text
panic: assignment to entry in nil map [recovered]
    panic: assignment to entry in nil map
FAIL    example.com/acoustic-calibration/internal/service
FAIL    example.com/acoustic-calibration/internal/store
FAIL
```

## 期望行为

省略限值配置的仪器应使用默认读数范围，后续更新读数范围时不应崩溃，并且仪器配置不应被调用方的可变数据影响。
