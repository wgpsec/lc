package aliyun

import (
	"context"
	"fmt"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/auth/credentials"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/ecs"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/rds"
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/projectdiscovery/gologger"
	"github.com/wgpsec/lc/pkg/schema"
	"github.com/wgpsec/lc/utils"
)

type Provider struct {
	id         string
	provider   string
	config     providerConfig
	ossClient  *oss.Client
	ecsRegions *ecs.DescribeRegionsResponse
	rdsRegions *rds.DescribeRegionsResponse
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
		rdsClient *rds.Client
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
	if okST {
		gologger.Debug().Msg("找到阿里云访问临时访问凭证")
	} else {
		gologger.Debug().Msg("找到阿里云访问永久访问凭证")
	}

	// oss client
	ossClient, err := oss.New(fmt.Sprintf("oss-%s.aliyuncs.com", region), accessKeyID, accessKeySecret)
	if err != nil {
		return nil, err
	}
	if okST {
		ossClient.Config.SecurityToken = sessionToken
	}
	gologger.Debug().Msg("阿里云 OSS 客户端创建成功")

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
	gologger.Debug().Msg("阿里云 ECS 客户端创建成功")
	// ecs regions
	ecsRegions, err := ecsClient.DescribeRegions(ecs.CreateDescribeRegionsRequest())
	if err != nil {
		return nil, err
	}
	gologger.Debug().Msg("阿里云 ECS 区域信息获取成功")

	// rds client
	rdsConfig := sdk.NewConfig()
	if okST {
		credential := credentials.NewStsTokenCredential(accessKeyID, accessKeySecret, sessionToken)
		rdsClient, err = rds.NewClientWithOptions(region, rdsConfig, credential)
		if err != nil {
			return nil, err
		}
	} else {
		credential := credentials.NewAccessKeyCredential(accessKeyID, accessKeySecret)
		rdsClient, err = rds.NewClientWithOptions(region, rdsConfig, credential)
		if err != nil {
			return nil, err
		}
	}
	gologger.Debug().Msg("阿里云 RDS 客户端创建成功")

	//rds regions
	rdsRegions, err := rdsClient.DescribeRegions(rds.CreateDescribeRegionsRequest())
	if err != nil {
		return nil, err
	}
	gologger.Debug().Msg("阿里云 RDS 区域信息获取成功")

	return &Provider{provider: utils.Aliyun, id: id, ossClient: ossClient, ecsRegions: ecsRegions, rdsRegions: rdsRegions, config: config}, nil
}

func (p *Provider) Resources(ctx context.Context) (*schema.Resources, error) {
	var err error
	ecsProvider := &instanceProvider{id: p.id, provider: p.provider, ecsRegions: p.ecsRegions, config: p.config}
	ecsList, err := ecsProvider.GetEcsResource(ctx)
	gologger.Info().Msgf("获取到 %d 条阿里云 ECS 信息", len(ecsList.GetItems()))
	if err != nil {
		return nil, err
	}
	rdsProvider := &dbInstanceProvider{id: p.id, provider: p.provider, rdsRegions: p.rdsRegions, config: p.config}
	rdsList, err := rdsProvider.GetRdsResource(ctx)
	if err != nil {
		return nil, err
	}
	gologger.Info().Msgf("获取到 %d 条阿里云 RDS 信息", len(rdsList.GetItems()))

	ossProvider := &ossProvider{ossClient: p.ossClient, id: p.id, provider: p.provider}
	buckets, err := ossProvider.GetResource(ctx)
	if err != nil {
		return nil, err
	}
	gologger.Info().Msgf("获取到 %d 条阿里云 OSS 信息", len(buckets.GetItems()))

	finalList := schema.NewResources()
	finalList.Merge(ecsList)
	finalList.Merge(rdsList)
	finalList.Merge(buckets)
	return finalList, nil
}

func (p *Provider) Name() string {
	return p.provider
}
func (p *Provider) ID() string {
	return p.id
}
