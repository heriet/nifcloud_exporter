package main

import (
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

// Instance represents RDB instance
type Instance struct {
	Name string `yaml:"name"`
}

// RDBEnv represents environments for RDB
type RDBEnv struct {
	Name            string     `yaml:"name"`
	Region          string     `yaml:"region"`
	AccessKeyId     string     `yaml:"accessKeyId"`
	SecretAccessKey string     `yaml:"secretAccessKey"`
	Instances       []Instance `yaml:"instances"`
}

// Config contain exporter config
type Config struct {
	RDBEnv []RDBEnv `yaml:"rdb"`
}

// Load loads config from file
func (cfg *Config) Load(filename string) error {
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}
	return yaml.Unmarshal(content, &cfg)
}
