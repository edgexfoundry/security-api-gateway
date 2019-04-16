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
	"fmt"
	"net/http"
	"os"

	"github.com/dghubble/sling"
)

func checkProxyStatus(url string, c *http.Client) {
	req, err := sling.New().Get(url).Request()
	resp, err := c.Do(req)
	if err != nil {
		lc.Error(fmt.Sprintf("The status of reverse proxy is unknown with error %s, the initialization is terminated.", err.Error()))
		os.Exit(0)
	} else {
		if resp.StatusCode == 200 {
			lc.Info("Reverse proxy is up successfully.")
		} else {
			lc.Error(fmt.Sprintf("The status of reverse proxy is unknown with error code %d, the initialization is terminated.", resp.StatusCode))
			os.Exit(0)
		}
	}
}

func checkSecretServiceStatus(url string, c *http.Client) {
	req, err := sling.New().Get(url).Request()
	resp, err := c.Do(req)
	if err != nil {
		lc.Error("The status of secret service is unknown, the initialization is terminated.")
		os.Exit(0)
	} else {
		if resp.StatusCode == 200 {
			lc.Info("Secret management service is up successfully.")
		} else {
			lc.Error(fmt.Sprintf("Secret management service is down. Please check the status of secret service with endpoint %s.", url))
			os.Exit(0)
		}
	}
}
