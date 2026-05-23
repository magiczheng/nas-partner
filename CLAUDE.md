# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

NAS Partner — NAS 后台管理系统。前后端分离架构：
- **frontend/**: React 19 + TypeScript + Vite 8 + Ant Design 6 + Tailwind CSS v4
- **backend/**: Go 1.25 + Gin + SQLite (modernc.org/sqlite) + JWT + bcrypt

## Commands

### Backend

```bash
cd backend && go run cmd/server/main.go
```

```bash
cd backend && go build ./...
```

```bash
cd backend && go mod tidy
```

### Frontend

```bash
cd frontend && npx vite
```

```bash
cd frontend && npx tsc --noEmit
```

```bash
cd frontend && ./node_modules/.bin/vite build
```

### Full stack (dev)

```bash
./dev.sh          # 正常启动
./dev.sh -c       # 清除数据库后启动（重新初始化）
```

## Architecture

### Auth Flow

App.tsx 管理 `AppState` = `'loading' | 'init' | 'login' | 'ready'`：

1. **loading**: 初始加载中，显示 Spin
2. **init**: 数据库无用户 → 渲染 InitPage（创建管理员账号）
3. **login**: 有用户但无有效 token → 渲染 LoginPage
4. **ready**: 已登录 → 渲染 AdminLayout + 子页面

状态切换：App.tsx 把 `onComplete` 回调传给 InitPage/LoginPage，子页面调用后 `setState` 触发路由重定向。详见 App.tsx 中路由守卫逻辑（每个路由检查 state，不匹配则 Navigate 重定向）。

### Backend API

所有 API 在 `/api` 下：

| 方法 | 路径 | 说明 | 认证 |
|------|------|------|------|
| GET | /health | 健康检查 | 无 |
| GET | /auth/status | 是否已初始化 | 无 |
| POST | /auth/init | 创建管理员账号 | 无 |
| POST | /auth/login | 登录，返回 JWT | 无 |
| GET | /me | 当前用户信息 | Bearer Token |

JWT token 有效期 7 天，存 localStorage。受保护路由用 `middleware.AuthRequired` 中间件。

### Backend Structure

```
backend/
├── cmd/server/main.go       — 入口
├── internal/
│   ├── config/config.go     — 环境变量配置
│   ├── database/db.go       — SQLite 初始化 + 自动建表
│   ├── handler/             — HTTP handlers（auth.go, health.go, me.go）
│   ├── middleware/           — CORS, JWT 认证
│   ├── model/user.go        — 数据模型 + 请求/响应结构体
│   └── router/router.go     — 路由注册
└── .env                     — 环境变量
```

### Frontend Structure

```
frontend/
├── src/
│   ├── api/
│   │   ├── client.ts        — fetch 封装，自动带 Authorization header
│   │   └── auth.ts          — 认证相关 API 调用
│   ├── components/
│   │   └── Layout.tsx       — AdminLayout（侧边栏 + 顶栏 + 内容区）
│   ├── pages/
│   │   ├── Home.tsx         — 控制台首页
│   │   ├── InitPage.tsx     — 首次初始化页
│   │   └── LoginPage.tsx    — 登录页
│   ├── App.tsx              — 路由 + 状态管理
│   ├── main.tsx             — 入口
│   └── index.css            — Tailwind + 基础样式
└── vite.config.ts           — Vite + Tailwind + API 代理配置
```

### Frontend Routes

- `/init` — 初始化管理员账号（仅在 state === 'init' 时可访问）
- `/login` — 登录（仅在 state === 'login' 时可访问）
- `/` — 后台首页（AdminLayout 包裹，仅在 state === 'ready' 时可访问）

### Config

环境变量配置（`backend/.env`）:

| 变量 | 默认值 | 说明 |
|------|--------|------|
| SERVER_PORT | 8080 | 后端端口 |
| DB_PATH | ./data/nas-partner.db | SQLite 文件路径 |
| JWT_SECRET | change-me-in-production | JWT 签名密钥 |

前端通过 Vite proxy 将 `/api` 请求代理到后端 `localhost:8080`。

### Notes

- SQLite 使用纯 Go 实现（modernc.org/sqlite），无需 CGo
- 前端使用 npmmirror 镜像加速安装：`npm install --registry=https://registry.npmmirror.com`
- Tailwind v4 通过 Vite 插件集成，不是独立 CLI
