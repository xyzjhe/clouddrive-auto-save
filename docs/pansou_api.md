# PanSou API 接口文档

> 来源项目: [quark-auto-save](https://github.com/Cp0219/quark-auto-save)
> SDK 文件: `app/sdk/pansou.py`

## 概述

PanSou（盘搜）是一个云盘资源搜索引擎，提供免认证的资源搜索能力。支持按网盘类型过滤，返回合并去重后的结果。

**默认服务地址:** `https://so.252035.xyz`

## 基础配置

| 配置项 | 类型 | 说明 |
|--------|------|------|
| `server` | string | PanSou 服务地址 |

> PanSou 无需认证，仅需配置服务地址即可使用。

## API 端点

### 1. 搜索资源

**GET** `/api/search`

搜索云盘分享资源，支持按网盘类型过滤和结果合并。

**查询参数:**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `kw` | string | 是 | 搜索关键词 |
| `cloud_types` | array | 否 | 网盘类型过滤，如 `["quark"]` |
| `res` | string | 否 | 结果模式，`merge` 表示合并去重 |
| `refresh` | boolean | 否 | 是否刷新缓存，默认 `false` |

**请求示例:**
```
GET /api/search?kw=哪吒&cloud_types=["quark"]&res=merge&refresh=false
```

**成功响应:**
```json
{
  "code": 0,
  "data": {
    "merged_by_type": {
      "quark": [
        {
          "url": "https://pan.quark.cn/s/abc123",
          "note": "哪吒之魔童降世【简介】国产动画电影巅峰之作",
          "datetime": "2025-01-15T10:30:00+08:00",
          "source": "channel_name"
        },
        {
          "url": "https://pan.quark.cn/s/def456",
          "note": "哪吒2之魔童闹海 4K高清",
          "datetime": "2025-02-20T14:00:00+08:00",
          "source": "another_channel"
        }
      ]
    }
  }
}
```

**失败响应:**
```json
{
  "code": -1,
  "message": "搜索失败"
}
```

> 当请求失败或异常时，SDK 直接返回空列表 `[]`，不抛出异常。

## 搜索结果格式化

PanSou 返回的 `note` 字段包含标题和描述，SDK 提供 `format_search_results` 方法进行结构化提取：

**解析规则:**

使用正则表达式从 `note` 字段中分离标题和描述：
```python
pattern = r'^(.*?)(?:[【\[]?(?:简介|介绍|描述)[】\]]?[:：]?)(.*)$'
```

- 匹配成功：`group(1)` 为标题，`group(2)` 为描述
- 匹配失败：整个 `note` 作为标题，描述为空

**处理逻辑:**
1. 解析 `note` 字段，分离标题和描述
2. 将 `datetime` 转换为 CST 时间格式 `YYYY-MM-DD HH:MM:SS`
3. 按 URL 去重（跳过重复链接）

**输出格式:**
```json
{
  "shareurl": "https://pan.quark.cn/s/abc123",
  "taskname": "哪吒之魔童降世",
  "content": "国产动画电影巅峰之作",
  "datetime": "2025-01-15 10:30:00",
  "channel": "channel_name",
  "source": "PanSou"
}
```

## 与 CloudSaver 输出格式对比

| 字段 | CloudSaver | PanSou | 说明 |
|------|-----------|--------|------|
| `shareurl` | ✅ | ✅ | 分享链接 |
| `taskname` | ✅ | ✅ | 任务名称/标题 |
| `content` | ✅ | ✅ | 描述内容 |
| `datetime` | ✅ | ✅ | 发布时间（CST） |
| `tags` | ✅ | ❌ | 标签（仅 CloudSaver） |
| `channel` | ✅ | ✅ | 来源频道 |
| `source` | `"CloudSaver"` | `"PanSou"` | 数据源标识 |

## 在 quark-auto-save 中的调用流程

```
用户搜索 → /task_suggestions API
                ↓
        读取 config.source.pansou 配置
                ↓
        检查 server 是否配置
                ↓
        创建 PanSou 实例
                ↓
        调用 search(keyword, refresh)
                ↓
        格式化结果 format_search_results()
                ↓
        与其他搜索源合并、去重、排序
                ↓
        返回统一格式结果
```

## 多源聚合搜索架构

quark-auto-save 的 `/task_suggestions` 端点使用 `ThreadPoolExecutor` 并发调用三个搜索源：

```
                    ┌─────────────┐
                    │  /task_suggestions  │
                    └──────┬──────┘
                           │
              ┌────────────┼────────────┐
              ▼            ▼            ▼
        ┌──────────┐ ┌──────────┐ ┌──────────┐
        │ net_search│ │cs_search │ │ps_search │
        │ (远程API) │ │(CloudSaver)│ │ (PanSou) │
        └────┬─────┘ └────┬─────┘ └────┬─────┘
             │            │            │
             └────────────┼────────────┘
                          ▼
                   ┌──────────────┐
                   │  合并 & 去重  │
                   │  按时间排序   │
                   └──────┬───────┘
                          ▼
                   ┌──────────────┐
                   │  返回 JSON   │
                   └──────────────┘
```

**并发策略:** `max_workers=3`，三个搜索源并行请求，任一失败不影响其他源。

**去重规则:** 按 `shareurl` 字段去重，保留先出现的结果。

**排序规则:** 按 `datetime` 字段降序排列（最新优先）。
