package cmd

import (
	"bytes"
	"context"
	"fmt"
	"github.com/projectdiscovery/gologger"
	"github.com/wgpsec/lc/pkg/inventory"
	"github.com/wgpsec/lc/pkg/schema"
	"github.com/wgpsec/lc/utils"
	"os"
)

type Runner struct {
	config  schema.Options
	options *Options
}

func New(options *Options) (*Runner, error) {
	if options.Config == "" {
		options.Config = defaultConfigLocation
		gologger.Print().Msgf("使用默认配置文件: %s\n", options.Config)
	}
	checkAndCreateConfigFile(options)
	config, err := utils.ReadConfig(options.Config)
	if err != nil {
		return nil, err
	}
	return &Runner{config: config, options: options}, nil
}

func (r *Runner) Enumerate() {
	var (
		err         error
		finalConfig schema.Options
	)

	if r.config, err = utils.ReadConfig(r.options.Config); err != nil {
		gologger.Fatal().Msgf("程序配置文件无效，请检查后重试，错误：%s", err)
	}

	for _, item := range r.config {
		if len(r.options.Provider) != 0 || len(r.options.Id) != 0 {
			if len(r.options.Provider) != 0 && !utils.Contains(r.options.Provider, item[utils.Provider]) {
				continue
			}
			if len(r.options.Id) != 0 && !utils.Contains(r.options.Id, item[utils.Id]) {
				continue
			}
			finalConfig = append(finalConfig, item)
		} else {
			finalConfig = append(finalConfig, item)
		}
	}

	inventory, err := inventory.New(finalConfig, r.options.CloudServices)
	if err != nil {
		gologger.Fatal().Msgf("%s", err)
	}
	var output *os.File
	if r.options.Output != "" {
		outputFile, err := os.Create(r.options.Output)
		if err != nil {
			gologger.Fatal().Msgf("无法创建导出的文件 %s: %s\n", r.options.Output, err)
		}
		output = outputFile
	}
	builder := &bytes.Buffer{}
	schema.SetThreads(r.options.Threads)
	for _, provider := range inventory.Providers {
		gologger.Info().Msgf("正在列出 %s (%s) 的资产\n", provider.Name(), provider.ID())
		instances, err := provider.Resources(context.Background(), r.options.CloudServices)
		if err != nil {
			gologger.Error().Msgf("无法获取 %s（%s）的资产: %s\n", provider.Name(), provider.ID(), err)
			continue
		}
		var Count int
		for _, instance := range instances.GetItems() {
			builder.Reset()
			if instance.DNSRecord != "" {
				Count++
				builder.WriteString("DNSRecord: ")
				builder.WriteString(instance.DNSRecord)
				builder.WriteString(` `)
				// 输出DNSName
				if instance.DNSName != "" {
					builder.WriteString("CNAME: ")
					builder.WriteString(instance.DNSName)
					builder.WriteRune('\n')
				}
				// 输出PublicIPv4
				if instance.PublicIPv4 != "" {
					builder.WriteString("PublicIPv4: ")
					builder.WriteString(instance.PublicIPv4)
					builder.WriteRune('\n')
				}
				// 输出PrivateIpv4
				if instance.PrivateIpv4 != "" {
					builder.WriteString("PrivateIPv4: ")
					builder.WriteString(instance.PrivateIpv4)
					builder.WriteRune('\n')
				}
				output.WriteString(builder.String())
				gologger.Silent().Msgf("%s", builder.String())
				continue
			}

			if instance.DNSName != "" {
				Count++
				builder.WriteString(instance.DNSName)
				builder.WriteRune('\n')
				output.WriteString(builder.String()) //nolint
				builder.Reset()
				gologger.Silent().Msgf("%s", instance.DNSName)
			}
			if instance.PublicIPv4 != "" {
				Count++
				builder.WriteString(instance.PublicIPv4)
				builder.WriteRune('\n')
				output.WriteString(builder.String())
				builder.Reset()
				gologger.Silent().Msgf("%s", instance.PublicIPv4)
			}
			if instance.PrivateIpv4 != "" && !r.options.ExcludePrivate {
				Count++
				builder.WriteString(instance.PrivateIpv4)
				builder.WriteRune('\n')
				output.WriteString(builder.String())
				builder.Reset()
				gologger.Silent().Msgf("%s", instance.PrivateIpv4)
			}
		}
		if Count == 0 {
			gologger.Info().Msgf("在 %s (%s) 下未发现资产，这可能是由于权限不足或没有资产，您可以在确认有相关权限后再进行尝试。", provider.Name(), provider.ID())
		}
		if !r.options.Silent {
			fmt.Println()
		}
	}
}
