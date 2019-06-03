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
	"errors"
	"fmt"
	"net/http"

	"github.com/dghubble/sling"
)

func deleteResource(id string, url string, path string, endpoint string, c *http.Client) error {
	req, err := sling.New().Base(url).Path(path).Delete(id).Request()
	resp, err := c.Do(req)
	if err != nil {
		s := fmt.Sprintf("Failed to delete %s at %s with error %s.", id, endpoint, err.Error())
		lc.Error(s)
		return errors.New(s)
	}
	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusCreated || resp.StatusCode == http.StatusNoContent {
		lc.Info(fmt.Sprintf("Successful to delete %s at %s.", id, endpoint))
		return nil
	}
	s := fmt.Sprintf("Failed to delete %s at %s with errocode %d.", id, endpoint, resp.StatusCode)
	lc.Error(s)
	return errors.New(s)
}
