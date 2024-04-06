package aliyun

import (
	"context"
	"fmt"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/auth/credentials"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/ecs"
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/wgpsec/lc/pkg/schema"
	"github.com/wgpsec/lc/utils"
)

type Provider struct {
	id        string
	provider  string
	config    providerConfig
	ossClient *oss.Client
	regions   *ecs.DescribeRegionsResponse
}

type providerConfig struct {
	accessKeyID     string
	accessKeySecret string
	sessionToken    string
	okST            bool
}

func New(options schema.OptionBlock) (*Provider, error) {
	var (
		region    = "cn-beijing"
		ecsClient *ecs.Client
		err       error
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

	config := providerConfig{
		accessKeyID:     accessKeyID,
		accessKeySecret: accessKeySecret,
		sessionToken:    sessionToken,
		okST:            okST,
	}

	// oss client
	ossClient, err := oss.New(fmt.Sprintf("oss-%s.aliyuncs.com", region), accessKeyID, accessKeySecret)
	if err != nil {
		return nil, err
	}
	if okST {
		ossClient.Config.SecurityToken = sessionToken
	}

	// ecs client
	ecsConfig := sdk.NewConfig()
	if okST {
		credential := credentials.NewStsTokenCredential(accessKeyID, accessKeySecret, sessionToken)
		ecsClient, err = ecs.NewClientWithOptions(region, ecsConfig, credential)
		if err != nil {
			return nil, err
		}
	} else {
		credential := credentials.NewAccessKeyCredential(accessKeyID, accessKeySecret)
		ecsClient, err = ecs.NewClientWithOptions(region, ecsConfig, credential)
		if err != nil {
			return nil, err
		}
	}
	// regions
	regions, err := ecsClient.DescribeRegions(ecs.CreateDescribeRegionsRequest())
	if err != nil {
		return nil, err
	}
	return &Provider{provider: utils.Aliyun, id: id, ossClient: ossClient, regions: regions, config: config}, nil
}

func (p *Provider) Resources(ctx context.Context) (*schema.Resources, error) {
	var err error
	ecsProvider := &instanceProvider{id: p.id, provider: p.provider, regions: p.regions, config: p.config}
	list, err := ecsProvider.GetResource(ctx)
	if err != nil {
		return nil, err
	}
	ossProvider := &ossProvider{ossClient: p.ossClient, id: p.id, provider: p.provider}
	buckets, err := ossProvider.GetResource(ctx)
	if err != nil {
		return nil, err
	}
	finalList := schema.NewResources()
	finalList.Merge(list)
	finalList.Merge(buckets)
	return finalList, nil
}

func (p *Provider) Name() string {
	return p.provider
}
func (p *Provider) ID() string {
	return p.id
}
