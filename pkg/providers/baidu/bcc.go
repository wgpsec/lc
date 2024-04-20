package baidu

import (
	"context"
	"github.com/baidubce/bce-sdk-go/auth"
	"github.com/baidubce/bce-sdk-go/services/bcc"
	"github.com/baidubce/bce-sdk-go/services/bcc/api"
	"github.com/wgpsec/lc/pkg/schema"
	"sync"
)

type instanceProvider struct {
	id       string
	provider string
	config   providerConfig
}

var list = schema.NewResources()

func (d *instanceProvider) GetResource(ctx context.Context) (*schema.Resources, error) {
	var (
		threads int
		err     error
		wg      sync.WaitGroup
	)
	var endpoints = []string{
		"https://bcc.bj.baidubce.com",
		"https://bcc.gz.baidubce.com",
		"https://bcc.su.baidubce.com",
		"https://bcc.hkg.baidubce.com",
		"https://bcc.fwh.baidubce.com",
		"https://bcc.bd.baidubce.com",
		"https://bcc.cd.baidubce.com",
		"https://bcc.nj.baidubce.com",
		"https://bcc.fsh.baidubce.com",
	}
	threads = schema.GetThreads()

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
	for _, item := range endpoints {
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
		bccClient *bcc.Client
	)
	for endpoint := range ch {
		if d.config.okST {
			bccClient, err = bcc.NewClient(d.config.accessKeyID, d.config.accessKeySecret, "")
			if err != nil {
				continue
			}
			stsCredential, err := auth.NewSessionBceCredentials(
				d.config.accessKeyID,
				d.config.accessKeySecret,
				d.config.sessionToken)
			if err != nil {
				continue
			}
			bccClient.Config.Credentials = stsCredential
		} else {
			bccClient, err = bcc.NewClient(d.config.accessKeyID, d.config.accessKeySecret, endpoint)
			if err != nil {
				continue
			}
		}
		listArgs := &api.ListInstanceArgs{}
		for {
			response, err := bccClient.ListInstances(listArgs)
			if err != nil {
				break
			}
			for _, instance := range response.Instances {
				var (
					ipv4        string
					privateIPv4 string
				)
				ipv4 = instance.PublicIP
				privateIPv4 = instance.InternalIP
				list.Append(&schema.Resource{
					ID:          d.id,
					Provider:    d.provider,
					PublicIPv4:  ipv4,
					PrivateIpv4: privateIPv4,
					Public:      ipv4 != "",
				})
			}
			if response.NextMarker == "" {
				break
			}
			listArgs.Marker = response.NextMarker
		}
	}
	return err
}
