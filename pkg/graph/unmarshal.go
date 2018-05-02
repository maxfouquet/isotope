package graph

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	units "github.com/docker/go-units"
	yaml "gopkg.in/yaml.v2"
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
		script, err := parseScript(yamlSettings.Script, defaults.PayloadSize)
		if err != nil {
			return err
		}
		g.Services[name] = Service{
			ServiceSettings: serviceSettings,
			Script:          script,
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
	ComputeUsage *string       `yaml:"computeUsage"`
	MemoryUsage  *string       `yaml:"memoryUsage"`
	ErrorRate    *string       `yaml:"errorRate"`
	Script       []interface{} `yaml:"script"`
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

func parseScript(script []interface{}, defaultPayloadSize int64) (
	commands []Executable, err error) {
	for _, step := range script {
		command, err := parseCommand(step, defaultPayloadSize)
		if err != nil {
			return nil, err
		}
		commands = append(commands, command)
	}
	return
}

// parseStep should be run on each step in the script the step may either be a
// command or list of commands.
// Step may be:
// - A command
// - A list of commands
// A command may be:
// - sleep: time.Duration
// - get|http...: string
// - get|http...: {service:string payloadSize:units.ByteSize}
func parseCommand(step interface{}, defaultPayloadSize int64) (
	command Executable, err error) {
	switch val := step.(type) {
	case map[interface{}]interface{}:
		command, err = parseSingleCommand(val, defaultPayloadSize)
	case []map[interface{}]interface{}:
		command, err = parseConcurrentCommand(val, defaultPayloadSize)
	default:
		err = fmt.Errorf("invalid step type %T", val)
	}
	return
}

func parseSingleCommand(
	yamlCmd map[interface{}]interface{}, defaultPayloadSize int64) (
	command Executable, err error) {
	if len(yamlCmd) != 1 {
		return nil, fmt.Errorf(
			"command must contain a single key: %v", yamlCmd)
	}
	// This for loop will only iterate once.
	for interfaceKey, val := range yamlCmd {
		key, ok := interfaceKey.(string)
		if !ok {
			return nil, fmt.Errorf("key is not a string")
		}
		if key == "sleep" {
			command, err = parseSleepCommand(val)
		} else {
			var httpMethod HTTPMethod
			httpMethod, err = HTTPMethodFromString(key)
			if err != nil {
				return nil, fmt.Errorf("unknown command: %s", key)
			}
			command, err = parseRequestCommand(
				val, httpMethod, defaultPayloadSize)
		}
	}
	return
}

func parseSleepCommand(yaml interface{}) (command SleepCommand, err error) {
	if s, ok := yaml.(string); ok {
		command.Duration, err = time.ParseDuration(s)
	} else {
		err = fmt.Errorf("expected a duration expressed as a string")
	}
	return
}

// A request command may be expressed as:
// - get|http...: string
// - get|http...: {service:string payloadSize:units.ByteSize}
func parseRequestCommand(
	impl interface{}, httpMethod HTTPMethod, defaultPayloadSize int64) (
	command RequestCommand, err error) {
	command.Method = httpMethod
	if s, ok := impl.(string); ok {
		command.ServiceName = s
		command.PayloadSize = defaultPayloadSize
	} else if _, ok := impl.(map[interface{}]interface{}); ok {
		// TODO: Should just use map directly.
		var request struct {
			ServiceName string `yaml:"service"`
			PayloadSize string `yaml:"payloadSize"`
		}
		yamlString, err := yaml.Marshal(impl)
		if err != nil {
			return command, err
		}
		err = yaml.Unmarshal(yamlString, &request)
		if err != nil {
			return command, err
		}
		command.ServiceName = request.ServiceName
		command.PayloadSize, err = units.FromHumanSize(request.PayloadSize)
	} else {
		err = fmt.Errorf("unknown type of request command: %s", impl)
	}
	return
}

func parseConcurrentCommand(impl interface{}, defaultPayloadSize int64) (
	command ConcurrentCommand, err error) {
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
