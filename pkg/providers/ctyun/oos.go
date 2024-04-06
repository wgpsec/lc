package ctyun

import (
	"context"
	"github.com/teamssix/oos-go-sdk/oos"
	"github.com/wgpsec/lc/pkg/schema"
	"strings"
)

type obsProvider struct {
	id        string
	provider  string
	oosClient *oos.Client
}

func (d *obsProvider) GetResource(ctx context.Context) (*schema.Resources, error) {
	var list = schema.NewResources()
	response, err := d.oosClient.ListBuckets()
	if err != nil {
		return nil, err
	}
	for _, bucket := range response.Buckets {
		endpointBuilder := &strings.Builder{}
		endpointBuilder.WriteString(bucket.Name)
		endpointBuilder.WriteString(".oos-cn.ctyunapi.cn")
		list.Append(&schema.Resource{
			ID:       d.id,
			Public:   true,
			DNSName:  endpointBuilder.String(),
			Provider: d.provider,
		})
	}
	return list, nil
}
