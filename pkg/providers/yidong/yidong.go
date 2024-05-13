package yidong

import (
	"context"
	"github.com/projectdiscovery/gologger"
	"github.com/wgpsec/lc/pkg/schema"
	"github.com/wgpsec/lc/utils"
)

type Provider struct {
	id       string
	provider string
	config   providerConfig
}

type providerConfig struct {
	accessKeyID     string
	accessKeySecret string
	sessionToken    string
}

func New(options schema.OptionBlock) (*Provider, error) {
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
		gologger.Debug().Msg("找到移动云临时访问凭证")
	} else {
		gologger.Debug().Msg("找到移动云永久访问凭证")
	}

	config := providerConfig{
		accessKeyID:     accessKeyID,
		accessKeySecret: accessKeySecret,
		sessionToken:    sessionToken,
	}
	return &Provider{id: id, provider: utils.YiDong, config: config}, nil
}

func (p *Provider) Name() string {
	return p.provider
}

func (p *Provider) ID() string {
	return p.id
}

func (p *Provider) Resources(ctx context.Context) (*schema.Resources, error) {
	eosProvider := &eosProvider{config: p.config, id: p.id, provider: p.provider}
	buckets, err := eosProvider.GetResource(ctx)
	if err != nil {
		return nil, err
	}
	gologger.Info().Msgf("获取到 %d 条移动云 EOS 信息", len(buckets.GetItems()))
	finalList := schema.NewResources()
	finalList.Merge(buckets)
	return finalList, nil
}
