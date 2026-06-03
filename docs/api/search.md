# 资源搜索 API

## 概述

资源搜索模块提供云盘资源聚合搜索能力，支持 CloudSaver 和 PanSou 两个搜索源。用户可以通过关键词搜索资源，并一键创建转存任务。

## API 端点

### 1. 搜索资源

**GET** `/api/search`

搜索云盘分享资源。

**查询参数:**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `q` | string | 是 | 搜索关键词 |
| `source` | array | 否 | 指定搜索源（如 `cloudsaver`、`pansou`），可多选 |
| `page` | int | 否 | 页码，默认 `1` |

**请求示例:**
```
GET /api/search?q=黑镜&source=cloudsaver&source=pansou&page=1
```

**成功响应:**
```json
{
  "code": 0,
  "data": {
    "items": [
      {
        "title": "黑镜 第七季",
        "url": "https://pan.quark.cn/s/abc123",
        "source": "CloudSaver",
        "platform": "quark",
        "summary": "黑镜第七季全集",
        "updated_at": "2025-01-15 10:30:00",
        "tags": ["科幻", "美剧"],
        "channel": "channel_001"
      }
    ],
    "total": 10,
    "page": 1
  }
}
```

**错误响应:**
```json
{
  "code": 400,
  "message": "请提供搜索关键词"
}
```

### 2. 列出搜索源

**GET** `/api/search/sources`

列出所有可用的搜索源。

**成功响应:**
```json
{
  "code": 0,
  "data": ["CloudSaver", "PanSou"]
}
```

### 3. 获取搜索配置

**GET** `/api/search/config`

获取搜索源配置（密码和 Token 脱敏显示）。

**成功响应:**
```json
{
  "code": 0,
  "data": {
    "cloudsaver": {
      "server": "http://localhost:8080",
      "username": "admin",
      "password": "***",
      "token": "***"
    },
    "pansou": {
      "server": "https://so.252035.xyz"
    }
  }
}
```

### 4. 更新搜索配置

**PUT** `/api/search/config`

更新搜索源配置，配置会持久化到 Setting 表并立即生效。

**请求体:**
```json
{
  "cloudsaver": {
    "server": "http://localhost:8080",
    "username": "admin",
    "password": "your_password"
  },
  "pansou": {
    "server": "https://so.252035.xyz"
  }
}
```

**成功响应:**
```json
{
  "code": 0,
  "message": "配置已更新"
}
```

**错误响应:**
```json
{
  "code": 400,
  "message": "请求参数错误"
}
```

## 搜索源说明

### CloudSaver

- **认证方式**: JWT Token（用户名/密码登录获取）
- **特性**: 支持自动登录续期，搜索结果自动清洗
- **配置项**: `server`、`username`、`password`
- **详细文档**: [CloudSaver API](../cloudsaver_api.md)

### PanSou

- **认证方式**: 无需认证
- **特性**: 支持按网盘类型过滤，结果合并去重
- **配置项**: `server`
- **详细文档**: [PanSou API](../pansou_api.md)

## 搜索结果字段说明

| 字段 | 类型 | 说明 |
|------|------|------|
| `title` | string | 任务名称/标题 |
| `url` | string | 分享链接 |
| `source` | string | 数据源标识（`CloudSaver` 或 `PanSou`） |
| `platform` | string | 平台（`quark`） |
| `summary` | string | 描述内容 |
| `updated_at` | string | 发布时间（CST 格式 `YYYY-MM-DD HH:MM:SS`） |
| `tags` | array | 标签（仅 CloudSaver） |
| `channel` | string | 来源频道 |

### 5. 验证链接有效性

**GET** `/api/search/validate`

验证分享链接是否有效（自动识别夸克/移动云盘平台）。

**查询参数:**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `url` | string | 是 | 分享链接 URL |

**成功响应:**
```json
{
  "code": 0,
  "data": {
    "valid": true,
    "message": "链接有效"
  }
}
```

**失效链接响应:**
```json
{
  "code": 0,
  "data": {
    "valid": false,
    "message": "链接已过期"
  }
}
```

## 预定义魔法匹配规则

**GET** `/api/magic_patterns`

获取所有预定义的正则匹配规则，任务中可直接用 `$名称` 引用。

**成功响应:**
```json
{
  "code": 0,
  "data": {
    "$TV": {
      "pattern": "(?i).*?([Ss]\\d{1,2})?...",
      "replacement": "$1E$2.$3",
      "description": "剧集标准化命名 (S01E01.mp4)"
    },
    "$BLACK_WORD": {
      "pattern": "^(?!.*纯享)(?!.*加更)...",
      "replacement": "$0",
      "description": "黑名单过滤"
    },
    "$SHOW_MAGIC": { "..." },
    "$TV_MAGIC": { "..." }
  }
}
```

## 错误码

| 错误码 | 说明 |
|--------|------|
| 0 | 成功 |
| 400 | 请求参数错误 |
| 500 | 服务器内部错误 |

## 使用示例

### 搜索资源

```bash
curl "http://localhost:8080/api/search?q=哪吒&source=cloudsaver&source=pansou"
```

### 获取搜索配置

```bash
curl "http://localhost:8080/api/search/config"
```

### 更新搜索配置

```bash
curl -X PUT "http://localhost:8080/api/search/config" \
  -H "Content-Type: application/json" \
  -d '{
    "cloudsaver": {
      "server": "http://localhost:8080",
      "username": "admin",
      "password": "your_password"
    },
    "pansou": {
      "server": "https://so.252035.xyz"
    }
  }'
```
