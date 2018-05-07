package main

import (
	"fmt"
	"time"

	yaml "gopkg.in/yaml.v2"
)

// Executable is the top-level interface for commands.
type Executable interface {
	Execute() error
}

type SleepCommand struct {
	Duration time.Duration `yaml:"duration"`
}

func (c SleepCommand) Execute() error {
	time.Sleep(c.Duration)
	return nil
}

func main2() {
	cmd := []SleepCommand{
		{
			Duration: 100 * time.Millisecond,
		},
		{
			Duration: 100 * time.Millisecond,
		},
	}
	s, _ := yaml.Marshal(cmd)
	fmt.Printf("%s\n", s)
}
