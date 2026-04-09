package provider

import (
	"fmt"
)

type ProviderFactory struct{}

func NewProviderFactory() *ProviderFactory {
	return &ProviderFactory{}
}

func (f *ProviderFactory) Create(providerType string, registry, namespace, username, password string) (Provider, error) {
	switch providerType {
	case "aliyun", "acr":
		return NewAliyunProvider(registry, namespace, username, password), nil
	case "huawei", "swr":
		return NewHuaweiProvider(registry, namespace, username, password), nil
	case "tencent", "tcr":
		return NewTencentProvider(registry, namespace, username, password), nil
	default:
		return nil, fmt.Errorf("unsupported provider: %s", providerType)
	}
}
