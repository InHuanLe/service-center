# service-center

一个基于 gRPC 的简单服务注册和发现中心实现。

## 架构概览

本项目采用类似 Kubernetes 的单仓库多二进制（monorepo multi-binary）架构：

- **`cmd/server`** - 服务注册中心服务端
- **`cmd/client`** - 命令行客户端工具
- **`pkg/`** - 共享库代码
- **`api/proto`** - gRPC 协议定义

## 快速开始

### 构建方式

#### 方式一：使用 Makefile（推荐）

```bash
# 查看所有可用命令
make help

# 构建所有二进制文件
make build

# 只构建服务端
make server

# 只构建客户端
make client

# 构建 Docker 镜像
make docker-build

# 构建特定平台的二进制
make build-linux
make build-windows
make build-darwin
```

#### 方式二：使用构建脚本

**Linux/macOS:**
```bash
# 构建所有
./build.sh

# 构建特定二进制
./build.sh server
./build.sh client

# 构建所有平台
./build.sh --all-platforms

# 指定版本
./build.sh --version v1.0.0
```

**Windows:**
```cmd
REM 构建所有
build.bat

REM 构建特定二进制
build.bat server
build.bat client

REM 指定版本
build.bat --version v1.0.0
```

#### 方式三：直接使用 Go

```bash
# 构建服务端
go build -o bin/server ./cmd/server

# 构建客户端
go build -o bin/client ./cmd/client
```

### Docker 构建

本项目 Dockerfile 支持多阶段构建，可以生成不同的镜像：

```bash
# 构建服务端镜像
docker build --target server -t service-center:server .

# 构建客户端镜像
docker build --target client -t service-center:client .

# 构建包含所有二进制的镜像
docker build --target all -t service-center:all .
```

### 运行服务

#### 本地运行

```bash
# 运行服务端
make run-server
# 或
./_output/bin/linux/amd64/server --address :50051

# 使用客户端
./_output/bin/linux/amd64/client register my-service localhost:8080
./_output/bin/linux/amd64/client discover my-service
./_output/bin/linux/amd64/client unregister <instance-id>
```

#### Docker Compose

```bash
# 启动服务注册中心和示例服务
docker-compose up -d

# 查看日志
docker-compose logs -f

# 停止服务
docker-compose down
```

#### Kubernetes 部署

```bash
# 应用配置（需要先创建 k8s manifests）
kubectl apply -f deploy/kubernetes/

# 查看服务状态
kubectl get pods -l app=service-center
```

## 项目结构

```
service-center/
├── cmd/                        # 可执行文件入口
│   ├── server/                 # 服务端主程序
│   │   └── main.go
│   └── client/                 # CLI 客户端
│       └── main.go
├── pkg/                        # 共享库
│   ├── api/pb/                 # 生成的 gRPC 代码
│   ├── service/registry/       # 服务注册逻辑
│   └── client/registry/        # 客户端库
├── api/proto/                  # Protobuf 定义
├── _output/                    # 构建输出（git ignored）
│   └── bin/
│       ├── linux/
│       ├── darwin/
│       └── windows/
├── Dockerfile                  # 多阶段 Docker 构建
├── docker-compose.yml          # 本地测试编排
├── Makefile                    # 构建自动化
├── build.sh                    # Linux/macOS 构建脚本
└── build.bat                   # Windows 构建脚本
```

## 客户端使用示例

```bash
# 注册服务实例
client register my-service 192.168.1.100:8080 --server localhost:50051

# 发现服务
client discover my-service --server localhost:50051

# 注销服务实例
client unregister <instance-id> --server localhost:50051

# 使用自定义超时
client discover my-service --server localhost:50051 --timeout 5s
```

## 当前实现分析

### 现有功能
- 基于内存存储（InMemoryStore）的服务注册
- 服务实例的存储和查询
- gRPC 协议支持

### 存在的问题

#### 1. **数据持久化缺失**
   - 仅使用内存存储，服务器重启后所有注册信息丢失
   - 需要添加数据库支持（如 SQLite、PostgreSQL）或本地文件持久化

#### 2. **心跳检测不完整**
   - `Available()` 方法有 TTL 超时检测逻辑，但存储层未实现心跳更新机制
   - 缺少定期清理过期实例的后台任务
   - 没有心跳失败时的自动反注册

#### 3. **并发控制问题**
   - `sync.Map` 的使用虽然提供了并发安全，但嵌套的 `map[string]struct{}` 操作仍存在竞态条件
   - 删除和读取操作间可能出现数据不一致

#### 4. **缺少错误处理**
   - 网络异常、客户端宕机等情况处理不足
   - 缺少客户端重试机制

#### 5. **功能不完整**
   - 没有服务实例的权重管理
   - 缺少健康检查端点
   - 没有服务版本管理
   - 缺少客户端的服务发现缓存机制

#### 6. **安全性问题**
   - 没有认证和授权机制
   - 缺少 TLS/SSL 加密传输
   - 没有访问控制和速率限制

#### 7. **可观测性缺陷**
   - 没有日志记录
   - 缺少性能指标采集（Prometheus）
   - 没有链路追踪支持

### 待改进清单

- [ ] 添加数据库持久化层（支持 PostgreSQL/SQLite）
- [ ] 实现完整的心跳检测和过期清理机制
- [ ] 修复并发操作的竞态条件
- [ ] 添加 TLS/SSL 支持
- [ ] 实现客户端 SDK 和发现缓存
- [ ] 添加日志系统（如 logrus）
- [ ] 集成 Prometheus 指标采集
- [ ] 添加单元测试和集成测试
- [ ] 实现优雅关闭机制
- [ ] 添加配置文件支持