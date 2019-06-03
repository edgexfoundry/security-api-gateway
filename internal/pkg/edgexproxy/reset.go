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
)

func ResetProxy(url string, client *http.Client) {
	paths := []string{RoutesPath, ServicesPath, ConsumersPath, PluginsPath, CertificatesPath}
	for _, p := range paths {
		d, err := getIDListFromEndpoint(url, p, client)
		if err != nil {
			lc.Error(err.Error())
		} else {
			for _, c := range d.Section {
				deleteResource(c.ID, url, p, p, client)
			}
		}
	}
}
