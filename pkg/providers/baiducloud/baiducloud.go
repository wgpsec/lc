package baiducloud

import (
	"context"
	"github.com/baidubce/bce-sdk-go/auth"
	"github.com/baidubce/bce-sdk-go/services/bos"
	"github.com/projectdiscovery/gologger"
	"github.com/wgpsec/lc/pkg/schema"
	"github.com/wgpsec/lc/utils"
)

type Provider struct {
	id        string
	provider  string
	bosClient *bos.Client
	config    providerConfig
}

type providerConfig struct {
	accessKeyID     string
	accessKeySecret string
	sessionToken    string
	okST            bool
}

func New(options schema.OptionBlock) (*Provider, error) {
	var (
		endpoint  = "https://bj.bcebos.com"
		err       error
		bosClient *bos.Client
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
		gologger.Debug().Msg("找到百度云访问临时访问凭证")
	} else {
		gologger.Debug().Msg("找到百度云访问永久访问凭证")
	}

	// bos client
	if okST {
		bosClient, err = bos.NewClient(accessKeyID, accessKeySecret, "")
		if err != nil {
			return nil, err
		}
		stsCredential, err := auth.NewSessionBceCredentials(
			accessKeyID,
			accessKeySecret,
			sessionToken)
		if err != nil {
			return nil, err
		}
		bosClient.Config.Credentials = stsCredential
	} else {
		clientConfig := bos.BosClientConfiguration{
			Ak:               accessKeyID,
			Sk:               accessKeySecret,
			Endpoint:         endpoint,
			RedirectDisabled: false,
		}
		bosClient, err = bos.NewClientWithConfig(&clientConfig)
	}
	if err != nil {
		return nil, err
	}

	config := providerConfig{
		accessKeyID:     accessKeyID,
		accessKeySecret: accessKeySecret,
		sessionToken:    sessionToken,
		okST:            okST,
	}

	return &Provider{provider: utils.BaiduCloud, id: id, bosClient: bosClient, config: config}, nil
}

func (p *Provider) Resources(ctx context.Context) (*schema.Resources, error) {
	var err error
	bccProvider := &instanceProvider{provider: p.provider, id: p.id, config: p.config}
	lists, err := bccProvider.GetResource(ctx)
	if err != nil {
		return nil, err
	}
	gologger.Info().Msgf("获取到 %d 条百度云 BCC 信息", len(lists.Items))
	bosProvider := &bosProvider{bosClient: p.bosClient, id: p.id, provider: p.provider}
	buckets, err := bosProvider.GetResource(ctx)
	if err != nil {
		return nil, err
	}
	gologger.Info().Msgf("获取到 %d 条百度云 BOS 信息", len(buckets.Items))
	finalList := schema.NewResources()
	finalList.Merge(lists)
	finalList.Merge(buckets)
	return finalList, nil
}

func (p *Provider) Name() string {
	return p.provider
}
func (p *Provider) ID() string {
	return p.id
}
