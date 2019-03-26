/*******************************************************************************
 * Copyright 2018 Dell Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software distributed under the License
 * is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express
 * or implied. See the License for the specific language governing permissions and limitations under
 * the License.
 *
 * @author: Tingyu Zeng, Dell
 * @version: 0.5.0
 *******************************************************************************/
package main

import (
	"github.com/BurntSushi/toml"
)

type tomlConfig struct {
	Title         string
	KongURL       kongurl
	KongAuth      kongauth
	KongACL       KongACLPlugin
	SecretService secretservice
	EdgexServices map[string]service
}

type kongurl struct {
	Server             string
	AdminPort          string
	AdminPortSSL       string
	ApplicationPort    string
	ApplicationPortSSL string
}

type kongauth struct {
	Name                    string
	Scopes                  string
	MandatoryScope          string
	EnableClientCredentials string
	ClientId                string
	ClientSecret            string
	RedirectUri             string
	GrantType               string
	ScopeGranted            string
	Resource                string
}

type kongacl struct {
	Name      string
	WhiteList string
}

type secretservice struct {
	Server          string
	Port            string
	HealthcheckPath string
	CertPath        string
	TokenPath       string
	CACertPath      string
	SNIS            string
}

type service struct {
	Name     string
	Host     string
	Port     string
	Protocol string
}

func LoadTomlConfig(path string) (*tomlConfig, error) {
	config := tomlConfig{}
	_, err := toml.DecodeFile(path, &config)
	return &config, err
}
