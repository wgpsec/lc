package qiniu

import (
	"context"
	"github.com/projectdiscovery/goflags"
	"github.com/projectdiscovery/gologger"
	"github.com/qiniu/go-sdk/v7/auth"
	"github.com/wgpsec/lc/pkg/schema"
	"github.com/wgpsec/lc/utils"
	"strings"
)

type Provider struct {
	id            string
	provider      string
	kodoClient    *auth.Credentials
	cloudServices []string
}

func New(options schema.OptionBlock, cs goflags.StringSlice) (*Provider, error) {
	var (
		kodoClient    *auth.Credentials
		cloudServices []string
	)
	accessKeyID, ok := options.GetMetadata(utils.AccessKey)
	if !ok {
		return nil, &utils.ErrNoSuchKey{Name: utils.AccessKey}
	}
	accessKeySecret, ok := options.GetMetadata(utils.SecretKey)
	if !ok {
		return nil, &utils.ErrNoSuchKey{Name: utils.SecretKey}
	}
	id, _ := options.GetMetadata(utils.Id)

	gologger.Debug().Msg("找到七牛云访问永久访问凭证")

	if cs[0] == "all" {
		cloudServicesResult, _ := options.GetMetadata(utils.CloudServices)
		cloudServices = strings.Split(cloudServicesResult, ",")
	} else {
		cloudServices = cs
	}
	for _, cloudService := range cloudServices {
		switch cloudService {
		case "kodo":
			// kodo client
			kodoClient = auth.New(accessKeyID, accessKeySecret)
		}
	}
	return &Provider{provider: utils.QiNiu, id: id, kodoClient: kodoClient, cloudServices: cloudServices}, nil
}

func (p *Provider) Resources(ctx context.Context, cs goflags.StringSlice) (*schema.Resources, error) {
	finalList := schema.NewResources()
	for _, cloudService := range p.cloudServices {
		switch cloudService {
		case "kodo":
			kodoProvider := &kodoProvider{kodoClient: p.kodoClient, id: p.id, provider: p.provider}
			buckets, err := kodoProvider.GetResource(ctx)
			if err != nil {
				return nil, err
			}
			gologger.Info().Msgf("获取到 %d 条七牛云 Kodo 对象存储信息", len(buckets.GetItems()))
			finalList.Merge(buckets)
		}
	}
	return finalList, nil
}

func (p *Provider) Name() string {
	return p.provider
}
func (p *Provider) ID() string {
	return p.id
}
