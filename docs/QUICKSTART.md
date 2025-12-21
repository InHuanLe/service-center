# 快速开始示例

演示如何使用 service-center 的多二进制架构。

## 场景一：本地开发和测试

### 步骤 1：构建所有组件

```bash
# 使用 Makefile（推荐）
make build

# 或使用 Go 命令
go build -o bin/server ./cmd/server
go build -o bin/client ./cmd/client
```

### 步骤 2：启动服务端

```bash
# Linux/macOS
./_output/bin/linux/amd64/server --address :50051

# Windows
_output\bin\windows\amd64\server.exe --address :50051
```

### 步骤 3：注册服务实例

在新终端中：

```bash
# 注册第一个服务实例
./_output/bin/linux/amd64/client register web-service 192.168.1.100 8080

# 注册第二个服务实例
./_output/bin/linux/amd64/client register web-service 192.168.1.101 8080

# 注册另一个服务
./_output/bin/linux/amd64/client register api-service 192.168.1.200 9090
```

### 步骤 4：发现服务

```bash
# 查找 web-service 的所有实例
./_output/bin/linux/amd64/client discover web-service

# 输出示例：
# Found 2 instance(s) for service web-service:
#   - ID: web-service-1734777600, Address: 192.168.1.100:8080
#   - ID: web-service-1734777605, Address: 192.168.1.101:8080
```

### 步骤 5：注销服务

```bash
# 使用实例 ID 注销
./_output/bin/linux/amd64/client unregister web-service web-service-1734777600
```

---

## 场景二：使用 Docker 容器

### 步骤 1：构建 Docker 镜像

```bash
# 构建所有镜像
docker build --target server -t service-center:server .
docker build --target client -t service-center:client .

# 或使用 Makefile
make docker-build
```

### 步骤 2：启动服务端容器

```bash
docker run -d \
  --name registry-server \
  -p 50051:50051 \
  service-center:server
```

### 步骤 3：使用客户端容器

```bash
# 注册服务
docker run --rm --network host \
  service-center:client \
  register my-app 192.168.1.50 8000 --server localhost:50051

# 发现服务
docker run --rm --network host \
  service-center:client \
  discover my-app --server localhost:50051

# 注销服务
docker run --rm --network host \
  service-center:client \
  unregister my-app <instance-id> --server localhost:50051
```

### 步骤 4：清理

```bash
docker stop registry-server
docker rm registry-server
```

---

## 场景三：使用 Docker Compose（完整环境）

### 步骤 1：启动环境

```bash
docker-compose up -d
```

这会启动：
- 1 个服务注册中心服务器
- 2 个示例服务实例（自动注册）

### 步骤 2：查看状态

```bash
# 查看运行的容器
docker-compose ps

# 查看服务端日志
docker-compose logs -f registry-server

# 查看示例服务日志
docker-compose logs example-service-1
docker-compose logs example-service-2
```

### 步骤 3：手动测试客户端

```bash
# 使用 Docker Compose 网络
docker run --rm --network service-center_service-net \
  service-center:client \
  discover example-service --server registry-server:50051
```

### 步骤 4：清理

```bash
docker-compose down
```

---

## 场景四：跨平台构建

### 构建 Linux 版本

```bash
make build-linux
# 输出在：_output/bin/linux/amd64/
```

### 构建 Windows 版本

```bash
make build-windows
# 输出在：_output/bin/windows/amd64/
```

### 构建 macOS 版本

```bash
make build-darwin
# 输出在：_output/bin/darwin/amd64/
```

### 构建所有平台

```bash
# 使用 Makefile
make build-all-platforms

# 或使用构建脚本
./build.sh --all-platforms

# 查看输出
tree _output/bin/
# _output/bin/
# ├── darwin/
# │   └── amd64/
# │       ├── server
# │       └── client
# ├── linux/
# │   └── amd64/
# │       ├── server
# │       └── client
# └── windows/
#     └── amd64/
#         ├── server.exe
#         └── client.exe
```

---

## 场景五：在 Kubernetes 中部署

### 步骤 1：推送镜像到仓库

```bash
# 标记镜像
docker tag service-center:server your-registry/service-center:server
docker tag service-center:client your-registry/service-center:client

# 推送镜像
docker push your-registry/service-center:server
docker push your-registry/service-center:client
```

### 步骤 2：创建部署清单

创建 `deploy/k8s/deployment.yaml`:

```yaml
apiVersion: v1
kind: Namespace
metadata:
  name: service-center

---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: registry-server
  namespace: service-center
spec:
  replicas: 3
  selector:
    matchLabels:
      app: registry-server
  template:
    metadata:
      labels:
        app: registry-server
    spec:
      containers:
      - name: server
        image: your-registry/service-center:server
        ports:
        - containerPort: 50051
          name: grpc
        resources:
          requests:
            memory: "64Mi"
            cpu: "100m"
          limits:
            memory: "128Mi"
            cpu: "200m"

---
apiVersion: v1
kind: Service
metadata:
  name: registry-service
  namespace: service-center
spec:
  selector:
    app: registry-server
  ports:
  - port: 50051
    targetPort: 50051
    name: grpc
  type: ClusterIP
```

### 步骤 3：部署

```bash
kubectl apply -f deploy/k8s/deployment.yaml

# 查看部署状态
kubectl get pods -n service-center
kubectl get svc -n service-center
```

### 步骤 4：测试连接

```bash
# 端口转发
kubectl port-forward -n service-center svc/registry-service 50051:50051

# 在另一个终端使用客户端
./bin/client register test-service 10.0.0.1 8080 --server localhost:50051
./bin/client discover test-service --server localhost:50051
```

---

## 场景六：开发新的二进制工具

假设我们要添加一个 `admin` 管理工具。

### 步骤 1：创建新的入口点

```bash
mkdir -p cmd/admin
```

创建 `cmd/admin/main.go`:

```go
package main

import (
    "context"
    "fmt"
    "time"
    "service-center/pkg/client/registry"
)

func main() {
    fmt.Println("Service Center Admin Tool v1.0")
    
    // 连接到服务注册中心
    client, err := registry.NewClient("localhost:50051")
    if err != nil {
        fmt.Printf("Failed to connect: %v\n", err)
        return
    }
    defer client.Close()
    
    // 列出所有服务
    // TODO: 实现管理功能
    fmt.Println("Admin operations...")
}
```

### 步骤 2：更新 Makefile

在 `Makefile` 中：

```makefile
BINARIES := server client admin
```

### 步骤 3：更新 Dockerfile

在 `Dockerfile` 中添加：

```dockerfile
# Build admin binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /bin/admin ./cmd/admin

# Admin image
FROM gcr.io/distroless/base-debian12 AS admin
COPY --from=builder /bin/admin /admin
ENTRYPOINT ["/admin"]
```

### 步骤 4：构建和测试

```bash
# 构建新二进制
make admin

# 运行
./_output/bin/linux/amd64/admin

# 构建 Docker 镜像
docker build --target admin -t service-center:admin .

# 运行容器
docker run --rm --network host service-center:admin
```

---

## 实用技巧

### 查看构建帮助

```bash
# Makefile 帮助
make help

# 构建脚本帮助
./build.sh --help
build.bat --help
```

### 快速开发循环

```bash
# 一次性完成：格式化、检查、测试、构建
make dev
```

### 清理构建产物

```bash
# 清理本地构建
make clean

# 清理 Docker 镜像
docker rmi service-center:server service-center:client service-center:all
```

### 查看版本信息

```bash
# 如果在构建时注入了版本信息
./_output/bin/linux/amd64/server --version
```

---

## 故障排查

### 问题：客户端无法连接到服务器

```bash
# 检查服务器是否运行
docker ps | grep registry-server

# 检查端口是否开放
netstat -an | grep 50051

# 使用 telnet 测试连接
telnet localhost 50051
```

### 问题：构建失败

```bash
# 清理并重新下载依赖
go clean -modcache
go mod download
go mod tidy

# 重新构建
make clean
make build
```

### 问题：Docker 镜像构建失败

```bash
# 清理 Docker 缓存
docker builder prune

# 重新构建（不使用缓存）
docker build --no-cache --target server -t service-center:server .
```

---

## 下一步

- 阅读 [多二进制构建详细文档](./docs/multi-binary-build.md)
- 查看 [API 文档](./api/proto/service.proto)
- 了解 [架构设计](./README.md)
- 贡献代码到项目

## 反馈

如有问题或建议，请提交 Issue 或 Pull Request。
