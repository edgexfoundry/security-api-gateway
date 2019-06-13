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
	"net/http"
	"testing"
)

type testRequestor struct {
}

func (tr *testRequestor) GetProxyBaseURL() string {
	return "test"
}

func (tr *testRequestor) GetSecretSvcBaseURL() string {
	return "test"
}

func (tr *testRequestor) GetHttpClient() *http.Client {
	return &http.Client{}
}

type testCertCfg struct {
}

func (tc *testCertCfg) GetCertPath() string {
	return "test"
}

func (tc *testCertCfg) GetTokenPath() string {
	return "test"
}

func TestGetSecret(t *testing.T) {
	path := "../../../test/test-resp-init.json"
	cs := Certs{&testRequestor{}, &testCertCfg{}}
	s, err := cs.getSecret(path)
	if err != nil {
		t.Errorf("failed to parse token file")
		t.Errorf(err.Error())
	}
	if s != "test-token" {
		t.Errorf("incorrect token")
		t.Errorf(s)
	}
}

func TestValidate(t *testing.T) {
	cc := CertCollect{CertPair{"private-cert", "private-key"}}
	cs := Certs{&testRequestor{}, &testCertCfg{}}
	_, err := cs.validate(cc)
	if err != nil {
		t.Errorf("failed to validate cert collection")
	}
}
