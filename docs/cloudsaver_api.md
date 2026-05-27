# CloudSaver API 接口文档

> 来源项目: [quark-auto-save](https://github.com/Cp0219/quark-auto-save)
> SDK 文件: `app/sdk/cloudsaver.py`

## 概述

CloudSaver 是一个云盘资源聚合搜索引擎，提供用户认证和资源搜索能力。采用 JWT Token 认证机制，支持自动登录续期。

## 基础配置

| 配置项 | 类型 | 说明 |
|--------|------|------|
| `server` | string | CloudSaver 服务地址（如 `http://localhost:8080`） |
| `username` | string | 登录用户名 |
| `password` | string | 登录密码 |
| `token` | string | JWT Token（登录后自动获取） |

## API 端点

### 1. 用户登录

**POST** `/api/user/login`

获取 JWT Token，后续所有请求需携带此 Token。

**请求头:**
```
Content-Type: application/json
```

**请求体:**
```json
{
  "username": "your_username",
  "password": "your_password"
}
```

**成功响应:**
```json
{
  "success": true,
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
  }
}
```

**失败响应:**
```json
{
  "success": false,
  "message": "CloudSaver登录用户名或密码错误"
}
```

### 2. 搜索资源

**GET** `/api/search`

搜索云盘分享资源。

**请求头:**
```
Authorization: Bearer <token>
```

**查询参数:**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `keyword` | string | 是 | 搜索关键词 |
| `lastMessageId` | string | 否 | 上一条消息 ID，用于分页翻页 |

**请求示例:**
```
GET /api/search?keyword=黑镜&lastMessageId=
```

**成功响应:**
```json
{
  "success": true,
  "data": [
    {
      "channelId": "channel_001",
      "list": [
        {
          "title": "名称：黑镜 第七季",
          "content": "描述：黑镜第七季全集 链接：https://pan.quark.cn/s/xxx 标签：#科幻",
          "pubDate": "2025-01-15T10:30:00+08:00",
          "tags": ["科幻", "美剧"],
          "cloudLinks": [
            {
              "cloudType": "quark",
              "link": "https://pan.quark.cn/s/abc123"
            },
            {
              "cloudType": "alipan",
              "link": "https://www.alipan.com/s/xyz789"
            }
          ]
        }
      ]
    }
  ]
}
```

**失败响应:**
```json
{
  "success": false,
  "message": "无效的 token"
}
```

## 认证机制

### Token 自动续期

SDK 提供 `auto_login_search` 方法，封装了 Token 过期自动重新登录的逻辑：

1. 先尝试使用现有 Token 搜索
2. 如果返回 `无效的 token` 或 `未提供 token` 错误
3. 自动调用登录接口获取新 Token
4. 使用新 Token 重新搜索
5. 返回结果中携带 `new_token` 字段，调用方应持久化保存

```python
result = cs.auto_login_search("关键词")
if result.get("new_token"):
    # 持久化保存新 token
    save_token(result["new_token"])
```

## 搜索结果清洗

CloudSaver 返回的原始数据结构较复杂，SDK 提供 `clean_search_results` 方法进行标准化处理：

**处理逻辑:**
1. 遍历所有 channel → item → cloudLinks
2. 仅保留 `cloudType == "quark"` 的链接
3. 从 `title` 中提取纯标题（去除"名称："、"标题："前缀）
4. 从 `content` 中提取纯描述（去除"描述："、"简介："前缀，去除 HTML 标签）
5. 将 `pubDate` 转换为 CST 时间格式 `YYYY-MM-DD HH:MM:SS`
6. 按链接去重

**输出格式:**
```json
{
  "shareurl": "https://pan.quark.cn/s/abc123",
  "taskname": "黑镜 第七季",
  "content": "黑镜第七季全集",
  "datetime": "2025-01-15 10:30:00",
  "tags": ["科幻", "美剧"],
  "channel": "channel_001",
  "source": "CloudSaver"
}
```

## 错误处理

| 错误消息 | 含义 | 处理建议 |
|----------|------|----------|
| `CloudSaver未设置用户名或密码` | 认证信息缺失 | 检查配置 |
| `无效的 token` | Token 已过期 | 调用 login 重新获取 |
| `未提供 token` | 请求未携带 Token | 先登录获取 Token |
| `CloudSaver登录xxx` | 登录失败 | 检查用户名密码 |

## 在 quark-auto-save 中的调用流程

```
用户搜索 → /task_suggestions API
                ↓
        读取 config.source.cloudsaver 配置
                ↓
        检查 server/username/password 是否配置
                ↓
        创建 CloudSaver 实例并设置认证
                ↓
        调用 auto_login_search(keyword)
                ↓
        清洗结果 clean_search_results()
                ↓
        与其他搜索源合并、去重、排序
                ↓
        返回统一格式结果
```
