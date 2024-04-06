package aliyun

import (
	"context"
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/wgpsec/lc/pkg/schema"
	"strings"
)

type ossProvider struct {
	id        string
	provider  string
	ossClient *oss.Client
}

func (d *ossProvider) GetResource(ctx context.Context) (*schema.Resources, error) {
	list = schema.NewResources()
	marker := oss.Marker("")
	for {
		response, err := d.ossClient.ListBuckets(oss.MaxKeys(1000), marker)
		if err != nil {
			break
		}
		marker = oss.Marker(response.NextMarker)
		for _, bucket := range response.Buckets {
			endpointBuilder := &strings.Builder{}
			endpointBuilder.WriteString(bucket.Name)
			endpointBuilder.WriteString(".oss-" + bucket.Region)
			endpointBuilder.WriteString(".aliyuncs.com")
			list.Append(&schema.Resource{
				ID:       d.id,
				Public:   true,
				DNSName:  endpointBuilder.String(),
				Provider: d.provider,
			})
		}
		if !response.IsTruncated {
			break
		}
	}
	return list, nil
}
