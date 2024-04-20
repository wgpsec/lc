package baidu

import (
	"context"
	"github.com/baidubce/bce-sdk-go/services/bos"
	"github.com/projectdiscovery/gologger"
	"github.com/wgpsec/lc/pkg/schema"
	"strings"
)

type bosProvider struct {
	id        string
	provider  string
	bosClient *bos.Client
}

func (d *bosProvider) GetResource(ctx context.Context) (*schema.Resources, error) {
	var list = schema.NewResources()
	gologger.Debug().Msg("正在获取百度云 BOS 资源信息")
	response, err := d.bosClient.ListBuckets()
	if err != nil {
		return nil, err
	}
	for _, bucket := range response.Buckets {
		endpointBuilder := &strings.Builder{}
		endpointBuilder.WriteString(bucket.Name)
		endpointBuilder.WriteString("." + bucket.Location)
		endpointBuilder.WriteString(".bcebos.com")
		list.Append(&schema.Resource{
			ID:       d.id,
			Public:   true,
			DNSName:  endpointBuilder.String(),
			Provider: d.provider,
		})
	}
	return list, nil
}
