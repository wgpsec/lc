package schema

import (
	"context"
	"fmt"
	"github.com/wgpsec/lc/pkg/schema/validate"
	"os"
	"strings"
	"sync"
)

var uniqueMap *sync.Map
var validator *validate.Validator
var Threads int

type Resources struct {
	items []*Resource
	sync.RWMutex
}

func (r *Resources) AppendItem(item *Resource) {
	r.Lock()
	defer r.Unlock()
	r.items = append(r.items, item)
}
func (r *Resources) GetItems() []*Resource {
	r.RLock()
	defer r.RUnlock()
	return r.items
}

type Provider interface {
	Name() string
	ID() string
	Resources(ctx context.Context) (*Resources, error)
}

type Resource struct {
	Public      bool   `json:"public"`
	Provider    string `json:"provider"`
	ID          string `json:"id,omitempty"`
	PublicIPv4  string `json:"public_ipv4,omitempty"`
	PrivateIpv4 string `json:"private_ipv4,omitempty"`
	DNSName     string `json:"dns_name,omitempty"`
}

type Options []OptionBlock
type OptionBlock map[string]string

func init() {
	uniqueMap = &sync.Map{}
	var err error
	validator, err = validate.NewValidator()
	if err != nil {
		panic(fmt.Sprintf("无法创建验证器: %s\n", err))
	}
}

// Resources
func (r *Resources) appendResource(resource *Resource, uniqueMap *sync.Map) {
	if _, ok := uniqueMap.Load(resource.DNSName); !ok && resource.DNSName != "" {
		resourceType := validator.Identify(resource.DNSName)
		r.appendResourceWithTypeAndMeta(resourceType, resource.DNSName, resource.ID, resource.Provider)
		uniqueMap.Store(resource.DNSName, struct{}{})
	}
	if _, ok := uniqueMap.Load(resource.PublicIPv4); !ok && resource.PublicIPv4 != "" {
		resourceType := validator.Identify(resource.PublicIPv4)
		r.appendResourceWithTypeAndMeta(resourceType, resource.PublicIPv4, resource.ID, resource.Provider)
		uniqueMap.Store(resource.PublicIPv4, struct{}{})
	}
	if _, ok := uniqueMap.Load(resource.PrivateIpv4); !ok && resource.PrivateIpv4 != "" {
		resourceType := validator.Identify(resource.PrivateIpv4)
		r.appendResourceWithTypeAndMeta(resourceType, resource.PrivateIpv4, resource.ID, resource.Provider)
		uniqueMap.Store(resource.PrivateIpv4, struct{}{})
	}
}

func (r *Resources) appendResourceWithTypeAndMeta(resourceType validate.ResourceType, item, id, provider string) {
	resource := &Resource{
		Provider: provider,
		ID:       id,
	}
	switch resourceType {
	case validate.DNSName:
		resource.Public = true
		resource.DNSName = item
	case validate.PublicIP:
		resource.Public = true
		resource.PublicIPv4 = item
	case validate.PrivateIP:
		resource.PrivateIpv4 = item
	default:
		return
	}
	r.AppendItem(resource)
}

func (r *Resources) Append(resource *Resource) {
	r.appendResource(resource, uniqueMap)
}

func (r *Resources) Merge(resources *Resources) {
	if resources == nil {
		return
	}
	mergeUniqueMap := &sync.Map{}
	for _, item := range resources.GetItems() {
		r.appendResource(item, mergeUniqueMap)
	}
}

// OptionBlock

func (o OptionBlock) GetMetadata(key string) (string, bool) {
	data, ok := o[key]
	if !ok || data == "" {
		return "", false
	}
	if data[0] == '$' {
		envData := os.Getenv(data[1:])
		if envData != "" {
			return strings.TrimSpace(envData), true
		}
	}
	return strings.TrimSpace(data), true
}

// Other

func NewResources() *Resources {
	return &Resources{items: make([]*Resource, 0)}
}

func SetThreads(threads int) {
	Threads = threads
}

func GetThreads() int {
	return Threads
}
