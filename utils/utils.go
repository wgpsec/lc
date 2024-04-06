package utils

import (
	"fmt"
	"github.com/wgpsec/lc/pkg/schema"
	"gopkg.in/yaml.v3"
	"os"
	"strings"
)

type ErrNoSuchKey struct {
	Name string
}

// 文本处理

func Contains(s []string, e string) bool {
	for _, a := range s {
		if strings.EqualFold(a, e) {
			return true
		}
	}
	return false
}

func (e *ErrNoSuchKey) Error() string {
	return fmt.Sprintf("no such key: %s", e.Name)
}

func DivideList(list []string, n int) [][]string {
	chunks := make([][]string, n)
	chunkSize := len(list) / n
	remaining := len(list) % n

	for i := 0; i < n; i++ {
		start := i * chunkSize
		end := start + chunkSize
		if i < remaining {
			end++
		}
		if i == n-1 {
			end = len(list)
		}
		chunks[i] = list[start:end]
	}
	return chunks
}

// 文件处理

func ReadConfig(configFile string) (schema.Options, error) {
	var config schema.Options

	file, err := os.Open(configFile)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	if err := yaml.NewDecoder(file).Decode(&config); err != nil {
		return nil, err
	}
	return config, nil
}
