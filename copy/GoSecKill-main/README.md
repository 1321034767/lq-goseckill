# GoSecKill

GoSecKill 是一个基于 Go 语言的电子商务闪购系统，专为高并发环境设计，采用微服务架构，使用 RabbitMQ 进行消息传递，使用 GORM 进行 MySQL 操作，使用 Redis 进行缓存，并使用 Iris 框架构建 Web 服务。

## 系统要求

- Go 1.22.0 或更高版本
- MySQL 5.7+ 或 MySQL 8.0+
- RabbitMQ 3.8+
- Redis（可选）

## 快速开始

### 1. 克隆项目

```bash
git clone <repository-url>
cd GoSecKill-main
```

### 2. 安装依赖

```bash
go mod download
```

### 3. 配置数据库和消息队列

复制配置文件示例并修改配置：

```bash
cp config/config.example.toml config/config.toml
```

编辑 `config/config.toml` 文件，配置以下信息：

```toml
[log]
developerMode = true
outputToConsole = true

[server]
serverPort = ":8080"    # 前端服务端口
adminPort = ":8081"     # 管理后台端口

[database]
username = "your_username"    # MySQL 用户名
password = "your_password"    # MySQL 密码
host = "localhost"            # MySQL 主机地址
port = "3306"                 # MySQL 端口
database = "your_database"    # 数据库名称

[rabbitmq]
url = "amqp://guest:guest@localhost:5672/"  # RabbitMQ 连接地址
```

### 4. 准备数据库

确保 MySQL 数据库已创建，系统会在首次运行时自动创建表结构。

### 5. 启动 RabbitMQ

确保 RabbitMQ 服务正在运行：

```bash
# 使用 Docker 启动 RabbitMQ（推荐）
docker run -d --name rabbitmq -p 5672:5672 -p 15672:15672 rabbitmq:3-management

# 或使用系统服务
sudo systemctl start rabbitmq-server
```

### 6. 运行项目

项目包含三个主要服务，需要在不同的终端窗口中分别运行：

#### 方式一：分别运行各个服务

**终端 1 - 启动管理后台服务（端口 8081）：**
```bash
go run cmd/admin/main.go
```

**终端 2 - 启动前端服务（端口 8080）：**
```bash
go run cmd/server/main.go
```

**终端 3 - 启动消息队列消费者：**
```bash
go run cmd/consumer.go
```

#### 方式二：编译后运行

**编译所有服务：**
```bash
# 编译管理后台
go build -o bin/admin cmd/admin/main.go

# 编译前端服务
go build -o bin/server cmd/server/main.go

# 编译消费者
go build -o bin/consumer cmd/consumer.go
```

**运行编译后的程序：**
```bash
# 终端 1
./bin/admin

# 终端 2
./bin/server

# 终端 3
./bin/consumer
```

### 7. 访问系统

- **前端服务（用户端）**: http://localhost:8080
- **管理后台**: http://localhost:8081

### 8. 用户注册和登录

系统没有默认用户，需要先注册账号才能登录使用。

#### 注册新用户

1. 访问注册页面：http://localhost:8080/user/register
2. 填写注册信息：
   - **用户名（Username）**：4-10个字符
   - **密码（Password）**：8-16个字符
3. 点击 "Register" 按钮完成注册
4. 注册成功后会自动跳转到登录页面

#### 登录系统

1. 访问登录页面：http://localhost:8080/user/login
   - 或者直接访问 http://localhost:8080 会自动重定向到登录页面
2. 使用注册时的用户名和密码登录
3. 登录成功后会自动跳转到产品页面

#### 示例账号

您可以注册任意符合要求的账号，例如：
- 用户名：`testuser`（4-10个字符）
- 密码：`password123`（8-16个字符）

## 项目结构

```
GoSecKill-main/
├── cmd/              # 可执行程序入口
│   ├── admin/        # 管理后台服务
│   ├── server/       # 前端服务
│   └── consumer.go   # 消息队列消费者
├── config/           # 配置文件
├── internal/         # 内部包
│   ├── config/       # 配置管理
│   ├── database/     # 数据库连接
│   └── routers/      # 路由定义
├── pkg/              # 公共包
│   ├── log/          # 日志管理
│   ├── mq/           # 消息队列
│   └── models/       # 数据模型
└── web/              # Web 资源文件
    ├── admin/        # 管理后台前端资源
    └── server/       # 前端服务资源
```

## 注意事项

1. **数据库连接**：确保 MySQL 服务正在运行，并且配置的数据库已创建
2. **RabbitMQ 连接**：确保 RabbitMQ 服务正在运行，否则消息队列功能无法使用
3. **端口占用**：确保 8080 和 8081 端口未被占用
4. **配置文件路径**：程序从 `./config` 目录读取 `config.toml` 文件，请确保在项目根目录运行

## 故障排查

- **数据库连接失败**：检查 MySQL 服务是否运行，用户名密码是否正确
- **RabbitMQ 连接失败**：检查 RabbitMQ 服务是否运行，连接地址是否正确
- **端口被占用**：修改 `config/config.toml` 中的端口配置
- **配置文件未找到**：确保在项目根目录运行程序，且 `config/config.toml` 文件存在