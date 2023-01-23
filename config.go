package main

import (
	"fmt"
	"io/ioutil"

	"github.com/creasty/defaults"
	"gopkg.in/yaml.v2"
)

type Token struct {
	Contract string `yaml:"contract"`
	Symbol   string `yaml:"symbol"`
	Decimal  int    `yaml:"decimal,omitempty" default:"18"`
}

type Wallet struct {
	Name     string              `yaml:"name"`
	Address  string              `yaml:"address"`
	TrackFor map[string][]string `yaml:"track_for"`
	Labels   map[string]string   `yaml:"labels"`
}

type ConfigYAML struct {
	Chains  map[string]string  `yaml:"chains"`
	Tokens  map[string][]Token `yaml:"tokens"`
	Wallets []Wallet           `yaml:"wallets"`
}

func ReadConfig(path string) (*ConfigYAML, error) {
	var c ConfigYAML
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("can't read file: %v", err)
	}
	if err := yaml.Unmarshal(b, &c); err != nil {
		return nil, fmt.Errorf("can't unmarshal: %v", err)
	}
	if err := defaults.Set(&c); err != nil {
		return nil, fmt.Errorf("can't sets defaults: %v", err)
	}
	return &c, nil
}
