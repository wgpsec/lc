package aliyun

import (
	"context"
	"fmt"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/auth/credentials"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/ecs"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/rds"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/sts"
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
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
	ossClient     *oss.Client
	ecsRegions    *ecs.DescribeRegionsResponse
	rdsRegions    *rds.DescribeRegionsResponse
	fcRegions     []FcRegion
	cloudServices []string
	identity      *sts.GetCallerIdentityResponse
}

type providerConfig struct {
	accessKeyID     string
	accessKeySecret string
	sessionToken    string
	okST            bool
}

func New(options schema.OptionBlock, cs goflags.StringSlice) (*Provider, error) {
	var (
		region    = "cn-beijing"
		ossClient *oss.Client
		ecsClient *ecs.Client
		rdsClient *rds.Client
		stsClient *sts.Client
		err       error

		identity   *sts.GetCallerIdentityResponse
		ecsRegions *ecs.DescribeRegionsResponse
		rdsRegions *rds.DescribeRegionsResponse
		fcRegions  []FcRegion

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
	if cs[0] == "all" {
		cloudServicesResult, _ := options.GetMetadata(utils.CloudServices)
		cloudServices = strings.Split(cloudServicesResult, ",")
	} else {
		cloudServices = cs
	}
	for _, cloudService := range cloudServices {
		switch cloudService {
		case "ecs":
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
			ecsRegions, err = ecsClient.DescribeRegions(ecs.CreateDescribeRegionsRequest())
			if err != nil {
				return nil, err
			}
			gologger.Debug().Msg("阿里云 ECS 区域信息获取成功")
		case "oss":
			// oss client
			ossClient, err = oss.New(fmt.Sprintf("oss-%s.aliyuncs.com", region), accessKeyID, accessKeySecret)
			if err != nil {
				return nil, err
			}
			if okST {
				ossClient.Config.SecurityToken = sessionToken
			}
			gologger.Debug().Msg("阿里云 OSS 客户端创建成功")
		case "rds":
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
			rdsRegions, err = rdsClient.DescribeRegions(rds.CreateDescribeRegionsRequest())
			if err != nil {
				return nil, err
			}
			gologger.Debug().Msg("阿里云 RDS 区域信息获取成功")
		case "fc":
			// sts GetCallerIdentity
			stsConfig := sdk.NewConfig()
			if okST {
				credential := credentials.NewStsTokenCredential(accessKeyID, accessKeySecret, sessionToken)
				stsClient, err = sts.NewClientWithOptions(region, stsConfig, credential)
			} else {
				credential := credentials.NewAccessKeyCredential(accessKeyID, accessKeySecret)
				stsClient, err = sts.NewClientWithOptions(region, stsConfig, credential)
			}
			if err != nil {
				return nil, err
			}

			stsReq := sts.CreateGetCallerIdentityRequest()
			stsReq.SetScheme("HTTPS")
			identity, err = stsClient.GetCallerIdentity(stsReq)
			if err != nil {
				return nil, err
			}
			gologger.Debug().Msg("阿里云 STS 信息获取成功")

			// fc regions
			fcRegions, err = GetFcRegions()
			if err != nil {
				return nil, err
			}

			gologger.Debug().Msgf("阿里云 FC 区域信息获取成功, 共 %d 个\n", len(fcRegions))
		}
	}

	return &Provider{
		provider: utils.Aliyun, id: id, config: config, identity: identity,
		ossClient: ossClient, ecsRegions: ecsRegions, rdsRegions: rdsRegions, fcRegions: fcRegions, cloudServices: cloudServices,
	}, nil
}

func (p *Provider) Resources(ctx context.Context, cs goflags.StringSlice) (*schema.Resources, error) {
	finalList := schema.NewResources()
	for _, cloudService := range p.cloudServices {
		switch cloudService {
		case "ecs":
			// ecs
			ecsProvider := &instanceProvider{id: p.id, provider: p.provider, ecsRegions: p.ecsRegions, config: p.config}
			ecsList, err := ecsProvider.GetEcsResource(ctx)
			gologger.Info().Msgf("获取到 %d 条阿里云 ECS 信息", len(ecsList.GetItems()))
			if err != nil {
				return nil, err
			}
			finalList.Merge(ecsList)
		case "rds":
			// rds
			rdsProvider := &dbInstanceProvider{id: p.id, provider: p.provider, rdsRegions: p.rdsRegions, config: p.config}
			rdsList, err := rdsProvider.GetRdsResource(ctx)
			if err != nil {
				return nil, err
			}
			gologger.Info().Msgf("获取到 %d 条阿里云 RDS 信息", len(rdsList.GetItems()))
			finalList.Merge(rdsList)
		case "oss":
			// oss
			ossProvider := &ossProvider{ossClient: p.ossClient, id: p.id, provider: p.provider}
			buckets, err := ossProvider.GetResource(ctx)
			if err != nil {
				return nil, err
			}
			gologger.Info().Msgf("获取到 %d 条阿里云 OSS 信息", len(buckets.GetItems()))
			finalList.Merge(buckets)
		case "fc":
			// fc
			fcProvider := &functionProvider{
				id: p.id, provider: p.provider, config: p.config,
				fcRegions: p.fcRegions, identity: p.identity,
			}
			fcList, err := fcProvider.GetResource()
			if err != nil {
				return nil, err
			}
			finalList.Merge(fcList)

			// fc 3.0
			fc3Provider := &function3Provider{
				id: p.id, provider: p.provider, config: p.config,
				fcRegions: p.fcRegions, identity: p.identity,
			}
			fc3List, err := fc3Provider.GetResource()
			if err != nil {
				return nil, err
			}

			gologger.Info().Msgf("获取到 %d 条阿里云 FC 信息", len(fcList.GetItems())+len(fc3List.GetItems()))
			finalList.Merge(fc3List)
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
