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
--- FAIL: TestRunSamplesAreIsolated
    stored readings were changed through a returned slice: [999 45]
FAIL    example.com/acoustic-calibration/internal/service
FAIL
```

## 期望行为

提交和查询校准读数时，调用方获得的数据副本应相互隔离；修改一次查询结果不应改变已保存的历史记录。
