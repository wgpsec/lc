package tencent

import (
	"context"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	cvm "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cvm/v20170312"
	lh "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/lighthouse/v20200324"
	"github.com/wgpsec/lc/pkg/schema"
	"sync"
)

type instanceProvider struct {
	id         string
	provider   string
	credential *common.Credential
	cvmRegions []*cvm.RegionInfo
	lhRegions  []*lh.RegionInfo
}

var cvmList = schema.NewResources()

func (d *instanceProvider) GetCVMResource(ctx context.Context) (*schema.Resources, error) {
	var (
		threads int
		//err     error
		wg      sync.WaitGroup
		regions []string
	)
	threads = schema.GetThreads()

	for _, region := range d.cvmRegions {
		regions = append(regions, *region.Region)
	}

	taskCh := make(chan string)
	for i := 0; i < threads; i++ {
		wg.Add(1)
		go func() {
			d.describeCVMInstances(taskCh, &wg)
			//if err != nil {
			//	return
			//}
		}()
	}
	for _, item := range regions {
		taskCh <- item
	}
	close(taskCh)
	wg.Wait()
	return cvmList, nil
}

func (d *instanceProvider) describeCVMInstances(ch <-chan string, wg *sync.WaitGroup) error {
	wg.Done()
	var (
		err       error
		cvmClient *cvm.Client
		response  *cvm.DescribeInstancesResponse
	)
	for region := range ch {
		cpf := profile.NewClientProfile()
		cpf.HttpProfile.Endpoint = "cvm.tencentcloudapi.com"
		cvmClient, err = cvm.NewClient(d.credential, region, cpf)
		if err != nil {
			wg.Done()
			continue
		}
		request := cvm.NewDescribeInstancesRequest()
		request.Limit = common.Int64Ptr(100)
		request.SetScheme("https")
		response, err = cvmClient.DescribeInstances(request)
		if err != nil {
			continue
		}
		for _, instance := range response.Response.InstanceSet {
			var (
				ipv4        []string
				privateIPv4 string
			)

			if len(instance.PublicIpAddresses) > 0 {
				for _, v := range instance.PublicIpAddresses {
					ipv4 = append(ipv4, *v)
				}
			}
			if len(instance.PrivateIpAddresses) > 0 {
				privateIPv4 = *instance.PrivateIpAddresses[0]
			}
			for _, v := range ipv4 {
				cvmList.Append(&schema.Resource{
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
