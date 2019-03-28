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
	"encoding/json"
	"io/ioutil"
)

type Auth struct {
	Secret Inner `json:"auth"`
}

type Inner struct {
	Token string `json:"client_token"`
}

type userTokenPair struct {
	User  string
	Token string
}

func getSecret(filename string) (string, error) {
	s := Auth{}
	raw, err := ioutil.ReadFile(filename)
	if err != nil {
		return s.Secret.Token, err
	}

	err = json.Unmarshal(raw, &s)
	return s.Secret.Token, err
}

func createTokenFile(u string, t string, filename string) error {

	data := userTokenPair{
		User:  u,
		Token: t,
	}

	jdata, err := json.MarshalIndent(data, "", " ")
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(filename, jdata, 0644)
	return err
}
