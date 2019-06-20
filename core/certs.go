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
package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/dghubble/sling"
)

func loadKongCerts(config *tomlConfig, url string, secretBaseURL string, c *http.Client) error {
	cert, key, err := getCertKeyPair(config, secretBaseURL, c)
	if err != nil {
		return err
	}
	body := &CertInfo{
		Cert: cert,
		Key:  key,
		Snis: []string{config.SecretService.SNIS},
	}

	lc.Info("Trying to upload cert to proxy server.")
	req, err := sling.New().Base(url).Post(CertificatesPath).BodyJSON(body).Request()
	resp, err := c.Do(req)
	if err != nil {
		lc.Error("Failed to upload cert to proxy server with error %s", err.Error())
		return err
	}

	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusCreated || resp.StatusCode == http.StatusConflict {
		lc.Info("Successful to add certificate to the reverse proxy.")
		return nil
	}

	return fmt.Errorf("failed to add certificate with errorcode %d", resp.StatusCode)
}

func getCertKeyPair(config *tomlConfig, secretBaseURL string, c *http.Client) (string, string, error) {

	t, err := getSecret(config.SecretService.TokenPath)
	if err != nil {
		return "", "", err
	}

	s := sling.New().Set(VaultToken, t)
	req, err := s.New().Base(secretBaseURL).Get(config.SecretService.CertPath).Request()
	resp, err := c.Do(req)
	if err != nil {
		errStr := fmt.Sprintf("Failed to retrieve certificate with path as %s with error %s", config.SecretService.CertPath, err.Error())
		return "", "", errors.New(errStr)
	}
	defer resp.Body.Close()

	collection := CertCollect{}
	json.NewDecoder(resp.Body).Decode(&collection)
	lc.Info(collection.Section.Cert)
	lc.Info(fmt.Sprintf("successful on retrieving certificate from %s.", config.SecretService.CertPath))
	return collection.Section.Cert, collection.Section.Key, nil
}
