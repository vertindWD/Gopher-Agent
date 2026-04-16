# Gopher-Agent 🤖

基于 Go 构建的分布式异步智能任务编排系统。

本项目旨在解决传统 Web 服务在调用大语言模型（LLM）时长耗时导致的 **HTTP 504 阻塞** 和 **Goroutine 泄露** 问题。通过引入 Kafka 消息总线与事件驱动架构，实现了 AI 任务的极速响应与后台高并发异步流转；同时深度集成 ReAct 范式，赋予 Agent 动态调用外部真实网络数据的能力。

## 🌟 核心特性

- 🚀 异步非阻塞架构：摒弃传统同步请求，API 接收指令即可实现毫秒级响应落库并投递 Kafka，Worker 节点异步拉取消费。
- 🧠 ReAct 动态工具调用：集成 langchaingo，基于 Function Calling 机制封装本地多态工具（如 Github 热门仓库抓取），打破 LLM 信息孤岛。
- 🛡️ 工业级容错治理：
  - 完备的分布式任务状态机（Pending / Running / Completed / Failed）。
  - 基于 Context Timeout 的长连接生命周期精准控制，杜绝协程泄露。
  - 大模型 JSON 解析失败重试与熔断机制。
- 📊 全链路可观测性：基于 Zap 的结构化日志记录与 DTO 参数严格校验。

## 🏗️ 系统架构设计

[用户请求] -> HTTP API (Gin) -> 1. 任务落库 (MySQL: Pending)
                            -> 2. 投递消息 (Kafka Topic: agent_tasks)
                                     |
                                     v
[后台系统] <- 4. 状态更新 (MySQL) <- 3. Worker 并发消费 (Goroutine)
             (Running/Completed)     -> 触发 Agent ReAct 思考链
                                     -> 动态调用 Tools (API) -> LLM (DeepSeek)

## 🛠️ 技术栈

- 语言框架: Go 1.21+, Gin, GORM
- 基础设施: MySQL 8.0, Redis 7.0, Kafka (KRaft 模式)
- AI 编排: LangChainGo, DeepSeek API (兼容 OpenAI 格式)
- 工程化: Zap (日志), Viper (配置)

## 🚀 快速启动

### 1. 环境准备
确保你的本地或服务器已安装并运行以下中间件：
- MySQL (默认端口 3306)
- Redis (默认端口 6379)
- Kafka (默认端口 9092)

### 2. 获取代码
git clone https://github.com/vertindWD/gopher-agent.git
cd gopher-agent

### 3. 修改配置
复制一份配置文件示例，并填入你的真实数据库密码和 LLM API Key：
cp configs/config.yaml.example configs/config.yaml

### 4. 运行服务
go run cmd/server/main.go

## 🔌 API 接口文档

### 1. 提交 Agent 任务 (异步)
POST /api/v1/agent/task

curl -X POST http://localhost:8080/api/v1/agent/task \
-H "Content-Type: application/json" \
-d '{"prompt": "帮我查一下 Github 上 Star 数量最高的 Go 语言 Web 框架是什么？"}'

### 2. 轮询查询任务结果
GET /api/v1/agent/task/:task_id

## 📄 许可证
[MIT License](LICENSE)
