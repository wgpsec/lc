package validate

import (
	"net"
	"regexp"
	"strings"
)

var ipv4PrivateRanges = []string{
	"0.0.0.0/8",       // 当前网络（仅作为源地址有效）
	"10.0.0.0/8",      // 私有网络
	"100.64.0.0/10",   // 共享地址空间
	"127.0.0.0/8",     // 回路地址
	"169.254.0.0/16",  // 本地链接（也是许多云提供商的元数据端点）
	"172.16.0.0/12",   // 私有网络
	"192.0.0.0/24",    // IETF 协议分配
	"192.0.2.0/24",    // TEST-NET-1，文档和示例
	"192.88.99.0/24",  // IPv6 到 IPv4 中继（包括2002::/16）
	"192.168.0.0/16",  // 私有网络
	"198.18.0.0/15",   // 网络基准测试
	"198.51.100.0/24", // TEST-NET-2，文档和示例
	"203.0.113.0/24",  // TEST-NET-3，文档和示例
	"224.0.0.0/4",     // IP组播（以前的 D 类网络）
	"240.0.0.0/4",     // 预留 （原 E 类网络）
}

var ipv6PrivateRanges = []string{
	"::1/128",       // 回路地址
	"64:ff9b::/96",  // IPv4/IPv6 转换（RFC 6052）
	"100::/64",      // 丢弃前缀（RFC 6666）
	"2001::/32",     // Teredo 隧道
	"2001:10::/28",  // 已弃用（先前为 ORCHID）
	"2001:20::/28",  // ORCHIDv2
	"2001:db8::/32", // 文档和示例源代码中使用的地址
	"2002::/16",     // 6to4
	"fc00::/7",      // 唯一本地地址
	"fe80::/10",     // 链接地址
	"ff00::/8",      // 多播
}

type Validator struct {
	ipv4      []*net.IPNet
	ipv6      []*net.IPNet
	rxDNSName *regexp.Regexp
}

func NewValidator() (*Validator, error) {
	validator := &Validator{}
	// regex from: https://github.com/asaskevich/govalidator
	validator.rxDNSName = regexp.MustCompile(`^(\*\.)?([a-zA-Z0-9_]{1}[a-zA-Z0-9_-]{0,62})(\.[a-zA-Z0-9_]{1}[a-zA-Z0-9_-]{0,62})*[\._]?$`)

	err := validator.appendIPv4Ranges(ipv4PrivateRanges)
	if err != nil {
		return nil, err
	}
	err = validator.appendIPv6Ranges(ipv6PrivateRanges)
	if err != nil {
		return nil, err
	}
	return validator, nil
}

type ResourceType int

const (
	DNSName ResourceType = iota + 1
	PublicIP
	PrivateIP
	None
)

func (v *Validator) Identify(item string) ResourceType {
	host, port, err := net.SplitHostPort(item)
	if err == nil && port != "" {
		item = host
	}
	parsed := net.ParseIP(item)
	if v.isDNSName(item, parsed) {
		return DNSName
	}
	if parsed == nil {
		return None
	}
	if strings.Contains(item, ":") {
		// 检查 ipv6 私网地址列表
		if v.containsIPv6(parsed) {
			return PrivateIP
		}
		return PublicIP
	}
	// 检查 ipv4 私网地址列表
	if v.containsIPv4(parsed) {
		return PrivateIP
	}
	return PublicIP
}

func (v *Validator) isDNSName(str string, parsed net.IP) bool {
	if str == "" || len(strings.Replace(str, ".", "", -1)) > 255 {
		return false
	}
	return !isIP(parsed) && v.rxDNSName.MatchString(str)
}

func isIP(parsed net.IP) bool {
	return parsed != nil
}

func (v *Validator) appendIPv4Ranges(ranges []string) error {
	for _, ip := range ranges {
		_, rangeNet, err := net.ParseCIDR(ip)
		if err != nil {
			return err
		}
		v.ipv4 = append(v.ipv4, rangeNet)
	}
	return nil
}

func (v *Validator) appendIPv6Ranges(ranges []string) error {
	for _, ip := range ranges {
		_, rangeNet, err := net.ParseCIDR(ip)
		if err != nil {
			return err
		}
		v.ipv6 = append(v.ipv6, rangeNet)
	}
	return nil
}

func (v *Validator) containsIPv4(IP net.IP) bool {
	for _, net := range v.ipv4 {
		if net.Contains(IP) {
			return true
		}
	}
	return false
}

func (v *Validator) containsIPv6(IP net.IP) bool {
	for _, net := range v.ipv6 {
		if net.Contains(IP) {
			return true
		}
	}
	return false
}
