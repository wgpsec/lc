package inventory

import (
	"fmt"
	"github.com/wgpsec/lc/pkg/providers/aliyun"
	"github.com/wgpsec/lc/pkg/providers/baiducloud"
	"github.com/wgpsec/lc/pkg/providers/ctyun"
	"github.com/wgpsec/lc/pkg/providers/cucloud"
	"github.com/wgpsec/lc/pkg/providers/huaweicloud"
	"github.com/wgpsec/lc/pkg/providers/tencentcloud"
	"github.com/wgpsec/lc/pkg/schema"
	"github.com/wgpsec/lc/utils"
)

type Inventory struct {
	Providers []schema.Provider
}

func New(options schema.Options) (*Inventory, error) {
	inventory := &Inventory{}

	for _, block := range options {
		value, ok := block.GetMetadata(utils.Provider)
		if !ok {
			continue
		}
		provider, err := nameToProvider(value, block)
		if err != nil {
			return nil, err
		}
		inventory.Providers = append(inventory.Providers, provider)
	}
	return inventory, nil
}

func nameToProvider(value string, block schema.OptionBlock) (schema.Provider, error) {
	switch value {
	case utils.Aliyun:
		return aliyun.New(block)
	case utils.TencentCloud:
		return tencentcloud.New(block)
	case utils.HuaweiCloud:
		return huaweicloud.New(block)
	case utils.Ctyun:
		return ctyun.New(block)
	case utils.BaiduCloud:
		return baiducloud.New(block)
	case utils.CuCloud:
		return cucloud.New(block)
	default:
		return nil, fmt.Errorf("发现无效的云服务商名: %s", value)
	}
}
