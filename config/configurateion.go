package config

import (
	"encoding/json"
	"fmt"
	"os"
)

// Config Main application Configuration
type Config struct {
	Debug    bool           `json:"debug"`
	Terminal TerminalConfig `json:"terminal"`
	Teams    []TeamConfig   `json:"teams"`
}

type TerminalConfig struct {
	Notify bool `json:"notify"`
	Title  bool `json:"title"`
}

// TeamConfig Configuration of a slack team
type TeamConfig struct {
	Name           string  `json:"name"`
	Token          string  `json:"slack_token"`
	AutoConnect    bool    `json:"auto_connect"`
	AutoJoin       bool    `json:"auto_join"`
	Colorize       bool    `json:"colorize"`
	ColorizeInline bool    `json:"colorize_inline"`
	History        History `json:"history"`
}

// History Configuration parameters regarding conversation history
type History struct {
	Fetch bool `json:"fetch"`
	Size  int  `json:"size"`
}

// Load Load configuration structures from a file
func Load(filename string) (*Config, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("unable to read %s: %v", filename, err)
	}
	cfg := defaultCfg()
	err = json.NewDecoder(file).Decode(cfg)
	if err != nil {
		return nil, fmt.Errorf("unable to dcode %s: %v", filename, err)
	}
	return cfg, nil
}

// GetTeamConfig Return configuration for a specific team
func (c *Config) GetTeamConfig(teamName string) (*TeamConfig, error) {
	for _, team := range c.Teams {
		if team.Name == teamName {
			return &team, nil
		}
	}
	return nil, fmt.Errorf("team %s not found inconfig", teamName)
}

func defaultCfg() *Config {
	// TODO: implement sane config defaults
	return &Config{}
}
