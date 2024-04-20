package tencentcloud

import (
	"context"
	"github.com/projectdiscovery/gologger"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/regions"
	cvm "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cvm/v20170312"
	lh "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/lighthouse/v20200324"
	"github.com/wgpsec/lc/pkg/schema"
	"github.com/wgpsec/lc/utils"
)

type Provider struct {
	id         string
	provider   string
	credential *common.Credential
	cvmRegions []*cvm.RegionInfo
	lhRegions  []*lh.RegionInfo
}

func New(options schema.OptionBlock) (*Provider, error) {
	var (
		err        error
		cvmRegions []*cvm.RegionInfo
		lhRegions  []*lh.RegionInfo
		credential *common.Credential
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

	if okST {
		gologger.Debug().Msg("找到腾讯云访问临时访问凭证")
	} else {
		gologger.Debug().Msg("找到腾讯云访问永久访问凭证")
	}

	if okST {
		credential = common.NewTokenCredential(accessKeyID, accessKeySecret, sessionToken)
	} else {
		credential = common.NewCredential(accessKeyID, accessKeySecret)
	}

	// cvm regions
	cvmCpf := profile.NewClientProfile()
	cvmCpf.HttpProfile.Endpoint = "cvm.tencentcloudapi.com"
	cvmClient, err := cvm.NewClient(credential, regions.Beijing, cvmCpf)
	cvmRequest := cvm.NewDescribeRegionsRequest()
	cvmRequest.SetScheme("https")
	cvmResponse, err := cvmClient.DescribeRegions(cvmRequest)
	if err != nil {
		return nil, err
	}
	cvmRegions = cvmResponse.Response.RegionSet

	// lh regions
	lhCpf := profile.NewClientProfile()
	lhCpf.HttpProfile.Endpoint = "lighthouse.tencentcloudapi.com"
	lhClient, err := lh.NewClient(credential, regions.Beijing, lhCpf)
	lhRequest := lh.NewDescribeRegionsRequest()
	lhResponse, err := lhClient.DescribeRegions(lhRequest)
	if err != nil {
		return nil, err
	}
	lhRegions = lhResponse.Response.RegionSet

	return &Provider{id: id, provider: utils.TencentCloud, credential: credential, cvmRegions: cvmRegions, lhRegions: lhRegions}, nil
}

func (p *Provider) Name() string {
	return p.provider
}
func (p *Provider) ID() string {
	return p.id
}

func (p *Provider) Resources(ctx context.Context) (*schema.Resources, error) {
	var err error
	cvmProvider := &instanceProvider{id: p.id, provider: p.provider, cvmRegions: p.cvmRegions, lhRegions: p.lhRegions, credential: p.credential}
	cvmList, err := cvmProvider.GetCVMResource(ctx)
	if err != nil {
		return nil, err
	}
	gologger.Info().Msgf("获取到 %d 条腾讯云 CVM 信息", len(cvmList.Items))

	lhProvider := &instanceProvider{id: p.id, provider: p.provider, cvmRegions: p.cvmRegions, lhRegions: p.lhRegions, credential: p.credential}
	lhList, err := lhProvider.GetLHResource(ctx)
	if err != nil {
		return nil, err
	}
	gologger.Info().Msgf("获取到 %d 条腾讯云 LH 信息", len(lhList.Items))

	finalList := schema.NewResources()
	finalList.Merge(cvmList)
	finalList.Merge(lhList)
	return finalList, nil
}
