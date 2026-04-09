package provider

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
)

// ParsedImage 表示解析后的镜像信息
type ParsedImage struct {
	Name      string // 镜像名称（不含标签）
	Tag       string // 标签
	Namespace string // 命名空间前缀（如果有）
}

// ParseImage 解析镜像地址，支持多种格式：
//   - python:3.11-slim
//   - nginx:latest
//   - jgraph/drawio:latest
//   - docker.io/library/nginx:latest
//   - gcr.io/google-containers/pause:3.9
func ParseImage(sourceImage string) ParsedImage {
	// 移除 digest 部分 (@sha256:...)
	if atIdx := strings.Index(sourceImage, "@"); atIdx != -1 {
		sourceImage = sourceImage[:atIdx]
	}

	// 解析镜像名称和标签
	var imageName, tag string
	if colonIdx := strings.LastIndex(sourceImage, ":"); colonIdx != -1 {
		// 检查是否是端口（如 localhost:5000/image）
		afterColon := sourceImage[colonIdx+1:]
		if !strings.Contains(afterColon, "/") {
			imageName = sourceImage[:colonIdx]
			tag = afterColon
		} else {
			imageName = sourceImage
			tag = "latest"
		}
	} else {
		imageName = sourceImage
		tag = "latest"
	}

	// 分割路径获取镜像名和命名空间
	parts := strings.Split(imageName, "/")
	var namePart string
	var namespaceParts []string

	if len(parts) == 1 {
		// 只有镜像名，如 "nginx"
		namePart = parts[0]
	} else if len(parts) == 2 {
		// 可能是 "jgraph/drawio"
		if !strings.Contains(parts[1], ":") {
			namespaceParts = []string{parts[0]}
			namePart = parts[1]
		} else {
			namePart = parts[0]
		}
	} else {
		// 多个部分，如 "docker.io/library/nginx" 或 "gcr.io/google-containers/pause"
		namePart = parts[len(parts)-1]
		namespaceParts = parts[1 : len(parts)-1]
	}

	namespace := ""
	if len(namespaceParts) > 0 {
		namespace = strings.Join(namespaceParts, "_")
	}

	return ParsedImage{
		Name:      namePart,
		Tag:       tag,
		Namespace: namespace,
	}
}

// BuildTargetImage 构建目标镜像地址
// 华为云/腾讯云格式: registry/namespace/prefix_name:tag
// 阿里云格式: registry/namespace/name:tag
func BuildTargetImage(registry, namespace string, img ParsedImage, usePrefix bool) string {
	var targetImageName string
	if usePrefix && img.Namespace != "" {
		targetImageName = img.Namespace + "_" + img.Name
	} else {
		targetImageName = img.Name
	}

	if img.Tag != "" && img.Tag != "latest" {
		return fmt.Sprintf("%s/%s/%s:%s", registry, namespace, targetImageName, img.Tag)
	}
	return fmt.Sprintf("%s/%s/%s", registry, namespace, targetImageName)
}

// checkImageExists 检查镜像是否已存在
func checkImageExists(ctx context.Context, image string) (bool, error) {
	cmd := exec.CommandContext(ctx, "skopeo", "inspect", fmt.Sprintf("docker://%s", image))
	output, err := cmd.CombinedOutput()
	if err != nil {
		if strings.Contains(string(output), "manifest unknown") ||
			strings.Contains(string(output), "not found") ||
			strings.Contains(string(output), "denied") {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// dockerLogin 执行 docker login
func dockerLogin(registry, username, password string) error {
	cmd := exec.Command("bash", "-c",
		fmt.Sprintf("echo '%s' | skopeo login --username '%s' --password-stdin %s",
			password, username, registry))
	return cmd.Run()
}

// skopeoCopy 复制镜像
func skopeoCopy(ctx context.Context, source, target string) error {
	cmd := exec.CommandContext(ctx, "skopeo", "copy",
		"--override-arch", "amd64",
		"--override-os", "linux",
		fmt.Sprintf("docker://%s", source),
		fmt.Sprintf("docker://%s", target))
	return cmd.Run()
}
