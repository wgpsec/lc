package qiniu

import (
	"context"
	"github.com/projectdiscovery/gologger"
	"github.com/qiniu/go-sdk/v7/auth"
	"github.com/qiniu/go-sdk/v7/storage"
	"github.com/wgpsec/lc/pkg/schema"
)

type kodoProvider struct {
	id         string
	provider   string
	kodoClient *auth.Credentials
}

func (d *kodoProvider) GetResource(ctx context.Context) (*schema.Resources, error) {
	var request storage.BucketV4Input
	var list = schema.NewResources()
	gologger.Debug().Msg("正在获取七牛云 Kodo 对象存储信息")
	cfg := storage.Config{
		UseHTTPS: true,
	}
	bucketManager := storage.NewBucketManager(d.kodoClient, &cfg)
	for {
		response, err := bucketManager.BucketsV4(&request)
		if err != nil {
			return nil, err
		}
		for _, bucket := range response.Buckets {
			list.Append(&schema.Resource{
				ID:       d.id,
				Public:   true,
				DNSName:  bucket.Name,
				Provider: d.provider,
			})
		}
		if response.IsTruncated {
			response.NextMarker = request.Marker
		} else {
			break
		}
	}
	return list, nil
}
