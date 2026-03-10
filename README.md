# 课堂分析系统 - 后端服务

基于 Go + Gin 的课堂专注度分析系统后端服务，提供视频/图片分析、专注度检测、情绪分析、疲劳度检测等功能。

## 技术栈

- **语言**: Go 1.21+
- **Web 框架**: Gin
- **数据库**: MongoDB
- **缓存**: Redis
- **消息队列**: Apache Kafka
- **对象存储**: MinIO
- **依赖注入**: Google Wire
- **监控**: Prometheus

## 项目结构

```
analysis/
├── internal/
│   ├── domain/          # 领域模型
│   ├── repository/      # 数据访问层
│   │   └── dao/        # DAO 层
│   ├── service/         # 业务逻辑层
│   ├── web/             # Web 控制器
│   ├── ioc/             # 依赖注入容器
│   └── util/            # 工具函数
├── config/              # 配置文件
├── app.go               # 应用入口
├── wire.go              # Wire 依赖注入配置
└── wire_gen.go          # Wire 生成的代码
```

## 快速开始

### 环境要求

- Go 1.21+
- MongoDB 5.0+
- Redis 6.0+
- Kafka 3.0+
- MinIO

### 安装依赖

```bash
go mod download
```

### 配置

复制配置文件模板并修改：

```bash
cp config/conf.yaml.example config/conf.yaml
```

编辑 `config/conf.yaml`，配置数据库、Redis、Kafka、MinIO 等连接信息。

### 运行

```bash
# 直接运行
go run .

# 或使用 Wire 生成依赖注入代码后运行
wire && go run .
```

服务默认运行在 `:8080` 端口。

## API 接口

### 认证
- `POST /teacher/login` - 教师登录
- `POST /teacher/register` - 教师注册

### 分析
- `POST /teacher/analyze` - 提交分析任务
- `GET /teacher/status` - 获取任务状态
- `GET /teacher/ws` - WebSocket 实时进度

### 数据查询
- `GET /teacher/dashboard` - 仪表盘数据
- `GET /teacher/students` - 学生列表
- `GET /teacher/class-analysis` - 班级分析
- `GET /teacher/history` - 历史记录
- `GET /teacher/emotion-analysis` - 情绪分析
- `GET /teacher/fatigue-analysis` - 疲劳度分析

### 报告
- `GET /teacher/report` - 分析报告
- `GET /teacher/report/list` - 报告列表

完整 API 文档请参考 [api.md](./api.md)

## 时区处理

### 问题
系统最初存在时区偏差问题：服务器使用 UTC 时间，而前端显示需要北京时间（UTC+8），导致时间显示相差 8 小时。

### 解决方案

1. **创建时区工具包** (`internal/util/timezone.go`)
   - `GetBeijingTime()` - 获取当前北京时间
   - `ToBeijingTime(t)` - 将时间转换为北京时间
   - `FormatBeijingTime(t, layout)` - 格式化北京时间

2. **统一时间存储**
   - 所有 DAO 层使用 `util.GetBeijingTime()` 替代 `time.Now()`
   - 确保数据库存储的时间都是北京时间

3. **时间格式化**
   - 使用 `util.FormatBeijingTime()` 替代 `.Format()`
   - 避免 UTC 和北京时间转换问题

### 修复文件
- `internal/util/timezone.go` (新建)
- `internal/repository/dao/*.go` (所有 DAO 文件)
- `internal/repository/analysis.go`
- `internal/service/teacher.go`
- `internal/web/reports.go`
- `internal/web/emotion.go`
- `app.go`

## 开发

### 生成 Wire 代码

```bash
wire
```

### 运行测试

```bash
go test ./...
```

### 构建

```bash
go build -o analysis .
```

## 部署

### Docker

```bash
docker build -t classroom-analysis .
docker run -p 8080:8080 classroom-analysis
```

## 许可证

MIT License
