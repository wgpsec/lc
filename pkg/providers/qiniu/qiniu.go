package qiniu

import (
	"context"
	"github.com/projectdiscovery/gologger"
	"github.com/qiniu/go-sdk/v7/auth"
	"github.com/wgpsec/lc/pkg/schema"
	"github.com/wgpsec/lc/utils"
)

type Provider struct {
	id         string
	provider   string
	kodoClient *auth.Credentials
}

func New(options schema.OptionBlock) (*Provider, error) {
	var (
		kodoClient *auth.Credentials
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

	// kodo client
	kodoClient = auth.New(accessKeyID, accessKeySecret)

	return &Provider{provider: utils.QiNiu, id: id, kodoClient: kodoClient}, nil
}

func (p *Provider) Resources(ctx context.Context) (*schema.Resources, error) {
	var err error
	kodoProvider := &kodoProvider{kodoClient: p.kodoClient, id: p.id, provider: p.provider}
	buckets, err := kodoProvider.GetResource(ctx)
	if err != nil {
		return nil, err
	}
	gologger.Info().Msgf("获取到 %d 条七牛云 Kodo 对象存储信息", len(buckets.GetItems()))
	finalList := schema.NewResources()
	finalList.Merge(buckets)
	return finalList, nil
}

func (p *Provider) Name() string {
	return p.provider
}
func (p *Provider) ID() string {
	return p.id
}
