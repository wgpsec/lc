// 感谢 github.com/projectdiscovery/cloudlist 项目， 得益
// 于 cloudlist 的 MIT 开源协议，为这个 Weekend Project 提
// 供了大量帮助， 此项目也将以 MIT 协议开源，共同助力人类开源项目发展。

// Thank you to the github.com/projectdiscovery/cloudlist
// project, which has provided substantial assistance to this
// Weekend Project thanks to its use of the MIT open-source
// license. As a result, this project will also adopt the MIT
// license, joining forces to promote the development of
// open-source initiatives for the benefit of humanity.

package main

import (
	"github.com/projectdiscovery/gologger"
	"github.com/wgpsec/lc/cmd"
	"io"
)

func main() {
	options := cmd.ParseOptions()
	runner, err := cmd.New(options)
	if err != nil {
		gologger.Info().Msg("使用 -h 或 --help 参数查看 lc 的帮助信息。")
		if err == io.EOF {
			gologger.Fatal().Msgf("配置文件为空，请在配置文件中填写上云服务商的访问配置，配置文件地址：%s\n", options.Config)
		} else {
			gologger.Fatal().Msgf("%s", err)
		}
	}
	runner.Enumerate()
}
