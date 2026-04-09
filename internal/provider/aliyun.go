package provider

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
)

type AliyunProvider struct {
	registry   string
	namespace  string
	username   string
	password   string
}

func NewAliyunProvider(registry, namespace, username, password string) *AliyunProvider {
	if registry == "" {
		registry = "registry.cn-hangzhou.aliyuncs.com"
	}
	return &AliyunProvider{
		registry:   registry,
		namespace:  namespace,
		username:   username,
		password:   password,
	}
}

func (p *AliyunProvider) Name() string {
	return "Aliyun ACR"
}

func (p *AliyunProvider) RegistryDomain() string {
	return p.registry
}

func (p *AliyunProvider) SyncImage(ctx context.Context, sourceImage string) (*SyncResult, error) {
	targetImage := p.buildTargetImage(sourceImage)
	result := &SyncResult{
		SourceImage: sourceImage,
		TargetImage: targetImage,
		Success:     false,
	}

	exists, err := p.CheckImageExists(ctx, targetImage)
	if err != nil {
		result.ErrorMessage = fmt.Sprintf("failed to check image: %v", err)
		return result, err
	}
	if exists {
		result.Success = true
		result.ErrorMessage = "already exists"
		return result, nil
	}

	if err := p.login(); err != nil {
		result.ErrorMessage = fmt.Sprintf("login failed: %v", err)
		return result, err
	}

	if err := p.skopeoCopy(ctx, sourceImage, targetImage); err != nil {
		result.ErrorMessage = fmt.Sprintf("copy failed: %v", err)
		return result, err
	}

	result.Success = true
	return result, nil
}

func (p *AliyunProvider) CheckImageExists(ctx context.Context, image string) (bool, error) {
	cmd := exec.CommandContext(ctx, "skopeo", "inspect", fmt.Sprintf("docker://%s", image))
	output, err := cmd.CombinedOutput()
	if err != nil {
		if strings.Contains(string(output), "manifest unknown") || strings.Contains(string(output), "not found") {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (p *AliyunProvider) buildTargetImage(sourceImage string) string {
	parts := strings.Split(sourceImage, "/")
	imageNameTag := parts[len(parts)-1]

	if tagParts := strings.Split(imageNameTag, "@"); len(tagParts) > 1 {
		imageNameTag = tagParts[0]
	}

	return fmt.Sprintf("%s/%s/%s", p.registry, p.namespace, imageNameTag)
}

func (p *AliyunProvider) login() error {
	cmd := exec.Command("bash", "-c", fmt.Sprintf("echo '%s' | skopeo login --username '%s' --password-stdin %s", p.password, p.username, p.registry))
	return cmd.Run()
}

func (p *AliyunProvider) skopeoCopy(ctx context.Context, source, target string) error {
	cmd := exec.CommandContext(ctx, "skopeo", "copy", "--override-arch", "amd64", "--override-os", "linux",
		fmt.Sprintf("docker://%s", source), fmt.Sprintf("docker://%s", target))
	return cmd.Run()
}
