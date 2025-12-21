# 多二进制构建指南

本项目采用 Kubernetes 风格的单仓库多二进制架构，支持构建多个独立的可执行文件和 Docker 镜像。

## 架构设计

### 目录结构

```
service-center/
├── cmd/
│   ├── server/        # 服务端入口
│   │   └── main.go
│   └── client/        # 客户端 CLI 入口
│       └── main.go
├── pkg/               # 共享库
├── Dockerfile         # 多阶段构建
├── Makefile          # 构建自动化
└── build.sh/bat      # 跨平台构建脚本
```

### 设计原则

1. **单一代码仓库** - 所有组件共享同一个代码库
2. **独立二进制** - 每个组件编译为独立的可执行文件
3. **共享库** - 通用代码放在 `pkg/` 目录下复用
4. **多目标构建** - 支持构建特定组件或全部组件

## 构建方式

### 1. 使用 Makefile（推荐）

```bash
# 查看所有命令
make help

# 构建所有二进制
make build

# 构建特定二进制
make server
make client

# 构建 Linux 版本
make build-linux

# 构建 Windows 版本
make build-windows

# 构建 macOS 版本
make build-darwin

# 构建所有平台
make build-all-platforms

# Docker 构建
make docker-build                # 所有镜像
make docker-build-server        # 仅服务端
make docker-build-client        # 仅客户端
```

### 2. 使用 Docker 多阶段构建

```bash
# 构建服务端镜像
docker build --target server -t service-center:server .

# 构建客户端镜像
docker build --target client -t service-center:client .

# 构建包含所有二进制的镜像
docker build --target all -t service-center:all .
```

**镜像大小对比：**
- `server`: ~36.6MB（仅包含服务端二进制）
- `client`: ~36.7MB（仅包含客户端二进制）
- `all`: ~52.4MB（包含两个二进制）

### 3. 使用构建脚本

**Linux/macOS:**
```bash
# 构建当前平台的所有二进制
./build.sh

# 构建特定二进制
./build.sh server
./build.sh client

# 构建所有平台
./build.sh --all-platforms

# 指定版本号
./build.sh --version v1.0.0

# 构建特定平台
./build.sh --platform linux/amd64 server
```

**Windows:**
```cmd
REM 构建所有二进制
build.bat

REM 构建特定二进制
build.bat server
build.bat client

REM 指定版本
build.bat --version v1.0.0
```

### 4. 直接使用 Go

```bash
# 构建服务端
go build -o bin/server ./cmd/server

# 构建客户端
go build -o bin/client ./cmd/client

# 带版本信息构建
go build -ldflags "-X main.Version=v1.0.0" -o bin/server ./cmd/server
```

## 运行示例

### 本地运行

```bash
# 1. 启动服务端
_output/bin/linux/amd64/server --address :50051

# 2. 使用客户端注册服务
_output/bin/linux/amd64/client register my-service 192.168.1.100 8080 --server localhost:50051

# 3. 发现服务
_output/bin/linux/amd64/client discover my-service --server localhost:50051

# 4. 注销服务
_output/bin/linux/amd64/client unregister my-service <instance-id> --server localhost:50051
```

### Docker 运行

```bash
# 启动服务端
docker run -d --name registry-server -p 50051:50051 service-center:server

# 使用客户端（一次性命令）
docker run --rm --network host service-center:client \
  register my-service 192.168.1.100:8080 --server localhost:50051

docker run --rm --network host service-center:client \
  discover my-service --server localhost:50051
```

### Docker Compose 运行

```bash
# 启动完整环境
docker-compose up -d

# 查看服务状态
docker-compose ps

# 查看日志
docker-compose logs -f registry-server

# 停止服务
docker-compose down
```

## Kubernetes 部署示例

创建 `deploy/kubernetes/server.yaml`:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: service-center-server
spec:
  replicas: 3
  selector:
    matchLabels:
      app: service-center-server
  template:
    metadata:
      labels:
        app: service-center-server
    spec:
      containers:
      - name: server
        image: service-center:server
        ports:
        - containerPort: 50051
          name: grpc
---
apiVersion: v1
kind: Service
metadata:
  name: service-center
spec:
  selector:
    app: service-center-server
  ports:
  - port: 50051
    targetPort: 50051
    name: grpc
  type: LoadBalancer
```

部署：
```bash
kubectl apply -f deploy/kubernetes/server.yaml
kubectl get pods -l app=service-center-server
```

## 添加新的可执行文件

如果需要添加新的二进制（例如 `admin` 管理工具）：

### 1. 创建入口文件

```bash
mkdir -p cmd/admin
```

创建 `cmd/admin/main.go`:
```go
package main

import (
    "fmt"
    "service-center/pkg/api/pb"
)

func main() {
    fmt.Println("Service Center Admin Tool")
    // 实现管理功能...
}
```

### 2. 更新 Dockerfile

在 `Dockerfile` 中添加构建步骤：

```dockerfile
# 在 builder 阶段添加
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /bin/admin ./cmd/admin

# 添加新的镜像目标
FROM gcr.io/distroless/base-debian12 AS admin
COPY --from=builder /bin/admin /admin
ENTRYPOINT ["/admin"]
```

### 3. 更新 Makefile

在 `Makefile` 中的 `BINARIES` 变量添加：

```makefile
BINARIES := server client admin
```

### 4. 更新构建脚本

在 `build.sh` 和 `build.bat` 中的 `BINARIES` 数组添加 `admin`。

### 5. 构建新二进制

```bash
# 使用 Makefile
make admin

# 使用 Docker
docker build --target admin -t service-center:admin .

# 使用构建脚本
./build.sh admin
```

## CI/CD 集成

### GitHub Actions 示例

创建 `.github/workflows/build.yml`:

```yaml
name: Build Multi-Binary

on:
  push:
    branches: [ main ]
  pull_request:
    branches: [ main ]

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v3
    
    - name: Set up Go
      uses: actions/setup-go@v4
      with:
        go-version: '1.25'
    
    - name: Build All Binaries
      run: make build-all-platforms
    
    - name: Run Tests
      run: make test
    
    - name: Build Docker Images
      run: make docker-build
    
    - name: Push to Registry
      if: github.ref == 'refs/heads/main'
      run: |
        echo "${{ secrets.DOCKER_PASSWORD }}" | docker login -u "${{ secrets.DOCKER_USERNAME }}" --password-stdin
        make docker-push
```

## 最佳实践

### 1. 代码组织

- `cmd/` - 每个可执行文件一个子目录
- `pkg/` - 可复用的库代码
- `internal/` - 私有库代码（不被外部导入）
- `api/` - API 定义（protobuf、OpenAPI 等）

### 2. 版本管理

使用 ldflags 在编译时注入版本信息：

```bash
VERSION=$(git describe --tags --always)
BUILD_DATE=$(date -u +'%Y-%m-%dT%H:%M:%SZ')
GIT_COMMIT=$(git rev-parse --short HEAD)

go build -ldflags "\
  -X main.Version=${VERSION} \
  -X main.BuildDate=${BUILD_DATE} \
  -X main.GitCommit=${GIT_COMMIT}" \
  -o bin/server ./cmd/server
```

在代码中使用：

```go
package main

var (
    Version   = "dev"
    BuildDate = "unknown"
    GitCommit = "unknown"
)

func main() {
    fmt.Printf("Version: %s\nBuild Date: %s\nGit Commit: %s\n",
        Version, BuildDate, GitCommit)
}
```

### 3. 依赖管理

所有二进制共享同一个 `go.mod`，确保依赖版本一致：

```bash
# 添加依赖
go get github.com/some/package@v1.2.3

# 清理未使用的依赖
go mod tidy
```

### 4. 测试策略

```bash
# 测试所有包
make test

# 测试特定包
go test ./pkg/service/registry/...

# 带覆盖率测试
make test-coverage
```

## 常见问题

### Q: 如何减小镜像大小？

使用多阶段构建和 distroless 基础镜像：
- 仅复制需要的二进制文件
- 使用 `CGO_ENABLED=0` 构建静态链接二进制
- 使用 `-ldflags "-w -s"` 去除调试信息

### Q: 如何支持更多平台？

在 Makefile 或构建脚本中添加平台：

```makefile
platforms=(
    "linux/amd64"
    "linux/arm64"
    "darwin/amd64"
    "darwin/arm64"
    "windows/amd64"
)
```

### Q: 如何处理共享配置？

将配置放在 `pkg/config` 中，所有二进制共享：

```go
package config

type Config struct {
    ServerAddr string
    Timeout    time.Duration
}

func Load() (*Config, error) {
    // 从环境变量或配置文件加载
}
```

## 参考资源

- [Kubernetes Build System](https://github.com/kubernetes/kubernetes/tree/master/build)
- [Go Multi-Binary Projects](https://github.com/golang-standards/project-layout)
- [Docker Multi-Stage Builds](https://docs.docker.com/build/building/multi-stage/)
