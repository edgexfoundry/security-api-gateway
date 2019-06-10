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
 * @version: 1.0.0
 *******************************************************************************/
package edgexproxy

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"github.com/dghubble/sling"
)

type CertPair struct {
	Cert string `json:"cert,omitempty"`
	Key  string `json:"key,omitempty"`	
	Request *Requestor
}

type Requestor struct {	
	ProxyBaseURL string
	SecretSvcBaseURL string
	Client *http.Client 
}

type CertCollect struct {
	Section CertPair `json:"data"`
}

type Auth struct {
	Secret Inner `json:"auth"`
}

type Inner struct {
	Token string `json:"client_token"`
}

func (cp *CertPair) init(config *tomlConfig) (string, string, error) {

	t, err := cp.getSecret(config.SecretService.TokenPath)
	if err != nil {
		return "", "", err
	}

	s := sling.New().Set(VaultToken, t)
	req, err := s.New().Base(cp.Request.SecretSvcBaseURL).Get(config.SecretService.CertPath).Request()
	resp, err := cp.Request.Client.Do(req)
	if err != nil {
		errStr := fmt.Sprintf("Failed to retrieve certificate with path as %s with error %s", config.SecretService.CertPath, err.Error())
		return "", "", errors.New(errStr)
	}
	defer resp.Body.Close()

	collection := CertCollect{}
	json.NewDecoder(resp.Body).Decode(&collection)
	lc.Info(collection.Section.Cert)
	lc.Info(fmt.Sprintf("successful on retrieving certificate from %s.", config.SecretService.CertPath))
	cp.Cert = collection.Section.Cert
	cp.Key = collection.Section.Key
	return collection.Section.Cert, collection.Section.Key, nil
}

func (cp *CertPair) getSecret(filename string) (string, error) {
	s := Auth{}
	raw, err := ioutil.ReadFile(filename)
	if err != nil {
		return s.Secret.Token, err
	}

	err = json.Unmarshal(raw, &s)
	return s.Secret.Token, err
}
