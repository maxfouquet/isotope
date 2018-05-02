package graph

import (
	"fmt"
	"strconv"
	"strings"

	units "github.com/docker/go-units"
)

// UnmarshalYAML implements the Unmarshaler interface and fills the ServiceGraph
// with the contents of a proper YAML document.
func (g *ServiceGraph) UnmarshalYAML(
	unmarshal func(interface{}) error) error {
	var document document
	if err := unmarshal(&document); err != nil {
		return err
	}
	defaults, err := parseDefaultSettings(document.DefaultSettings)
	if err != nil {
		return err
	}
	g.Services = make(map[string]Service)
	for name, yamlSettings := range document.Services {
		serviceSettings, err := parseServiceSettings(
			yamlSettings, defaults.ServiceSettings)
		if err != nil {
			return err
		}
		g.Services[name] = Service{
			ServiceSettings: serviceSettings,
			Name:            name,
		}
	}
	return nil
}

// document is an intermediate, easily unmarshaled struct.
type document struct {
	APIVersion      string                         `yaml:"apiVersion"`
	DefaultSettings yamlDefaultSettings            `yaml:"default"`
	Services        map[string]yamlServiceSettings `yaml:"services"`
}

type yamlDefaultSettings struct {
	yamlServiceSettings `yaml:",inline"`
	yamlRequestSettings `yaml:",inline"`
}

type yamlServiceSettings struct {
	ComputeUsage *string `yaml:"computeUsage"`
	MemoryUsage  *string `yaml:"memoryUsage"`
	ErrorRate    *string `yaml:"errorRate"`
}

type yamlRequestSettings struct {
	PayloadSize *string `yaml:"payloadSize"`
}

type defaultSettings struct {
	ServiceSettings
	RequestSettings
}

func parsePercentageWithDefault(s *string, d float64) (f float64, err error) {
	if s == nil {
		f = d
	} else {
		f, err = parsePercentage(*s)
	}
	return
}

func parseDefaultSettings(
	yaml yamlDefaultSettings) (settings defaultSettings, err error) {
	if yaml.ComputeUsage != nil {
		settings.ComputeUsage, err = parsePercentage(*yaml.ComputeUsage)
		if err != nil {
			return
		}
	}
	if yaml.MemoryUsage != nil {
		settings.MemoryUsage, err = parsePercentage(*yaml.MemoryUsage)
		if err != nil {
			return
		}
	}
	if yaml.ErrorRate != nil {
		settings.ErrorRate, err = parsePercentage(*yaml.ErrorRate)
		if err != nil {
			return
		}
	}
	if yaml.PayloadSize != nil {
		settings.PayloadSize, err = units.FromHumanSize(*yaml.PayloadSize)
		if err != nil {
			return
		}
	}
	return
}

func parseServiceSettings(
	yaml yamlServiceSettings,
	defaults ServiceSettings) (settings ServiceSettings, err error) {
	settings.ComputeUsage, err = parsePercentageWithDefault(
		yaml.ComputeUsage, defaults.ComputeUsage)
	if err != nil {
		return
	}
	settings.MemoryUsage, err = parsePercentageWithDefault(
		yaml.MemoryUsage, defaults.MemoryUsage)
	if err != nil {
		return
	}
	settings.ErrorRate, err = parsePercentageWithDefault(
		yaml.ErrorRate, defaults.ErrorRate)
	return
}

func parsePercentage(s string) (f float64, err error) {
	percentIndex := strings.Index(s, "%")
	if percentIndex < 0 {
		err = fmt.Errorf("Could not parse percentage: '%%' not found")
		return
	}
	percentageFloat, err := strconv.ParseFloat(s[:percentIndex], 64)
	if err != nil {
		return
	}
	f = percentageFloat / 100
	return
}
