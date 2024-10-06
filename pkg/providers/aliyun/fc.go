package aliyun

import (
	"encoding/json"
	"errors"
	"fmt"
	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	fc "github.com/alibabacloud-go/fc-open-20210406/v2/client"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/sts"
	"github.com/projectdiscovery/gologger"
	"github.com/wgpsec/lc/pkg/schema"
	"io"
	"net/http"
	"strings"
	"sync"
)

type functionProvider struct {
	id        string
	identity  *sts.GetCallerIdentityResponse
	provider  string
	config    providerConfig
	fcRegions []FcRegion
}

type FcRegionsResp struct {
	Code int `json:"code"`
	Data struct {
		Type      string     `json:"type"`
		Endpoints []FcRegion `json:"endpoints"`
	} `json:"data"`
}

type FcRegion struct {
	RegionId   string `json:"regionId"`
	RegionName string `json:"regionName"`
	AreaId     string `json:"areaId"`
	AreaName   string `json:"areaName"`
	Public     string `json:"public"`
	VPC        string `json:"vpc"`
}

type FcTriggerConfig struct {
	Method             []string `json:"method"`
	AuthType           string   `json:"authType"`
	DisableURLInternet bool     `json:"disableURLInternet"`
}

var fcList = schema.NewResources()
var fcResourceMap = sync.Map{}

func (f *functionProvider) GetResource() (*schema.Resources, error) {
	var (
		threads int
		err     error
		wg      sync.WaitGroup
		regions []string
	)

	for _, region := range f.fcRegions {
		regions = append(regions, region.RegionId)
	}

	threads = schema.GetThreads()
	taskCh := make(chan string, threads)
	for i := 0; i < threads; i++ {
		wg.Add(1)
		go func() {
			err = f.describeFcService(taskCh, &wg)
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

	taskCh = make(chan string, threads)
	for i := 0; i < threads; i++ {
		wg.Add(1)
		go func() {
			err = f.describeFcCustomDomains(taskCh, &wg)
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

	return fcList, nil
}

func (f *functionProvider) newFcConfig(region string) *openapi.Config {
	endpoint := fmt.Sprintf("%s.%s.fc.aliyuncs.com", f.identity.AccountId, region)
	return &openapi.Config{
		AccessKeyId:     &f.config.accessKeyID,
		AccessKeySecret: &f.config.accessKeySecret,
		SecurityToken:   &f.config.sessionToken,
		Endpoint:        &endpoint,
		RegionId:        &region,
	}
}

// describeFcCustomDomains 经测试, 就算 fc 禁用公网访问, 如有自定义域名, 能自定义域名+路由直接访问函数
func (f *functionProvider) describeFcCustomDomains(ch <-chan string, wg *sync.WaitGroup) error {
	defer wg.Done()
	var (
		err       error
		fcClient  *fc.Client
		domainRes *fc.ListCustomDomainsResponse
	)

	for region := range ch {

		if _, ok := fcResourceMap.Load(region); !ok {
			gologger.Debug().Msgf("%s 区域下的阿里云无 FC 函数, 跳过获取自定义域名", region)
			continue
		}

		gologger.Debug().Msgf("正在获取 %s 区域下的阿里云 FC 自定义域名资源信息", region)
		fcConfig := f.newFcConfig(region)
		fcClient, err = fc.NewClient(fcConfig)
		if err != nil {
			gologger.Debug().Msgf("%s endpoint NewClient err: %s", *fcConfig.Endpoint, err)
			break
		}

		lcdReq := &fc.ListCustomDomainsRequest{}
		for {
			domainRes, err = fcClient.ListCustomDomains(lcdReq)
			if err != nil {
				gologger.Debug().Msgf("%s endpoint ListCustomDomains err: %s", *fcClient.Endpoint, err)
				continue
			}
			for _, cd := range domainRes.Body.CustomDomains {
				fcList.Append(&schema.Resource{
					ID:       f.id,
					Provider: f.provider,
					// FIXME 目前 lc 输出结果并没有 region 区分, 但在控制台 FC 中很难识别是哪个区
					// 因为控制台鼠标指针放到可用区并不会显示数量.... 所以目前先这样显示
					// 此外, 有的 url 不会拼接 cn-shanghai 之类的
					DNSName: fmt.Sprintf("%s://%s#%s", strings.ToLower(*cd.Protocol), *cd.DomainName, region),
					// 如果想判断内外网, 目前接口没有字段能表示是公网还是内网, 只能 dns 查询 CNAME
					// 结果是否为 -internal.fc.aliyuncs.com 结尾
				})
			}
			if domainRes.Body.NextToken == nil {
				break
			}
			gologger.Debug().Msgf("NextToken 不为空，正在获取下一页数据")
			lcdReq.NextToken = domainRes.Body.NextToken
		}
	}
	return err
}

func (f *functionProvider) describeFcService(ch <-chan string, wg *sync.WaitGroup) error {
	defer wg.Done()
	var (
		err      error
		fcClient *fc.Client
	)

	for region := range ch {
		gologger.Debug().Msgf("正在获取 %s 区域下的阿里云 FC 资源信息", region)

		fcConfig := f.newFcConfig(region)
		fcClient, err = fc.NewClient(fcConfig)
		if err != nil {
			gologger.Debug().Msgf("%s endpoint NewClient err: %s", *fcConfig.Endpoint, err)
			break
		}

		err = f.processFcService(fcClient)
		if err != nil {
			gologger.Debug().Msgf("%s endpoint ListServices err: %s", *fcClient.Endpoint, err)
		}
	}

	return err
}

func (f *functionProvider) processFcService(fcClient *fc.Client) error {
	lsReq := &fc.ListServicesRequest{}
	for {
		serviceRes, err := fcClient.ListServices(lsReq)
		if err != nil {
			return err
		}

		for _, s := range serviceRes.Body.Services {
			err = f.processFcFunction(fcClient, s)
			if err != nil {
				gologger.Debug().Msgf("%s endpoint ListFunctions err: %s", *fcClient.Endpoint, err)
				break
			}
		}

		if serviceRes.Body.NextToken == nil {
			break
		}
		gologger.Debug().Msgf(
			"%s region's serviceRes NextToken 不为空 %s，正在获取下一页数据",
			*fcClient.RegionId, *serviceRes.Body.NextToken,
		)
		lsReq.NextToken = serviceRes.Body.NextToken
	}

	return nil
}

func (f *functionProvider) processFcFunction(fcClient *fc.Client, s *fc.ListServicesResponseBodyServices) error {
	lfReq := &fc.ListFunctionsRequest{}
	for {
		funcRes, err := fcClient.ListFunctions(s.ServiceName, lfReq)
		if err != nil {
			return err
		}

		// speed up for describeFcCustomDomains
		if len(funcRes.Body.Functions) > 0 {
			fcResourceMap.Store(*fcClient.RegionId, true)
		}

		for _, ft := range funcRes.Body.Functions {
			err = f.processFcTrigger(fcClient, s, ft)
			if err != nil {
				gologger.Debug().Msgf(
					"%s endpoint [%s]-[%s] ListTriggers err: %s",
					*fcClient.Endpoint, *s.ServiceName, *ft.FunctionName, err,
				)
			}
		}

		if funcRes.Body.NextToken == nil {
			break
		}
		gologger.Debug().Msgf(
			"%s service's funcRes NextToken 不为空 %s，正在获取下一页数据",
			*s.ServiceName, *funcRes.Body.NextToken,
		)
		lfReq.NextToken = funcRes.Body.NextToken
	}

	return nil
}

func (f *functionProvider) processFcTrigger(
	fcClient *fc.Client, s *fc.ListServicesResponseBodyServices, ft *fc.ListFunctionsResponseBodyFunctions,
) error {
	ltReq := &fc.ListTriggersRequest{}
	for {
		triggerRes, err := fcClient.ListTriggers(s.ServiceName, ft.FunctionName, ltReq)
		if err != nil {
			return err
		}

		for _, t := range triggerRes.Body.Triggers {
			if t.TriggerType != nil && strings.ToLower(*t.TriggerType) == "http" {
				var ftc FcTriggerConfig
				err = json.Unmarshal([]byte(*t.TriggerConfig), &ftc)
				if err != nil {
					gologger.Debug().Msgf("%s endpoint Unmarshal FcTriggerConfig err: %s", *fcClient.Endpoint, err)
					continue
				}
				if ftc.DisableURLInternet {
					continue
				}
				if t.UrlInternet == nil {
					gologger.Debug().Msgf(
						"%s endpoint %s - %s  enable internet access but url not found, skip",
						*fcClient.Endpoint, *s.ServiceName, *ft.FunctionName,
					)
					continue
				}
				fcList.Append(&schema.Resource{
					ID:       f.id,
					Provider: f.provider,
					// FIXME 目前 lc 输出结果并没有分区一说, 但在 fc 中很难识别是哪个区
					// 因为控制台鼠标指针放到可用区并不会显示数量.... 所以目前先这样显示
					// 此外, 有的 url 不会拼接 cn-shanghai 之类的
					DNSName: fmt.Sprintf("%s#%s", *t.UrlInternet, *fcClient.RegionId),
					Public:  ftc.DisableURLInternet,
				})
			}
		}

		if triggerRes.Body.NextToken == nil {
			break
		}

		gologger.Debug().Msgf(
			"%s service %s function triggerRes NextToken 不为空 %s，正在获取下一页数据",
			*s.ServiceName, *ft.FunctionName, *triggerRes.Body.NextToken,
		)
		ltReq.NextToken = triggerRes.Body.NextToken
	}

	return nil
}

// GetFcRegions 貌似阿里云没有提供 SDK 获取可用区, 只能抓接口拿了
func GetFcRegions() ([]FcRegion, error) {
	resp, err := http.Get("https://next.api.aliyun.com/meta/v1/products/FC-Open/endpoints.json")
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Error fetching URL: %v\n", err))
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Error reading response body: %v\n", err))
	}

	var endpoints FcRegionsResp
	err = json.Unmarshal(body, &endpoints)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Error decoding JSON: %v\n", err))
	}

	return endpoints.Data.Endpoints, nil
}
