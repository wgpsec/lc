package yidong

import (
	"context"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/projectdiscovery/gologger"
	"github.com/wgpsec/lc/pkg/schema"
	"strings"
	"sync"
)

type eosProvider struct {
	id       string
	provider string
	config   providerConfig
}

type regions struct {
	region   string
	endpoint string
}

var list = schema.NewResources()
var resourcePools = []regions{
	{region: "shanghai1", endpoint: "eos-shanghai-1.cmecloud.cn"},
	{region: "shanghai2", endpoint: "eos-shanghai-2.cmecloud.cn"},
	{region: "shanghai3", endpoint: "eos.shanghai-3.cmecloud.cn"},
	{region: "wuxi1", endpoint: "eos-wuxi-1.cmecloud.cn"},
	{region: "suzhou4", endpoint: "eos-suzhou-4.cmecloud.cn"},
	{region: "fenhu1", endpoint: "eos.fenhu-1.cmecloud.cn"},
	{region: "wuxi5", endpoint: "eos-wuxi-5.cmecloud.cn"},
	{region: "ningbo1", endpoint: "eos-ningbo-1.cmecloud.cn"},
	{region: "ningbo6", endpoint: "eos.ningbo-6.cmecloud.cn"},
	{region: "jinan1", endpoint: "eos-jinan-1.cmecloud.cn"},
	{region: "jinan4", endpoint: "eos.jinan-4.cmecloud.cn"},
	{region: "guangzhou1", endpoint: "eos-guangzhou-1.cmecloud.cn"},
	{region: "dongguan1", endpoint: "eos-dongguan-1.cmecloud.cn"},
	{region: "dongguan7", endpoint: "eos-dongguan-7.cmecloud.cn"},
	{region: "dongguan8", endpoint: "eos-dongguan-8.cmecloud.cn"},
	{region: "chengdu1", endpoint: "eos-chengdu-1.cmecloud.cn"},
	{region: "chengdu6", endpoint: "eos-chengdu-6.cmecloud.cn"},
	{region: "guiyang1", endpoint: "eos-guiyang-1.cmecloud.cn"},
	{region: "guiyang4", endpoint: "eos-guiyang-4.cmecloud.cn"},
	{region: "chongqing1", endpoint: "eos-chongqing-1.cmecloud.cn"},
	{region: "chongqing3", endpoint: "eos-chongqing-3.cmecloud.cn"},
	{region: "xian1", endpoint: "eos-xian-1.cmecloud.cn"},
	{region: "xian2", endpoint: "eos.xian-2.cmecloud.cn"},
	{region: "beijing1", endpoint: "eos-beijing-1.cmecloud.cn"},
	{region: "beijing2", endpoint: "eos-beijing-2.cmecloud.cn"},
	{region: "beijing4", endpoint: "eos-beijing-4.cmecloud.cn"},
	{region: "beijing7", endpoint: "eos-beijing-7.cmecloud.cn"},
	{region: "huhehaote1", endpoint: "eos-huhehaote-1.cmecloud.cn"},
	{region: "huhehaote6", endpoint: "eos-huhehaote-6.cmecloud.cn"},
	{region: "hunan1", endpoint: "eos-hunan-1.cmecloud.cn"},
	{region: "zhuzhou1", endpoint: "eos-zhuzhou-1.cmecloud.cn"},
	{region: "zhengzhou1", endpoint: "eos-zhengzhou-1.cmecloud.cn"},
	{region: "zhengzhou4", endpoint: "eos.zhengzhou-4.cmecloud.cn"},
	{region: "tianjin1", endpoint: "eos-tianjin-1.cmecloud.cn"},
	{region: "tianjin2", endpoint: "eos.tianjin-2.cmecloud.cn"},
	{region: "jilin1", endpoint: "eos-jilin-1.cmecloud.cn"},
	{region: "jilin2", endpoint: "eos.jilin-2.cmecloud.cn"},
	{region: "hubei1", endpoint: "eos-hubei-1.cmecloud.cn"},
	{region: "hubei2", endpoint: "eos.hubei-2.cmecloud.cn"},
	{region: "jiangxi1", endpoint: "eos-jiangxi-1.cmecloud.cn"},
	{region: "jiangxi2", endpoint: "eos.jiangxi-2.cmecloud.cn"},
	{region: "gansu1", endpoint: "eos-gansu-1.cmecloud.cn"},
	{region: "gansu2", endpoint: "eos.gansu-2.cmecloud.cn"},
	{region: "shanxi1", endpoint: "eos-shanxi-1.cmecloud.cn"},
	{region: "shanxi2", endpoint: "eos.shanxi-2.cmecloud.cn"},
	{region: "shanxi3", endpoint: "eos.shanxi-3.cmecloud.cn"},
	{region: "liaoning1", endpoint: "eos-liaoning-1.cmecloud.cn"},
	{region: "liaoning2", endpoint: "eos.liaoning-2.cmecloud.cn"},
	{region: "yunnan", endpoint: "eos-yunnan.cmecloud.cn"},
	{region: "yunnan2", endpoint: "eos-yunnan-2.cmecloud.cn"},
	{region: "hebei1", endpoint: "eos-hebei-1.cmecloud.cn"},
	{region: "hebei2", endpoint: "eos.hebei-2.cmecloud.cn"},
	{region: "fujian1", endpoint: "eos-fujian-1.cmecloud.cn"},
	{region: "fujian2", endpoint: "eos.fujian-2.cmecloud.cn"},
	{region: "fujian3", endpoint: "eos.fujian-3.cmecloud.cn"},
	{region: "guangxi1", endpoint: "eos-guangxi-1.cmecloud.cn"},
	{region: "anhui1", endpoint: "eos-anhui-1.cmecloud.cn"},
	{region: "anhui2", endpoint: "eos.anhui-2.cmecloud.cn"},
	{region: "hainan1", endpoint: "eos-hainan-1.cmecloud.cn"},
	{region: "heilongjiang1", endpoint: "eos-heilongjiang-1.cmecloud.cn"},
	{region: "heilongjiang2", endpoint: "eos.heilongjiang-2.cmecloud.cn"},
	{region: "xinjiang1", endpoint: "eos-xinjiang-1.cmecloud.cn"},
	{region: "xinjiang2", endpoint: "eos.xinjiang-2.cmecloud.cn"},
	{region: "ningxia1", endpoint: "eos.ningxia-1.cmecloud.cn"},
	{region: "qinghai1", endpoint: "eos.qinghai-1.cmecloud.cn"},
	{region: "xizang1", endpoint: "eos.xizang-1.cmecloud.cn"},
}

func (d *eosProvider) GetResource(ctx context.Context) (*schema.Resources, error) {
	var (
		threads int
		err     error
		wg      sync.WaitGroup
		buckets []string
	)

	config := aws.NewConfig()
	config.WithRegion("beijing1")
	config.WithEndpoint("https://eos-beijing-1.cmecloud.cn")
	config.WithCredentials(credentials.NewStaticCredentials(d.config.accessKeyID, d.config.accessKeySecret, d.config.sessionToken))
	session, err := session.NewSession(config)
	if err != nil {
		return nil, err
	}
	s3Client := s3.New(session)

	listBucketsOutput, err := s3Client.ListBuckets(nil)
	if err != nil {
		return nil, err
	}
	for _, bucket := range listBucketsOutput.Buckets {
		buckets = append(buckets, *bucket.Name)
	}
	gologger.Debug().Msgf("找到 %d 个移动云 EOS 资源", len(buckets))

	threads = schema.GetThreads()

	taskCh := make(chan string, threads)
	for i := 0; i < threads; i++ {
		wg.Add(1)
		go func() {
			err = d.listBuckets(taskCh, &wg, s3Client)
			if err != nil {
				return
			}
		}()
	}
	for _, item := range buckets {
		taskCh <- item
	}
	close(taskCh)
	wg.Wait()
	return list, nil

}

func (d *eosProvider) listBuckets(ch <-chan string, wg *sync.WaitGroup, s3Client *s3.S3) error {
	defer wg.Done()
	var err error
	for bucket := range ch {
		bucketLocation, err := s3Client.GetBucketLocation(&s3.GetBucketLocationInput{
			Bucket: aws.String(bucket),
		})
		if err != nil {
			continue
		}
		gologger.Debug().Msgf("%s 的 Location 值为 %s", bucket, *bucketLocation.LocationConstraint)
		endpointBuilder := &strings.Builder{}
		endpointBuilder.WriteString(bucket)
		for _, resourcePool := range resourcePools {
			if *bucketLocation.LocationConstraint == resourcePool.region {
				endpointBuilder.WriteString("." + resourcePool.endpoint)
			}
		}

		list.Append(&schema.Resource{
			ID:       d.id,
			Public:   true,
			DNSName:  endpointBuilder.String(),
			Provider: d.provider,
		})
	}
	return err
}
