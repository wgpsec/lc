package cucloud

import (
	"context"
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
	sessionToken, _ := options.GetMetadata(utils.SessionToken)

	config := providerConfig{
		accessKeyID:     accessKeyID,
		accessKeySecret: accessKeySecret,
		sessionToken:    sessionToken,
	}
	return &Provider{id: id, provider: utils.CuCloud, config: config}, nil
}

func (p *Provider) Name() string {
	return p.provider
}

func (p *Provider) ID() string {
	return p.id
}

func (p *Provider) Resources(ctx context.Context) (*schema.Resources, error) {
	ossProvider := &s3Provider{config: p.config, id: p.id, provider: p.provider}
	buckets, err := ossProvider.GetResource(ctx)
	if err != nil {
		return nil, err
	}
	finalList := schema.NewResources()
	finalList.Merge(buckets)
	return finalList, nil
}
