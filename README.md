<h1 align="center">Welcome to CertMagic-COS</h1>

<p align="center">
    <img src="https://goreportcard.com/badge/github.com/yikotee/certmagic-cos" />
    <img src="https://godoc.org/github.com/yikotee/certmagic-cos?status.svg" />
    <a href="https://github.com/yikotee/certmagic-cos/README.md">
    <img src="https://img.shields.io/badge/Docs-使用文档-blue?style=flat-square&logo=readthedocs" alt="Docs" /></a>
    <a href="https://github.com/yikotee/certmagic-cos/LICENSE.md">
    <img src="https://img.shields.io/badge/License-MIT-green?style=flat-square&logo=github" alt="License" /></a>
</p>
<h3 align="center">一款简洁的腾讯云COS连接工具 </h3>
<p>现有的 Caddy 插件：<a href="https://github.com/ss098/certmagic-s3">CertMagic-S3</a> 仅支持 path-style 域名格式，无法直接用于腾讯云 COS（path-style 已被腾讯云弃用，见<a href="https://cloud.tencent.com/document/product/436/96243">腾讯云COS桶安全通知</a>。本项目提供 CertMagic 的 Storage 接口，用于连接到腾讯云 COS，证书的申请与续期仍由 CertMagic 负责，方便于 SSL 证书安全存储于 COS，并支持跨实例共享与分布式续期锁。<p>


## ✨ 特性

- ✅ 提供 `CertMagic` 的 `Storage` 接口，持久化腾讯云 `COS`
- ✅ 兼容 腾讯云 `COS` `virtual-hosted-style` 域名地址格式
- ✅ 与 `CertMagic` 无缝配合，支持证书的自动申请与续期（由 `CertMagic` 负责）
- ✅ 可作为独立 `Go` 库使用并集成到 `Caddy`（插件）
- ✅ 支持多实例分布式部署（分布式锁）

## 📦 安装

### 作为 Go 库

```bash
go get github.com/yikotee/certmagic-cos
```

### 作为 Caddy 插件

使用 [xcaddy](https://github.com/caddyserver/xcaddy) 构建：

```bash
xcaddy build --with github.com/yikotee/certmagic-cos
```

## 🚀 快速开始

### 方式一：Go 库

```go
package main

import (
    "log"
    "net/http"
    "github.com/yikotee/certmagic-cos/cos"
    "github.com/caddyserver/certmagic"
)

func main() {
    // 创建 COS 存储
    storage, err := cos.NewStorage(cos.Config{
        Bucket:    "your-bucketName",
        Region:    "ap-xxx",
        SecretID:  "your-secret-id",
        SecretKey: "your-secret-key",
        Prefix:    "certmagic",
    })
    if err != nil {
        log.Fatal(err)
    }

    // 配置 CertMagic
    certmagic.Default.Storage = storage
    certmagic.DefaultACME.Email = "admin@example.com"
    certmagic.DefaultACME.Agreed = true

    // 你的业务逻辑
    mux := http.NewServeMux()
    mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte("Hello HTTPS!"))
    })

    // 启动 HTTPS 服务器（自动申请证书）
    log.Println("启动 HTTPS 服务器...")
    if err := certmagic.HTTPS([]string{"example.com"}, mux); err != nil {
        log.Fatal(err)
    }
}
```

### 方式二：Caddy 插件

#### Caddyfile 配置

```caddyfile
{
    storage cos {
        bucket your-bucketName
        region ap-xxx
        secret_id your-secret-id
        secret_key your-secret-key
        prefix certmagic
    }
}

example.com {
    respond "Hello HTTPS!"
}
```

#### JSON 配置

```json
{
  "apps": {
    "http": {
      "servers": {
        "srv0": {
          "listen": [":443"],
          "routes": [{
            "match": [{"host": ["example.com"]}],
            "handle": [{
              "handler": "static_response",
              "body": "Hello HTTPS!"
            }]
          }]
        }
      }
    },
    "tls": {
      "automation": {
        "policies": [{
          "subjects": ["example.com"],
          "storage": {
            "module": "cos",
            "bucket": "your-bucketName",
            "region": "ap-xxx",
            "secret_id": "your-secret-id",
            "secret_key": "your-secret-key",
            "prefix": "certmagic"
          }
        }]
      }
    }
  }
}
```

## ⚙️ 参数说明

| 参数         | 说明                                                         | 示例              | 必填 |
| ------------ | ------------------------------------------------------------ | ----------------- | ---- |
| `subjects`   | 域名                                                         | `example.com`     | ✅    |
| `bucket`     | COS 存储桶名称（含 APPID）                                   | `cert-1234567890` | ✅    |
| `region`     | 腾讯云地域，ap-xxx，见：[更多地域](https://cloud.tencent.com/document/product/436/6224) | `ap-nanjing`      | ✅    |
| `secret_id`  | 腾讯云 API 密钥 ID                                           | -                 | ✅    |
| `secret_key` | 腾讯云 API 密钥 Key                                          | -                 | ✅    |
| `prefix`     | 存储路径前缀                                                 | `certmagic`       | ❌    |

## 🛠️ COS 存储桶设置

### 1. 创建存储桶

1. 登录 [腾讯云 COS 控制台](https://console.cloud.tencent.com/cos)
2. 点击"创建存储桶"
3. 填写配置：
   - **名称**：自定义（如 `cert`）
   - **地域**：选择与服务器相同或相近的地域
   - **访问权限**：私有读写
4. 创建完成后，记录完整的存储桶名称（含 APPID），如 `cert-1234567890`

### 2. 获取 API 密钥

1. 访问 [API 密钥管理](https://console.cloud.tencent.com/cam/capi)
2. 点击"新建密钥"
3. 保存 `SecretId` 和 `SecretKey`

⚠️ **安全提示**：请勿将密钥硬编码在代码中，建议使用环境变量或密钥管理服务。

## 📝 工作流程

```
┌─────────────┐
│  域名请求    │
└──────┬──────┘
       │
       ▼
┌─────────────┐
│ CertMagic   │ ← 自动申请/续期证书
└──────┬──────┘
       │
       ▼
┌─────────────┐
│  Storage    │ ← 你的实现
│  接口       │
└──────┬──────┘
       │
       ▼
┌─────────────┐
│ 腾讯云 COS  │ ← 证书存储
└─────────────┘
```

**证书生命周期管理：**

1. **首次启动**：CertMagic 检测无证书，向 Let's Encrypt 申请
2. **验证域名**：通过 HTTP-01 或 TLS-ALPN-01 验证
3. **保存证书**：调用 `Storage.Store()` 保存到 COS
4. **后续启动**：调用 `Storage.Load()` 从 COS 读取证书
5. **自动续期**：证书到期前 30 天自动续期

## 🔒 分布式锁

支持多实例部署，通过 COS 实现分布式锁：

```go
// 自动处理并发申请
// 实例 A 获取锁 → 申请证书 → 释放锁
// 实例 B 等待锁 → 读取已申请的证书
```

## 💡 使用场景

### 场景 1：微服务集群

```go
// 服务 A
storage, _ := cos.NewStorage(cos.Config{...})
certmagic.HTTPS([]string{"api.example.com"}, handlerA)

// 服务 B
certmagic.HTTPS([]string{"admin.example.com"}, handlerB)

// 共享同一个 COS 存储的证书池
```

### 场景 2：负载均衡多实例

```
┌──────────┐
│ 实例 1    │ ─┐
└──────────┘  │
              ├─→ 共享 COS 证书存储
┌──────────┐  │
│ 实例 2    │ ─┤
└──────────┘  │
              │
┌──────────┐  │
│ 实例 N    │ ─┘
└──────────┘
```

### 场景 3：通用 COS 存储库

```go
// 不仅用于证书，还可以存储其他文件
storage, _ := cos.NewStorage(cos.Config{...})

// 存储配置
storage.Store(ctx, "config/app.json", configData)

// 存储文件
storage.Store(ctx, "uploads/file.pdf", fileData)
```

## 📌 注意事项

1. **域名解析**：域名必须解析到运行程序的服务器
2. **端口开放**：服务器需要开放 80 和 443 端口
3. **测试环境**：建议先使用 Let's Encrypt 测试环境避免速率限制

```go
// 使用测试环境
certmagic.DefaultACME.CA = certmagic.LetsEncryptStagingCA
```

4. **生产环境**：确认测试通过后再切换到生产环境

```go
// 生产环境（默认）
certmagic.DefaultACME.CA = certmagic.LetsEncryptProductionCA
```

## 🤝 贡献

欢迎提交 Issue 和 Pull Request！

## 🔗 相关项目

- [CertMagic](https://github.com/caddyserver/certmagic)
- [CertMagic-S3](https://github.com/ss098/certmagic-s3)
- [Caddy](https://github.com/caddyserver/caddy)
- [腾讯云 COS Go SDK](https://github.com/tencentyun/cos-go-sdk-v5)

---

如果这个项目对你有帮助，欢迎 ⭐ Star 支持！