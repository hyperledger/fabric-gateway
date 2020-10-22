/*
Copyright 2020 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/kelseyhightower/envconfig"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

type config struct {
	Gateway struct {
		ID             string   `yaml:"id"`
		ListenAddress  string   `yaml:"listenAddress"`
		MspID          string   `yaml:"mspId`
		BootstrapPeers []string `yaml:"bootstrapPeers"`
		Cert           struct {
			File string
		}
		Key struct {
			File string
		}
		TLS struct {
			Enabled            bool
			ClientAuthRequired bool
			RootCert           struct {
				File string
			}
			Cert struct {
				File string
			}
			Key struct {
				File string
			}
		} `yaml:"tls"`
	}
	certificatePEM []byte
	keyPEM         []byte
	tlsRootCertPEM []byte
}

func loadConfig() (*config, error) {
	conf := &config{}
	cfgFile := os.Getenv("FABRIC_CFG_PATH")
	if cfgFile != "" {
		cfg, err := ioutil.ReadFile(cfgFile + "/gateway.yaml")
		if err != nil {
			errors.Wrap(err, "No config yaml found at location: "+cfgFile)
		}
		err = yaml.Unmarshal(cfg, &conf)
		if err != nil {
			errors.Wrap(err, "failed to parse gateway config")
		}
	}
	// apply any env-var overrides
	err := envconfig.Process("GATEWAY", &conf.Gateway)
	if err != nil {
		errors.Wrap(err, "Failed to apply env-var overrides")
	}
	fmt.Println(conf)

	return conf, nil
}

func (c *config) BootstrapPeers() []string {
	return c.Gateway.BootstrapPeers
}

func (c *config) MspID() string {
	return c.Gateway.MspID
}

func (c *config) Certificate() []byte {
	if c.certificatePEM == nil {
		c.certificatePEM, _ = ioutil.ReadFile(c.Gateway.Cert.File)
	}
	return c.certificatePEM
}

func (c *config) Key() []byte {
	if c.keyPEM == nil {
		c.keyPEM, _ = ioutil.ReadFile(c.Gateway.Key.File)
	}
	return c.keyPEM
}

func (c *config) TLSRootCert() []byte {
	if c.tlsRootCertPEM == nil {
		c.tlsRootCertPEM, _ = ioutil.ReadFile(c.Gateway.TLS.RootCert.File)
	}
	return c.tlsRootCertPEM
}
