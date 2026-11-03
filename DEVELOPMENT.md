# CraftShelf 开发文档

> **文档版本**：v1.5（**S5 开工前缺陷修复后校对**）
> **校对日期**：2026-11-03
> **校对依据**：`F:\project\CraftShelf` 全部源文件、`go.mod`、迁移文件、`git status`、实际命令输出
> **当前阶段**：S0–S4 代码已落地；P-09/P-10/P-11 等阻断项已修复；T1 架构守护测试已按用户决定移除，测试统一延后到 S5.5；S4 仍因 auth/jwt 测试缺失而部分验收（DEV-001）；S5 未开始
> **项目性质**：个人学习 + 求职作品项目（Go 后端 + AI Agent）

> ## ⚠️ 阅读前必读
>
> **本文档已按仓库实际代码逐文件核对，一律以代码为准。**
>
> 上一版（v1.3）存在多处与代码不符的事实性错误（迁移文件名、约束名、测试状态等），已全部修正，修正清单见 **§2.5**。
>
> **若本文档与代码再次冲突**：以代码为准，并须在 §2.5 登记修正。
>
> **核对基线（当前必须成立的事实）**
>
> | 项 | 实际值 |
> |---|---|
> | 模块路径 | `github.com/IssierHM/CraftShelf` |
> | Go 版本 | `go 1.26.0` |
> | 唯一迁移 | `migrations/20261027151811_create_works.sql`（**文件名是 works，内容建 `users` 表**） |
> | 测试 | `internal/pkg/config/config_test.go`（2 个函数）；架构守护测试已移除（S5.5 统一补） |
> | 已补齐文件 | `.golangci.yml`（v2 格式）、`.env.example` |
> | 空文件 | `README.md`（0 字节） |
> | 未提交 git | 无（2026-11-03 已提交，见 §2.5） |
> | lint 工具 | golangci-lint **v2.15.0**（v1.64.8 在 Go 1.26 下 panic，已升级） |

---

## 0. 如何使用这份文档

### 0.1 文档目的

本文档是 CraftShelf 项目的**唯一权威状态记录**。它的设计目标是：

1. **上下文恢复**：任何人（或 AI）读完本文档即可准确知道"做到哪、下一步做什么、有哪些坑"
2. **决策追溯**：每个技术选择都记录了理由与被放弃的方案，避免重复讨论
3. **契约固化**：接口、数据模型、命名规范一旦写入，后续实现必须遵守

### 0.2 推荐阅读顺序

| 目的 | 读哪几节 |
|---|---|
| 快速恢复工作 | **§2 进度总览** → **§2.5 文档修正记录** → **§13.4 如何恢复工作** |
| 理解架构 | §3 整体架构 → §4 模块清单 → §5 接口契约 |
| 开始写下一阶段代码 | §10 技术难点设计 → §6 数据模型 → §9 工程规范 |
| 准备面试 | §12 面试口径 → §7 决策记录 → §13.1 术语表 |
| 排查环境问题 | §13.3 常见故障 → §8 配置参考 |
| **接手前确认真实状态** | **§2.4 已知缺口** → **§2.5 修正记录** |

### 0.3 状态标记约定

| 标记 | 含义 |
|---|---|
| ✅ **已落地** | **已按实际代码核实存在且可用** |
| 🟡 **部分验收** | 代码存在，但验收标准未达成（见偏离记录） |
| 🔶 **部分交付** | 设计已完成，代码未完整给出 |
| ⬜ **未开始** | 仅有设计概要 |
| ⚠️ **待核实** | 依赖外部信息，必须在实施时确认 |
| 🚧 **已知缺口** | 已识别的不一致、未完成项或缺陷 |

> **v1.4 起，✅ 仅表示"代码中确实存在"，不再表示"文档中已写出代码"。** 二者在旧版中曾被混用，是造成混乱的根因之一。

---

## 1. 项目概览

### 1.1 定位

**CraftShelf** 是面向**游戏模组作者与视觉创作者**的 UGC 内容发布与分发平台。

创作者上传并持续维护作品；用户订阅、检索、下载并参与互动；平台内置 AI Agent 流水线，承担内容审核、元数据生成与语义检索。

抽象上，平台把「模组资源包」与「绘画作品」统一为同一套**作品模型**：

```
作者 (User) ──1:N──► 作品 (Work) ──1:N──► 版本 (Version) ──1:1──► 文件 (File, 内容寻址)
```

### 1.2 目标问题

项目必须实际解决以下三个问题，而非仅作功能演示：

| 问题 | 现状 | 设计要求 |
|---|---|---|
| **依赖关系不可校验** | 用户装完资源 A 才发现缺 B，或 A 与 C 冲突；作者无法声明版本范围 | 版本模型支持依赖声明（含可选依赖与版本范围）；发布前完成依赖存在性、范围匹配、循环依赖校验；失败须给出可定位的错误 |
| **重复内容浪费存储与带宽** | 同一热门资源被大量用户重复上传下载 | 文件实体以内容哈希（SHA-256）为唯一标识，实现跨用户去重与零字节秒传；生命周期由引用计数管理 |
| **内容缺乏结构化组织** | 标签靠人工、分类靠猜测、检索靠关键字字面匹配 | 提供 Schema 注册表支撑品类扩展；检索同时支持结构化筛选、关键字检索、语义检索 |

### 1.3 与既有平台的差异

| 参照 | 借鉴 | 差异 |
|---|---|---|
| Steam 创意工坊 | 发布/订阅、版本管理、依赖解析 | 后端完全自建；不绑定单一游戏或引擎 |
| Pixiv / ArtStation | 创作者主页、关注流、标签体系 | 资源文件与画作共用同一作品模型与同一分发链路 |
| 通用网盘 | 分片上传、断点续传 | 面向**分发**而非存储：预签名直链、引用计数、内容寻址去重 |

### 1.4 泛化能力的实现方式（强制约束）

平台**不得**预设品类字段。每个品类注册一份属性定义（字段名、类型、是否必填、枚举取值、长度约束）：

- 上传时按定义**动态生成表单**并在服务端校验
- 检索时按定义**暴露可筛选维度**
- 新增品类只需注册定义，**不得**要求修改表结构或新增领域结构体

**这是"可泛化"诉求的唯一合法实现路径。** 判断标准：加一个新品类（例如"音乐素材"）时，若需要写新的 Go struct 或改 `CREATE TABLE`，则设计失败。

### 1.5 项目背景与目标优先级

开发者画像：软件工程专业学生，具备 Java/Python 经验，**首次以 Go 完成中大型项目**。

由此确定两条目标，优先级不可颠倒：

| 优先级 | 目标 | 含义 |
|---|---|---|
| **P0** | 学会 Go 并做出能讲透的作品 | 架构选择偏向"可讲透"而非"追逐复杂度" |
| P1 | 产出可在秋招中量化陈述的项目经历 | 每个技术点须有可验证的结果或数据 |

架构选择因此偏向**单体分层**而非微服务——不是因为微服务不好，而是 9 周内做不扎实，且会引入面试中无法自圆其说的追问风险。

---

## 2. 进度总览

### 2.1 阶段状态表

| 阶段 | 周期 | 目标 | 状态 | 实际代码状态（2026-11-03 核对） |
|---|---|---|---|---|
| **S0** 工具链 | 1–2 天 | Go / Docker / 五个 CLI 就绪 | ✅ 已落地 | 五项工具已通过 `tool` 指令固化在 `go.mod` |
| **S1** 语言基础 | 第 1 周 | Go 语法 + 并发 SHA-256 CLI | ⬜ 无仓库产物 | 练习代码未纳入本项目（独立练习） |
| **S2** 并发与上下文 | 第 2 周 | context / errgroup / 流式处理 | ⬜ 无仓库产物 | 练习代码未纳入本项目（独立练习） |
| **S3** 环境与首个接口 | 第 3 周 | 三服务 + 注册登录鉴权 | ✅ 已落地 | 26 个源文件全部存在，接口可运行 |
| **S4** 工程化分层 | 第 4 周 | wire / 测试 / 架构检查 / lint | 🟡 部分验收 | wire、`.golangci.yml`、Makefile 已落地；**架构守护测试已移除，auth/jwt/service 测试仍缺**（DEV-001） |
| **S5** 分片上传与秒传 | 第 5 周 | 内容寻址 + S3 Multipart | ⬜ 未开始 | 设计完成；`config.go` 的 S3 字段**已就绪** |
| **S5.5** 测试整合 | 第 5 周后 | 补齐全部单测 | ⬜ 未开始 | **硬性检查点**（DEV-001 收口） |
| **S6** 版本与依赖 DAG | 第 6 周 | 拓扑排序 + 环路径还原 | ⬜ 未开始 | 设计完成，代码未给 |
| **S7** AI Agent | 第 7 周 | 三段式编排 + 评测集 | ⬜ 未开始 | 设计完成，代码未给 |
| **S8** 可观测与部署 | 第 8 周 | 指标 / 链路 / 压测 / 上线 | ⬜ 未开始 | 概要 |
| **S9** 收尾与陈述 | 第 9 周 | 文档 / 视频 / 面试口径 | ⬜ 未开始 | 概要 |
| **S10** 缓冲 | 第 10 周 | 吸收偏差 | ⬜ 未开始 | — |

> **S1 / S2 的判定说明**：这两个阶段交付的是**独立练习程序**（并发 SHA-256 CLI、超时 worker、并发下载器），从未纳入 `CraftShelf` 仓库。因此仓库中不存在对应产物，但这不代表练习未做。**接手的 Agent 不应在仓库中寻找 S1/S2 产物。**

**总投入估算**：约 185 小时（每日 3 小时 × 9 周）。每日 5 小时可压至 6 周。

### 2.2 已完成阶段明细

#### S0 工具链 ✅

**任务**：安装并验证 Go、Docker Desktop、sqlc、goose、wire、golangci-lint。

**实际做法（与设计不同，已按代码修正）**

仓库**没有**用裸 `go install`，而是把工具声明进 `go.mod` 的 **`tool` 指令**（Go 1.24+ 特性）：

```
tool (
	github.com/golangci/golangci-lint/cmd/golangci-lint
	github.com/google/uuid
	github.com/google/wire/cmd/wire
	github.com/pressly/goose/v3/cmd/goose
	github.com/sqlc-dev/sqlc/cmd/sqlc
)
```

**好处**：工具版本随 `go.mod` 走，队友与 CI 拿到完全一致的版本；无需额外 Makefile 锁版本。

**注意两个问题**：

| 问题 | 说明 |
|---|---|
| `github.com/google/uuid` 在 `tool` 块里 | uuid 是**运行库**不是工具，放在 `tool` 块是配置错误。它同时也出现在 `require` 中，功能不受影响，但语义错误 |
| 调用方式 | 声明后必须用 `go tool sqlc generate`、`go tool wire ./cmd/api`、`go tool goose ...`，不能再直接敲 `sqlc` |

**模块代理**

```powershell
go env -w GOPROXY=https://goproxy.cn,direct
```

**验证**：`go tool sqlc version`、`go tool goose -version`、`go tool wire help` 均可执行。

**已知障碍**

| 症状 | 成因 | 处置 |
|---|---|---|
| 直接敲 `sqlc` 命令找不到 | 工具已改为 `go tool` 调用 | 用 `go tool sqlc ...` |
| 下载停滞 | 未配模块代理 | 配置 `GOPROXY` |

**排错心法**：报"找不到命令"时，先看 `%GOPATH%\bin` 里**文件在不在**。在 → `PATH` 问题；不在 → 安装失败（或该项目改用 `go tool` 了）。

#### S1 语言基础 ⬜ 无仓库产物

> **⚠️ 更正**：S1 的练习程序（并发 SHA-256 CLI）是**独立练习**，未纳入 `CraftShelf` 仓库。仓库中不存在该产物。旧版文档标记为"✅ 已交付 / 完整可运行"会被误读为"仓库里有这个文件"。**接手的 Agent 不要去找它。**

**学习范围**：变量、切片、map、结构体、方法、接口、错误处理、包与可见性。

**明确排除**：反射、底层编程、汇编、cgo、高级泛型（与项目目标无关）。

**练习内容**：并发计算多文件 SHA-256 的 CLI（独立目录，未提交到本项目）。

**引入的构件**（均为 S5 直接依赖）：

| 构件 | S5 对应位置 |
|---|---|
| `os.Open` + `defer Close` | 读取上传临时文件 |
| `io.Copy` + `hash.Hash` | 合并后**流式**重算哈希 |
| `crypto/sha256` | 内容寻址核心算法 |
| goroutine + `sync.WaitGroup` | 并发上传分片 |
| 预分配结果切片按索引写入 | 分片结果汇总（免锁） |

**核心教学设计**：

- 错误是返回值不是异常；每个 `err` 都要接住并用 `%w` 包装上下文
- `io.Reader` 抽象让同一函数可处理文件/网络/内存
- 预分配切片 + 按索引写入 = 无需加锁（不同 goroutine 写不同内存位置）
- `wg.Add` 必须在 `go` 之前；不能放进 goroutine 内部

**验收**：与系统 `sha256sum` 结果一致；`go run -race` 无告警；10 文件并发耗时显著小于串行。

#### S2 并发与上下文 ⬜ 无仓库产物

> **⚠️ 更正**：S2 的三个练习同样是**独立程序，未纳入仓库**。仓库中不存在对应文件。

**本阶段为关键路径，未通过不得进入 S3。**

**学习范围**

| 概念 | 须掌握 |
|---|---|
| goroutine | 轻量但非无成本，须以并发上限控制 |
| channel | 无缓冲/有缓冲差异、`select`、关闭语义 |
| `context` | 取消、超时、传值 |
| `sync` | `Mutex`、`WaitGroup`、`Once`、`atomic` |
| `errgroup` | 并发执行、错误收集、取消传播 |
| `semaphore` | 并发数量上限 |
| `io` 接口 | `io.Copy`、`io.MultiWriter` |

**三个练习（独立程序，未纳入仓库）**

| 练习 | 练的能力 | 对应后续 |
|---|---|---|
| 超时控制 worker | `context` + `select` + `ctx.Done()` | S7 步骤超时 |
| 流式合并哈希 | `io.MultiWriter` + 固定缓冲 | S5 合并校验 |
| 并发下载器 | `errgroup.SetLimit` + 取消传播 | **S5 分片并发原型** |

**核心纪律**

- `ctx` 必须是第一个参数
- `defer cancel()` 必须写（否则定时器与父引用泄漏）
- `select` 里必须监听 `ctx.Done()`
- 只有发送方能关闭 channel
- `http.Get` 会让取消失效，必须用 `http.NewRequestWithContext`

**验收**：`-race` 无告警；能回答五问（channel 关闭后读写、ctx 位置、WaitGroup.Add 位置、有缓冲满后发送、ctx.Done 是关闭还是发值）。

#### S3 环境与首个接口 ✅

**目标**：跑通一条完整链路 请求 → 业务 → 数据库 → 响应，并把三个依赖服务全部验证到位。

**实际文件清单（2026-11-03 核对，以 `git ls-files` + 磁盘为准）**

```
CraftShelf/
├── deploy/
│   ├── docker-compose.yml                    ✅ 存在
│   ├── postgres/init/01_extensions.sql       ✅ 存在
│   └── seaweedfs/s3.json                     ✅ 存在
├── migrations/
│   └── 20261027151811_create_works.sql       ✅ 存在（日期命名，内容建 users 表）
├── sqlc.yaml                                 ✅ 存在
├── Makefile                                  ✅ 2.3 KB（build/test/lint/generate/check）
├── README.md                                 🚧 存在但为 0 字节空文件
├── .gitignore                                ✅ 存在（仅 3 类规则）
├── .env                                      ✅ 存在（未跟踪，正确）
├── .env.example                              ✅ 已创建（P-13）
├── .golangci.yml                             ✅ 已创建（v2 格式，P-12）
├── api.exe                                   ✅ 已删除（P-18）
├── cmd/api/
│   ├── main.go                               ✅ 已提交
│   ├── app.go                                ✅ 已提交
│   ├── providers.go                          ✅ 已提交
│   ├── wire.go                               ✅ 已提交
│   └── wire_gen.go                           ✅ 已提交
└── internal/
    ├── domain/
    │   ├── user.go                           ✅（json tag 已移除）
    │   ├── errors.go                         ✅
    │   └── architecture_test.go              ❌ 已移除（T1 延后到 S5.5）
    ├── pkg/
    │   ├── config/config.go                  ✅（S3 字段齐备，P-09 已修）
    │   ├── config/config_test.go             ✅ 2 个测试函数
    │   ├── db/pool.go                        ✅
    │   └── jwt/jwt.go                        ✅（NewSigner 返回 error）
    ├── repository/
    │   ├── queries/users.sql                 ✅ 4 条查询
    │   ├── sqlc/                             ✅ 生成（sqlc v1.31.1）
    │   └── user.go                           ✅
    ├── service/auth.go                       ✅（P-10 已修）
    ├── handler/
    │   ├── dto/common.go                     ✅
    │   ├── dto/user.go                       ✅
    │   ├── respond.go                        ✅
    │   ├── errors.go                         ✅
    │   ├── auth.go                           ✅
    │   ├── user.go                           ✅
    │   ├── health.go                         ✅
    │   └── router.go                         ✅
    └── middleware/
        ├── auth.go                           ✅
        └── recovery.go                       ✅
```

**已实现的端点**

| 方法 | 路径 | 说明 |
|---|---|---|
| GET | `/healthz` | 存活探针，不检查依赖 |
| GET | `/readyz` | 就绪探针，**实际 ping 数据库** |
| POST | `/api/v1/auth/register` | 注册 |
| POST | `/api/v1/auth/login` | 登录，返回令牌 |
| GET | `/api/v1/users/me` | 需鉴权，返回当前用户 |

**关键实现要点（按实际代码）**

| 主题 | 实际实现 |
|---|---|
| 密码 | bcrypt，cost 由 `APP_BCRYPT_COST` 控制，限定 `[10,14]`，默认 12；哈希定长 60 字节，`password_hash` 用 `TEXT` |
| 登录失败 | 两种情形返回**同一个** `domain.ErrInvalidCredentials`；用户不存在时执行一次 `CompareHashAndPassword(s.dummyHash, ...)` 防时间侧信道。**P-10 已修：仓储返回其他错误时直接返回，不再解引用 nil user** |
| JWT | 只放 `sub`/`iat`/`exp`；校验签名算法为 HMAC（防 `alg=none`）；`NewSigner` **自身**也校验密钥（≥32 字节、无空白/控制字符），与 config 层形成**双层校验** |
| 唯一约束翻译 | 仓储层捕获 `pgconn.PgError` 的 `23505`，按 `ConstraintName` 映射。**实际约束名是 `user_email_key` / `user_username_key`（单数 `user_`）** |
| 连接池 | `MaxConns=20`、`MinConns=4`、`MaxConnLifetime=1h`、`MaxConnIdleTime=30m`、`HealthCheckPeriod=1m`、`statement_timeout=5000` |
| 健康检查 | `pg_isready` 必须带 `-U` `-d`，否则误判 |
| 环境变量加载 | `main.go` 用 `github.com/joho/godotenv` 加载 `.env`（`os.ErrNotExist` 时忽略） |
| 日志级别 | `slog.LevelDebug`（非 Info） |
| 错误码 | 参数校验失败返回 `invalid_request`（**非** `invalid_argument`） |

**S3 建立的四条纪律**

1. **三层数据形态分离**：DTO / domain / Row（sqlc），必须显式转换
2. **配置集中**：只在 `main` 读取一次，密钥无默认值，缺失即启动失败
3. **两种健康检查**：存活与就绪分离，就绪必须真探测
4. **数据库约束兜底**：应用层查重仅用于友好提示，唯一约束才是正确性保证

> **⚠️ 纪律 1 的部分违反**：`domain.User` 当前带 `json` tag，且包含 `password_hash`。虽然现有 handler 走 DTO 转换未泄漏，但 domain 一旦被直接序列化就会泄露。见 **P-14**。

#### S4 工程化分层 🟡

> **验收状态**：wire 工程化已落地；T1 架构守护测试曾于 2026-11-03 重建并通过，随后按用户决定移除，测试统一延后到 S5.5。auth/jwt 测试仍缺。S4 验收（`go test -race ./...` 全绿、service 覆盖率 ≥60%）**未达成**，由 **S5.5** 关闭，详见 **DEV-001**（§2.2.1）。

**目标**：把"能跑"变成"可维护、可测试、边界可验证"。

**实际文件状态（按磁盘与 git 核对）**

| 文件 | 作用 | 实际状态 |
|---|---|---|
| `cmd/api/app.go` | `App` 结构体（`Server` + `Log`） | ✅ 已提交（2026-11-03） |
| `cmd/api/providers.go` | `newLogger`、`newPool`、`newSigner`、`newAuthService`、`newHTTPServer` | ✅ 已提交（2026-11-03） |
| `cmd/api/wire.go` | 注入器声明（`//go:build wireinject`），函数名 **`InitializeApp`** | ✅ 已提交（2026-11-03） |
| `cmd/api/wire_gen.go` | wire 生成（`//go:build !wireinject`） | ✅ 已提交（2026-11-03） |
| `cmd/api/main.go` | 生命周期 + `godotenv` 加载 `.env` | ✅ 已提交（2026-11-03） |
| `internal/pkg/config/config_test.go` | `TestValidateJWTSecret`（10 例）+ `TestLoadTrimsJWTSecret` | ✅ 2 个测试函数 |
| `internal/domain/architecture_test.go` | `TestDomainHasNoFrameworkDependency` | ❌ **已移除（2026-11-03，延后到 S5.5）**；`make verify-layers` 保留为手动检查 |
| `internal/service/auth_test.go` | fake 仓储 + 表驱动 + 安全断言 | ❌ 不存在（S5.5） |
| `internal/pkg/jwt/jwt_test.go` | 往返、篡改、换密钥、负 TTL 过期 | ❌ 不存在（S5.5） |
| `.golangci.yml` | 静态检查配置 | ✅ v2 格式（golangci-lint v2.15.0） |
| `.env.example` | 环境变量模板 | ✅ 已创建（占位值，不含真实密钥） |
| `Makefile` | 统一入口 | ✅ 2.3 KB，含 build/test/lint/generate/check 等目标 |

> **⚠️ 测试现状（最容易误判的地方）**
>
> 仓库中**只有一个测试文件**：`internal/pkg/config/config_test.go`，它只测 `validateJWTSecret` 与 `Load` 的密钥裁剪逻辑。架构守护测试已移除（S5.5 统一补）。
>
> | 包 | 测试状态 |
> |---|---|
> | `internal/pkg/config` | ✅ 2 个测试函数 |
> | `internal/domain` | ❌ 无（架构守护已移除） |
> | `internal/service` | ❌ 无 |
> | `internal/pkg/jwt` | ❌ 无 |
> | `internal/repository` | ❌ 无 |
> | `internal/handler` / `middleware` | ❌ 无 |
>
> **`go test ./...` 会因 config 包而"通过"，但核心业务逻辑仍未被验证。** 详见 DEV-001 与 P-06。

**四项工程化能力**

| 能力 | 实现 | 实际状态 |
|---|---|---|
| 依赖注入 | wire：Provider / Bind / Injector / `wire.Struct` | ✅ 已落地，`InitializeApp` 已生成 |
| 可测试性 | 手写 fake（不用 mockgen） | 🟡 仅有 config 包的测试；service 无 fake、无测试 |
| 架构守护 | `TestDomainHasNoFrameworkDependency` | 🚧 **已移除（2026-11-03）**；仅剩 `make verify-layers` 手动检查 |
| 质量门禁 | golangci-lint + `make check` | 🟡 `.golangci.yml` 与 Makefile 已落地；本机未安装 make，底层命令已逐条验证 |

**S4 的关键教学点**

| 点 | 内容 |
|---|---|
| **接口定义在消费方** | `service` 定义 `UserRepository`/`TokenSigner`；`middleware` 定义 `TokenParser`；`handler` 定义 `Pinger`。同一个 `*jwt.Signer` 满足两个不同小接口 |
| **wire 三处 Bind 的原因** | Go 的接口是隐式的，但 wire 生成代码时必须**显式知道**用哪个实现 |
| **wire 不能拆结构体** | `jwt.NewSigner(string, time.Duration)` 的两个参数都在 `*config.Config` 里 → 必须写 `newSigner` 适配函数 |
| **不要包装 logger** | 包装会丢失 `With`/`Enabled`，且接口注定不完整。测试用 `io.Discard` 即可 |
| **测试用 `bcrypt.MinCost`** | 生产 cost 12（250ms），测试 cost 4（~1ms），否则几十个测试要跑十几秒 |
| **负 TTL 测过期** | `NewSigner(secret, -time.Minute)` 可制造过期令牌，**不必等待，也不必引入 Clock 抽象** |
| **先验证测试会失败** | 把密码改成明文存储，确认安全测试失败，再改回 |
| **实际注入器名** | 代码里是 **`InitializeApp`（导出）**，不是 `initializeApp`；`newSigner` 返回 `(*jwtpkg.Signer, error)` |

#### 2.2.1 偏离记录 DEV-001：S4 测试延后整合

> **开发者原话**：「在 S4 阶段我并没有将测试整合进项目中，目的是之后将所有测试整合进大任务中」

| 项 | 内容 |
|---|---|
| **编号** | DEV-001 |
| **发生阶段** | S4 工程化分层 |
| **偏离内容** | S4 验收要求 `go test -race ./...` 全绿、service 覆盖率 ≥60%。实际：测试**未整合进项目**，计划后续统一整合 |
| **决策理由** | 集中处理测试，减少阶段性上下文切换 |
| **决策状态** | 已确认，**接受该偏离并记录风险** |
| **确认情况** | **情况 A**：测试文件**从未写入项目**（T1 后补后删除；2026-11-03 重建后再次按用户决定移除） |
| **当前进展** | 2026-11-03 决定：`architecture_test.go` **已移除**，仓库只剩 `config_test.go`；**T1/T2/T3 统一延后到 S5.5** |

##### 后果一：`make check` 给出虚假信心（最实际的陷阱）

`go test ./...` 对**没有测试的包**会以退出码 0 通过：

```
?   github.com/IssierHM/CraftShelf/internal/service   [no test files]
```

**退出码为 0**。即使一个包完全没有测试，质量门禁也会显示"通过"。

**当前实际状态（2026-11-03 决定后）**：T1 已移除，仓库只剩 `config_test.go`；service 等核心包仍无测试。

| 包 | 测试状态 |
|---|---|
| `internal/pkg/config` | ✅ 有（2 个函数，仅测密钥校验与裁剪） |
| `internal/domain` | ❌ 无（架构守护已移除） |
| `internal/service` | ❌ 无 |
| `internal/pkg/jwt` | ❌ 无 |
| `internal/repository` | ❌ 无 |

**即 `make check` / `go test ./...` 的"通过"只说明 config 包正常，核心业务逻辑仍未被验证。**

这比"测试失败"更危险——你会误以为安全网已经齐全。

**处置**：Makefile 已重建（P-11 已修），`make test` / `make check` 会真实执行测试。但在 S5.5 前，它们只覆盖 config 包，仍会对 service 的 `[no test files]` 返回 0。

##### 后果二：架构守护已移除（风险重新出现）

`TestDomainHasNoFrameworkDependency` 是**唯一能自动拦截"领域层被污染"的机制**。约 25 行、无外部依赖、执行毫秒级，守护着整个分层契约（§3.4）。

**风险场景**：S5 会引入 `Storage` 接口与 SeaweedFS 实现（AWS SDK v2），这是领域层最容易被第三方 SDK 污染的时刻。若无守护，污染会静默发生，等问题暴露时代码量已经很大。

**当前状态（2026-11-03 用户决定）**：T1 已移除，测试统一延后到 S5.5。S5 期间**没有任何自动化机制**拦截领域层污染。

**唯一替代手段**：每次改动 `domain` 后手动执行（`make verify-layers` 也已提供）

```powershell
go list -deps ./internal/domain | Select-String "gin-gonic|jackc|google/uuid|aws"
```

输出必须为空。

##### 后果三：安全断言无保护（**仍然存在**）

S5 会扩展用户模型（头像、简介），可能触及口令与登录流程。当前**没有任何自动化机制**验证以下两条静默失效点：

| 不变量 | 被破坏时的表现 | 当前保护 |
|---|---|---|
| 口令不以明文落库 | **不报错**，数据静默泄露 | ❌ 无 |
| 登录失败不区分"邮箱不存在"与"密码错误" | **不报错**，提供邮箱枚举接口 | ❌ 无 |

这类缺陷不会被 `go build`、`golangci-lint` 或手工点测发现。

T2 对应的口令/登录风险仍在（架构污染风险见上一条），对应 §13.4.1 的 T2（约 1.5 小时）。

##### 风险控制要求（按情况 A 修正成本）

因测试文件从未写入，"提前补测试"的成本比原先估计高（需要先写 fake 与辅助函数）。以下按投入分三级，可按剩余时间选择：

| 级别 | 内容 | 成本 | 覆盖的风险 | 状态 |
|---|---|---|---|---|
| **T1 强烈建议** | `internal/domain/architecture_test.go` | ~25 行 / 15 分钟 | **S5 引入 AWS SDK 导致领域层污染**（S5 的最大架构风险） | ⬜ 已移除，S5.5 统一补 |
| **T2 建议** | T1 + `fakeUserRepo` + 两个安全断言测试 | ~200 行 / 1.5 小时 | 口令明文落库、登录失败泄露邮箱存在性 | ⬜ 待 S5.5 |
| **T3 可延后** | S4 全部测试（含 jwt、表驱动全用例） | ~400 行 / 4 小时 | 交付质量与覆盖率 | ⬜ 待 S5.5 |

**T1 已移除（2026-11-03）**：按用户要求统一测试，`internal/domain/architecture_test.go` 已删除，S5.5 再补。S5 期间只能用 `make verify-layers` 手动检查领域层依赖。

**T2 的取舍**：需要先建 fake（~80 行），但这两个断言保护的是"坏了不会报错"的静默失效点。若 S5 要动用户模型与口令流程，建议做。

**T3 的豁免**：`jwt_test.go` 与表驱动全用例不涉及 S5 的改动风险，可安全延后到 S5.5。

**测试整合任务的硬性时间约束**

延后成本随阶段数量**超线性增长**：

| 整合时机 | 待覆盖模块数 | 复杂度 |
|---|---|---|
| S4 结束（原计划） | 1（auth） | 低 |
| S5 结束 | 3（auth / work / upload） | 中 |
| S6 结束 | 5（+ version / depgraph） | 高 |
| **S7 之后** | **8（+ agent / search / moderation）** | **很高，且 Agent 的并发调度与降级分支最难测** |

> **硬性要求**：测试整合任务**不得晚于 S6 结束**。
>
> 理由：S7 的 Agent 模块含并发编排、独立超时、三级降级——这是全项目最需要测试的部分。若在它之后才建测试体系，等于**在最难的地方补最基础的保障**。

**排期决策**

| 方案 | 位置 | 评价 |
|---|---|---|
| **A** | **S5 结束后插入 S5.5** | ✅ **采用**。仅 3 个模块，且刚写完上传逻辑印象最深 |
| B | S6 结束后插入 | 可接受，但需覆盖 5 个模块 |
| C | S8 一并处理 | ❌ 超出硬性要求 |
| D | S9 收尾时 | ❌ 不可接受 |

**本方案采用 A**：新增 **S5.5「测试整合」**（见 §2.3）。S5.5 通过后，S4 状态由 🟡 改为 ✅。

### 2.3 未完成阶段明细

#### S5 分片上传与内容寻址秒传 ⬜

**周期**：第 5 周｜**优先级**：核心，不可裁剪

> **✅ 前置条件已就绪**：`internal/pkg/config/config.go` **已包含全部 S3 字段**（`S3Endpoint`/`S3Region`/`S3AccessKey`/`S3SecretKey`/`S3BucketPublic`/`S3BucketPrivate`/`S3UsePathStyle`）。旧版文档的 P-04「config 未解析 S3 字段」**已失效**。
>
> **✅ P-09 已修复（2026-11-03）**：`config.go` 字段已改为 `S3BucketWorks`/`S3BucketAvatars`，环境变量为 `APP_S3_BUCKET_WORKS`/`APP_S3_BUCKET_AVATARS`，值保持 `craftshelf-works`/`craftshelf-avatars`（与 compose 已建桶一致）。S5 可以开工。

**实现顺序（绝不跳步）**

| 步 | 内容 | 为什么这个顺序 |
|---|---|---|
| 1 | 单文件直传 | 先建立"文件如何存储"的完整认知 |
| 2 | 拆成分片：init / upload part / complete | 一次只引入一个新变量 |
| 3 | 秒传预检：内容哈希查重 | |
| 4 | 并发分片：`errgroup` + 信号量 | |
| 5 | 合并校验：流式重算哈希，不匹配则清理 | |
| 6 | 引用计数 + 延迟回收 | |

**设计要求**

| 要求 | 说明 |
|---|---|
| 秒传预检原子化 | "检查并占用引用"须在**单次原子往返**内完成（`INSERT ... ON CONFLICT ... RETURNING` + `xmax <> 0` 判断是否命中） |
| 分片接收幂等 | 客户端网络重试导致的重传是**正常现象**；分片号去重 |
| **标准 S3 Multipart** | `CreateMultipartUpload` → `UploadPart` → `CompleteMultipartUpload`；**不用 MinIO 专有的 `ComposeObject`** |
| 流式校验 | 合并后流式重算 SHA-256，与声明值比对；失败则 `AbortMultipartUpload` |
| 延迟回收 | `ref_count = 0` 的文件须经安全期（如 7 天）才清理 |
| 下载直链 | 预签名 URL，不经过应用服务器 |
| 会话生命周期 | 上传会话带过期时间，后台任务清理过期会话与临时分片 |
| **优先分片直传** | 客户端经预签名 URL 直接上传分片到对象存储，API 只协调与鉴权，**不转发字节** |

**验收清单**

- [ ] 100 MB 文件上传成功，下载后哈希一致
- [ ] 同一文件二次上传返回零字节（查日志确认未写入）
- [ ] 上传中途终止客户端，重连可续传
- [ ] 故意提交哈希不符的文件，服务端拒绝并清理残留
- [ ] 上传 1 GB 文件时进程内存保持平稳
- [ ] `go test -race` 覆盖 service 层（fake Storage）

**产物**：上传时序图 + 实测耗时与去重率数据

#### S5.5 测试整合 ⬜

**周期**：S5 结束后（**情况 A 下约 3–4 天**）｜**优先级**：**硬性检查点**，不可跳过

**背景**：S4 的测试**从未写入项目**（DEV-001 情况 A）。本阶段从零编写 S3–S5 的全部单元测试，关闭 DEV-001。

**工作量估算（情况 A）**

| 模块 | 内容 | 估算 | 状态 |
|---|---|---|---|
| `domain` | 架构守护测试 | 0.25 小时 | ⬜ 已移除，待重做 |
| `pkg/jwt` | 往返、篡改、换密钥、负 TTL 过期 | 2 小时 | ⬜ |
| `service/auth` | `fakeUserRepo`（~80 行）+ 表驱动 Register/Login + 2 个安全断言 + GetByID | 6 小时 | ⬜ |
| `service/upload` | `fakeStorage` + `fakeFileRepo` + 5–6 个场景 | 8 小时 | ⬜ |
| `service/work` | 状态机流转、权限校验 | 3 小时 | ⬜ |
| 门禁恢复 | Makefile 合并 `check-arch`/`test`、覆盖率阈值 | 1 小时 | ⬜ |
| **合计（剩余）** | | **20.25 小时 ≈ 3–4 天** | 累计完成 0 小时 |

> 原先估计"2–3 天"基于"已有测试只需跑通"。情况 A 下需从零编写，含 fake 与辅助函数，**上调为 3–4 天**。这意味着总工期约增加半周，由 S10 缓冲吸收（缓冲从 1 周压缩为约 2–3 天剩余）。

**为什么放在这里而不是更晚**

| 整合时机 | 待覆盖模块数 | 复杂度 |
|---|---|---|
| **S5 结束后** | **3（auth / work / upload）** | **中** |
| S6 结束后 | 5（+ version / depgraph） | 高 |
| S7 之后 | 8（+ agent / search / moderation） | 很高，且 Agent 最难测 |

延后成本超线性增长，且 S7 的并发编排与三级降级是全项目最需要测试的部分。

**任务清单**

1. **`internal/domain/architecture_test.go`**：`TestDomainHasNoFrameworkDependency`（2026-11-03 已移除，S5.5 重做；代码见 §9.3 方式 B）
2. **`internal/pkg/jwt/jwt_test.go`**：往返、结构损坏、篡改载荷、换密钥、**负 TTL 测过期**
3. **`internal/service/auth_test.go`**：
   - 先建 `fakeUserRepo`（内存 map + 可注入错误 + `sync.Mutex`）
   - `TestAuthService_Register` 表驱动 5 例（成功 / 邮箱重复 / 大小写重复 / 唯一约束冲突 / 仓储故障）
   - `TestAuthService_Register_StoresHashedPassword`（4 项断言：非明文、长度 60、可验证、盐生效）
   - `TestAuthService_Login` 表驱动 4 例 + `TestAuthService_Login_UniformError`
   - `TestAuthService_GetByID`
4. **`internal/service/upload_test.go`**（fake `Storage` + fake `FileRepo`，**不起 SeaweedFS**）：
   - 秒传命中返回零字节、`ref_count` +1
   - 分片重传幂等（重复分片号只记一次）
   - 哈希不符触发 `AbortMultipartUpload` 并清理残留
   - 引用计数：创建 +1、删除 -1、归零进入待回收
   - 上传会话过期清理
5. **`internal/service/work_test.go`**：创建/发布状态机流转、非作者不可修改
6. **恢复 `make check` 完整门禁**：移除临时的 `exit 1`，恢复 `go test -race`
7. **在 `Makefile` 中锁定覆盖率门槛**：

   ```makefile
   test-cover:
   	go test -race -coverprofile=coverage.out ./...
   	@go tool cover -func=coverage.out | tail -1
   	@go tool cover -func=coverage.out | awk '/total:/ {gsub(/%/,"",$$3); if ($$3+0 < 60) {print "❌ 覆盖率低于 60%"; exit 1}}'
   ```

**验收清单**

- [ ] `go test -race ./...` 全绿，且**至少 20 个测试函数**（不是 "no test files"）
- [ ] `go test -cover ./internal/service` ≥ 60%
- [ ] `make check` 在无测试时**必须失败**（验证门禁真实有效）
- [ ] 三个安全断言测试在位，且**故意破坏代码时确实失败**
- [ ] upload 测试不依赖容器（纯 fake，秒级完成）
- [ ] `TestDomainHasNoFrameworkDependency` 在位，且故意污染 domain 时失败


**产物**：`go tool cover -html` 的覆盖率报告截图

**关闭条件**：本阶段验收全部通过后，将 §2.1 中 S4 状态由 🟡 改为 ✅，并在本节标注"DEV-001 已关闭"。

#### S6 版本管理与依赖 DAG ⬜

**周期**：第 6 周｜**优先级**：核心，不可裁剪

**学习范围**：邻接表、拓扑排序、环检测与环路径还原、语义化版本范围匹配、事务与并发发布竞态。

**实现顺序**

| 步 | 内容 |
|---|---|
| 1 | 版本模型 + CRUD（不含依赖） |
| 2 | 版本号语义化比较（纯函数，易测） |
| 3 | 依赖声明（存边） |
| 4 | 拓扑排序**仅判断有无环** |
| 5 | **环路径还原** ← 本阶段核心价值点 |
| 6 | 校验纳入发布事务 |

**发布校验（同一事务内，顺序不可变）**

```
版本号唯一性
  → 依赖作品存在性 + 满足范围的已发布版本存在性
  → 全量依赖图环检测
  → 文件引用计数 +1
  → 写 work_versions + version_deps
  → 事务内投递 River 任务（embedding / 审核 / 通知订阅者）
```

任一环节失败整体回滚，**不得产生半发布状态**。

**环路径还原的价值**：仅报告"存在循环依赖"对使用者无帮助；须返回 `A@1.0 → B@2.0 → C@1.0 → A@1.0`，使作者可定位需拆解的边。

**验收清单**

- [ ] 三节点依赖链发布成功
- [ ] 双节点环被拒，错误信息可辨识完整环路
- [ ] 三节点环同样可定位
- [ ] 依赖作品不存在返回明确错误
- [ ] 发布中数据库报错，回滚后无半发布状态

**产物**：依赖图可视化

#### S7 AI Agent 模块 ⬜

**周期**：第 7 周｜**优先级**：核心，不可裁剪

**设计立场**：Agent 须齐备五要素——**编排骨架、能力契约、失败降级、成本控制、效果评测**。缺失任一即退化为 API 调用包装。

**分层**

```
Agent 运行时
├── 内核层        规划器 · 执行器 · 校验器 · 预算控制（与业务无关）
├── 模型接入层    统一接口 + 真实实现 + 模拟实现
├── 能力层        内容审核 · 语义检索 · 元数据生成
└── 规则层        降级兜底（敏感词、正则、黑白名单）
```

**三段式编排**

| 阶段 | 职责 | 关键设计 |
|---|---|---|
| 规划 | 目标 → 有序步骤 | 步骤间声明依赖，构成 DAG |
| 执行 | 按依赖调度 | 无依赖步骤并发；**每步独立超时**；失败可重试 |
| 校验 | 检查产物是否符合契约 | 不合格时**携带失败原因回炉一次**，而非直接报错 |

**三级降级（强制）**

| 级 | 触发 | 行为 |
|---|---|---|
| 1 重试 | 网络抖动 / 限流 | 指数退避 + 随机抖动；参数类错误不重试 |
| 2 回炉 | 校验不通过 | 把失败原因作为反馈重新规划一次 |
| 3 降级 | 重试耗尽 | 切规则引擎完成基本判定，任务标记**"降级"而非"失败"** |

**"降级"而非"失败"的语义意义**：任务在业务上已完成，仅质量下降。使模型不可用时系统仍可服务，且监控能区分真失败与降级。

**三项业务能力**

| 能力 | 触发 | 输入 | 输出 | 方式 |
|---|---|---|---|---|
| 内容审核 | 作品/评论提交后 | 标题、简介、封面、历史 | 风险等级、命中标签、理由 | 异步 |
| 元数据生成 | 审核通过后 | 简介与说明文档 | 推荐标签、品类、摘要 | 异步 |
| 语义检索 | 搜索请求 | 自然语言 | 检索计划与结果 | 同步 + 缓存 |

**高风险内容进入待人工复核队列——AI 仅做初筛，最终判定权保留给人。**

**成本控制（须可量化）**

- **规则前置过滤**：明确违规直接拦截，信誉良好作者直接放行，**仅中间地带送模型**
- 全局 + 单用户双层令牌桶限流
- 单任务与时间窗口开销上限，超出即降级
- 相同输入结果缓存

**检索融合**：结构化筛选走索引；关键词走全文打分；语义走向量近邻。两路分数**量纲不同不可直接比较**，用**倒数排名融合（RRF）**——仅以排名位置算综合分，避免归一化调参。

**效果评测（强制）**

- 人工标注评测集：审核约 200 条（正负均衡）、语义检索约 100 条
- 独立评测命令，支持模拟模型（CI，零费用）与真实模型（产出指标）
- 指标：准确率、召回率、F1、延迟分位数、平均开销、降级率

**评测有无决定陈述质量量级**

| 无评测 | 有评测 |
|---|---|
| "实现了 AI 内容审核功能" | "自建 200 条标注评测集，审核准确率 X%、召回率 Y%，规则前置过滤使模型调用量下降 Z%" |

**验收清单**

- [ ] 模拟模型跑通全流程，单测全绿且零费用
- [ ] 模型返回非法 JSON 时校验器捕获并触发回炉
- [ ] 断网状态下审核任务**降级**执行而非失败
- [ ] 自然语言检索"能配合光影使用的建筑模组"返回合理结果
- [ ] 评测命令输出全部指标

#### S8 可观测性与部署 ⬜

**周期**：第 8 周｜**优先级**：**最先可裁剪**

**任务**

1. 通用指标（请求量、延迟、错误率）+ **业务指标**（秒传命中数、上传字节总量、降级率、模型调用量）
2. 链路标识自 HTTP 入口透传至 Agent 执行
3. 容器化集成测试（testcontainers）
4. 压测获取真实数据
5. 部署云服务器 + 反向代理

**缓存一致性要求**：旁路缓存。读未命中查库回填；写操作**先写库再删缓存**，不得直接更新缓存。配合请求合并防击穿、空值占位防穿透。

**验收**

- [ ] 指标接口可查见自定义业务指标
- [ ] 单次请求日志可追踪其 Agent 内执行步骤
- [ ] 集成测试可在干净环境从零跑通
- [ ] 有实测压测数据（测量而非估算）
- [ ] 公网地址可访问

#### S9 收尾与陈述 ⬜

**产物**

- `README.md`：一句话简介、架构图、一键启动、演示动图
- 架构决策文档：每项选择记录理由与被放弃方案
- **3 分钟演示视频**：秒传 → 发布带依赖的版本 → 触发审核 → 自然语言检索
- 面试口径成稿

#### S10 缓冲 ⬜

**不可省略。** 前序估算必然因环境问题、概念偏差、疑难缺陷而偏移。缓冲用于吸收该偏差。

### 2.4 待核实与已知缺口

> **本节是接手者的第一站。** 所有与"文档理想状态"不同的地方都在这里，且已按 2026-11-03 的代码核对。

#### 2.4.1 阻断性缺陷（已于 2026-11-03 全部修复）

| 编号 | 内容 | 证据 | 处置 |
|---|---|---|---|
| **P-09** | **桶名不一致**：`config.go` 默认值 与 `.env` 实际值都是 `craftshelf-public`/`craftshelf-private`；而 `docker-compose.yml` 初始化容器创建的桶是 `craftshelf-works`/`craftshelf-avatars` | `config.go:46-47`、`.env`、`docker-compose.yml:88-89` | ✅ **已修复（方案 A）**：字段改为 `S3BucketWorks`/`S3BucketAvatars`，变量改为 `APP_S3_BUCKET_WORKS`/`APP_S3_BUCKET_AVATARS`，值保持 `craftshelf-works`/`craftshelf-avatars`；compose 未改、桶未重建 |
| **P-10** | **`AuthService.Login` nil 指针 panic**：若 `FindByEmailForAuth` 返回的错误**不是** `ErrUserNotFound`（如数据库故障），代码不 return，继续访问 `user.PasswordHash`，而 `user` 为 nil | `service/auth.go:66-72` | ✅ **已修复**：`err != nil` 时先判 `ErrUserNotFound`，其余错误直接返回；临时单测验证不再 panic |
| **P-11** | **`Makefile` 为 0 字节空文件**：所有 `make` 目标（迁移、生成、测试、门禁）均不存在 | 文件大小 0 | ✅ **已修复**：Makefile 约 2.3 KB，含 build/test/lint/generate/migrate/check 等目标 |

> **结论**：三项阻断项已关闭，S5 可开工。验证命令与输出见 §2.5.1。

**P-09 修复记录（2026-11-03）**

采用方案 A（改名 + 改值）。理由：S5 的 `Storage` 代码尚未编写，当时只有 `config.go` 与 `.env`/`.env.example` 引用这两个字段，是最便宜的重命名时机；同时消除"`PUBLIC` 变量指向头像桶"的语义错配。

```dotenv
# .env
APP_S3_BUCKET_WORKS=craftshelf-works
APP_S3_BUCKET_AVATARS=craftshelf-avatars
```

```go
// internal/pkg/config/config.go
S3BucketWorks:   getEnv("APP_S3_BUCKET_WORKS",   "craftshelf-works"),
S3BucketAvatars: getEnv("APP_S3_BUCKET_AVATARS", "craftshelf-avatars"),
```

`docker-compose.yml` 初始化容器保持创建 `craftshelf-works` / `craftshelf-avatars`，**未反向改 compose、未重建桶**。

**为什么按对象分桶（works/avatars）比按访问级别分桶（public/private）在本项目更合适**

存储层不需要根据 `visibility` 字段做"该进哪个桶"的路由分支。公开作品与私有作品的文件同在 `works` 桶，访问控制统一由预签名 URL 的 TTL 完成。**按访问级别分桶反而多一层耦合。**

**将来需要按级别分离时**：新增桶的成本远低于重命名。所以现在选什么都不锁死。

#### 2.4.2 工程质量缺口

| 编号 | 内容 | 证据 | 处置 |
|---|---|---|---|
| **P-12** | `.golangci.yml` **不存在**，静态检查无配置 | 目录中无此文件 | ✅ **已修复**：创建 v2 格式配置；因 v1.64.8 在 Go 1.26 下 `buildssa` panic，工具已升级到 **v2.15.0**，`go tool golangci-lint run ./...` = `0 issues.` |
| **P-13** | `.env.example` **不存在**，只有未跟踪的 `.env` | 顶层文件列表 | ✅ **已修复**：创建 `.env.example`，只含变量名与占位值 |
| **P-14** | `domain.User` **带 json tag 且含 `password_hash`** | `domain/user.go:20` | ✅ **已修复**：移除 domain 的全部 json tag，序列化统一走 DTO |
| **P-15** | `app.go`/`providers.go`/`wire.go`/`wire_gen.go`/`DEVELOPMENT.md` **均未提交 git**；`main.go` 有未提交修改 | `git status --short` | ✅ **已修复**：2026-11-03 提交；`wire_gen.go` 已入库 |
| **P-16** | `godotenv` 被 `main.go` 直接 import，但在 `go.mod` 中标记为 `// indirect` | `main.go:14` vs `go.mod:143` | ✅ **已修复**：`go mod tidy` 后成为直接依赖 |
| **P-17** | `google/uuid` 被放入 `go.mod` 的 `tool` 块（它是运行库不是工具） | `go.mod:5-11` | ✅ **已修复**：从 `tool` 块移除；仍保留在 `require`（`sqlc/models.go` 使用） |
| **P-18** | 仓库根目录有构建产物 `api.exe` | 文件列表 | ✅ **已修复**：`api.exe` 已删除（`.gitignore` 仍覆盖 `*.exe`） |

#### 2.4.3 环境与版本待核实

| 编号 | 类型 | 内容 | 处置 |
|---|---|---|---|
| P-01 | ⚠️ 待核实 | SeaweedFS 镜像 tag | 现状：`chrislusf/seaweedfs:4.51`（旧命名空间但版本真实存在）。安装时用 `docker manifest inspect` 复核 |
| P-02 | ⚠️ 待核实 | `s3.json` 格式 | 已按官方 identity 结构编写，首次启动如报错请对照官方 wiki |
| P-03 | ⚠️ 已确定 | Redis 镜像 | 现状 `redis:8-alpine`，已落地 |
| P-05 | 🚧 已知缺口 | `APP_REDIS_ADDR` 已解析但 Redis 未被使用 | S7 接入缓存/限流时启用，当前无害 |
| P-06 | 🚧 已知缺口 | 技能文档 `.dsh/skills/forgehub-go-project/` 仍为旧名（ForgeHub）与旧选型（MinIO、minio-go、ComposeObject） | 需同步更新为 CraftShelf + SeaweedFS + AWS SDK + Multipart |
| P-08 | ✅ 已关闭 | `golangci-lint` 版本与配置格式 | v1.64.8 在 Go 1.26 下 `buildssa` panic；已升级 **v2.15.0**，`.golangci.yml` 已迁移为 v2 格式，lint 为 0 issues |
| P-19 | 🚧 环境限制 | 本机 `CGO_ENABLED=0` 且 PATH 中无 gcc，`go test -race` 报 `-race requires cgo` | S5.5 前安装 MinGW-w64 并设 `CGO_ENABLED=1`；或改在 CI/Linux 上跑 race。`go test ./...` 与 `go vet` 当前正常 |

> **P-04 / P-07 已关闭**：
> - P-04（config 未解析 S3 字段）→ **已解决**，`config.go` 已含全部 S3 字段
> - P-07（版本号硬编码易过期）→ 已在 §8.5 改为"策略 + 核实命令"

### 2.5 文档修正记录（v1.5）

#### v1.3 → v1.4

本节记录 v1.3 → v1.4 因"与代码不符"而做的全部修正。**目的是防止下一位读者再被旧信息误导**，也是"文档必须服从代码"的落地证据。

| # | 位置 | v1.3 错误内容 | 实际代码 | 影响 |
|---|---|---|---|---|
| 1 | §2.2 S3 文件清单 | `migrations/00001_create_users.sql` | **`20261027151811_create_works.sql`**（日期时间戳命名） | 高：找不到迁移文件 |
| 2 | §2.2 / §6.1 / §9 | 约束名 `users_email_key`、`users_username_key`、`users_public_id_key`、`users_role_check` | **`user_email_key`、`user_username_key`、`user_public_id_key`、`user_role_check`**（单数） | **高：按文档改迁移会导致 `mapUniqueViolation` 全部失效** |
| 3 | §2.2 S4 | `architecture_test.go`「已落地并通过」 | **文件已删除** | 高：误以为有架构守护 |
| 4 | §2.2 S4 | 测试覆盖表只列 architecture_test | 实际唯一测试是 **`config_test.go`**（文档从未提及） | 高：漏报真实测试 |
| 5 | §2.2 S4 | `.golangci.yml`「已交付」 | **不存在** | 中 |
| 6 | §2.2 S4 / §13.2 | `Makefile`「统一入口，含 test/lint/verify-layers/wire」 | **0 字节空文件** | 高：`make` 全部失败 |
| 7 | §2.2 S3 | `.env.example`「已交付」 | **不存在** | 中 |
| 8 | §2.1 | S1/S2「✅ 已交付，完整可运行」 | **仓库中无对应产物**（独立练习） | 中：会去仓库里找不存在的东西 |
| 9 | §5.2 | `jwt.NewSigner(...) *Signer` | **`(*Signer, error)`**，且内部校验密钥 | 高：签名不符编译失败 |
| 10 | §5.2 | 错误变量 `ErrInvalidToken` | **`ErrorInvalidToken`** | 中 |
| 11 | §4.1 / §5.3 | 模块路径 `github.com/<user>/craftshelf` | **`github.com/IssierHM/CraftShelf`** | 高：import 全错 |
| 12 | §4 / §13.2 | Go 1.22+ | **`go 1.26.0`** | 低 |
| 13 | §8.1 | `APP_S3_BUCKET_WORKS` / `APP_S3_BUCKET_AVATARS` | **`APP_S3_BUCKET_PUBLIC` / `APP_S3_BUCKET_PRIVATE`** | 高：环境变量名不符 |
| 14 | §5.5 错误码表 | 参数错误 `invalid_argument` | **`invalid_request`** | 中 |
| 15 | §8.5 | 工具用 Makefile 或裸 `go install` | **`go.mod` 的 `tool` 指令** | 中 |
| 16 | §9.3 分层验证 | 提供"自动化测试"方式 | 该测试**已删**，只剩命令行方式 | 高 |
| 17 | §2.2 S4 | 注入器名 `initializeApp` | **`InitializeApp`（导出）** | 低 |
| 18 | 全文 | 日志级别 `LevelInfo` | **`LevelDebug`** | 低 |
| 19 | §2.2 S3 | 未提及 `.env` 自动加载 | `main.go` 用 **`godotenv`** 加载 | 中 |
| 20 | §6.1 | 迁移 SQL 原文 | 实际用 `TIMESTAMP WITH TIME ZONE`、`LOWER(email)`，role CHECK 顺序为 `'user','admin','moderator'` | 低（语义等价） |
| **21** | **§2.4 P-09 处置** | **"推荐改 compose 与 config 对齐到 `public`/`private`，语义更贴合"** | **应为：改 `.env` 与 `config.go` 对齐到 compose 已建的 `works`/`avatars`** | **中：按原建议需重建桶（无谓成本）** |

> **第 21 条的特殊性**：前 20 条是"文档与代码不符"，这一条是**"建议本身错了"**——我按语义偏好而非成本做决策，忽略了迁移成本。教训：**处置建议必须先算成本，再谈语义**。已定的判据是"能让已建好的设施继续用的方案优先"。

**修正原则（此后必须遵守）**

1. **文档只写"已核实的事实"**，不写"设计意图"当作事实
2. **写 ✅ 前必须实际看到文件**——`git ls-files` 或读文件
3. **区分三件事**：设计（可能未实现）／代码（可能未提交）／测试（可能已删）
4. **任何改动后**：先改代码 → 再改文档 → 在 §2.5 登记

#### v1.4 → v1.5（2026-11-03，S5 开工前修复）

| # | 位置 | v1.4 状态 | v1.5 实际状态 | 验证方式 |
|---|---|---|---|---|
| 1 | §2.4.1 P-09 | 桶名不一致（config/env 为 public/private，compose 为 works/avatars） | `config.go` 字段改为 `S3BucketWorks`/`S3BucketAvatars`，env 改为 `APP_S3_BUCKET_WORKS`/`APP_S3_BUCKET_AVATARS`，默认值与 `.env` 均为 works/avatars；compose 未改 | `s3.bucket.list` 仅见 works/avatars |
| 2 | §2.4.1 P-10 | `Login` 在非 `ErrUserNotFound` 错误时 nil panic | `err != nil` 时先判 `ErrUserNotFound`，其余错误直接返回 | 临时单测（仓储返回 `database unavailable`）通过后删除 |
| 3 | §2.4.1 P-11 | `Makefile` 0 字节 | 约 2.3 KB，含 build/test/test-race/lint/generate/migrate/verify-layers/check | 逐条执行底层 `go` 命令；本机未安装 make |
| 4 | §2.4.2 P-12 | `.golangci.yml` 不存在 | v2 格式配置已创建；golangci-lint 由 v1.64.8 升级到 **v2.15.0** | `go tool golangci-lint run ./...` → `0 issues.` |
| 5 | §2.4.2 P-13 | `.env.example` 不存在 | 已创建，仅含占位值 | 文件存在，无真实密钥 |
| 6 | §2.4.2 P-14 | `domain.User` 含 json tag 与 `password_hash` | 全部 json tag 已移除 | 代码核对；handler 仍走 DTO |
| 7 | §2.4.2 P-15 | 5 个文件未提交 | 已提交（含 `wire_gen.go`） | `git status --short` 为空 |
| 8 | §2.4.2 P-16 | `godotenv` 为 `// indirect` | `go mod tidy` 后为直接依赖 | `go.mod` require 块 |
| 9 | §2.4.2 P-17 | `google/uuid` 在 `tool` 块 | 已从 `tool` 块移除，仍在 `require`（`sqlc/models.go` 使用） | `go.mod` 核对 |
| 10 | §2.4.2 P-18 | 根目录有 `api.exe` | 已删除 | 文件不存在 |
| 11 | §9.3 T1 | 架构守护测试已删除 | 重建后按用户决定再次移除；S5.5 统一补 | 本轮未保留自动化守护 |
| 12 | 阅读前基线 | 唯一测试为 config（2 个函数） | 仍为 1 个测试文件（config 2 个函数） | `go test ./...` |
| 13 | §8.1 | 变量名 `APP_S3_BUCKET_PUBLIC`/`PRIVATE` | 改为 `APP_S3_BUCKET_WORKS`/`AVATARS`（v1.3→v1.4 第 13 条是当时事实，v1.5 已再次变更） | `config.go` 与 `.env.example` |
| 14 | §9.5 / §13.2 | golangci-lint 版本待核实 | v1.64.8 在 Go 1.26 下 `buildssa` panic；升级 v2.15.0 并迁移配置 | `--version` + `run ./...` |
| 15 | §2.4.3 P-19 | race 未验证 | 本机无 CGO/gcc，`-race` 无法运行；普通测试与 vet 通过 | `go env CGO_ENABLED` = 0；`go test ./...` 全绿 |

**v1.5 验证命令与实测结果**

```powershell
go build ./...                    # 无输出
go vet ./...                      # 无输出
go test ./...                     # config ok；其余 [no test files]
go tool golangci-lint run ./...   # 0 issues.
go tool wire ./cmd/api            # wrote cmd/api/wire_gen.go
go tool sqlc generate             # 无输出，无 diff
go tool goose -dir migrations status
# 20261027151811_create_works.sql 已应用
docker exec craftshelf-seaweedfs sh -c "echo 's3.bucket.list' | weed shell -master=localhost:9333"
# craftshelf-works、craftshelf-avatars

# 冒烟：go run ./cmd/api 后
Invoke-WebRequest http://localhost:8080/healthz   # {"status":"alive"}
Invoke-WebRequest http://localhost:8080/readyz    # {"status":"ready"}
# register → login → /users/me 均返回 200/201，响应不含 password_hash
```

> **注意 1**：本机（Windows）未安装 `make`，因此 `make check` 未直接执行；Makefile 中的每条底层命令已单独运行并通过。
>
> **注意 2**：本机 `CGO_ENABLED=0` 且无 gcc，`go test -race ./...` 报 `-race requires cgo`（P-19）。S5.5 前需补 C 工具链，或在 CI/Linux 上执行 race。

#### v1.5.1（2026-11-03，用户决定）

- 移除 `internal/domain/architecture_test.go`；T1/T2/T3 统一延后到 S5.5（DEV-001）。
- `make verify-layers` 保留为 S5 期间唯一的分层检查手段；每次改动 `domain` 后必须手动执行。
- 影响：S5 引入 AWS SDK 时，领域层污染不会被 `go test` 自动拦截。

---

## 3. 整体架构

### 3.1 部署形态

```
                        ┌──────────────────────────────┐
   Browser / CLI ────►  │  Nginx   限流 · TLS · 静态   │
                        └──────────────┬───────────────┘
                                       │
                     ┌─────────────────▼──────────────────┐
                     │      craftshelf-api  (Go / Gin)     │
                     │  中间件 → 接口层 → 业务层 → 数据访问  │
                     └──┬───────────┬───────────┬─────────┘
                        │           │           │
          ┌─────────────▼──┐  ┌─────▼─────┐  ┌──▼──────────────┐
          │ PostgreSQL 16  │  │  Redis 8  │  │  SeaweedFS (S3) │
          │ ─ 业务关系表    │  │ ─ 缓存    │  │ ─ 作品文件       │
          │ ─ JSONB 属性    │  │ ─ 限流    │  │ ─ 封面/头像      │
          │ ─ 全文检索      │  │ ─ 跨实例  │  │ ─ 分片暂存        │
          │ ─ 向量检索      │  └───────────┘  └─────────────────┘
          │ ─ 任务队列      │
          └──────┬──────────┘
                 │  Worker 轮询同一队列表
      ┌──────────▼──────────────────────────────┐
      │   craftshelf-worker  (独立进程)          │
      │   Agent 运行时：规划 → 执行 → 校验        │
      └─────────────────────────────────────────┘
```

### 3.2 PostgreSQL 的六种角色

本方案最核心的架构判断。评审时须能完整陈述收益与代价。

| 角色 | 承载 |
|---|---|
| 关系型业务库 | 用户、作品、版本、评论、互动 |
| 文档型属性存储 | 品类扩展属性（JSONB + GIN 索引） |
| 全文检索引擎 | `tsvector` + `ts_rank` |
| 向量检索引擎 | pgvector + HNSW 索引 |
| 任务队列 | River 队列表与业务表同库 |
| 统计与聚合 | 窗口函数、CTE、`generate_series` |

**收益**

- 本地依赖仅三个容器
- 不存在"业务库与检索库不同步"的一致性缺口
- 异步任务与业务写入可置于同一事务，获得**原子性**

**代价（须主动说明）**

单一存储承载多种负载，规模增长后需按读写特征拆分。当前量级下集中带来的简单性显著大于分散带来的灵活性，此为**有意识的取舍**。

### 3.3 为什么是单体分层，且只拆一个 Worker

| 维度 | 单体分层 | 微服务 |
|---|---|---|
| 可讲透的深度 | 分层边界、事务、依赖注入 | 分布式事务、服务发现、链路追踪 |
| 引入故障面 | 低 | 高 |
| 周期内交付完整度 | 高 | 通常为半成品 |

**唯一允许的进程级拆分是 Agent Worker。**

依据：Agent 任务为长耗时、可重试、以等待外部模型响应为主的负载，与短平快的 HTTP 请求解耦更合理；两者**共享同一份领域模型代码**，因而不产生跨服务数据一致性问题。

**陈述要求**：须同时说明"为何拆这一个"与"为何不多拆"。后者是评审重点。

### 3.4 分层契约（强制）

| 层 | 包 | 职责 | 可以 import | 禁止 import |
|---|---|---|---|---|
| 领域 | `internal/domain` | 业务模型、纯逻辑、哨兵错误 | 仅标准库 | **任何第三方、任何本项目其他包** |
| 基础设施 | `internal/pkg/*` | 配置、连接池、令牌、缓存、存储封装 | 各自所需第三方 | `service`、`handler`、`repository` |
| 持久化 | `internal/repository` | SQL、行映射、错误翻译 | `domain`、sqlc、pgx | gin、service、handler |
| 业务 | `internal/service` | 流程编排、事务边界、安全策略 | `domain`、标准库、bcrypt | **pgx、gin、repository** |
| 接口 | `internal/handler` | 路由、参数校验、响应组装、错误映射 | `service`、`domain`、`middleware`、`dto`、gin | **pgx、repository、sqlc** |
| 中间件 | `internal/middleware` | 鉴权、恢复、追踪、限流 | `dto`、标准库、gin | service、repository |

**领域层零框架依赖是架构生命线。**

实际收益：对象存储、任务队列、时间源、模型接入均可替换而业务代码不变。

**可机械验证**（两种方式，都须通过）：

```powershell
# 方式一：命令行
go list -deps ./internal/domain | Select-String "gin-gonic|jackc|google/uuid|sqlc"
# 输出必须为空

# 方式二：自动化测试（推荐，已实现）
go test ./internal/domain -run TestDomainHasNoFrameworkDependency
```

**任何违反此约束的变更视为架构违规，须回退而非"暂时放行"。**

### 3.5 依赖方向图

```
                        cmd/api
         ┌───────────────┬──────────────┬──────────────┐
         ▼               ▼              ▼              ▼
    handler          pkg/config    pkg/db        repository
    │   │               │              │              │
    │   ├──► middleware │              └──────► domain
    │   │       └──► dto                     ▲
    │   └──► dto ──► domain                   │
    │                                          │
    └──► service ──────────────────────────────┘
              │
              └──► domain（含接口定义处）

    pkg/jwt  ◄── service（通过 TokenSigner 接口）
             ◄── middleware（通过 TokenParser 接口）
```

**关键：`service` 只依赖"接口"，不依赖 `repository` 包。** 所以 `service` 的代码里看不到任何 `pgx`。

### 3.6 运行时调用链（以注册为例）

```
POST /api/v1/auth/register
  │
  ▼ gin 路由命中 handler.NewRouter 中注册的 d.Auth.Register
  ▼ middleware.Recovery 的 defer 就绪（捕获后续 panic）
  ▼ gin.Logger() 记录访问日志
  │
AuthHandler.Register(c)                       internal/handler/auth.go
  │ c.ShouldBindJSON(&req) 校验格式
  │ 失败 → respondError(400, "invalid_argument")
  ▼
AuthService.Register(ctx, email, username, password)   internal/service/auth.go
  │
  ├─► users.ExistsByEmail(ctx, email)          ← 接口调用
  │      ▼ UserRepo.ExistsByEmail → q.ExistsUserByEmail → SQL
  │
  ├─ 已存在 → domain.ErrEmailExists
  ├─► bcrypt.GenerateFromPassword(password, cost)   ≈250ms
  └─► users.Create(ctx, email, username, hash)
         ▼ UserRepo.Create → q.CreateUser → INSERT ... RETURNING *
           违反唯一约束 → mapUniqueViolation → domain.ErrEmailExists
           成功 → toDomainUser(row) → *domain.User
  │
  ▼ 回到 handler
respondData(c, 201, dto.ToUserResponse(u))     ← DTO 丢弃 PasswordHash
  │
  ▼ 写回 HTTP 响应
```

**错误分支**：

```
AuthService 返回 domain.ErrEmailExists
  ▼ mapDomainError → (409, "email_exists", "该邮箱已被注册")
  ▼ respondError → 不记 Error 日志（4xx 不是故障）
```

---

## 4. 模块清单与职责

### 4.1 完整目录结构（含状态）

> ✅ = **已按实际代码核实存在**；🚧 = 存在但有缺陷；❌ = **文档曾声称存在，实际不存在**；⬜ = 未开始

```
CraftShelf/                          module: github.com/IssierHM/CraftShelf  (go 1.26.0)
├── cmd/
│   ├── api/
│   │   ├── main.go                 ✅ 生命周期管理 + godotenv（已提交）
│   │   ├── app.go                  ✅ App 结构体（已提交）
│   │   ├── providers.go            ✅ wire 适配函数（已提交）
│   │   ├── wire.go                 ✅ 注入器 InitializeApp（wireinject）（已提交）
│   │   └── wire_gen.go             ✅ wire 生成（!wireinject）（已提交）
│   ├── worker/main.go              ⬜ S7 Agent 消费者
│   ├── migrate/main.go             ⬜ S8 可选：迁移执行器
│   └── eval/main.go                ⬜ S7 Agent 评测命令
│
├── internal/
│   ├── domain/                     【领域层 · 应为零框架依赖】
│   │   ├── user.go                 ✅ User, Role（json tag 已移除，P-14）
│   │   ├── errors.go               ✅ 4 个哨兵错误
│   │   ├── architecture_test.go    ❌ 已移除（S5.5 补）
│   │   ├── work.go                 ⬜ S5
│   │   ├── version.go              ⬜ S6
│   │   ├── file.go                 ⬜ S5
│   │   ├── storage.go              ⬜ S5 对象存储接口
│   │   ├── queue.go                ⬜ S7 任务队列接口
│   │   ├── clock.go                ⬜ S5 时间源接口
│   │   ├── semver/                 ⬜ S6 版本号解析与范围匹配
│   │   ├── depgraph/               ⬜ S6 拓扑排序 + 环还原
│   │   └── schema/                 ⬜ S5 品类属性注册表
│   │
│   ├── pkg/                        【基础设施】
│   │   ├── config/config.go        ✅ 环境变量 → 结构体（**已含全部 S3 字段**）
│   │   ├── config/config_test.go   ✅ config 包测试（2 个函数）
│   │   ├── db/pool.go              ✅ pgxpool
│   │   ├── jwt/jwt.go              ✅ Signer（NewSigner 返回 error）
│   │   ├── jwt/jwt_test.go         ❌ 不存在
│   │   ├── cache/                  ⬜ S7 Redis 封装 + singleflight
│   │   ├── queue/                  ⬜ S7 Queue 接口 + 各实现
│   │   ├── storage/                ⬜ S5 Storage 接口 + 实现
│   │   ├── hash/                   ⬜ S5 流式 SHA-256
│   │   └── observability/          ⬜ S8 prometheus + otel
│   │
│   ├── repository/                 【持久化】
│   │   ├── queries/users.sql       ✅ 4 条查询
│   │   ├── sqlc/                   ✅ 生成（sqlc v1.31.1，永不手改）
│   │   ├── user.go                 ✅ UserRepo + 错误翻译（约束名 user_*）
│   │   ├── queries/works.sql       ⬜ S5
│   │   ├── queries/versions.sql    ⬜ S6
│   │   ├── work.go                 ⬜ S5
│   │   └── version.go              ⬜ S6
│   │
│   ├── service/                    【业务编排】
│   │   ├── auth.go                 ✅ AuthService + 接口定义（P-10 已修复）
│   │   ├── auth_test.go            ❌ 不存在
│   │   ├── work.go                 ⬜ S5
│   │   ├── upload.go               ⬜ S5
│   │   ├── version.go              ⬜ S6
│   │   └── search.go               ⬜ S7
│   │
│   ├── handler/                    【接口层】
│   │   ├── dto/common.go           ✅ Response, Error
│   │   ├── dto/user.go             ✅ 请求/响应 DTO
│   │   ├── respond.go              ✅ 统一响应助手
│   │   ├── errors.go               ✅ 领域错误 → HTTP
│   │   ├── auth.go                 ✅ AuthHandler（错误码 invalid_request）
│   │   ├── user.go                 ✅ UserHandler
│   │   ├── health.go               ✅ HealthHandler
│   │   ├── router.go               ✅ NewRouter + RouterDeps
│   │   └── sse.go                  ⬜ S7 SSE 推送
│   │
│   ├── middleware/                 【横切】
│   │   ├── auth.go                 ✅ Auth + UserIDFrom + TokenParser
│   │   ├── recovery.go             ✅ panic 恢复
│   │   ├── trace.go                ⬜ S8 链路标识
│   │   └── ratelimit.go            ⬜ S7 令牌桶
│   │
│   ├── agent/                      ⬜ S7
│   │   ├── core/                   规划器/执行器/校验器/预算
│   │   ├── llm/                    统一接口 + 真实 + 模拟
│   │   ├── moderation/             内容审核
│   │   ├── search/                 语义检索
│   │   ├── metadata/               元数据生成
│   │   └── rules/                  降级规则引擎
│   │
│   └── realtime/                   ⬜ S7 SSE Hub
│
├── migrations/
│   └── 20261027151811_create_works.sql   ✅ 唯一迁移（建 users 表，文件名与内容不符）
├── deploy/
│   ├── docker-compose.yml          ✅ 三服务（SeaweedFS 用 chrislusf/seaweedfs:4.51）
│   ├── postgres/init/
│   │   └── 01_extensions.sql       ✅ vector + pg_trgm + uuid-ossp
│   ├── seaweedfs/s3.json           ✅ 单身份凭据
│   └── nginx/                      ⬜ S8
├── testdata/eval/                  ⬜ S7 评测集
├── docs/                           ⬜ S9
├── go.mod / go.sum                 ✅ tool 块仅保留工具；godotenv 为直接依赖（P-16/P-17）
├── sqlc.yaml                       ✅ database.uri 模式
├── DEVELOPMENT.md                  ✅ 本文档（已提交）
├── README.md                       🚧 **0 字节空文件**（S9 补）
├── .env                            ✅ 存在（未跟踪，正确）
├── .env.example                    ✅ 环境变量模板（P-13）
├── .gitignore                      ✅ 存在（仅 .env / bin / *.exe / *.test / *.out）
├── .golangci.yml                   ✅ v2 格式（P-12）
└── Makefile                        ✅ build/test/lint/generate/check 等目标（P-11）
```

### 4.2 模块职责

| 模块 | 职责 | 不该做的事 |
|---|---|---|
| `domain` | 定义业务对象、纯逻辑（版本比较、图算法）、哨兵错误 | **不 import 任何框架**；不做 IO |
| `pkg/config` | 环境变量 → 结构体；缺失必需项即报错 | 不被下层直接调用（只由 main 读） |
| `pkg/db` | 连接池构造与配置；启动即验证连通 | 不含业务查询 |
| `pkg/jwt` | 令牌签发与解析；算法校验 | 不含业务规则 |
| `repository` | SQL 执行、行 → 领域对象映射、**数据库错误 → 领域错误翻译** | 不含业务判断；不返回 sqlc 类型 |
| `service` | 流程编排、事务边界、安全策略（哈希、统一错误）、**定义所需接口** | 不 import pgx/gin；不做 SQL |
| `handler` | 解析请求、校验格式、调用服务、组装响应、**领域错误 → HTTP 状态** | 不含业务逻辑；不直接访问数据库 |
| `middleware` | 鉴权、panic 恢复、追踪、限流 | 不含业务逻辑 |
| `agent` | 目标 → 计划 → 执行 → 校验；降级 | 与业务解耦（通过 Tool 注册接入） |

### 4.3 三层数据形态（强制）

```
HTTP JSON
   ↓ 解析
 DTO（internal/handler/dto）
   ↓ 转换（显式手写）
domain（internal/domain）
   ↓ 映射（repository 内）
 Row（internal/repository/sqlc，自动生成）
   ↓
PostgreSQL
```

| 形态 | 位置 | 强制约束 | 实际状态 |
|---|---|---|---|
| **DTO** | `handler/dto` | **绝不含敏感字段**（PasswordHash） | ✅ `UserResponse` 确实无 PasswordHash |
| **domain** | `domain` | **零第三方依赖**；可空列折叠为零值 | 🚧 **带 json tag 且含 password_hash**（P-14） |
| **Row** | `repository/sqlc` | **不手改**；可能含 `uuid.UUID`、`*string` | ✅ `sqlc.User` 中 `PublicID` 是 `uuid.UUID`，`Bio`/`AvatarUrl` 是 `*string` |

**混用的后果**

| 混用 | 后果 |
|---|---|
| 把 Row 当响应返回 | `password_hash` 泄漏到 API |
| 把 domain 传给 sqlc | `uuid.UUID` 等第三方类型渗入领域层，架构检查失败 |
| 把 DTO 传进 service | 业务逻辑与 HTTP 耦合，换协议要重写业务 |

**映射方向**

- `repository` 内：`sqlc.User` → `domain.User`（`uuid.UUID` → `string`，`pgx.ErrNoRows` → `domain.ErrUserNotFound`）
- `handler` 内：`domain.User` → `dto.UserResponse`（**丢弃 PasswordHash**）

---

## 5. 接口与契约

### 5.1 已实现的领域层对象

#### `domain.User`

```go
type User struct {
    ID           int64
    PublicID     string      // UUID 的字符串形式，避免领域层依赖 google/uuid
    Email        string
    Username     string
    PasswordHash string      // 敏感：绝不进入 DTO
    Bio          string      // 可空列折叠为 ""
    AvatarURL    string
    Role         Role
    CreatedAt    time.Time
    UpdatedAt    time.Time
}
```

#### `domain.Role`

```go
type Role string
const (
    RoleUser      Role = "user"
    RoleModerator Role = "moderator"
    RoleAdmin     Role = "admin"
)
```

#### 哨兵错误

```go
var (
    ErrUserNotFound       = errors.New("user not found")
    ErrEmailExists        = errors.New("email already registered")
    ErrUsernameExists     = errors.New("username already taken")
    ErrInvalidCredentials = errors.New("invalid credentials")
)
```

**约定**：所有领域错误用 `errors.New`，以 `Err` 前缀命名，由接口层用 `errors.Is` 判断并映射为 HTTP 状态。领域层不认识 `net/http`。

### 5.2 消费方定义的接口（已实现）

| 接口 | 定义于 | 方法 | 由谁实现 |
|---|---|---|---|
| `service.UserRepository` | `service/auth.go` | `Create` / `FindByEmailForAuth` / `FindByID` / `ExistsByEmail` | `*repository.UserRepo` |
| `service.TokenSigner` | `service/auth.go` | `Sign(userID int64) (string, error)` | `*jwt.Signer` |
| `middleware.TokenParser` | `middleware/auth.go` | `Parse(token string) (int64, error)` | `*jwt.Signer` |
| `handler.Pinger` | `handler/health.go` | `Ping(ctx context.Context) error` | `*pgxpool.Pool` |

**Go 接口是隐式的**：实现方不需要 `implements`，也不需要知道谁在用它。`*jwt.Signer` 同时满足两个不同的小接口，这是接口最小化的直接收益。

**`jwt.Signer` 的实际签名（与旧文档不同，务必以代码为准）**

```go
// 注意：返回 (*Signer, error)，且内部校验密钥
func NewSigner(secret string, ttl time.Duration) (*Signer, error)
func (s *Signer) Sign(userID int64) (string, error)
func (s *Signer) Parse(token string) (int64, error)

// 错误变量名是 ErrorInvalidToken（不是 ErrInvalidToken）
var ErrorInvalidToken = errors.New("invalid token")
```

**密钥校验是双层的**：

| 层 | 校验内容 |
|---|---|
| `config.validateJWTSecret` | 非空、≥32 字节、无空白/控制字符、**不等于默认占位符** |
| `jwt.validateSecret` | ≥32 字节、无空白/控制字符 |

这是**有意的冗余**：`jwt` 包不能假定调用方一定先做了校验。

**为什么接口定义在消费方**

| 定义位置 | 效果 |
|---|---|
| 实现方（`repository` 定义 `UserRepository`） | 实现方被迫预测所有使用方式，接口会变胖 |
| **消费方（`service` 定义）** | 接口**恰好是使用所需的最小集合**，且实现方无需感知 |

### 5.3 待实现的接口（S5 / S7）

> **接口放置位置待定（S5 开工时决定）**
>
> ADR-011 规定"接口定义在消费方"。但 `Storage` 预计会被 **upload / work 两个 service 共用**，此时有两种做法：
>
> | 方案 | 适用 | 评价 |
> |---|---|---|
> | 定义在 `service` 包（消费方） | 只有一个消费方 | 符合 ADR-011 |
> | 定义在 `domain` 包 | **多个 service 共用** | 可接受，`domain` 仍是零依赖的（接口本身不需要第三方库） |
>
> **倾向**：`Storage` 放 `domain`（多消费者），`UserRepository` 这类单一消费者的继续放 `service`。**判断标准是"消费方数量"，不是教条。**

#### `domain.Storage`（S5）

```go
// Storage 抽象对象存储。所有实现必须遵守 S3 语义。
type Storage interface {
    // 单次上传（适合小文件）
    Put(ctx context.Context, key string, r io.Reader, size int64, contentType string) error
    Get(ctx context.Context, key string) (io.ReadCloser, error)
    Remove(ctx context.Context, key string) error

    // 分片上传（标准 S3 Multipart）
    CreateMultipart(ctx context.Context, key, contentType string) (uploadID string, err error)
    UploadPart(ctx context.Context, key, uploadID string, partNum int32, r io.Reader, size int64) (etag string, err error)
    CompleteMultipart(ctx context.Context, key, uploadID string, parts []CompletedPart) error
    AbortMultipart(ctx context.Context, key, uploadID string) error

    // 预签名（分发链路核心）
    PresignGet(ctx context.Context, key string, ttl time.Duration) (string, error)
    PresignPutPart(ctx context.Context, key, uploadID string, partNum int32, ttl time.Duration) (string, error)
}
```

**实现计划**

| 实现 | 用途 |
|---|---|
| `storage.LocalFS` | S4/S5 单测与本地开发，无外部依赖 |
| `storage.SeaweedFS`（AWS SDK v2） | 集成与生产 |
| `storage.Memory` | 单元测试 |

### 5.4 HTTP API 清单

#### 已实现 ✅

| 方法 | 路径 | 鉴权 | 说明 | 响应 |
|---|---|---|---|---|
| GET | `/healthz` | 否 | 存活探针 | `{"status":"alive"}` |
| GET | `/readyz` | 否 | 就绪探针（真 ping 库） | 200 / 503 |
| POST | `/api/v1/auth/register` | 否 | 注册 | 201 + UserResponse |
| POST | `/api/v1/auth/login` | 否 | 登录 | 200 + `{token}` |
| GET | `/api/v1/users/me` | **是** | 当前用户 | 200 + UserResponse |

#### 规划 ⬜

```
# 用户与社交（S5）
PATCH  /api/v1/users/me
GET    /api/v1/users/:public_id
GET    /api/v1/users/:public_id/works
POST   /api/v1/users/:public_id/follow
DELETE /api/v1/users/:public_id/follow
GET    /api/v1/feed                       关注流（游标分页）

# 作品（S5）
POST   /api/v1/works
GET    /api/v1/works                      筛选 + 游标分页
GET    /api/v1/works/:public_id
PATCH  /api/v1/works/:public_id
DELETE /api/v1/works/:public_id
POST   /api/v1/works/:public_id/publish
POST   /api/v1/works/:public_id/like
DELETE /api/v1/works/:public_id/like
GET    /api/v1/works/:public_id/comments
POST   /api/v1/works/:public_id/comments

# 版本与依赖（S6）
POST   /api/v1/works/:public_id/versions
GET    /api/v1/works/:public_id/versions
GET    /api/v1/versions/:public_id/deps
GET    /api/v1/versions/:public_id/deps/graph

# 上传（S5）
POST   /api/v1/uploads/check              秒传预检
POST   /api/v1/uploads/init
PUT    /api/v1/uploads/:id/parts/:n
POST   /api/v1/uploads/:id/complete
DELETE /api/v1/uploads/:id

# 检索（S7）
GET    /api/v1/search                     关键词（tsvector）
POST   /api/v1/search/semantic            自然语言（向量 + RRF）

# 通知与实时（S7）
GET    /api/v1/notifications
POST   /api/v1/notifications/read
GET    /api/v1/stream                     SSE

# 运维（S8）
GET    /metrics
```

### 5.5 统一响应与错误码

#### 响应结构

```json
// 成功
{ "data": { ... } }

// 失败
{ "error": { "code": "email_exists", "message": "该邮箱已被注册" } }
```

**为什么不用 `{code, msg, data}` 三合一**：前端每次都得判 `code === 0`，容易漏。成功与失败用不同字段，前端直接判 `error` 是否存在。

#### 错误码表（按实际代码）

| code | HTTP | 场景 | 状态 |
|---|---|---|---|
| `invalid_request` | 400 | 请求参数校验失败（**注意：不是 `invalid_argument`**） | ✅ 已实现 |
| `unauthorized` | 401 | 未登录或令牌无效 | ✅ 已实现 |
| `invalid_credentials` | 401 | 邮箱或密码错误 | ✅ 已实现 |
| `user_not_found` | 404 | 用户不存在 | ✅ 已实现 |
| `email_exists` | 409 | 邮箱已被注册 | ✅ 已实现 |
| `username_exists` | 409 | 用户名已被占用 | ✅ 已实现 |
| `internal` | 500 | 服务内部错误 | ✅ 已实现 |
| `forbidden` | 403 | 已登录但无权限 | ⬜ 未实现 |
| `not_found` | 404 | 通用资源不存在 | ⬜ 未实现 |
| `dependency_cycle` | 422 | 依赖成环（S6，**须返回具体路径**） | ⬜ S6 |
| `hash_mismatch` | 409 | 上传哈希不符（S5） | ⬜ S5 |

#### 两条硬规则

1. **错误码是稳定的字符串常量，不是数字** —— 支持前端分支处理与国际化
2. **5xx 才记 Error 日志** —— 4xx 是预期内的业务分支，记成 Error 会淹没真故障并触发无意义告警

#### 敏感信息规则

**绝不把原始错误返回给客户端**（可能泄漏 SQL、路径、栈）。详细错误写日志，客户端只收到通用提示。

**循环依赖错误要携带可操作细节**：

```json
{
  "error": {
    "code": "dependency_cycle",
    "message": "检测到循环依赖",
    "details": { "path": ["create@1.0.0", "flywheel@2.1.0", "create@1.0.0"] }
  }
}
```

---

## 6. 数据模型

### 6.1 已实现：`users`

**迁移文件**：`migrations/20261027151811_create_works.sql`

> ⚠️ **两处必须注意**：
> 1. 文件名是 **日期时间戳** `20261027151811`（goose 的 timestamp 命名），**不是** `00001` 序号式
> 2. 文件名写的是 **`create_works`**，但内容建的是 **`users` 表**——文件名与内容不符，属历史遗留。**改表结构时不要被文件名误导**

**实际 SQL（逐字摘自仓库）**

```sql
-- +goose Up
CREATE TABLE users (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    public_id UUID NOT NULL DEFAULT uuid_generate_v4(),
    username TEXT NOT NULL,
    email TEXT NOT NULL,
    password_hash TEXT NOT NULL,
    bio TEXT,
    avatar_url TEXT,
    role TEXT NOT NULL DEFAULT 'user',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

    CONSTRAINT user_email_key UNIQUE (email),
    CONSTRAINT user_username_key UNIQUE (username),
    CONSTRAINT user_public_id_key UNIQUE (public_id),
    CONSTRAINT user_role_check CHECK (role IN ('user', 'admin', 'moderator'))
);

CREATE INDEX idx_users_email_lower ON users (LOWER(email));

-- +goose Down
DROP TABLE users;
```

#### 设计理由

| 写法 | 理由 |
|---|---|
| `BIGINT GENERATED ALWAYS AS IDENTITY` | SQL 标准写法，等价于 `BIGSERIAL` 但更规范 |
| 另设 `public_id UUID` | 对外暴露用 UUID，**不泄漏用户规模**；内部关联用整数，索引更小 |
| `TEXT` 而非 `VARCHAR(n)` | PG 中两者性能相同；`TEXT` 无长度限制更灵活。**密码哈希定长 60 字节，用 `VARCHAR(50)` 会截断且报错信息不指向原因** |
| `TIMESTAMP WITH TIME ZONE` | 等价 `TIMESTAMPTZ`，存绝对时刻。**永不用无时区的 `TIMESTAMP`** |
| 显式 `CONSTRAINT 名字` | 出错可按名定位；**`mapUniqueViolation` 依赖这些名字做分支** |
| `CHECK ... IN (...)` | 数据库层枚举校验；比原生 `ENUM` 灵活（加值不需 `ALTER TYPE`） |
| `idx_users_email_lower` | **函数索引**。不建则 `WHERE lower(email)=lower($1)` 会全表扫描 |

#### ⚠️ 约束名是 `user_*` 而非 `users_*`（最易出错处）

实际约束名是**单数** `user_` 前缀：

| 约束名（实际） | 旧文档曾写（错误） |
|---|---|
| `user_email_key` | ~~`users_email_key`~~ |
| `user_username_key` | ~~`users_username_key`~~ |
| `user_public_id_key` | ~~`users_public_id_key`~~ |
| `user_role_check` | ~~`users_role_check`~~ |

**为什么这件事重要**：`internal/repository/user.go` 的 `mapUniqueViolation` 精确匹配这些字符串：

```go
switch pgErr.ConstraintName {
case "user_email_key":     return domain.ErrEmailExists
case "user_username_key":  return domain.ErrUsernameExists
default:                   return nil
}
```

**如果按旧文档把迁移里的约束名改成 `users_*`，唯一约束冲突将不再被翻译成领域错误，注册重复邮箱会变成 500。** 迁移与代码必须同步修改。

**命名建议**：`user_*`（单数）与表名 `users`（复数）不一致，不符合"表名_列名_key"的常见惯例。**若要统一，必须同时改迁移与 `mapUniqueViolation`，且需一次新的迁移（ALTER TABLE RENAME CONSTRAINT）。当前不是优先事项，先记录在此。**

#### 文件名与内容不符的处置

`20261027151811_create_works.sql` 建的是 `users`。两个选择：

| 选项 | 做法 | 评价 |
|---|---|---|
| **保留** | 接受历史遗留，在文件头加注释说明 | ✅ 推荐。迁移已应用，改名会导致 goose 版本表对不上 |
| 改名 | 重命名并同步 `goose_db_version` | ❌ 风险高，收益低 |

**结论**：保留文件名，但**必须**在文件首行加一行注释，避免下一个人误判。建议内容：

```sql
-- 注意：本迁移实际创建 users 表，文件名沿用了早期命名，勿据文件名推断内容
```

### 6.2 规划表（S5 / S6 / S7）

#### S5：作品与文件

```sql
CREATE TYPE work_category AS ENUM ('mod','art','asset','tool');
CREATE TYPE work_status   AS ENUM ('draft','pending','published','rejected','taken_down');
CREATE TYPE visibility    AS ENUM ('public','unlisted','private');

CREATE TABLE works (
    id             BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    public_id      UUID NOT NULL DEFAULT uuid_generate_v4() UNIQUE,
    author_id      BIGINT NOT NULL REFERENCES users(id),
    title          TEXT NOT NULL,
    summary        TEXT,
    cover_url      TEXT,
    category       work_category NOT NULL,
    status         work_status NOT NULL DEFAULT 'draft',
    visibility     visibility  NOT NULL DEFAULT 'public',
    tags           TEXT[] NOT NULL DEFAULT '{}',
    like_count     INT    NOT NULL DEFAULT 0,
    view_count     BIGINT NOT NULL DEFAULT 0,
    download_count BIGINT NOT NULL DEFAULT 0,
    published_at   TIMESTAMPTZ,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_works_tags      ON works USING gin (tags);
CREATE INDEX idx_works_fts       ON works USING gin (
    setweight(to_tsvector('simple', coalesce(title,'')),  'A') ||
    setweight(to_tsvector('simple', coalesce(summary,'')), 'B'));
CREATE INDEX idx_works_published ON works (category, published_at DESC)
    WHERE status = 'published' AND visibility = 'public';
CREATE INDEX idx_works_author    ON works (author_id, updated_at DESC);

-- 品类扩展层（元数据驱动）
CREATE TABLE work_metadata (
    work_id    BIGINT PRIMARY KEY REFERENCES works(id) ON DELETE CASCADE,
    schema_key TEXT NOT NULL,
    schema_ver INT  NOT NULL,
    attrs      JSONB NOT NULL DEFAULT '{}',
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_wm_attrs ON work_metadata USING gin (attrs jsonb_path_ops);

-- 文件（内容寻址）
CREATE TABLE files (
    id          BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    hash_sha256 CHAR(64) NOT NULL UNIQUE,   -- 去重 + 秒传的支点
    size_bytes  BIGINT   NOT NULL CHECK (size_bytes >= 0),
    mime        TEXT,
    storage_key TEXT NOT NULL,
    ref_count   INT NOT NULL DEFAULT 0,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_files_gc ON files (created_at) WHERE ref_count = 0;

CREATE TABLE work_versions (
    id             BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    public_id      UUID NOT NULL DEFAULT uuid_generate_v4() UNIQUE,
    work_id        BIGINT NOT NULL REFERENCES works(id) ON DELETE CASCADE,
    semver         TEXT NOT NULL,
    changelog      TEXT,
    file_id        BIGINT NOT NULL REFERENCES files(id),
    game_version   TEXT,
    loader         TEXT,
    status         work_status NOT NULL DEFAULT 'draft',
    download_count BIGINT NOT NULL DEFAULT 0,
    published_at   TIMESTAMPTZ,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (work_id, semver)
);

-- 依赖 DAG 的边
CREATE TABLE version_deps (
    version_id        BIGINT NOT NULL REFERENCES work_versions(id) ON DELETE CASCADE,
    dep_work_id       BIGINT NOT NULL REFERENCES works(id),
    dep_version_range TEXT NOT NULL,
    optional          BOOLEAN NOT NULL DEFAULT false,
    PRIMARY KEY (version_id, dep_work_id)
);

-- 分片上传会话
CREATE TABLE upload_sessions (
    id             UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    uploader_id    BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    file_hash      CHAR(64) NOT NULL,
    file_size      BIGINT NOT NULL,
    part_size      INT NOT NULL,
    total_parts    INT NOT NULL,
    s3_upload_id   TEXT NOT NULL,
    storage_key    TEXT NOT NULL,
    received       INT[] NOT NULL DEFAULT '{}',
    status         TEXT NOT NULL DEFAULT 'active',
    expires_at     TIMESTAMPTZ NOT NULL DEFAULT now() + interval '24 hours',
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_us_expire ON upload_sessions (expires_at) WHERE status = 'active';
```

#### S6 / S7：社交、AI 任务、向量、通知

```sql
CREATE TABLE follows (
    follower_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    followee_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (follower_id, followee_id),
    CHECK (follower_id <> followee_id)
);

CREATE TABLE likes (
    user_id    BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    work_id    BIGINT NOT NULL REFERENCES works(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, work_id)
);

CREATE TABLE comments (
    id         BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    work_id    BIGINT NOT NULL REFERENCES works(id) ON DELETE CASCADE,
    version_id BIGINT REFERENCES work_versions(id) ON DELETE SET NULL,
    user_id    BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    parent_id  BIGINT REFERENCES comments(id) ON DELETE CASCADE,
    content    TEXT NOT NULL CHECK (length(content) BETWEEN 1 AND 2000),
    like_count INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TYPE task_status AS ENUM ('queued','running','succeeded','failed','degraded');

CREATE TABLE agent_tasks (
    id          BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    task_type   TEXT NOT NULL,
    target_type TEXT NOT NULL,
    target_id   BIGINT NOT NULL,
    status      task_status NOT NULL DEFAULT 'queued',
    input       JSONB NOT NULL DEFAULT '{}',
    output      JSONB,
    degraded    BOOLEAN NOT NULL DEFAULT false,
    trace_id    TEXT,
    token_used  INT NOT NULL DEFAULT 0,
    latency_ms  INT,
    error       TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    finished_at TIMESTAMPTZ
);

CREATE TABLE embeddings (
    id         BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    owner_type TEXT NOT NULL,
    owner_id   BIGINT NOT NULL,
    chunk_no   INT NOT NULL,
    content    TEXT NOT NULL,          -- 原文，检索后直接返回，免回表
    embedding  vector(768) NOT NULL,   -- 维度必须与 embedding 模型对齐
    model      TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (owner_type, owner_id, chunk_no, model)
);
CREATE INDEX idx_emb_hnsw ON embeddings
    USING hnsw (embedding vector_cosine_ops) WITH (m = 16, ef_construction = 64);

CREATE TABLE notifications (
    id         BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id    BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type       TEXT NOT NULL,
    payload    JSONB NOT NULL DEFAULT '{}',
    read_at    TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_notif_unread ON notifications (user_id, created_at DESC)
    WHERE read_at IS NULL;
```

### 6.3 属性存储的分界线（关键约束）

"元数据驱动"若被理解为"所有字段计入文档"，将同时引发三个问题：数据库约束丧失、多属性过滤性能劣化、应用层失去类型安全。

**判定规则**：

> **会出现在筛选、排序、关联条件中的字段，必须是独立列。仅用于展示、且结构随品类变化的字段，才存入文档。**

| 归属 | 字段 |
|---|---|
| **独立列** | 标题、作者、品类、状态、可见性、标签数组、游戏版本与加载器、各类计数 |
| **文档字段** | 模组的依赖结构细节、插画的尺寸与媒介、AI 生成的标签候选与风险标注、任务输入输出 |

**真正的"元数据驱动"位于 Schema 注册表，而非存储层。**

### 6.4 索引策略

| 目标 | 策略 |
|---|---|
| 标签检索 | `TEXT[]` + GIN，支持包含与重叠查询 |
| 全文检索 | 组合 `tsvector`，标题权重 `A`、简介 `B` |
| 列表页 | **部分索引**仅覆盖"已发布且公开"，显著缩小索引体积 |
| 文档属性 | `jsonb_path_ops`，体积与查询效率优于默认配置 |
| 向量检索 | HNSW 图索引，查询快且内存可控 |
| 高频过滤 | 部分索引，不为不查的数据付维护成本 |
| 大小写不敏感查询 | **函数索引** `lower(email)` |

### 6.5 分页策略（强制）

列表接口**一律使用游标分页**。

```sql
-- 复合游标：稳定且走索引
SELECT * FROM works
WHERE status = 'published'
  AND (published_at, id) < ($1, $2)     -- 上一页最后一条
ORDER BY published_at DESC, id DESC
LIMIT 20;
```

| 方案 | 深翻页复杂度 | 并发写一致性 |
|---|---|---|
| 偏移分页 `OFFSET n` | O(n)，线性劣化 | 会重复/遗漏 |
| **游标分页** | O(log n) | 稳定 |

---

## 7. 技术选型与决策记录

### 7.1 选型总表

| 层面 | 选择 | 核心理由 |
|---|---|---|
| 语言 | Go 1.22+ | 高并发 I/O；1.22 起循环变量语义修正 |
| Web 框架 | Gin | 中间件生态完整，路由与绑定成熟 |
| 主库 | PostgreSQL 16 | 单库覆盖关系/文档/检索/队列六角色 |
| 连接与事务 | pgx (`pgxpool`) | 建池、事务控制、健康检查 |
| 查询生成 | sqlc | 由 SQL 生成类型安全查询函数 |
| 数据库迁移 | goose | 纯 SQL 迁移，版本可控 |
| 向量检索 | pgvector（HNSW） | 与业务数据同事务域 |
| 全文检索 | PostgreSQL `tsvector` | 免引入独立搜索中间件 |
| 任务队列 | **River**（PG 后端） | 支持**事务性入队** |
| 缓存与限流 | **Redis 8** | 缓存、令牌桶、跨实例协调 |
| 对象存储 | **SeaweedFS** | Apache 2.0、Go 实现、活跃维护 |
| 存储 SDK | **AWS SDK for Go v2** | 上游为 AWS，供应商无关 |
| 实时推送 | **SSE 优先**，WebSocket 后置 | 九成场景为单向推送 |
| 依赖注入 | wire | 编译期校验，生成普通 Go 代码 |
| 日志 | **`log/slog`**（标准库） | 零依赖、结构化、够用 |
| 链路追踪 | OpenTelemetry | trace_id 贯穿 HTTP → service → Agent |
| 监控 | Prometheus | 含自定义业务指标 |
| 测试 | testify + testcontainers | 真实依赖的集成测试 |
| 质量工具 | golangci-lint + Makefile | 统一入口，CI 友好 |

### 7.2 明确排除的选项

| 不选 | 理由 |
|---|---|
| **NSQ** | 已停止维护；引入即为技术判断力负分；需额外守护进程 |
| **MinIO** | 2025 年转向商业产品，2026 年仓库归档（详见 ADR-004） |
| **`minio-go` SDK** | 上游项目已归档，长期依赖风险 |
| **重型 ORM（GORM 等）** | 会屏蔽 JSONB 查询与向量距离运算，而这两者正是核心能力 |
| 独立向量数据库 | PostgreSQL 已覆盖，多一套运维与一致性缺口 |
| 独立搜索中间件 | 同上 |
| **Kafka** | 吞吐量级不符，运维负担大 |
| 微服务 / 容器编排 | 周期内做不扎实，追问风险高 |
| 反射、底层编程、cgo、高级泛型 | 与项目目标无关 |

### 7.3 ADR-001｜单体分层而非微服务

| 项 | 内容 |
|---|---|
| **背景** | 9 周周期，开发者首次用 Go 做中大型项目 |
| **选项** | 单体分层 / 微服务 / 模块化单体 |
| **决定** | 单体分层，仅 Agent Worker 拆为独立进程 |
| **理由** | 微服务引入分布式事务、服务发现、链路追踪三大难题，周期内无法做扎实，且面试追问"跨服务一致性怎么保证"时无法自圆其说 |
| **代价** | 单进程资源隔离弱；未来若拆分需重构 |
| **状态** | 已采纳 |

### 7.4 ADR-002｜PostgreSQL 承担六种角色

| 项 | 内容 |
|---|---|
| **背景** | 需要关系存储、灵活属性、全文检索、向量检索、任务队列 |
| **选项** | MySQL + 独立向量库 + 独立搜索引擎 + 独立 MQ / PostgreSQL 单库 |
| **决定** | PostgreSQL 16 单库 |
| **理由** | ① 向量与业务数据同事务域，无跨存储一致性缺口 ② 少 3 套组件运维 ③ JSONB 表达力优于 MySQL JSON ④ 全文检索自带相关度打分 |
| **代价** | 单一存储承载多种负载，规模增长后需按读写特征拆分 |
| **状态** | 已采纳；代价须在评审中主动说明 |

### 7.5 ADR-003｜pgx + sqlc 而非 GORM

| 项 | 内容 |
|---|---|
| **背景** | 核心查询依赖 JSONB 操作符与 pgvector 距离运算 |
| **选项** | GORM / pgx + sqlc / database/sql 手写 |
| **决定** | pgx（连接池）+ sqlc（代码生成） |
| **理由** | ① ORM 抽象会挡住 JSONB 与向量查询 ② sqlc 生成类型安全代码，列名/类型错误在编译期暴露 ③ SQL 可见可控，便于讲解与优化 ④ 对无 ORM 经验的开发者学习曲线更平缓 |
| **澄清** | **pgx 与 sqlc 不是替代关系**：pgx 是运行时驱动与连接池，sqlc 是构建时代码生成器，生成代码底层即调用 pgx |
| **代价** | 需维护 `sqlc.yaml`；生成代码目录禁止手改 |
| **状态** | 已采纳 |

### 7.6 ADR-004｜SeaweedFS 替代 MinIO

| 项 | 内容 |
|---|---|
| **背景** | MinIO 曾是最常用的自建 S3 兼容存储，但已从开源撤退 |
| **事实时间线** | 2025-05 社区版移除大部分管理功能与 LDAP/OIDC → 2025-10 **停止发布 Docker 镜像与预编译二进制** → 2025-12-03 README 宣布维护模式 → 2026-02 仓库归档（短暂解档）→ 2026-04-25 再次归档，read-only。官方引导转向商业产品 AIStor（专有许可，按容量计费） |
| **选项** | 继续用 MinIO CE / SeaweedFS / Garage / RustFS / Ceph RGW |
| **决定** | **SeaweedFS** |
| **理由** | ① Apache 2.0，无 AGPL 网络分发义务 ② Go 实现，与项目同语言 ③ 活跃维护，生产可用 ④ `server -s3` 一条命令起全套，学习成本低 ⑤ `multipart upload` 与预签名 URL 均支持（本项目唯一需要的两项） ⑥ 空闲内存 60–100 MB，适合与 PG/Redis 同机 |
| **被放弃选项** | Garage（需配置集群 layout + key 授权，单机多一层概念，且 AGPL）；RustFS（状态有冲突信息，历史短，不宜作为 9 周项目地基）；Ceph RGW（最少 3 节点，16 GB+ 内存，明显过重） |
| **代价** | 生态与文档不如 MinIO 广；S3 API 覆盖为"中等"（不支持对象锁、桶版本控制——本项目不需要） |
| **状态** | 已采纳 |

### 7.7 ADR-005｜AWS SDK for Go v2 替代 minio-go

| 项 | 内容 |
|---|---|
| **背景** | 原本计划用 `minio-go`，但 MinIO 项目已归档 |
| **决定** | `github.com/aws/aws-sdk-go-v2/service/s3` |
| **理由** | ① 上游为 AWS，不会消失 ② 与任何 S3 兼容后端兼容 ③ 行业标准，简历价值更高 |
| **关键配置** | **必须设置 `UsePathStyle = true`**。SDK 默认虚拟主机式寻址（`bucket.s3.amazonaws.com`），自建后端无泛域名 DNS，会报 `lookup xxx.localhost: no such host` —— 错误信息完全指不到真正原因 |
| **代价** | API 较啰嗦 |
| **状态** | 已采纳 |

### 7.8 ADR-006｜标准 S3 Multipart 替代 ComposeObject

| 项 | 内容 |
|---|---|
| **背景** | 早期设计使用服务端合并（`ComposeObject`），但那是 **MinIO 专有扩展**，非 S3 标准 |
| **决定** | 改用标准 S3 Multipart：`CreateMultipartUpload` → `UploadPart` → `CompleteMultipartUpload` / `AbortMultipartUpload` |
| **理由** | ① 所有 S3 兼容后端都支持，可移植 ② 换后端时上传逻辑零改动 ③ 面试讲"标准 S3 Multipart"比讲"MinIO 的 ComposeObject"价值更高 |
| **衍生优化** | 优先**分片直传**：客户端经预签名 URL 直接上传分片，API 只协调与鉴权，不转发字节 → 带宽不成为瓶颈 |
| **状态** | 已采纳（S5 实施） |

### 7.9 ADR-007｜River 替代 NSQ

| 项 | 内容 |
|---|---|
| **背景** | 需要异步任务队列（内容审核、embedding 生成、通知） |
| **选项** | NSQ / asynq（Redis） / River（PostgreSQL） / Kafka |
| **决定** | River |
| **理由** | ① NSQ 已停止维护，引入即负分 ② River 后端在 PostgreSQL，支持**事务性入队**——业务写入与任务投递同一事务，获得原子性 ③ 免去 outbox 表 + 定时补偿的复杂度 ④ 复用已有 PG，零新增组件 |
| **强制要求** | 队列在业务层必须是**抽象接口**，业务代码只依赖"投递任务"契约，不依赖 River 类型 |
| **代价** | 队列负载压在 PG 上，高吞吐场景不如专用 MQ |
| **状态** | 已采纳（S7 实施） |

### 7.10 ADR-008｜SSE 优先于 WebSocket

| 项 | 内容 |
|---|---|
| **背景** | 需要实时推送（通知、AI 任务进度、版本更新广播） |
| **决定** | SSE 覆盖单向推送；WebSocket 仅保留给真正的双向场景（作为可选增强） |
| **理由** | 场景分析显示约九成实时需求是单向的。SSE 走标准 HTTP（无协议升级）、浏览器原生自动重连、对网关与负载均衡友好、`curl` 可调试 |
| **场景判定表** | 上传进度 → 前端本地计算；审核结果 → SSE 或轮询；AI 任务进度 → SSE；站内通知 → SSE；版本广播 → SSE；私信/协同编辑 → WebSocket |
| **衍生价值** | **主动选择不用 WebSocket 本身即为判断力的体现**，评审时若被问"为何用轮询不用 WS"，此处即口径 |
| **状态** | 已采纳（S7 实施） |

### 7.11 ADR-009｜Redis 8 而非 Redis 7

| 项 | 内容 |
|---|---|
| **背景** | 需要缓存、令牌桶限流、跨实例 Pub/Sub |
| **决定** | Redis 8.x |
| **理由** | ① 本项目用到的能力（GET/SET+TTL、Lua 脚本、PUBLISH/SUBSCRIBE、SET NX PX）在两版之间 API 完全一致，无功能性理由必须用 7 ② **许可证更好**：7.4 为 RSALv2/SSPL（源码可见但非 OSI 认证），8.0 起新增 AGPLv3 成为 OSI 认证开源 ③ 8.0 于 2025-05 GA，成熟度已足够 |
| **备选** | Valkey（Linux Foundation 分支，BSD-3，API 兼容）。本项目选 Redis 8 因生态与文档更成熟，但**知道 Valkey 的存在是判断力的一部分** |
| **状态** | 已采纳（S7 实施） |

### 7.12 ADR-010｜`log/slog` 替代 zap

| 项 | 内容 |
|---|---|
| **背景** | 原计划 zap |
| **决定** | 标准库 `log/slog` |
| **理由** | ① 零新增依赖 ② 本项目日志量不需要 zap 的极致性能 ③ S8 接 OpenTelemetry 时 slog 有官方 bridge |
| **重要约束** | **不要为日志定义接口**。包装会丢失 `With`/`Enabled`/`Handler()`，且接口注定不完整。测试用 `slog.New(slog.NewTextHandler(io.Discard, nil))` 静默即可 |
| **换 zap 的改动面** | 仅三处：`providers.go` 的 `newLogger`；`handler/*` 的字段类型与调用写法；`middleware/*` 同上。**业务层与领域层完全不受影响** —— 这是分层的证明 |
| **状态** | 已采纳 |

### 7.13 ADR-011｜接口定义在消费方

| 项 | 内容 |
|---|---|
| **决定** | 接口由**使用方**定义，而非实现方 |
| **已实施** | `service.UserRepository`、`service.TokenSigner`（定义于 `service`）；`middleware.TokenParser`（定义于 `middleware`）；`handler.Pinger`（定义于 `handler`） |
| **理由** | ① 接口恰好是使用所需的最小集合，不会变胖 ② 实现方无需感知谁在用它 ③ 同一个实现可满足多个不同小接口（`*jwt.Signer` 同时满足两个） |
| **对 wire 的影响** | Go 的接口是隐式的，但 wire 生成代码时必须显式 `wire.Bind` 才知道用哪个实现 |
| **状态** | 已采纳 |

### 7.14 ADR-012｜内容寻址存储

| 项 | 内容 |
|---|---|
| **决定** | 文件实体以 `hash_sha256` 为唯一标识；对象存储 key 由哈希派生 |
| **理由** | ① 天然去重：同一内容全平台只存一份 ② 秒传：哈希命中即零字节返回 ③ 校验：合并后可重算哈希验证完整性 |
| **关键细节** | 秒传预检须**原子化**（`INSERT ... ON CONFLICT DO UPDATE ... RETURNING`，用 `xmax <> 0` 判断是否命中冲突分支），避免 check-then-act 竞态 |
| **防碰撞** | 哈希 + `size_bytes` 双重校验；分片哈希仅用于传输幂等，不作为全局标识 |
| **状态** | 已采纳（S5 实施） |

### 7.15 ADR-013｜项目改名 ForgeHub → CraftShelf

| 项 | 内容 |
|---|---|
| **决定** | 项目名称、模块路径、数据库名、桶名、容器名统一为 `craftshelf` |
| **映射** | 模块 **`github.com/IssierHM/CraftShelf`**；库/用户 `craftshelf`；容器 `craftshelf-*`；桶 **`craftshelf-works` / `craftshelf-avatars`**（compose、`.env`、`config.go` 已对齐） |
| **遗留** | 技能文档仍为旧名，见 P-06 |
| **状态** | 已采纳 |

---

## 8. 配置参考

### 8.1 环境变量（按 `config.go` 实际读取的键名）

| 变量 | 必需 | 默认 | 实际约束 |
|---|---|---|---|
| `APP_HTTP_ADDR` | 否 | `:8080` | — |
| `APP_POSTGRES_DSN` | **是** | — | 缺失即 `Load()` 失败 |
| `APP_REDIS_ADDR` | 否 | `localhost:6379` | 已解析，S7 前未使用（P-05） |
| `APP_JWT_SECRET` | **是** | — | 非空、≥32 字节、无空白/控制字符、**不等于默认占位符 `changemechangemechangemechangeme12`** |
| `APP_JWT_TTL` | 否 | `24h` | `time.ParseDuration` 可解析 |
| `APP_BCRYPT_COST` | 否 | `12` | 限定 `[10,14]` |
| `APP_S3_ENDPOINT` | 否 | `http://localhost:8333` | 带协议前缀 |
| `APP_S3_REGION` | 否 | `us-east-1` | SDK 强制要求，后端不校验 |
| `APP_S3_ACCESS_KEY` | 否 | 空 | 与 `s3.json` 的 `accessKey` 一致 |
| `APP_S3_SECRET_KEY` | 否 | 空 | 与 `s3.json` 的 `secretKey` 一致 |
| **`APP_S3_BUCKET_WORKS`** | 否 | **`craftshelf-works`** | 作品文件桶（与 compose 一致） |
| **`APP_S3_BUCKET_AVATARS`** | 否 | **`craftshelf-avatars`** | 头像桶（与 compose 一致） |
| `APP_S3_USE_PATH_STYLE` | 否 | `true` | 自建后端必须 `true` |

> ✅ **P-04 已关闭**：`config.go` **已完整解析**全部 `APP_S3_*` 字段。
>
> ✅ **P-09 已关闭（2026-11-03）**：`config.go` 字段为 `S3BucketWorks`/`S3BucketAvatars`，环境变量为 `APP_S3_BUCKET_WORKS`/`APP_S3_BUCKET_AVATARS`，默认值与 `.env` 均为 `craftshelf-works`/`craftshelf-avatars`，与 compose 初始化容器创建的两桶一致。SeaweedFS 实测 `s3.bucket.list` 可见两桶。
>
> ✅ **P-13 已关闭**：`.env.example` 已创建，仅含变量名与占位值。

### 8.2 密钥管理规则

| 规则 | 原因 |
|---|---|
| 密钥**只能**从环境变量读，源码中不出现 | 源码会提交 git |
| 密钥**无默认值**，缺失即启动失败 | 带弱密钥运行比启动失败更危险 |
| `.env` **绝不提交**；`.env.example` **必须提交** | 前者含真实值，后者是文档 |
| `APP_JWT_SECRET` 长度 ≥32 字节并强制校验 | HS256 安全性依赖密钥熵 |

### 8.3 DSN 结构

```
postgres://craftshelf:craftshelf_dev_pwd@localhost:5432/craftshelf?sslmode=disable
└──┬─────┘ └───┬────┘ └──────┬───────┘ └──┬───┘ └┬─┘ └───┬────┘ └──────┬───────┘
  协议       用户          密码          主机   端口   库名        查询参数
```

`sslmode` 取值

| 值 | 行为 | 适用 |
|---|---|---|
| `disable` | 不加密 | **本地开发（Docker 内 PG 未配 TLS）** |
| `require` | 加密但不验证书 | 内网 |
| `verify-full` | 全验证 | **生产** |

**生产用 `disable` 是严重安全缺陷；本地用 `require` 会直接连不上。**

### 8.4 Docker 服务

| 服务 | 实际镜像 | 端口 | 健康检查 |
|---|---|---|---|
| PostgreSQL | `pgvector/pgvector:pg16` | 5432 | `pg_isready -U craftshelf -d craftshelf` |
| Redis | `redis:8-alpine` | 6379 | `redis-cli ping` |
| SeaweedFS | **`chrislusf/seaweedfs:4.51`** | 8333 (S3) / 9333 (master) / 8888 (filer UI) | `curl -f http://localhost:9333/cluster/status` |

> **关于镜像命名空间**：`4.51` 是真实存在的版本（GitHub releases 可查，发布日期 2026-10-25），但仓库用的是**旧命名空间 `chrislusf/`** 而非新命名空间 `seaweedfs/`。两者目前都可用，**保持现状即可**（已在本地验证可跑），但升级时需确认目标 tag 在哪个命名空间下存在。

**关键约束**

- 必须用**预装 vector 扩展的 PG 镜像**，官方 `postgres:16` 没有
- PG 初始化脚本放 `/docker-entrypoint-initdb.d`，**仅在数据卷为空时执行一次**
- `depends_on` 用 `condition: service_healthy`
- `start_period` 给足（PG 首次初始化需 20–30s）
- 无自定义 `networks:` 段 —— Compose 默认网络已内置服务名 DNS

**容器内 vs 宿主机寻址**

| 程序跑在 | 连数据库 | 连 SeaweedFS |
|---|---|---|
| Windows 宿主机（`go run`） | `localhost:5432` | `http://localhost:8333` |
| 容器内 | `postgres:5432` | `http://seaweedfs:8333` |

### 8.5 版本锁定策略

**本项目已采用 `go.mod` 的 `tool` 指令**（Go 1.24+ 特性），工具版本随模块走：

```
tool (
	github.com/golangci/golangci-lint/cmd/golangci-lint
	github.com/google/uuid
	github.com/google/wire/cmd/wire
	github.com/pressly/goose/v3/cmd/goose
	github.com/sqlc-dev/sqlc/cmd/sqlc
)
```

**调用方式**（不能直接敲命令）：

```powershell
go tool sqlc generate
go tool wire ./cmd/api
go tool goose -dir migrations up
go tool golangci-lint run ./...
```

| 类型 | 策略 | 实际状态 |
|---|---|---|
| 运行时依赖 | `go.mod` + `go.sum` 提交；阶段收尾 `go mod tidy` | 🚧 `godotenv` 被直接 import 却标 `// indirect`（P-16） |
| 开发工具 | **`tool` 指令** | 🚧 `google/uuid` 被误放入 `tool` 块（P-17） |
| 容器镜像 | pin 到具体小版本，不用 `latest` | ✅ `4.51` / `8-alpine` / `pg16` 均已 pin |
| 生产对象存储 | 切云存储只改配置，不改代码 | ⬜ S8 |

**核实命令**

```powershell
# SeaweedFS 当前版本
(Invoke-RestMethod 'https://api.github.com/repos/seaweedfs/seaweedfs/releases/latest').tag_name

# 确认某 tag 存在（不拉镜像，注意两个命名空间）
docker manifest inspect chrislusf/seaweedfs:4.51
docker manifest inspect seaweedfs/seaweedfs:4.51
docker manifest inspect redis:8-alpine
```

**原则**：文档只写版本**策略与核实方法**，不写具体版本号——除了已冻结不变的（如 MinIO 的归档状态）。

---

## 9. 工程规范

### 9.1 命名规范

| 场景 | 规范 | 示例 |
|---|---|---|
| 模块路径 | 完整仓库地址 | **`github.com/IssierHM/CraftShelf`** |
| 包名 | 全小写、单数、不用 `util`/`common`/`base` | `config`、`repository`、`jwt` |
| 可见性 | 首字母大小写 | `UserRepo` / `toDomainUser` |
| 构造函数 | `New` + 类型名 | `NewUserRepo`、`NewAuthService`、`NewRouter` |
| 错误变量 | `Err` 前缀 + 语义 | `ErrEmailExists`（**例外：`jwt.ErrorInvalidToken`**） |
| 错误信息 | **全小写、无结尾标点** | `fmt.Errorf("create user: %w", err)` |
| 路径格式化 | 用 `%q` 而非 `%s` | `fmt.Errorf("open %q: %w", path, err)` |
| 接口（单方法） | 方法名 + `er` | `Storage`、`Pinger`、`TokenParser` |
| context 参数 | 第一个，固定名 `ctx` | `func (s *AuthService) Login(ctx context.Context, ...)` |
| 上下文 key | **未导出常量** + 访问函数 | `const ctxUserIDKey = "auth.user_id"` + `UserIDFrom()` |
| 缩写词 | 全大写或全小写 | `userID`、`parseURL`、`S3Endpoint` |
| 布尔 | `is`/`has`/`can` 前缀 | `allOK`、`isValid` |
| 接收者 | 1–2 字母，同类型内一致 | `func (r *UserRepo)` |
| 数据库名/用户 | 小写下划线 | `craftshelf` |
| 约束名 | `<表名>_<列名>_key` | **实际是 `user_email_key`（单数，不统一，见 §6.1）** |
| 桶名 | 小写连字符 | **`craftshelf-works` / `craftshelf-avatars`**（按 compose 已建桶为准，见 P-09） |
| 渠道参数 | 明确方向 | `jobs <-chan Job`、`out chan<- Result` |

**已发现的实际命名偏差（记录，不强制立即修）**

| 位置 | 实际写法 | 建议 | 原因 |
|---|---|---|---|
| `service/auth.go` | `NewAuthService(..., bcryptcost int)` | `bcryptCost` | Go 惯例是驼峰，且首字母不应全小写连写 |
| `service/auth.go:32` | `dummyHash, error := ...` | 改名 `err` | `error` 遮蔽内置类型名，可读性差 |
**Go 的可见性规则**：无 `public`/`private` 关键字，首字母大小写即访问修饰符。配合 `internal/` 目录可在**编译期**强制封装。

### 9.2 代码生成流程（顺序不可变）

```
① goose 建表
      ↓  必须在 ② 之前
② sqlc 读表结构 → 生成类型安全查询代码
      ↓
③ wire 生成依赖装配代码
      ↓
④ go build / go test
```

**硬约束**：sqlc 需要**真实表结构**。因此 **`goose up` 必须成功后才能 `sqlc generate`**。

**推荐做法**：`sqlc.yaml` 用 `database.uri` 直连数据库读 schema，而非解析 `migrations/` 目录。

| 方式 | 原理 | 问题 |
|---|---|---|
| `schema: migrations` | 解析 SQL 文件构建表结构 | **会把 `-- +goose Down` 里的 `DROP TABLE` 一起解析**，导致表被删、查询报"表不存在" |
| **`database.uri`** | 连真实数据库读 schema | 需库在跑；DSN 写在文件里 |

**生成目录规则**

| 目录 | 能否手改 | 状态 |
|---|---|---|
| `internal/repository/sqlc/` | ❌ **永不手改**，改 `queries/*.sql` 后重新生成 | ✅ 生成文件头标注 `sqlc v1.31.1` |
| `cmd/api/wire_gen.go` | ❌ 生成，但**必须提交 git** | ✅ 已提交（P-15） |

### 9.3 分层验证

#### 现状：无自动化守护，仅手动检查

`internal/domain/architecture_test.go` 已于 2026-11-03 按用户决定移除，测试统一延后到 S5.5。**S5 期间每次改动 `domain` 后必须手动执行**：

```powershell
go list -deps ./internal/domain | Select-String "gin-gonic|jackc|google/uuid|sqlc|aws|redis"
```

**输出必须为空。** 注意 `aws` 是新加的检查项——S5 会引入 AWS SDK，这是当前最大的污染风险源。

#### 两种可选落地方式（方式 A 已保留；方式 B 待 S5.5）

**方式 A：Makefile 的 `verify-layers`（已添加）**

```makefile
verify-layers:
	@echo "检查领域层依赖..."
	@go list -deps ./internal/domain | grep -E "gin-gonic|jackc|google/uuid|sqlc|aws|redis" \
	  && { echo "❌ 领域层被污染"; exit 1; } || echo "✅ 领域层干净"
```

**方式 B：T1 测试（已移除，S5.5 重做时使用）**

```go
package domain_test

import (
	"os/exec"
	"strings"
	"testing"
)

func TestDomainHasNoFrameworkDependency(t *testing.T) {
	out, err := exec.Command("go", "list", "-deps", ".").CombinedOutput()
	if err != nil {
		t.Fatalf("go list 执行失败: %v\n%s", err, out)
	}
	forbidden := []string{
		"github.com/gin-gonic/gin",
		"github.com/jackc/pgx",
		"github.com/google/uuid",
		"github.com/aws/aws-sdk-go-v2",
		"github.com/redis/go-redis",
		"/internal/repository",
		"/internal/service",
		"/internal/handler",
		"/internal/pkg",
	}
	deps := string(out)
	for _, f := range forbidden {
		if strings.Contains(deps, f) {
			t.Errorf("领域层引入了禁止的依赖 %q —— 架构边界被破坏", f)
		}
	}
}
```

**验证它有效的办法**：临时在 `domain/user.go` 加 `import _ "github.com/google/uuid"`，确认测试**失败**，再删掉。**看到失败一次，才能确认守护不是空壳。**（2026-11-03 曾按此验证；文件已移除，S5.5 重做时再启用。）

### 9.4 测试策略

| 层次 | 范围 | 方式 |
|---|---|---|
| 单元 | 领域层纯逻辑（版本范围、环检测、属性校验） | 表驱动，无外部依赖 |
| 服务 | 业务编排 | **手写 fake**（不用 mockgen） |
| 集成 | 上传全流程、发布校验、任务执行 | testcontainers 起真实依赖 |
| 评测 | Agent 效果 | 模拟模型跑基线，真实模型产指标 |

#### fake / stub / mock

| 类型 | 作用 | 本项目 |
|---|---|---|
| Stub | 返回预设值，不验证调用 | 返回固定用户 |
| **Fake** | 有真实但简化的实现（内存 map 代替数据库） | **主要内容** |
| Mock | 预设期望，验证"是否被调用 N 次" | **不需要** |

**为什么用 fake 不用 mock**：fake 更接近真实行为；mock 的断言常绑死实现细节，重构即崩；手写 fake 约 40 行，比引入 mockgen 简单。

#### 关键测试技巧

| 技巧 | 说明 |
|---|---|
| **测试用 `bcrypt.MinCost`** | 生产 cost 12（≈250ms），测试 cost 4（≈1ms）。否则几十个测试要跑十几秒 |
| **负 TTL 测过期** | `NewSigner(secret, -time.Minute)` 制造过期令牌，**不必等待，也不必引入 Clock 抽象** |
| **断言"测试会失败"** | 把密码改成明文存储，确认安全测试失败，再改回。验证测试真的在守护 |
| **断言同一个错误值** | `errNoUser != errWrongPass` 才说明防邮箱枚举生效（而不是两个内容相同的新错误） |

#### 覆盖率

```powershell
go test -race -coverprofile=coverage.out ./...
go tool cover -func=coverage.out
go tool cover -html=coverage.out -o coverage.html
```

**覆盖率不是目标。** 60% 覆盖但包含全部安全断言与错误分支的套件，价值高于 95% 只测 happy path 的。

**优先覆盖**：错误分支、安全断言（哈希/统一错误/过期）、边界条件。

### 9.5 质量门禁

```powershell
gofmt -l .                         # 应无输出
go vet ./...                       # 应无输出
go tool golangci-lint run ./...    # 应无 error（当前 0 issues）
go test -race ./...                # 应全绿（本机需 CGO_ENABLED=1 + gcc，见 P-19）
```

**三条必须理解的 linter**

| linter | 抓什么 | 为什么重要 |
|---|---|---|
| `errcheck` | 忽略 `error` 返回值 | Go 的错误是返回值，忘了接住是最高频 bug |
| `bodyclose` | 未关闭 `resp.Body` | 连接泄漏，高并发下耗尽连接池 |
| `contextcheck` | 用了 `context.Background()` 而非传入 ctx | 取消信号断裂 |

> ✅ **P-08 已关闭（2026-11-03）**：`golangci-lint v1.64.8` 在 Go 1.26 下 `buildssa` panic，已升级到 **v2.15.0**；`.golangci.yml` 已用 `golangci-lint migrate` 转为 v2 格式，`go tool golangci-lint run ./...` 输出 `0 issues.`。

---

## 10. 技术难点设计（S5–S7 详细方案）

### 10.1 分片上传与秒传（S5）

#### 完整流程

```
① POST /uploads/check {hash, size}
   INSERT INTO files ... ON CONFLICT (hash_sha256) DO UPDATE SET ref_count = ref_count + 1
   RETURNING id, (xmax <> 0) AS existed
   ├─ existed = true  → 秒传，零字节返回
   └─ existed = false → 继续

② POST /uploads/init {hash, size, part_size}
   → CreateMultipartUpload 拿到 uploadID
   → 写 upload_sessions
   → 返回 {upload_id, total_parts, presigned_urls[]}

③ 客户端并发 PUT 分片到预签名 URL（API 不转发字节）
   服务端记录 received 分片号（幂等，重传不报错）

④ POST /uploads/complete
   → CompleteMultipartUpload
   → 流式重算 SHA-256 与声明值比对
     ├─ 一致 → 更新 files。storage_key，ref_count 生效
     └─ 不一致 → AbortMultipartUpload + 清理，返回 409 hash_mismatch

⑤ 后台任务
   → 清理过期 upload_sessions 及其临时分片
   → ref_count = 0 且超过安全期（7 天）的文件 GC
```

#### 关键技术点

| 点 | 实现 |
|---|---|
| **秒传原子化** | `xmax <> 0` 判断是否命中 `ON CONFLICT` 分支，一次往返完成"检查 + 占用引用"，避免 check-then-act 竞态 |
| **分片幂等** | `received` 用 `INT[]` + `array_agg(DISTINCT ...)` 去重；重传是正常现象 |
| **流式校验** | `io.Copy(h, reader)` 边读边算，固定 32KB 缓冲，内存与文件大小无关 |
| **并发控制** | `errgroup.SetLimit(n)` 或信号量；客户端侧限制 4–8 分片并发 |
| **延迟回收** | `ref_count = 0 AND created_at < now() - interval '7 days'` 才 GC，避免"刚上传尚未关联"被误删 |
| **下载直链** | `PresignGet` 签名有时效的 URL，文件不经过应用服务器 |

#### 可量化指标

- 大文件上传平均耗时与吞吐
- 秒传命中率、存储去重率
- 1 GB 文件上传时的进程内存峰值（须平稳）
- 下载接口 QPS 与 P99

### 10.2 版本与依赖 DAG（S6）

#### 环检测与路径还原

普通"检测到有环"价值有限。**能返回具体环路路径才是亮点**——用户看到 `A@1.0 → B@2.0 → C@1.0 → A@1.0` 才知道该拆哪条边。

**算法步骤**

```
① Kahn 拓扑排序：不断摘掉入度为 0 的节点
② 若入拓扑序节点数 < 总数 → 存在环
③ 剩下入度 > 0 的节点必在环内
④ 从任一环内节点出发 DFS，用 step 表记录首次访问步号
⑤ 遇到已访问节点 → 截取 [step[d], len(path)) 得到环本体
```

**面试追问预案**

| 追问 | 要点 |
|---|---|
| 为什么用 Kahn 而非 DFS 三色标记 | 两者都能检测环；Kahn 的副产物是"已入选拓扑序的节点集合"，可直接定位环成员；三色标记需额外回溯栈 |
| 环可能有多条 | 当前返回一条代表性环路（够定位问题）；全部环需 Johnson 算法，复杂度高，当前不做 |

#### 语义化版本范围匹配

`^1.2.0` / `~1.2.0` / `>=1.0.0 <2.0.0` 须自己解析。

**不引第三方 semver 库做核心逻辑**：① 自己实现 `^` `~` 的展开是很好的面试题 ② 体现理解"依赖范围"而非"精确版本"。

#### 发布校验（同一事务）

任一环节失败整体回滚，**不得产生半发布状态**。

```
版本号唯一性
  → 依赖作品存在性 + 满足范围的已发布版本存在性
  → 全量依赖图环检测（失败须返回具体路径）
  → files.ref_count + 1
  → 写 work_versions + version_deps
  → 事务内投递 River 任务
```

### 10.3 Agent 编排（S7）

#### 核心契约

```go
type Tool interface {
    Name() string
    Description() string                      // 给 Planner 看
    Schema() json.RawMessage                  // 参数 JSON Schema，约束 LLM 输出
    Run(ctx context.Context, args json.RawMessage) (Result, error)
    Idempotent() bool                         // 决定失败后能否安全重试
}

type Step struct {
    ID        string
    Tool      string
    Args      json.RawMessage
    DependsOn []string
    Timeout   time.Duration
    Retries   int
}
```

#### 执行器设计

```
① 构建入度表，检测步骤图有环（纯内存图）
② 入度为 0 的步骤可并发执行
③ 每步独立 context.WithTimeout —— 单步卡死不得拖垮整个计划
④ 信号量约束并发数 —— 避免打爆外部接口配额
⑤ errgroup 的取消传播：任一步失败，同批其余步骤被取消
⑥ g.Wait() 返回后，results 读取天然安全（happens-before）
```

**关键前提**：所有阻塞调用必须接受 `context`。链条上有一环断了，取消就传不到底。

#### 三级降级实现要点

```
runWithRetry:
  for attempt := 0; attempt <= Retries; attempt++ {
      res, err := runOnce(ctx, step)
      if err == nil {
          if Verify(res) == nil { return res }
          step = refineWithFeedback(step, verifyErr)   // 回炉：把失败原因带回
          continue
      }
      if !isRetryable(err) { break }                   // 400 类错误不重试
      backoff(attempt)                                 // 指数退避 + 抖动
  }
  return Result{}, lastErr

调用方：
  业务 Agent 的 Run 在 exec 返回错误时 → rules.Fallback() → 返回 (结果, ErrDegraded)
  任务标记 degraded=true 而非 failed
```

#### 成本控制

| 措施 | 效果 |
|---|---|
| 规则前置过滤（AC 自动机敏感词 + 正则 + 白名单） | 明确违规直接拦、信誉良好直接放行，**仅中间地带送模型** |
| 全局 + 单用户令牌桶 | 防异常调用 |
| 单任务/时间窗口 token 上限 | 超出即降级 |
| 结果缓存 | 相同输入复用 |

#### 检索融合（RRF）

```go
// score = Σ 1/(k + rank)，k 通常取 60
// 好处：免去向量分数与全文分数不同量纲的归一化问题
func fuseRRF(lists [][]Hit, k int) []Hit
```

#### 评测集

```
testdata/eval/
├── moderation_cases.jsonl    # 200 条：100 违规 + 100 正常，人工标注
├── search_cases.jsonl        # 100 条：query → 期望命中的 work_id 集合
└── README.md                 # 标注规范
```

**评测命令支持两种模式**：`--mock`（CI，零费用）/ `--real`（本地手动，产出简历数字）。

---

## 11. 风险与待决问题

### 11.1 实施风险

| 风险 | 对策 |
|---|---|
| 环境搭建超预期 | S3 作为硬卡点，先打通依赖再写业务；容器化一条命令启动 |
| 范围蔓延做不完 | 严格按裁剪顺序：**S8 可观测性 → 实时推送 → 社交细节 → 检索融合**；S5/S6/S7 不可裁 |
| 模型成本与不稳定 | 规则前置过滤、结果缓存、预算上限、**模拟实现**；开发期不依赖真实模型 |
| 并发量级未验证 | S8 压测实测；宁可数字保守，不空口宣称 |
| 追问超出准备 | §12 成稿；未实施部分**明确说"当前不做及理由"**，比含糊更可信 |
| 陈述沦为功能罗列 | 每条描述绑定一个技术决策 + 一个可量化结果，禁止"实现了 XX 功能"式表述 |
| 版本号过期（已发生） | 见 P-01/P-03/P-07：文档只写策略与核实命令 |
| **测试债务累积（已发生）** | **DEV-001**：S4 测试延后。三项必做测试须在进 S5 前补；全部测试须在 **S6 结束前**整合完毕（S5.5）。**不得拖到 S8/S9** |
| **门禁假通过（已发生）** | 无测试的包 `go test` 退出码为 0。T1 已移除，功能测试仍全缺。S5.5 前 `make check` 只能证明 config 包通过；架构边界需手动执行 `make verify-layers` |

### 11.2 待决问题

| 编号 | 问题 | 需谁决定 |
|---|---|---|
| Q-01 | 生产对象存储选 Cloudflare R2 / 阿里云 OSS / 腾讯云 COS | 开发者（S8 前） |
| Q-02 | Agent 用哪家模型 API（成本 vs 效果） | 开发者（S7 前） |
| Q-03 | embedding 模型与维度（须与 `vector(768)` 一致） | 开发者（S7 前） |
| Q-04 | 是否实现 WebSocket（当前优先级最低） | 开发者（S8 视余力） |
| Q-05 | 是否保留 `APP_REDIS_ADDR` 在 S3 阶段（当前未使用） | 开发者（无害，可保留） |

---

## 12. 面试口径

### 12.1 可写入简历的四条

每条须同时包含**技术决策**与**可量化结果**。

1. **架构与选型**
   > 基于 Go 构建分层后端，领域层与框架完全解耦（以自动化测试守护依赖边界）；以 PostgreSQL 单库承载关系存储、文档属性、全文检索、向量检索与任务队列六种角色，在减少组件的同时获得跨模块事务一致性。

2. **内容寻址的上传体系**
   > 实现秒传预检、分片并发上传、标准 S3 Multipart 合并与流式哈希校验、引用计数与延迟回收；重复文件零字节上传，存储去重率与上传吞吐具备压测数据支撑。

3. **依赖图与版本管理**
   > 设计版本与依赖模型，发布前在同一事务内完成依赖存在性、版本范围匹配与循环依赖校验；以拓扑排序检测环并**还原具体环路路径**反馈使用者。

4. **Agent 编排与工程化**
   > 设计"规划—执行—校验"三段式 Agent 运行时，支持步骤级并发调度、独立超时、失败重试与三级降级；落地内容审核、元数据生成与语义检索三项能力；自建标注评测集产出准确率、召回率与成本下降指标。

### 12.2 六个必答问题

| 问题 | 回答要点 |
|---|---|
| 为何选用 Go 而非 Python | 上传下载属高并发 I/O 场景，Go 的并发模型与内存效率更适配；Agent 侧为等待型任务，同样受益于轻量并发 |
| 为何用 PostgreSQL | 向量与业务数据同事务域，避免跨存储一致性缺口与额外组件运维；业务模型两库均可承载，迁移成本主要在 JSONB 查询 |
| 秒传如何防止哈希碰撞 | 哈希 + 文件大小双重校验；分片哈希仅用于传输幂等，不作为全局标识 |
| 循环依赖如何检测与定位 | 拓扑排序得出环内节点集合，再自任一环内节点出发还原完整路径 |
| 任务重复消费如何处置 | 应用层幂等检查以节省资源，数据库唯一约束以保证正确性，双层防护 |
| 模型不可用如何处置 | 限流、超时、重试、回炉，最终降级至规则引擎并标记为"降级"而非失败 |

### 12.3 扩展追问预案

| 追问 | 要点 |
|---|---|
| 为何不用 MinIO | 2025 年移除管理功能、停止发布镜像，2025-12 进入维护模式，2026 年仓库归档；选 SeaweedFS 因其 Apache 2.0、活跃维护。更重要的是**存储后端在设计中一直是可替换实现细节**，本地 SeaweedFS、生产可切云存储只改配置 |
| 为何用 River 而非 NSQ | NSQ 已停维护且需额外守护进程；River 后端在 PG，支持事务性入队，业务写入与任务投递同事务获得原子性，免去 outbox 表与补偿任务 |
| 为何用 SSE 而非 WebSocket | 场景分析显示九成实时需求为单向；SSE 走标准 HTTP、浏览器原生重连、网关友好。**主动选择不用 WS 本身即为判断力体现** |
| 为何不用微服务 | 周期内无法做扎实；追问"跨服务一致性怎么保证"风险高。只拆 Agent Worker，因其负载特征不同且共享同一份领域模型 |
| 连接池应设多大 | 并非越大越好。参考式 `连接数 ≈ CPU 核数 × (1 + 等待/处理时间)`，且总数受 `max_connections` 约束；盲目增大因上下文切换降低吞吐 |
| 步骤超时后如何清理 | `context` 逐层传播，底层网络调用收到取消信号并关闭连接；前提是所有阻塞调用都接受 `context` |
| 为何用 Kahn 而非三色标记 | 两者都能检测环；Kahn 副产物是"入选拓扑序的节点集合"，可直接定位环成员 |
| pgx 和 sqlc 是什么关系 | **不是替代关系**：pgx 是驱动+连接池（运行时），sqlc 是代码生成器（构建时），生成代码底层调用 pgx |
| 为什么不包装 logger | 包装丢失 `With`/`Enabled`/`Handler()`，接口注定不完整；测试用 `io.Discard` 静默即可 |

### 12.4 可主动展示的五个证据

1. 打开 `internal/domain`，展示**无任何框架依赖**，并运行 `TestDomainHasNoFrameworkDependency`
2. 说明**为何只拆一个服务**，以及为何不多拆
3. 展示评测指标表，证明 AI 模块**有量化效果**而非演示脚本
4. 说明**为何主动选择不用 WebSocket**
5. 展示依赖环错误的返回结构，说明错误信息如何做到**可操作**

### 12.5 陈述合格判据（四问自检）

任一技术表述若不能通过以下追问，视为不达标：

- 该选择的**替代方案**是什么？
- **放弃**替代方案的具体理由是什么？
- 该选择带来的**代价**是什么？
- 有无**可量化**的结果支撑？

---

## 13. 附录

### 13.1 术语表（保持陈述一致性）

| 建议表述 | 避免表述 |
|---|---|
| 内容寻址 / 以内容哈希为唯一标识 | "用哈希做文件名" |
| 事务性入队 | "消息队列保证不丢" |
| 环路径还原 | "检测到循环依赖" |
| 步骤级并发调度 | "并发跑 Agent" |
| 三级降级 | "加了容错" |
| 游标分页 | "分页查询" |
| 领域层零框架依赖 | "代码分层了" |
| 评测集与指标 | "效果还不错" |
| 标准 S3 Multipart | "服务端合并分片" |
| 预签名直传 | "上传走服务端转发" |

### 13.2 命令速查（按本项目实际工具配置）

> **本项目工具走 `go.mod` 的 `tool` 指令，必须用 `go tool <name>` 调用**，不能直接敲 `sqlc` / `goose` / `wire` / `golangci-lint`。

```powershell
# ── 环境 ──
cd deploy; docker compose up -d          # 启动三服务
docker compose ps                        # 查看健康状态
docker compose down -v                   # 清空数据（会丢数据）

# ── 代码生成（顺序固定：goose → sqlc → wire）──
$env:GOOSE_DRIVER   = "postgres"
$env:GOOSE_DBSTRING = "postgres://craftshelf:craftshelf_dev_pwd@localhost:5432/craftshelf?sslmode=disable"
go tool goose -dir migrations up
go tool sqlc generate
go tool wire ./cmd/api

# ── 构建与运行 ──
go build ./...
go run ./cmd/api

# ── 测试 ──
go test ./...                            # 当前仅覆盖 config 包
go test -race -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html

# ── 质量门禁 ──
gofmt -l .
go vet ./...
go tool golangci-lint run ./...          # v2 配置，当前 0 issues

# ── 分层验证（T1 已移除；手动执行）──
go list -deps ./internal/domain | Select-String "gin-gonic|jackc|google/uuid|sqlc|aws|redis"

# ── 依赖整理 ──
go mod tidy                              # P-16 已修正：godotenv 为直接依赖
```

**S3 连通性验证**

```powershell
$env:AWS_ACCESS_KEY_ID     = "craftshelf"
$env:AWS_SECRET_ACCESS_KEY = "craftshelf_dev_pwd"
$env:AWS_DEFAULT_REGION    = "us-east-1"

# 列出桶（应能看到 compose 创建的两个桶）
aws --endpoint-url http://localhost:8333 s3 ls
```

**git 状态（P-15 已于 2026-11-03 提交）**

```powershell
git status --short               # 预期无输出
git log --oneline -1             # 查看本轮修复提交
```

### 13.3 常见故障对照

| 症状 | 原因 | 处置 |
|---|---|---|
| CLI 命令找不到 | `%GOPATH%\bin` 未入 PATH | 加入后重开终端 |
| `go test -race` 报 `-race requires cgo` | `CGO_ENABLED=0` 或无 C 编译器 | 安装 MinGW-w64 后设 `CGO_ENABLED=1`，或在 Linux/CI 上跑（P-19） |
| 依赖下载停滞 | 未配 `GOPROXY` | 配置代理 |
| postgres 一直 unhealthy | `pg_isready` 未带 `-U`/`-d` | 补全参数 |
| 改初始化 SQL 不生效 | **数据卷已存在，脚本只跑一次** | `docker compose down -v` |
| 端口被占用 | 本机已有同名服务 | 设 `POSTGRES_PORT` 等环境变量 |
| seaweedfs unhealthy | 镜像无 `curl` | 换 healthcheck 命令 |
| **psql 能连但程序连不上** | host 用了 `localhost` 而程序在容器内 | 容器内用服务名 |
| `sqlc generate` 报找不到表 | 未先跑 `goose up` | 顺序：goose → sqlc |
| 生成字段是 `pgtype.Text` | 未开 `emit_pointers_for_null_types` | 打开配置重新生成 |
| 密码字段写入报错 | 迁移用了 `VARCHAR(<60)` | 改 `TEXT` |
| 令牌一直无效 | 密钥每次重启都变 | 从环境变量固定 |
| 请求 401 但令牌看起来正常 | 中间件未校验过期 | 检查 `exp` 校验 |
| S3 请求 403 | `s3.json` 凭据与客户端不一致 | 对齐两处 |
| **S3 报 `lookup xxx.localhost: no such host`** | 未开 `UsePathStyle` | 设 `UsePathStyle = true` |
| wire 报 `no provider found for X` | 缺 provider 或 `wire.Bind` | 从缺失类型倒推"该由谁产出" |
| `wire_gen.go` 重复定义 | 忘了 `wireinject` build tag | 检查 `wire.go` 首行 |
| 测试跑十几秒 | 用了生产 bcrypt cost | 测试用 `bcrypt.MinCost` |
| `golangci-lint` 配置报错 | v2 格式与 v1 不同 | `golangci-lint migrate` |

### 13.4 如何恢复工作（上下文丢失时的入口）

#### 当前精确状态（2026-11-03 按代码核对）

| 项 | 实际状态 |
|---|---|
| 仓库位置 | `F:\project\CraftShelf` |
| S3 功能 | ✅ 可运行（注册/登录/me/健康检查均已实现） |
| S4 工程化 | 🟡 wire、`.golangci.yml`、Makefile 已落地；架构守护测试已移除，auth/jwt/service 测试仍缺 |
| **S4 验收** | ❌ **未达成**：auth/jwt 测试从未编写、service 覆盖率未达标（DEV-001） |
| 测试 | 仅 `config_test.go`（2 个函数）；架构守护测试已移除 |
| 阻断 S5 的缺陷 | ✅ **无**（P-09/P-10/P-11 已于 2026-11-03 修复） |
| 技术债 | **DEV-001**：S5.5 从零编写全部测试，约 20.25 小时 |
| 下一个动作 | **进入 S5**：定义 `domain.Storage` + `storage.Memory`/`LocalFS`，再写单文件直传 |

> **⚠️ 恢复工作时的三条纪律**
>
> 1. **核心业务仍无测试**（service/repository/jwt）。`go test ./...` 只覆盖 config 包
> 2. **`.golangci.yml` 与 Makefile 已就位**；本机若未安装 make，直接运行底层 `go` 命令
> 3. **wire 装配代码已提交 git**，但 `wire_gen.go` 仍必须随 `wire.go` 变更重新生成并提交

### 13.4.1 进入 S5 前的三项测试投入（DEV-001 情况 A）

| 级别 | 内容 | 成本 | 保护的不变量 | 状态 |
|---|---|---|---|---|
| **T1 强烈建议** | `internal/domain/architecture_test.go` | 15 分钟 | 领域层零框架依赖 → **S5 引入 AWS SDK 时的最大架构风险** | ⬜ 已移除，S5.5 补 |
| **T2 建议** | T1 + `fakeUserRepo` + 两个安全断言 | 1.5 小时 | 口令不以明文落库；登录失败不区分原因（两者失效均**不报错**） | ⬜ 待做 |
| **T3 可延后** | S4 全部测试（jwt、表驱动全用例） | 4 小时 | 交付质量与覆盖率 → 由 S5.5 一并完成 | ⬜ 待做 |

**T1 已移除**：按用户要求统一测试，S5.5 再补；代码模板见 §9.3「方式 B」。

**T1/T2/T3 均延后**：从 S5 开始到 S5.5 结束期间，架构边界只有 `make verify-layers` 手动检查；口令流程与业务逻辑无任何自动化断言。

#### 开始新一轮工作前必读四节

1. **§2.1 阶段状态表** —— 确认做到哪
2. **§2.4 已知缺口** —— P-09/P-10/P-11 已修复；关注 DEV-001 与 P-01/P-02 等剩余项
3. **§2.5 文档修正记录** —— 了解上一版文档有哪些错误（避免再次被误导）
4. **§13.3 常见故障** —— 报错先查这张表

#### S5 启动检查清单（按实际命令修正）

进入 S5 之前必须确认：

- [x] 三个容器均 `healthy`：`docker compose -f deploy/docker-compose.yml ps`
- [x] S3 协议连通：`s3.bucket.list` 能列出桶
- [x] **桶名已与 `config.go` 对齐**（P-09，2026-11-03）
- [x] 迁移已应用：`go tool goose -dir migrations status` 显示 `20261027151811` 已应用
- [x] `go tool sqlc generate` 无报错
- [x] `go tool wire ./cmd/api` 无报错，`go build ./...` 通过
- [x] `go test ./...` 通过（仅 config 包）
- [x] `go run ./cmd/api` 后五个端点行为正确（见 §2.2 S3）——healthz/readyz/register/login/me 已点测
- [x] **`Login` 的 nil panic 已修**（P-10）——临时单测验证仓储故障时返回 error，不 panic

#### S5 的第一步（已修正）

~~补 `config.go` 的 S3 字段解析（P-04）~~ → **已不需要，config 已完整**。

正确的第一步顺序：

| 序 | 动作 | 原因 | 状态 |
|---|---|---|---|
| 1 | **修 P-09 桶名不一致** | 不修则上传必失败于 `NoSuchBucket` | ✅ 2026-11-03 |
| 2 | **修 P-10 `Login` nil panic** | 安全缺陷，且 S5 会改用户模型 | ✅ 2026-11-03 |
| 3 | **重建 T1 架构守护测试** | S5 引入 AWS SDK，防污染 | ⬜ 已移除，S5.5 补；当前用 `make verify-layers` |
| 4 | 空 Makefile 填入最小目标（P-11） | 后续所有命令都靠它 | ✅ 2026-11-03 |
| 5 | **定义 `domain.Storage` 接口 + `LocalFS` 实现** | 先接口后实现，便于测试 | **下一步** |
| 6 | 写 `storage.Memory` 假实现 | 让 service 测试不依赖容器 | ⬜ |
| 7 | 接 AWS SDK v2（配 `UsePathStyle`） | 真实实现 | ⬜ |

**顺序原则不变**：自底向上，每层用上一层验证。

#### 文档维护约定

| 时机 | 更新哪节 |
|---|---|
| 完成一个阶段 | §2.1 状态表；把该阶段从 §2.3 移到 §2.2 |
| 新增技术决策 | §7 增加 ADR 条目 |
| 发现不一致 | §2.4 增加 P-xx 或 Q-xx |
| 阶段验收后 | 在 §2.2 补充实测数据与遗留问题 |

### 13.5 项目文件索引（按角色）

| 角色 | 文件 |
|---|---|
| 配置入口 | `internal/pkg/config/config.go`、`.env.example` |
| 装配入口 | `cmd/api/main.go`、`cmd/api/providers.go`、`cmd/api/wire.go` |
| 业务核心 | `internal/service/auth.go` |
| 数据模型 | `internal/domain/user.go`、`migrations/*.sql` |
| SQL | `internal/repository/queries/*.sql`、`sqlc.yaml` |
| 接口契约 | `internal/handler/router.go`、`internal/handler/dto/*` |
| 分层检查 | `make verify-layers`（T1 测试已移除，S5.5 补） |
| 测试范例 | `internal/service/auth_test.go`、`internal/pkg/jwt/jwt_test.go` |
| 环境 | `deploy/docker-compose.yml`、`deploy/postgres/init/*`、`deploy/seaweedfs/s3.json` |
| 质量门禁 | `.golangci.yml`、`Makefile` |

---

**文档结束**

> 维护提示：本文档随阶段推进更新。若与代码冲突，**以代码为准并修正文档**，同时在 §2.4 记录冲突原因。
