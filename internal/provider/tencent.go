package provider

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
)

type TencentProvider struct {
	registry   string
	namespace  string
	username   string
	password   string
}

func NewTencentProvider(registry, namespace, username, password string) *TencentProvider {
	return &TencentProvider{
		registry:   registry,
		namespace:  namespace,
		username:   username,
		password:   password,
	}
}

func (p *TencentProvider) Name() string {
	return "Tencent TCR"
}

func (p *TencentProvider) RegistryDomain() string {
	return p.registry
}

func (p *TencentProvider) SyncImage(ctx context.Context, sourceImage string) (*SyncResult, error) {
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

func (p *TencentProvider) CheckImageExists(ctx context.Context, image string) (bool, error) {
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

func (p *TencentProvider) buildTargetImage(sourceImage string) string {
	slashParts := strings.Split(sourceImage, "/")
	imageNameTag := slashParts[len(slashParts)-1]

	if tagParts := strings.Split(imageNameTag, "@"); len(tagParts) > 1 {
		imageNameTag = tagParts[0]
	}

	var namespacePrefix string
	if len(slashParts) > 2 {
		namespacePrefix = slashParts[len(slashParts)-2]
	} else if len(slashParts) == 2 {
		namespacePrefix = slashParts[0]
	}

	var fullImageName string
	if namespacePrefix != "" {
		fullImageName = fmt.Sprintf("%s_%s", namespacePrefix, imageNameTag)
	} else {
		fullImageName = imageNameTag
	}

	return fmt.Sprintf("%s/%s/%s", p.registry, p.namespace, fullImageName)
}

func (p *TencentProvider) login() error {
	cmd := exec.Command("bash", "-c", fmt.Sprintf("echo '%s' | skopeo login --username '%s' --password-stdin %s", p.password, p.username, p.registry))
	return cmd.Run()
}

func (p *TencentProvider) skopeoCopy(ctx context.Context, source, target string) error {
	cmd := exec.CommandContext(ctx, "skopeo", "copy", "--override-arch", "amd64", "--override-os", "linux",
		fmt.Sprintf("docker://%s", source), fmt.Sprintf("docker://%s", target))
	return cmd.Run()
}
