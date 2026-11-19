# OpenDataBuddy

开源数据智能体工作台：用对话完成取数、开发、治理与分析。

后端使用 **GoFrame v2.10.4**（2026-11-10 核对的最新稳定版），本地 **SQLite + DuckDB** 存储；前端为 **Vue 3 + Vite + Naive UI + Pinia + TypeScript**（`frontend/`，参照 InnoData 工程结构），构建产物由 Go 单进程直接提供。Python 仅执行 Notebook / 工作流 Python 节点，不承担 HTTP、鉴权、调度或业务服务。

## 启动

需要 Go 1.24+、启用 CGO 的 C/C++ 编译环境。macOS 可使用 Xcode Command Line Tools；Linux 可使用 GCC。只有执行 Python cell / Python 节点时才需要 Python 3.10+，不需要安装 pip 包。

```bash
go run ./cmd/opendatabuddy
# 或
./start.sh --port 8000
```

Windows：`start.bat`。打开 <http://127.0.0.1:8000>，默认账号 `admin@local` / `admin123`。

首次启动会创建示例工作空间和半导体数仓（Doris 语义对齐）：5,200 行晶圆产出工单、600 位客户、产品（制程节点）、晶圆厂、良率/营收指标、看板与计算资源。示例工单日期会随首次启动日期平移。

```bash
# 构建后运行，不再需要 Go 编译器
make build
./bin/opendatabuddy --host 127.0.0.1 --port 8000

# 隔离开发数据
DATABUDDY_DATA_DIR=/tmp/odb-dev go run ./cmd/opendatabuddy --port 8042
```

进程退出会停止调度并关闭数据库、Python 内核。`--root` 指定项目根目录；相对 `--data-dir` 相对于该目录解析。运行时需要 `frontend/dist/`（前端构建产物）和可选 Python 内核脚本。

### 前端开发

前端工程在 `frontend/`，需要 Node.js 24+。首次构建：

```bash
cd frontend
npm ci          # 安装依赖
npm run build   # 构建到 frontend/dist/，Go 直接提供；保持单进程部署
```

开发模式（Vite dev server 代理 `/api` 到 Go 8000）：`npm run dev`，访问 <http://127.0.0.1:3082>。类型检查：`npm run typecheck`。

## 功能

| 能力 | 实现 |
| --- | --- |
| Buddy | Go 实现的 OpenAI 兼容工具循环、22 个工具、**三场景按模式分流工具与提示词**、工具白名单、规则兜底、记忆、知识检索、执行轨迹与审批恢复、**任务执行计划卡片** |
| 数据接入 | CSV / JSON / Parquet / XLSX 接入、SQL 建表、卷管理、60 个数据源模板及连接诊断、**离线同步任务（全量 / 水位增量）**、**数据对账**、**CSV / Parquet 出湖** |
| 数据工程 | Notebook 跨 cell 变量、13 个魔法命令、9 类 DAG 节点、12 种条件、**周期 / 一次性 / 事件三种触发**、重试、时区 cron、持续运行、节点拖拽与位置保存 |
| 发布管理 | Bundle YAML 的 init / validate / deploy / run / summary / destroy，重复发布幂等，只销毁本 Bundle 创建的资源 |
| 数据科学 | 特征定义、版本、计算、物化、训练集；模型注册、版本比较、冠军晋级、服务启停、A/B 权重、伸缩建议与**真实推理端点** |
| 数据分析 | 指标 SQL 编译、问数、贡献归因、四章节报告、报告定时运行、**看板卡片编辑与拖拽搭建**、空间嵌出配置 |
| 数据治理 | 六类质量规则、**基于剖析统计的规则智能推荐**、YAML 导入导出、阻断预检查、三种剖析模式、五个漂移指标、异常确认、表级/字段级血缘、本体和术语表 |
| 运维与安全 | 四层运行视图、规则诊断、**六类告警渠道真实投递**（邮件/Webhook/Teams/企微/Slack/飞书）、双向 Guardrails、**OBO 主体校验与工具三要素**、只读 SQL 护栏、会话与空间权限、读写审计 |

保留原 v0.3 的 **220 条 API 路径 / 280 个 HTTP 操作**，另补充离线 API 文档与 MCP HTTP 入口。接口存在性、主要读取接口和重点写入链路均有可重复测试；测试范围见 [HANDOFF.md](HANDOFF.md)。

- API 文档：<http://127.0.0.1:8000/api/docs>（无外部 CDN）
- OpenAPI：<http://127.0.0.1:8000/api/openapi.json>
- 健康检查：<http://127.0.0.1:8000/api/health>
- 历史兼容契约：[openapi.json](internal/platform/assets/openapi.json)

## 模型配置

未配置模型时，可使用注册指标问数、目录查询、表结构、敏感扫描、质量检查、**质量规则智能推荐**（`GET /api/quality/suggest`）和知识问答。开放式编排需要在「平台设置」接入模型。

```dotenv
DEEPSEEK_API_KEY=your-key
DEEPSEEK_BASE_URL=https://api.deepseek.com/v1
DEEPSEEK_MODEL=deepseek-chat
# 或任意 OpenAI 兼容端点
# LLM_BASE_URL=http://127.0.0.1:11434/v1
# LLM_API_KEY=ollama
# LLM_MODEL=qwen2.5-coder:7b
```

配置可来自进程环境、项目 `.env` 或平台设置。环境变量不会被 `.env` 覆盖。其余选项见 [.env.example](.env.example)。默认业务时区为 `Asia/Shanghai`，可用 `DATABUDDY_TIMEZONE` 修改；Python 解释器可用 `DATABUDDY_PYTHON` 指定。

### 三类模型分别接入

除 Chat/Reasoning 外，还可独立接入 Embedding 与 OCR（均为 OpenAI 兼容端点）：

```dotenv
# Chat/Reasoning（qwen3.6 等推理模型建议关闭思考，避免 reasoning 占满 max_tokens）
LLM_BASE_URL=http://10.30.6.200:18019/v1
LLM_MODEL=qwen3.6-35b-a3b
LLM_API_KEY=your-key
LLM_ENABLE_THINKING=false
LLM_MAX_TOKENS=2048

# Embedding：配置后语义检索自动改用 bge-m3 稠密向量（1024 维），
# 未配置则回退到内置哈希稀疏向量
EMBEDDING_BASE_URL=http://10.30.6.200:18001/v1
EMBEDDING_MODEL=bge-m3
EMBEDDING_API_KEY=your-key
EMBEDDING_DIM=1024

# OCR：图片文字识别，通过 /api/ocr 或 /api/knowledge/ocr-ingest 调用
OCR_BASE_URL=http://10.30.6.200:18009/v1
OCR_MODEL=deepseek-ocr
OCR_API_KEY=your-key
```

对应接口：`POST /api/settings/llm/test`、`POST /api/settings/embedding/test`、`POST /api/settings/ocr/test` 可分别做连通性测试；`GET /api/settings/models` 查看三类模型的脱敏配置。注意：模型服务地址若为内网/私有 IP，请求会绕过系统 HTTP 代理直连。

## MCP

先启动控制台 HTTP 服务并登录，再构建 Go 程序。独立 stdio 进程通过 GoFrame `/api/mcp` 复用同一套工具、身份与数据库，避免多个进程争用 DuckDB 文件。

```bash
make build
DATABUDDY_TOKEN='<控制台会话 token>' ./bin/opendatabuddy --mcp
```

客户端配置示例（替换为本机绝对路径）：

```json
{
  "mcpServers": {
    "opendatabuddy": {
      "command": "/path/to/OpenDataBuddy/bin/opendatabuddy",
      "args": ["--mcp"],
      "env": {
        "DATABUDDY_TOKEN": "your-console-session-token",
        "DATABUDDY_API_URL": "http://127.0.0.1:8000"
      }
    }
  }
}
```

可用 `DATABUDDY_WORKSPACE_ID` 指定有权访问的空间。需要审批的写操作会被 MCP 拒绝，须回到 Buddy 控制台处理。

## 数据与迁移

默认数据保存在 `data/metadata.db`、`data/warehouse.duckdb`、`data/uploads/`、`data/volumes/`、`data/egress/`（出湖文件，按工作空间分目录）。停止服务后复制整个数据目录即可备份。

旧 SQLite 表采用增量补列方式迁移；已有用户、密码、会话和工作空间保留。已有目录继续使用原物理命名；新建目录使用独立物理前缀，避免不同空间的同名表互相覆盖。不要同时运行旧 Python 服务和新 Go 服务访问同一个 DuckDB 文件。

## 开发与验证

```bash
make test        # 原生引擎、迁移、权限、调度、审批、内核回调及接口清单测试
make test-http   # 启动随机本地端口，使用临时数据，执行真实 HTTP 验收
make vet
make build
```

测试不会使用项目 `data/`。真实模型调用用本地模拟服务验证；真实供应商的密钥、网络和模型表现需要在配置后验证。

```text
cmd/opendatabuddy/        Go 程序入口，HTTP / stdio MCP
internal/platform/       鉴权、存储、查询及各业务引擎
internal/platform/assets 表结构、增量迁移、示例数据、元信息和历史契约
scripts/python_kernel.py 唯一 Python 运行时：执行用户代码，通过回调请求 Go 查询
frontend/                Vue 3 + Vite + Naive UI 前端工程（当前实现），构建到 dist/
web/                     旧版原生前端（保留存档，Go 已不再提供）
```

架构见 [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md)，与商业产品的差异见 [docs/OSS_DIFF.md](docs/OSS_DIFF.md)。

## 当前边界

告警渠道真实出网投递（设 `DATABUDDY_ALERT_SIMULATE=1` 可退回离线模拟），投递失败只记录状态、不自动重投；模型服务除版本与流量管理外，`POST /api/serving/services/{id}/predict` 以登记的 SQL 推理语句真实执行，不是模型框架；离线同步任务目前只支持 SQLite 源，外部目录中其他类型数据源需导出后接入；目标表仍是整表重建，不做增量列演进。XLSX 使用首个工作表。SQL 血缘采用投影规则推断，复杂子查询需要人工核对。Notebook 内核为单机进程状态，重启后变量重置；本地用户代码没有内存隔离，默认 30 秒执行超时。

本项目为独立实现，不包含腾讯云代码、SDK、计费、席位或云资源依赖。MIT License。
