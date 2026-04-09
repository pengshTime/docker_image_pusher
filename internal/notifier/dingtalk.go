package notifier

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"
)

// DingTalkNotifier 钉钉推送通知器
type DingTalkNotifier struct {
	token   string
	enabled bool
}

// DingTalkMessage 钉钉消息结构
type DingTalkMessage struct {
	MsgType  string           `json:"msgtype"`
	Markdown DingTalkMarkdown `json:"markdown"`
}

// DingTalkMarkdown Markdown 内容
type DingTalkMarkdown struct {
	Title string `json:"title"`
	Text  string `json:"text"`
}

// NewDingTalkNotifier 创建钉钉通知器
func NewDingTalkNotifier() *DingTalkNotifier {
	token := os.Getenv("DINGTALK_TOKEN")
	enabled := token != ""

	if !enabled {
		fmt.Fprintf(os.Stderr, "[WARN] DINGTALK_TOKEN not found in environment variables, DingTalk notification disabled\n")
	}

	return &DingTalkNotifier{
		token:   token,
		enabled: enabled,
	}
}

// IsEnabled 检查是否启用
func (d *DingTalkNotifier) IsEnabled() bool {
	return d.enabled
}

// SyncResult 同步结果
type SyncResult struct {
	Success      bool
	SourceImage  string
	TargetImage  string
	ErrorMessage string
}

// BuildMessage 构建钉钉 Markdown 消息
func (d *DingTalkNotifier) BuildMessage(provider string, results []SyncResult) *DingTalkMessage {
	// 统计结果
	successCount := 0
	failCount := 0
	skipCount := 0

	for _, r := range results {
		if r.Success {
			if r.ErrorMessage == "already exists" {
				skipCount++
			} else {
				successCount++
			}
		} else {
			failCount++
		}
	}

	// 构建 Markdown 文本
	var sb strings.Builder

	// 标题（必须包含关键词"镜像同步"）
	sb.WriteString("## 🚀 镜像同步任务报告\n\n")

	// 基本信息
	sb.WriteString(fmt.Sprintf("**提供商**: %s\n\n", provider))
	sb.WriteString(fmt.Sprintf("**时间**: %s\n\n", time.Now().Format("2006-01-02 15:04:05")))

	// 统计摘要
	sb.WriteString("### 📊 统计摘要\n\n")
	sb.WriteString(fmt.Sprintf("- 总计: %d 个镜像\n", len(results)))
	sb.WriteString(fmt.Sprintf("- ✅ 成功: %d\n", successCount))
	sb.WriteString(fmt.Sprintf("- ⏭️ 跳过: %d\n", skipCount))
	if failCount > 0 {
		sb.WriteString(fmt.Sprintf("- **❌ 失败: %d**\n", failCount))
	} else {
		sb.WriteString(fmt.Sprintf("- ❌ 失败: %d\n", failCount))
	}
	sb.WriteString("\n")

	// 镜像清单（包含所有镜像的云端地址）
	sb.WriteString("### 📦 镜像清单\n\n")
	for _, r := range results {
		if r.Success {
			if r.ErrorMessage == "already exists" {
				// 已存在的镜像
				sb.WriteString(fmt.Sprintf("- ⏭️ [已存在] `%s`\n", r.TargetImage))
			} else {
				// 同步成功的镜像
				sb.WriteString(fmt.Sprintf("- ✅ [成功] `%s`\n", r.TargetImage))
			}
		} else {
			// 失败的镜像
			sb.WriteString(fmt.Sprintf("- ❌ [失败] `%s` (%s)\n", r.TargetImage, r.ErrorMessage))
		}
	}
	sb.WriteString("\n")

	// 页脚
	sb.WriteString("---\n")
	sb.WriteString("*由 docker-image-sync 自动发送*")

	return &DingTalkMessage{
		MsgType: "markdown",
		Markdown: DingTalkMarkdown{
			Title: "镜像同步报告",
			Text:  sb.String(),
		},
	}
}

// ToJSON 将消息转换为 JSON 字符串
func (d *DingTalkNotifier) ToJSON(msg *DingTalkMessage) (string, error) {
	data, err := json.Marshal(msg)
	if err != nil {
		return "", fmt.Errorf("failed to marshal DingTalk message: %w", err)
	}
	return string(data), nil
}

// OutputForEnv 输出 JSON 到环境变量文件
func (d *DingTalkNotifier) OutputForEnv(jsonStr string) error {
	githubEnv := os.Getenv("GITHUB_ENV")
	if githubEnv == "" {
		// 非 GitHub Actions 环境，输出到 stdout
		fmt.Printf("DINGTALK_MESSAGE=%s\n", jsonStr)
		return nil
	}

	// 写入 GITHUB_ENV 文件（需要转义）
	content := fmt.Sprintf("DINGTALK_MESSAGE=%s\n", jsonStr)
	if err := os.WriteFile(githubEnv, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write to GITHUB_ENV: %w", err)
	}

	return nil
}
