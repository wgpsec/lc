package aliyun

import (
	"context"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/auth/credentials"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/rds"
	"github.com/projectdiscovery/gologger"
	"github.com/wgpsec/lc/pkg/schema"
	"github.com/wgpsec/lc/utils"
	"sync"
)

type dbInstanceProvider struct {
	id         string
	provider   string
	config     providerConfig
	rdsRegions *rds.DescribeRegionsResponse
}

type rdsInstance struct {
	dbId   string
	region string
}

var rdsInstances []rdsInstance
var rdsList = schema.NewResources()

func (d *dbInstanceProvider) GetRdsResource(ctx context.Context) (*schema.Resources, error) {
	var (
		threads int
		err     error
		wg      sync.WaitGroup
		regions []string
	)
	threads = schema.GetThreads()

	for _, region := range d.rdsRegions.Regions.RDSRegion {
		regions = append(regions, region.RegionId)
	}
	regions = utils.RemoveRepeatedElement(regions)

	taskCh := make(chan string)
	for i := 0; i < threads; i++ {
		wg.Add(1)
		go func() {
			err = d.describeRdsInstances(taskCh, &wg)
			if err != nil {
				return
			}
		}()
	}
	for _, item := range regions {
		taskCh <- item
	}
	close(taskCh)
	wg.Wait()
	err = d.GetRdsConnectionString(ctx)
	if err != nil {
		return nil, err
	}
	return rdsList, nil
}

func (d *dbInstanceProvider) describeRdsInstances(ch <-chan string, wg *sync.WaitGroup) error {
	wg.Done()
	var (
		err       error
		rdsClient *rds.Client
		response  *rds.DescribeDBInstancesResponse
	)
	for region := range ch {
		rdsConfig := sdk.NewConfig()
		if d.config.okST {
			credential := credentials.NewStsTokenCredential(d.config.accessKeyID, d.config.accessKeySecret, d.config.sessionToken)
			rdsClient, err = rds.NewClientWithOptions(region, rdsConfig, credential)
			if err != nil {
				continue
			}
		} else {
			credential := credentials.NewAccessKeyCredential(d.config.accessKeyID, d.config.accessKeySecret)
			rdsClient, err = rds.NewClientWithOptions(region, rdsConfig, credential)
			if err != nil {
				continue
			}
		}
		gologger.Debug().Msgf("正在获取 %s 区域下的阿里云 RDS 资源信息", region)
		request := rds.CreateDescribeDBInstancesRequest()
		for {
			response, err = rdsClient.DescribeDBInstances(request)
			if err != nil {
				break
			}
			if len(response.Items.DBInstance) > 0 {
				gologger.Warning().Msgf("在 %s 区域下获取到 %d 条 RDS 资源", region, len(response.Items.DBInstance))
			}
			for _, DBInstance := range response.Items.DBInstance {
				rdsInstances = append(rdsInstances, rdsInstance{
					dbId:   DBInstance.DBInstanceId,
					region: region,
				})
			}
			if response.NextToken == "" {
				gologger.Debug().Msgf("NextToken 为空，已终止获取")
				break
			}
			gologger.Debug().Msgf("NextToken 不为空，正在获取下一页数据")
			request.NextToken = response.NextToken
		}
	}
	return err
}

func (d *dbInstanceProvider) GetRdsConnectionString(ctx context.Context) error {
	var (
		private   string
		public    string
		err       error
		rdsClient *rds.Client
		response  *rds.DescribeDBInstanceNetInfoResponse
	)
	for _, dbInstance := range rdsInstances {
		gologger.Debug().Msgf("正在获取 %s RDS 实例的连接信息", dbInstance.dbId)
		rdsConfig := sdk.NewConfig()
		if d.config.okST {
			credential := credentials.NewStsTokenCredential(d.config.accessKeyID, d.config.accessKeySecret, d.config.sessionToken)
			rdsClient, err = rds.NewClientWithOptions(dbInstance.region, rdsConfig, credential)
			if err != nil {
				continue
			}
		} else {
			credential := credentials.NewAccessKeyCredential(d.config.accessKeyID, d.config.accessKeySecret)
			rdsClient, err = rds.NewClientWithOptions(dbInstance.region, rdsConfig, credential)
			if err != nil {
				continue
			}
		}
		request := rds.CreateDescribeDBInstanceNetInfoRequest()
		request.DBInstanceId = dbInstance.dbId

		response, err = rdsClient.DescribeDBInstanceNetInfo(request)
		if err != nil {
			return nil
		}
		for _, DBInstanceNetInfo := range response.DBInstanceNetInfos.DBInstanceNetInfo {
			if DBInstanceNetInfo.IPType == "Private" {
				private = DBInstanceNetInfo.ConnectionString
			} else if DBInstanceNetInfo.IPType == "Public" {
				public = DBInstanceNetInfo.ConnectionString
			}
		}
		rdsList.Append(&schema.Resource{
			ID:          d.id,
			Provider:    d.provider,
			PublicIPv4:  public,
			PrivateIpv4: private,
			Public:      public != "",
		})
	}
	return err
}
