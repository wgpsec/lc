package aliyun

import (
	"context"
	dns "github.com/alibabacloud-go/alidns-20150109/v4/client"
	util "github.com/alibabacloud-go/tea-utils/v2/service"
	"github.com/alibabacloud-go/tea/tea"
	"github.com/projectdiscovery/gologger"
	"github.com/wgpsec/lc/pkg/schema"
	"net"
)

type dnsProvider struct {
	id        string
	provider  string
	dnsClient *dns.Client
}

func (d *dnsProvider) GetResource(ctx context.Context, domain string) (*schema.Resources, error) {

	var allRespResult []*dns.DescribeDomainRecordsResponseBodyDomainRecordsRecord
	dnsList := schema.NewResources()
	gologger.Debug().Msg("正在获取阿里云 DNS 资源信息")

	pageNumber := int64(1)
	pageSize := int64(100)
	totalCount := int64(0)
	runtime := &util.RuntimeOptions{}

	for {
		describeDomainRecordsRequest := &dns.DescribeDomainRecordsRequest{
			DomainName: tea.String(domain),
			PageNumber: tea.Int64(pageNumber),
			PageSize:   tea.Int64(pageSize),
		}

		response, err := d.dnsClient.DescribeDomainRecordsWithOptions(describeDomainRecordsRequest, runtime)
		if err != nil {
			return nil, err
		}

		// Append current page records to the result list
		for _, item := range response.Body.DomainRecords.Record {
			allRespResult = append(allRespResult, item)
		}
		totalCount = *response.Body.TotalCount

		// Check if we have retrieved all records
		if int64(len(allRespResult)) >= totalCount {
			break
		}

		// Move to the next page
		pageNumber++
	}

	// Add records to dnsList
	for _, item := range allRespResult {
		if item.RR != nil && item.DomainName != nil {
			if *item.Type == "A" || *item.Type == "AAAA" || *item.Type == "CNAME" {
				result := *item.RR + `.` + *item.DomainName
				pub := ""
				pri := ""
				cname := ""

				if net.ParseIP(*item.Value) != nil && isPrivateIP(net.ParseIP(*item.Value)) {
					pri = *item.Value
				} else if net.ParseIP(*item.Value) != nil && !isPrivateIP(net.ParseIP(*item.Value)) {
					pub = *item.Value
				} else {
					cname = *item.Value
				}

				dnsList.Append(&schema.Resource{
					ID:          d.id,
					Public:      true,
					DNSRecord:   result,
					DNSName:     cname,
					PublicIPv4:  pub,
					PrivateIpv4: pri,
					Provider:    d.provider,
				})
			}
		}
	}
	return dnsList, nil
}

func isPrivateIP(ip net.IP) bool {
	return ip.IsPrivate()
}
