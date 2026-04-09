# Docker 镜像同步工具

📦 使用 GitHub Actions 将 Docker 镜像从 DockerHub 同步到阿里云容器镜像服务（ACR）

---

## ✨ 功能特性

- 🚀 **Skopeo 直接复制** - 无需本地存储，流式传输
- 🎯 **多架构支持** - 支持 AMD64、ARM64 等多种架构
- 🔄 **智能去重** - 跨云自动检测，避免重复拉取
- 📧 **邮件通知** - 同步完成后自动发送详细报告
- 💪 **自动重试** - 网络错误自动重试，最多 3 次
- ⚡ **并发同步** - 支持多个镜像并行处理
- 🎨 **进度显示** - 实时显示同步进度

---

## 📝 使用说明

### 1. 配置 Secrets

在 GitHub 仓库的 Settings → Secrets and variables → Actions 中添加以下机密信息：

**阿里云配置（必需）：**

| 机密名称 | 说明 | 示例 |
|---------|------|------|
| `ALIYUN_NAME_SPACE` | 阿里云命名空间 | `my-namespace` |
| `ALIYUN_REGISTRY_USER` | 阿里云用户名 | `myusername` |
| `ALIYUN_REGISTRY_PASSWORD` | 阿里云密码 | `mypassword` |
| `ALIYUN_REGISTRY` | 阿里云仓库地址 | `registry.cn-hangzhou.aliyuncs.com` |

**邮件通知（可选）：**

| 机密名称 | 说明 | 示例 |
|---------|------|------|
| `EMAIL_USERNAME` | 163 邮箱账号 | `example@163.com` |
| `EMAIL_PASSWORD` | 163 邮箱授权码 | `ABC123DEF` |

### 2. 添加镜像列表

编辑 `images.txt` 文件，按以下格式添加镜像：

```txt
# Docker镜像列表
# 格式说明：按云商分组，每行一个镜像地址
# 支持的云商: aliyun

[aliyun]
jgraph/drawio:latest
corentinth/it-tools:latest
fnsys/dockhand:latest
```

### 3. 触发同步

进入 GitHub 仓库的 Actions 页面，手动触发 "Sync Docker Images" 工作流。

---

## � 邮件通知格式

同步完成后，您将收到包含以下信息的邮件：

```
Docker镜像同步任务已完成

提供商: 阿里云 (aliyun)
时间: 2026-04-09 08:30:00

总计: 3 个镜像
- 已同步: 0
- 已跳过: 3
- 失败: 0

[已跳过]
  ⏭ docker.io/jgraph/drawio:latest
    registry.cn-hangzhou.aliyuncs.com/my-namespace/drawio
  ⏭ docker.io/corentinth/it-tools:latest
    registry.cn-hangzhou.aliyuncs.com/my-namespace/it_tools
  ⏭ docker.io/fnsys/dockhand:latest
    registry.cn-hangzhou.aliyuncs.com/my-namespace/dockhand
```

---

## ⚙️ 高级配置

### 环境变量

| 变量名 | 默认值 | 说明 |
|-------|--------|------|
| `LOG_LEVEL` | `INFO` | 日志级别 (DEBUG/INFO/WARN/ERROR) |
| `TIMEOUT` | `300` | 同步超时时间（秒） |
| `MAX_RETRIES` | `3` | 失败重试次数 |
| `CONCURRENCY` | `3` | 并发同步数量 |

---

## 📄 开源协议

本项目基于 MIT 协议开源。

## 🙏 致谢

感谢以下开源项目的启发：
- [docker_image_pusher](https://github.com/tech-shrimp/docker_image_pusher) by 技术爬爬虾
