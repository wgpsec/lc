package liantong

import (
	"context"
	"github.com/projectdiscovery/goflags"
	"github.com/projectdiscovery/gologger"
	"github.com/wgpsec/lc/pkg/schema"
	"github.com/wgpsec/lc/utils"
	"strings"
)

type Provider struct {
	id            string
	provider      string
	config        providerConfig
	cloudServices []string
}

type providerConfig struct {
	accessKeyID     string
	accessKeySecret string
	sessionToken    string
	cloudServices   []string
}

func New(options schema.OptionBlock, cs goflags.StringSlice) (*Provider, error) {
	var cloudServices []string
	accessKeyID, ok := options.GetMetadata(utils.AccessKey)
	if !ok {
		return nil, &utils.ErrNoSuchKey{Name: utils.AccessKey}
	}
	accessKeySecret, ok := options.GetMetadata(utils.SecretKey)
	if !ok {
		return nil, &utils.ErrNoSuchKey{Name: utils.SecretKey}
	}
	id, _ := options.GetMetadata(utils.Id)
	sessionToken, stsOk := options.GetMetadata(utils.SessionToken)
	if stsOk {
		gologger.Debug().Msg("找到联通云临时访问凭证")
	} else {
		gologger.Debug().Msg("找到联通云永久访问凭证")
	}
	if cs[0] == "all" {
		cloudServicesResult, _ := options.GetMetadata(utils.CloudServices)
		cloudServices = strings.Split(cloudServicesResult, ",")
	} else {
		cloudServices = cs
	}
	config := providerConfig{
		accessKeyID:     accessKeyID,
		accessKeySecret: accessKeySecret,
		sessionToken:    sessionToken,
	}
	return &Provider{id: id, provider: utils.LianTong, config: config, cloudServices: cloudServices}, nil
}

func (p *Provider) Name() string {
	return p.provider
}

func (p *Provider) ID() string {
	return p.id
}

func (p *Provider) Resources(ctx context.Context, cs goflags.StringSlice) (*schema.Resources, error) {
	finalList := schema.NewResources()
	for _, cloudService := range p.cloudServices {
		switch cloudService {
		case "oss":
			ossProvider := &ossProvider{config: p.config, id: p.id, provider: p.provider}
			buckets, err := ossProvider.GetResource(ctx)
			if err != nil {
				return nil, err
			}
			gologger.Info().Msgf("获取到 %d 条联通云 OSS 信息", len(buckets.GetItems()))
			finalList.Merge(buckets)
		}
	}
	return finalList, nil
}
