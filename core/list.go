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
	"errors"
	"fmt"
	"net/http"

	"github.com/dghubble/sling"
)

func getIDListFromEndpoint(url string, path string, c *http.Client) (DataCollect, error) {
	req, err := sling.New().Get(url).Path(path).Request()
	resp, err := c.Do(req)
	if err != nil {
		s := fmt.Sprintf("Failed to get list of %s with error %s.", path, err.Error())
		lc.Error(s)
		return DataCollect{}, errors.New(s)
	}
	defer resp.Body.Close()
	collection := DataCollect{}
	json.NewDecoder(resp.Body).Decode(&collection)
	return collection, nil
}
