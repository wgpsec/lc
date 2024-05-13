package aliyun

import (
	"context"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/auth/credentials"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/ecs"
	"github.com/projectdiscovery/gologger"
	"github.com/wgpsec/lc/pkg/schema"
	"sync"
)

type instanceProvider struct {
	id         string
	provider   string
	config     providerConfig
	ecsRegions *ecs.DescribeRegionsResponse
}

var ecsList = schema.NewResources()

func (d *instanceProvider) GetEcsResource(ctx context.Context) (*schema.Resources, error) {
	var (
		threads int
		err     error
		wg      sync.WaitGroup
		regions []string
	)
	threads = schema.GetThreads()

	for _, region := range d.ecsRegions.Regions.Region {
		regions = append(regions, region.RegionId)
	}

	taskCh := make(chan string, threads)
	for i := 0; i < threads; i++ {
		wg.Add(1)
		go func() {
			err = d.describeEcsInstances(taskCh, &wg)
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
	return ecsList, nil
}

func (d *instanceProvider) describeEcsInstances(ch <-chan string, wg *sync.WaitGroup) error {
	defer wg.Done()
	var (
		err       error
		ecsClient *ecs.Client
		response  *ecs.DescribeInstancesResponse
	)
	for region := range ch {
		ecsConfig := sdk.NewConfig()
		if d.config.okST {
			credential := credentials.NewStsTokenCredential(d.config.accessKeyID, d.config.accessKeySecret, d.config.sessionToken)
			ecsClient, err = ecs.NewClientWithOptions(region, ecsConfig, credential)
			if err != nil {
				continue
			}
		} else {
			credential := credentials.NewAccessKeyCredential(d.config.accessKeyID, d.config.accessKeySecret)
			ecsClient, err = ecs.NewClientWithOptions(region, ecsConfig, credential)
			if err != nil {
				continue
			}
		}
		gologger.Debug().Msgf("正在获取 %s 区域下的阿里云 ECS 资源信息", region)
		request := ecs.CreateDescribeInstancesRequest()
		for {
			response, err = ecsClient.DescribeInstances(request)
			if err != nil {
				break
			}
			if len(response.Instances.Instance) > 0 {
				gologger.Warning().Msgf("在 %s 区域下获取到 %d 条 ECS 资源", region, len(response.Instances.Instance))
			}
			for _, instance := range response.Instances.Instance {
				var (
					ipv4        []string
					privateIPv4 string
				)
				if len(instance.PublicIpAddress.IpAddress) > 0 {
					ipv4 = append(ipv4, instance.PublicIpAddress.IpAddress...)
				}
				if len(instance.NetworkInterfaces.NetworkInterface) > 0 && len(instance.NetworkInterfaces.NetworkInterface[0].PrivateIpSets.PrivateIpSet) > 0 {
					privateIPv4 = instance.NetworkInterfaces.NetworkInterface[0].PrivateIpSets.PrivateIpSet[0].PrivateIpAddress
				}
				for _, v := range ipv4 {
					ecsList.Append(&schema.Resource{
						ID:          d.id,
						Provider:    d.provider,
						PublicIPv4:  v,
						PrivateIpv4: privateIPv4,
						Public:      v != "",
					})
				}
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
