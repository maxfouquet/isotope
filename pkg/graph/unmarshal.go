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

func parseDefaultSettings(
	yaml yamlDefaultSettings) (settings defaultSettings, err error) {
	if yaml.ComputeUsage != nil {
		settings.ComputeUsage, err = parseFloat(*yaml.ComputeUsage)
		if err != nil {
			return
		}
	}
	if yaml.MemoryUsage != nil {
		settings.MemoryUsage, err = parseFloat(*yaml.MemoryUsage)
		if err != nil {
			return
		}
	}
	if yaml.ErrorRate != nil {
		settings.ErrorRate, err = parseFloat(*yaml.ErrorRate)
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
	settings.ComputeUsage, err = parseFloatWithDefault(
		yaml.ComputeUsage, defaults.ComputeUsage)
	if err != nil {
		return
	}
	settings.MemoryUsage, err = parseFloatWithDefault(
		yaml.MemoryUsage, defaults.MemoryUsage)
	if err != nil {
		return
	}
	settings.ErrorRate, err = parseFloatWithDefault(
		yaml.ErrorRate, defaults.ErrorRate)
	return
}

// parseFloatWithDefault tries to parse s, which is either a percentage of the
// form "X.X%" or a float. If s is nil, return d.
func parseFloatWithDefault(s *string, d float64) (f float64, err error) {
	if s == nil {
		f = d
	} else {
		f, err = parseFloat(*s)
	}
	return
}

func parseFloat(s string) (f float64, err error) {
	if strings.Contains(s, "%") {
		f, err = parsePercentage(s)
	} else {
		f, err = strconv.ParseFloat(s, 64)
	}
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
