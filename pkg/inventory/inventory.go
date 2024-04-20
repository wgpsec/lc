package inventory

import (
	"fmt"
	"github.com/wgpsec/lc/pkg/providers/aliyun"
	"github.com/wgpsec/lc/pkg/providers/baidu"
	"github.com/wgpsec/lc/pkg/providers/huawei"
	"github.com/wgpsec/lc/pkg/providers/liantong"
	"github.com/wgpsec/lc/pkg/providers/qiniu"
	"github.com/wgpsec/lc/pkg/providers/tencent"
	"github.com/wgpsec/lc/pkg/providers/tianyi"
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
	case utils.Tencent:
		return tencent.New(block)
	case utils.Huawei:
		return huawei.New(block)
	case utils.TianYi:
		return tianyi.New(block)
	case utils.Baidu:
		return baidu.New(block)
	case utils.LianTong:
		return liantong.New(block)
	case utils.QiNiu:
		return qiniu.New(block)
	default:
		return nil, fmt.Errorf("发现无效的云服务商名: %s", value)
	}
}
