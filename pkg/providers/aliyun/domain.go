package aliyun

import (
	"context"
	domain "github.com/alibabacloud-go/domain-20180129/v4/client"
	util "github.com/alibabacloud-go/tea-utils/v2/service"
	"github.com/alibabacloud-go/tea/tea"
	"github.com/projectdiscovery/gologger"
	"github.com/wgpsec/lc/pkg/schema"
)

type domainProvider struct {
	id           string
	provider     string
	domainClient *domain.Client
}

func (d *domainProvider) GetResource(ctx context.Context) (*schema.Resources, error) {
	domainList := schema.NewResources()
	gologger.Debug().Msg("正在获取阿里云 Domain 资源信息")
	queryDomainListRequest := &domain.QueryDomainListRequest{
		PageNum:  tea.Int32(1),
		PageSize: tea.Int32(10),
	}
	runtime := &util.RuntimeOptions{}
	response, err := d.domainClient.QueryDomainListWithOptions(queryDomainListRequest, runtime)
	if err != nil {
		return nil, err
	}
	for _, domainResult := range response.Body.Data.Domain {
		domainList.Append(&schema.Resource{
			ID:       d.id,
			Public:   true,
			DNSName:  *domainResult.DomainName,
			Provider: d.provider,
		})
	}
	return domainList, nil
}
