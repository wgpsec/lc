package aliyun

import (
	"fmt"
	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	fc "github.com/alibabacloud-go/fc-20230330/v4/client"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/sts"
	"github.com/projectdiscovery/gologger"
	"github.com/wgpsec/lc/pkg/schema"
	"strings"
	"sync"
)

type function3Provider struct {
	id        string
	identity  *sts.GetCallerIdentityResponse
	provider  string
	config    providerConfig
	fcRegions []FcRegion
}

var fc3List = schema.NewResources()

func (f *function3Provider) GetResource() (*schema.Resources, error) {
	var (
		threads int
		err     error
		wg      sync.WaitGroup
		regions []string
	)

	for _, region := range f.fcRegions {
		if !strings.Contains(region.RegionId, "finance") {
			regions = append(regions, region.RegionId)
		}
	}

	threads = schema.GetThreads()
	taskCh := make(chan string, threads)
	for i := 0; i < threads; i++ {
		wg.Add(1)
		go func() {
			err = f.listCustomDomains(taskCh, &wg)
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

	return fc3List, nil
}

func (f *function3Provider) newFcConfig(region string) *openapi.Config {
	endpoint := fmt.Sprintf("%s.%s.fc.aliyuncs.com", f.identity.AccountId, region)
	return &openapi.Config{
		AccessKeyId:     &f.config.accessKeyID,
		AccessKeySecret: &f.config.accessKeySecret,
		SecurityToken:   &f.config.sessionToken,
		Endpoint:        &endpoint,
		RegionId:        &region,
	}
}

func (f *function3Provider) listCustomDomains(ch <-chan string, wg *sync.WaitGroup) error {
	defer wg.Done()
	var (
		err       error
		fcClient  *fc.Client
		domainRes *fc.ListCustomDomainsResponse
	)

	for region := range ch {
		gologger.Debug().Msgf("正在获取 %s 区域下的阿里云 FC 3.0 自定义域名资源信息", region)
		fcConfig := f.newFcConfig(region)
		fcClient, err = fc.NewClient(fcConfig)
		if err != nil {
			gologger.Debug().Msgf("%s endpoint NewClient err: %s", *fcConfig.Endpoint, err)
			break
		}

		lcdReq := &fc.ListCustomDomainsRequest{}
		domainRes, err = fcClient.ListCustomDomains(lcdReq)
		if err != nil {
			gologger.Debug().Msgf("%s endpoint ListCustomDomains err: %s", *fcClient.Endpoint, err)
			continue
		}
		for _, cd := range domainRes.Body.CustomDomains {
			fc3List.Append(&schema.Resource{
				ID:       f.id,
				Provider: f.provider,
				DNSName:  fmt.Sprintf("%s://%s", strings.ToLower(*cd.Protocol), *cd.DomainName),
			})
		}
	}
	return err
}
