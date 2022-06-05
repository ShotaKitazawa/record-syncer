package replace

import (
	"encoding/json"
	"os"
	"regexp"
	"strings"

	"sigs.k8s.io/yaml"
)

type Replacer struct {
	env     map[string]string
	filters []FilterExp
}

type FilterExp struct {
	regexp      *regexp.Regexp
	replacement string
}

type Filter struct {
	MatchRule   string `json:"matchRule"`
	Replacement string `json:"replacement"`
}

func New(filterFile string) (*Replacer, error) {
	r := &Replacer{env: make(map[string]string)}
	for _, v := range os.Environ() {
		split_v := strings.Split(v, "=")
		r.env[split_v[0]] = split_v[1]
	}
	f, err := os.ReadFile(filterFile)
	if err != nil {
		return nil, err
	}
	js, err := yaml.YAMLToJSON(f)
	if err != nil {
		return nil, err
	}
	var filters []Filter
	if err := json.Unmarshal(js, &filters); err != nil {
		return nil, err
	}
	for _, filter := range filters {
		r.filters = append(r.filters, FilterExp{regexp.MustCompile(filter.MatchRule), filter.Replacement})
	}
	return r, nil
}

func (r Replacer) ReplaceRecord(record string) string {
	result := record
	for _, filter := range r.filters {
		result = filter.regexp.ReplaceAllString(result, filter.replacement)
		for k, v := range r.env {
			result = strings.ReplaceAll(result, "__"+k+"__", v)
		}
	}
	return string(result)
}
