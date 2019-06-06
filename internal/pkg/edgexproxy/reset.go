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
	"errors"
	"fmt"
	"encoding/json"
	"github.com/dghubble/sling"
)

func ResetProxy(url string, client *http.Client) {
	paths := []string{RoutesPath, ServicesPath, ConsumersPath, PluginsPath, CertificatesPath}
	for _, p := range paths {
		d, err := getIDListFromEndpoint(url, p, client)
		if err != nil {
			lc.Error(err.Error())
		} else {
			for _, c := range d.Section {
				err = deleteResource(c.ID, url, p, p, client)
				if err != nil {
					lc.Error(err.Error())
				}
			}
		}
	}
}

func getIDListFromEndpoint(url string, path string, c *http.Client) (DataCollect, error) {
	collection := DataCollect{}

	req, err := sling.New().Get(url).Path(path).Request()
	resp, err := c.Do(req)
	if err != nil {
		s := fmt.Sprintf("Failed to get list of %s with error %s.", path, err.Error())		
		return collection, errors.New(s)
	}
	defer resp.Body.Close()

	json.NewDecoder(resp.Body).Decode(&collection)
	return collection, nil
}

func deleteResource(id string, url string, path string, endpoint string, c *http.Client) error {
	req, err := sling.New().Base(url).Path(path).Delete(id).Request()
	resp, err := c.Do(req)
	if err != nil {
		s := fmt.Sprintf("Failed to delete %s at %s with error %s.", id, endpoint, err.Error())		
		return errors.New(s)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusCreated || resp.StatusCode == http.StatusNoContent {
		lc.Info(fmt.Sprintf("Successful to delete %s at %s.", id, endpoint))
		return nil
	}
	s := fmt.Sprintf("Failed to delete %s at %s with errocode %d.", id, endpoint, resp.StatusCode)	
	return errors.New(s)
}
