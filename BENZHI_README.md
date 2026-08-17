# acoustic-calibration__004 Docker 交付说明

## 项目概览
- Acoustic Calibration Record Service stores calibration instruments and evaluates
- Go module: `example.com/acoustic-calibration`

## 标准命令

```bash
go build ./...
go test ./...
```

## 实际启动入口

```bash
go run ./cmd/calibrated
```

## Docker 构建

```bash
./build_benzhi_docker.sh acoustic-calibration__004-benzhi linux/amd64
docker run --rm -it acoustic-calibration__004-benzhi bash
```

## 环境

- 基础镜像: `golang:1.22`
- 依赖在镜像构建阶段预下载，容器内可直接执行 Go 构建和测试命令。
- 代码目录: `/app`
- 源码中检测到的服务端口: `8080`
