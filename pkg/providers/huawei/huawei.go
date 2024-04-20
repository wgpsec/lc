package huawei

import (
	"context"
	"github.com/huaweicloud/huaweicloud-sdk-go-obs/obs"
	"github.com/projectdiscovery/gologger"
	"github.com/wgpsec/lc/pkg/schema"
	"github.com/wgpsec/lc/utils"
)

type Provider struct {
	id        string
	provider  string
	obsClient *obs.ObsClient
}

func New(options schema.OptionBlock) (*Provider, error) {
	var (
		region    = "cn-north-4"
		err       error
		obsClient *obs.ObsClient
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
	sessionToken, okST := options.GetMetadata(utils.SessionToken)

	if okST {
		gologger.Debug().Msg("找到华为云访问临时访问凭证")
	} else {
		gologger.Debug().Msg("找到华为云访问永久访问凭证")
	}

	// obs client
	if okST {
		obsClient, err = obs.New(accessKeyID, accessKeySecret, "https://obs."+region+".myhuaweicloud.com", obs.WithSecurityToken(sessionToken))
	} else {
		obsClient, err = obs.New(accessKeyID, accessKeySecret, "https://obs."+region+".myhuaweicloud.com")
	}
	if err != nil {
		return nil, err
	}

	return &Provider{provider: utils.Huawei, id: id, obsClient: obsClient}, nil
}

func (p *Provider) Resources(ctx context.Context) (*schema.Resources, error) {
	var err error
	obsProvider := &obsProvider{obsClient: p.obsClient, id: p.id, provider: p.provider}
	buckets, err := obsProvider.GetResource(ctx)
	if err != nil {
		return nil, err
	}
	gologger.Info().Msgf("获取到 %d 条华为云 OBS 信息", len(buckets.Items))
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
