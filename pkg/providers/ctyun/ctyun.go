package ctyun

import (
	"context"
	"github.com/teamssix/oos-go-sdk/oos"
	"github.com/wgpsec/lc/pkg/schema"
	"github.com/wgpsec/lc/utils"
)

type Provider struct {
	id        string
	provider  string
	oosClient *oos.Client
}

func New(options schema.OptionBlock) (*Provider, error) {
	var (
		err       error
		oosClient *oos.Client
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

	// oos client
	clientOptionV4 := oos.V4Signature(true)
	isEnableSha256 := oos.EnableSha256ForPayload(true)
	oosClient, err = oos.New("https://oos-cn.ctyunapi.cn", accessKeyID, accessKeySecret, clientOptionV4, isEnableSha256)
	if err != nil {
		return nil, err
	}

	return &Provider{provider: utils.Ctyun, id: id, oosClient: oosClient}, nil
}

func (p *Provider) Resources(ctx context.Context) (*schema.Resources, error) {
	var err error
	oosProvider := &obsProvider{oosClient: p.oosClient, id: p.id, provider: p.provider}
	buckets, err := oosProvider.GetResource(ctx)
	if err != nil {
		return nil, err
	}
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
