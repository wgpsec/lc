package huawei

import (
	"context"
	"github.com/huaweicloud/huaweicloud-sdk-go-obs/obs"
	"github.com/wgpsec/lc/pkg/schema"
	"strings"
)

type obsProvider struct {
	id        string
	provider  string
	obsClient *obs.ObsClient
}

func (d *obsProvider) GetResource(ctx context.Context) (*schema.Resources, error) {
	var list = schema.NewResources()
	response, err := d.obsClient.ListBuckets(&obs.ListBucketsInput{QueryLocation: true})
	if err != nil {
		return nil, err
	}
	for _, bucket := range response.Buckets {
		endpointBuilder := &strings.Builder{}
		endpointBuilder.WriteString(bucket.Name)
		endpointBuilder.WriteString(".obs." + bucket.Location)
		endpointBuilder.WriteString(".myhuaweicloud.com")
		list.Append(&schema.Resource{
			ID:       d.id,
			Public:   true,
			DNSName:  endpointBuilder.String(),
			Provider: d.provider,
		})
	}
	return list, nil
}
