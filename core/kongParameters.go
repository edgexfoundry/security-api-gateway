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
 * @version: 0.1.0
 *******************************************************************************/
package main

import jwt "github.com/dgrijalva/jwt-go"

type KongService struct {
	Name     string `url:"name,omitempty"`
	Host     string `url:"host,omitempty"`
	Port     string `url:"port,omitempty"`
	Protocol string `url:"protocol,omitempty"`
}

type KongRoute struct {
	Paths []string `url:"paths[],omitempty"`
	Hosts []string `url:"hosts[],omitempty"`
}

type KongPlugin struct {
	Name string `url:"name,omitempty"`
}

type KongBasicAuthPlugin struct {
	Name            string `url:"name,omitempty"`
	HideCredentials string `url:"config.hide_credentials,omitempty"`
}

type KongUser struct {
	UserName string `url:"username,omitempty"`
	Password string `url:"password,omitempty"`
}

type CertPair struct {
	Cert string `json:"cert,omitempty"`
	Key  string `json:"key,omitempty"`
}

type CertCollect struct {
	Section CertPair `json:"data"`
}

type CertInfo struct {
	Cert string   `json:"cert,omitempty"`
	Key  string   `json:"key,omitempty"`
	Snis []string `json:"snis,omitempty"`
}

type JWTCred struct {
	ConsumerID string `json:"consumer_id, omitempty"`
	CreatedAt  int    `json:"created_at, omitempty"`
	ID         string `json:"id, omitempty"`
	Key        string `json:"key, omitempty"`
	Secret     string `json:"secret, omitempty"`
}

type KongJWTClaims struct {
	ISS  string `json:"iss"`
	Acct string `json:"account"`
	jwt.StandardClaims
}

type Item struct {
	ID string `json:"id"`
}

type DataCollect struct {
	Section []Item `json:"data"`
}
