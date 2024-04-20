package tencent

import (
	"context"
	"github.com/projectdiscovery/gologger"
	"github.com/tencentyun/cos-go-sdk-v5"
	"github.com/wgpsec/lc/pkg/schema"
	"strings"
)

type cosProvider struct {
	id        string
	provider  string
	cosClient *cos.Client
}

func (d *cosProvider) GetResource(ctx context.Context) (*schema.Resources, error) {
	cosList := schema.NewResources()
	gologger.Debug().Msg("正在获取腾讯云 COS 资源信息")
	response, _, err := d.cosClient.Service.Get(context.Background())
	if err != nil {
		return nil, err
	}
	for _, bucket := range response.Buckets {
		endpointBuilder := &strings.Builder{}
		endpointBuilder.WriteString(bucket.Name)
		endpointBuilder.WriteString("." + bucket.BucketType)
		endpointBuilder.WriteString("." + bucket.Region)
		endpointBuilder.WriteString(".myqcloud.com")
		cosList.Append(&schema.Resource{
			ID:       d.id,
			Public:   true,
			DNSName:  endpointBuilder.String(),
			Provider: d.provider,
		})
	}
	return cosList, nil
}
