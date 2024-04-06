package cmd

import (
	"github.com/projectdiscovery/gologger/levels"
	fileutil "github.com/projectdiscovery/utils/file"
	"os"
	"os/user"
	"path/filepath"

	"github.com/projectdiscovery/goflags"
	"github.com/projectdiscovery/gologger"
)

type Options struct {
	Threads        int                 // Threads 设置线程数量
	Silent         bool                // Silent 只展示结果
	Version        bool                // Version 返回工具版本
	ExcludePrivate bool                // ExcludePrivate 从结果中排除私有 IP
	Config         string              // Config 指定配置文件路径
	Output         string              // Output 将结果写入到文件中
	Provider       goflags.StringSlice // Provider 指定要列出的云服务商
	Id             goflags.StringSlice // Id 指定要列出的对象
}

var (
	defaultConfigLocation = filepath.Join(userHomeDir(), ".config/lc/config.yaml")
)

func ParseOptions() *Options {
	options := &Options{}
	flagSet := goflags.NewFlagSet()
	flagSet.SetDescription(`lc (list cloud) 是一个多云攻击面资产梳理工具`)

	flagSet.CreateGroup("config", "配置",
		flagSet.StringVarP(&options.Config, "config", "c", defaultConfigLocation, "指定配置文件路径"),
		flagSet.IntVarP(&options.Threads, "threads", "t", 3, "指定扫描的线程数量"),
	)
	flagSet.CreateGroup("filter", "过滤",
		flagSet.StringSliceVarP(&options.Id, "id", "i", nil, "指定要使用的配置（以逗号分隔）", goflags.NormalizedStringSliceOptions),
		flagSet.StringSliceVarP(&options.Provider, "provider", "p", nil, "指定要使用的云服务商（以逗号分隔）", goflags.NormalizedStringSliceOptions),
		flagSet.BoolVarP(&options.ExcludePrivate, "exclude-private", "ep", false, "从输出的结果中排除私有 IP"),
	)
	flagSet.CreateGroup("output", "输出",
		flagSet.StringVarP(&options.Output, "output", "o", "", "将结果输出到指定的文件中"),
		flagSet.BoolVarP(&options.Silent, "silent", "s", false, "只输出结果"),
		flagSet.BoolVarP(&options.Version, "version", "v", false, "输出工具的版本"),
	)
	_ = flagSet.Parse()
	options.configureOutput()
	showBanner()
	if options.Version {
		gologger.Info().Msgf("当前版本：%s, 发布日期：%s", version, versionDate)
		os.Exit(0)
	}
	checkAndCreateConfigFile(options)
	return options
}

func (options *Options) configureOutput() {
	if options.Silent {
		gologger.DefaultLogger.SetMaxLevel(levels.LevelSilent)
	}
}

func userHomeDir() string {
	usr, err := user.Current()
	if err != nil {
		gologger.Fatal().Msgf("Could not get user home directory: %s\n", err)
	}
	return usr.HomeDir
}

func checkAndCreateConfigFile(options *Options) {
	if options.Config == "" || !fileutil.FileExists(defaultConfigLocation) {
		err := os.MkdirAll(filepath.Dir(options.Config), os.ModePerm)
		if err != nil {
			gologger.Warning().Msgf("无法创建配置文件：%s\n", err)
		}
		if !fileutil.FileExists(defaultConfigLocation) {
			if writeErr := os.WriteFile(defaultConfigLocation, []byte(defaultConfigFile), os.ModePerm); writeErr != nil {
				gologger.Warning().Msgf("Could not write default output to %s: %s\n", defaultConfigLocation, writeErr)
			}
		}
	}
}
