package liantong

import (
	"context"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/wgpsec/lc/pkg/schema"
	"strings"
	"sync"
)

type ossProvider struct {
	id       string
	provider string
	config   providerConfig
}

type regions struct {
	region   string
	endpoint string
}

var list = schema.NewResources()

func (d *ossProvider) GetResource(ctx context.Context) (*schema.Resources, error) {
	var (
		threads int
		err     error
		wg      sync.WaitGroup
	)

	zones := []regions{
		{region: "cn-langfang-2", endpoint: "obs-helf.cucloud.cn"},
		{region: "cn-xiamen-1", endpoint: "obs-fjxm.cucloud.cn"},
		{region: "cn-nanping-1", endpoint: "obs-fjnp.cucloud.cn"},
		{region: "cn-ningde-1", endpoint: "obs-fjnd.cucloud.cn"},
		{region: "cn-huhehaote-2", endpoint: "obs-nmhhht2.cucloud.cn"},
		{region: "cn-guiyang-2", endpoint: "obs-gzgy2.cucloud.cn"},
		{region: "cn-chongqing-1", endpoint: "obs-cq.cucloud.cn"},
		{region: "cn-shenzhen-1", endpoint: "obs-gdsz.cucloud.cn"},
		{region: "cn-shengyang-1", endpoint: "obs-lnsy.cucloud.cn"},
		{region: "cn-harbin-1", endpoint: "obs-hlhrb.cucloud.cn"},
		{region: "cn-shanghai-1", endpoint: "obs-sh.cucloud.cn"},
		//{region: "cn-huhehaote-3", endpoint: "obs-nmhhht3.cucloud.cn"},
		//{region: "cn-shijiazhuang-1", endpoint: "obs-hesjz.cucloud.cn"},
		{region: "cn-changsha-1", endpoint: "obs-hncs.cucloud.cn"},
	}
	threads = schema.GetThreads()

	taskCh := make(chan regions, threads)
	for i := 0; i < threads; i++ {
		wg.Add(1)
		go func() {
			err = d.listBuckets(taskCh, &wg)
			if err != nil {
				return
			}
		}()
	}
	for _, item := range zones {
		taskCh <- item
	}
	close(taskCh)
	wg.Wait()
	return list, nil

}

func (d *ossProvider) listBuckets(ch <-chan regions, wg *sync.WaitGroup) error {
	defer wg.Done()
	var err error
	for region := range ch {
		config := aws.NewConfig()
		config.WithRegion(region.region)
		config.WithEndpoint("https://" + region.endpoint)
		config.WithCredentials(credentials.NewStaticCredentials(d.config.accessKeyID, d.config.accessKeySecret, d.config.sessionToken))
		session, err := session.NewSession(config)

		if err != nil {
			continue
		}
		s3Client := s3.New(session)

		listBucketsOutput, err := s3Client.ListBuckets(nil)
		if err != nil {
			continue
		}
		for _, bucket := range listBucketsOutput.Buckets {
			endpointBuilder := &strings.Builder{}
			endpointBuilder.WriteString(aws.StringValue(bucket.Name))
			endpointBuilder.WriteString("." + region.endpoint)
			list.Append(&schema.Resource{
				ID:       d.id,
				Public:   true,
				DNSName:  endpointBuilder.String(),
				Provider: d.provider,
			})
		}
	}
	return err
}
