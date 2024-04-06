package tencentcloud

import (
	"context"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	lh "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/lighthouse/v20200324"
	"github.com/wgpsec/lc/pkg/schema"
	"sync"
)

var lhList = schema.NewResources()

func (d *instanceProvider) GetLHResource(ctx context.Context) (*schema.Resources, error) {
	var (
		threads int
		err     error
		wg      sync.WaitGroup
		regions []string
	)
	threads = schema.GetThreads()

	for _, region := range d.lhRegions {
		regions = append(regions, *region.Region)
	}
	taskCh := make(chan string)
	for i := 0; i < threads; i++ {
		wg.Add(1)
		go func() {
			err = d.describeLHInstances(taskCh, &wg)
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
	return lhList, nil
}

func (d *instanceProvider) describeLHInstances(ch <-chan string, wg *sync.WaitGroup) error {
	wg.Done()
	var (
		err      error
		lhClient *lh.Client
		response *lh.DescribeInstancesResponse
	)
	for region := range ch {
		cpf := profile.NewClientProfile()
		cpf.HttpProfile.Endpoint = "lighthouse.tencentcloudapi.com"
		lhClient, err = lh.NewClient(d.credential, region, cpf)
		if err != nil {
			continue
		}
		request := lh.NewDescribeInstancesRequest()
		request.Limit = common.Int64Ptr(100)
		request.SetScheme("https")
		response, err = lhClient.DescribeInstances(request)
		if err != nil {
			continue
		}
		for _, instance := range response.Response.InstanceSet {
			var (
				ipv4        []string
				privateIPv4 string
			)

			if len(instance.PublicAddresses) > 0 {
				for _, v := range instance.PublicAddresses {
					ipv4 = append(ipv4, *v)
				}
			}
			if len(instance.PrivateAddresses) > 0 {
				privateIPv4 = *instance.PrivateAddresses[0]
			}
			for _, v := range ipv4 {
				lhList.Append(&schema.Resource{
					ID:          d.id,
					Provider:    d.provider,
					PublicIPv4:  v,
					PrivateIpv4: privateIPv4,
					Public:      v != "",
				})
			}
		}
	}
	return err
}
