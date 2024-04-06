package aliyun

import (
	"context"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/auth/credentials"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/ecs"
	"github.com/wgpsec/lc/pkg/schema"
	"sync"
)

type instanceProvider struct {
	id       string
	provider string
	config   providerConfig
	regions  *ecs.DescribeRegionsResponse
}

var list = schema.NewResources()

func (d *instanceProvider) GetResource(ctx context.Context) (*schema.Resources, error) {
	var (
		threads int
		err     error
		wg      sync.WaitGroup
		regions []string
	)
	threads = schema.GetThreads()

	for _, region := range d.regions.Regions.Region {
		regions = append(regions, region.RegionId)
	}

	taskCh := make(chan string)
	for i := 0; i < threads; i++ {
		wg.Add(1)
		go func() {
			err = d.describeInstances(taskCh, &wg)
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
	return list, nil
}

func (d *instanceProvider) describeInstances(ch <-chan string, wg *sync.WaitGroup) error {
	wg.Done()
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
		for {
			request := ecs.CreateDescribeInstancesRequest()
			response, err = ecsClient.DescribeInstances(request)
			if err != nil {
				break
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
					list.Append(&schema.Resource{
						ID:          d.id,
						Provider:    d.provider,
						PublicIPv4:  v,
						PrivateIpv4: privateIPv4,
						Public:      v != "",
					})
				}
			}
			if response.NextToken == "" {
				break
			}
			request.NextToken = response.NextToken
		}
	}
	return err
}
