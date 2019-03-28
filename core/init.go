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

func initSecurityServices(config *tomlConfig, baseURL string, secretBaseURL string, client *http.Client) {
	// load the certificates into kong first so that when the services' routes are
	// created with the host specified as the snis, it matches the snis
	// set for the certificate and thus kong will serve up these certificates when the
	// proxies are being accessed (and also the https client supports SNI)
	err := loadKongCerts(config, baseURL, secretBaseURL, client)
	if err != nil {
		lc.Error(err.Error())
	}

	for _, service := range config.EdgexServices {
		serviceParams := &KongService{
			Name:     service.Name,
			Host:     service.Host,
			Port:     service.Port,
			Protocol: service.Protocol,
		}

		// create the kong service first so that we get the service ID that is associated with the route
		serviceObject, err := initKongService(baseURL, client, serviceParams)
		if err != nil {
			lc.Info(err.Error())
			continue
		}

		lc.Info(serviceObject.ID)

		// create the route using the Host as the same thing as the configured sni
		routeParams := &KongRoute{
			Paths: []string{"/" + service.Name},
			Name:  service.Name,
		}
		initKongRoutes(baseURL, client, routeParams, service.Name)

		initAuthmethodForRoute(config, baseURL, client, service.Name)
		initACLForRoute(config, baseURL, client, service.Name)
	}

	lc.Info("Finishing initialization for reverse proxy.")
}

func initKongService(url string, c *http.Client, service *KongService) (*KongServiceResponse, error) {
	req, err := sling.New().Base(url).Post(ServicesPath).BodyForm(service).Request()
	resp, err := c.Do(req)

	if err != nil {
		s := fmt.Sprintf("Failed to set up proxy service for %s.", service.Name)
		defer resp.Body.Close()
		return nil, errors.New(s)
	} else {
		if resp.StatusCode == 201 {
			lc.Info(fmt.Sprintf("Successful to set up proxy service for %s.", service.Name))

			serviceObj := KongServiceResponse{}
			err = json.NewDecoder(resp.Body).Decode(&serviceObj)
			if err != nil {
				return nil, err
			}
			return &serviceObj, nil
		} else if resp.StatusCode == 409 {
			return nil, fmt.Errorf("proxy service for %s has been set up", service.Name)
		} else {
			return nil, fmt.Errorf("failed to set up proxy service for %s", service.Name)
		}
	}
}

func initACLForRoute(config *tomlConfig, url string, c *http.Client, service string) {
	lc.Info("Enabling ACL for api gateway route.")
	aclParams := &KongACLPlugin{
		Name:      config.KongACL.Name,
		WhiteList: config.KongACL.WhiteList,
	}
	req, err := sling.New().Base(url + "routes/" + service + "/").Post(PluginsPath).BodyForm(aclParams).Request()
	resp, err := c.Do(req)
	if err != nil {
		s := fmt.Sprintf("Failed to set up acl.")
		lc.Error(s)
		defer resp.Body.Close()
	} else {
		if resp.StatusCode == 200 || resp.StatusCode == 201 || resp.StatusCode == 409 {
			lc.Info("Successful to set up acl.")
		} else {
			s := fmt.Sprintf("Failed to set up acl with errorcode %d.", resp.StatusCode)
			lc.Error(s)
		}
	}
}

func initAuthmethodForRoute(config *tomlConfig, url string, c *http.Client, service string) {
	lc.Info(fmt.Sprintf("selected auth method as %s.", config.KongAuth.Name))
	if config.KongAuth.Name == "jwt" {
		initJWTAuthForRoute(config, url, c, service)
	} else if config.KongAuth.Name == "oauth2" {
		initOauth2ForService(config, url, c, service)
	}
}

func initOauth2ForService(config *tomlConfig, url string, c *http.Client, service string) {
	oauth2Params := &KongOAuth2Plugin{
		Name:                    config.KongAuth.Name,
		Scope:                   config.KongAuth.Scopes,
		MandatoryScope:          config.KongAuth.MandatoryScope,
		EnableClientCredentials: config.KongAuth.EnableClientCredentials,
	}

	req, err := sling.New().Base(url + "services/" + service + "/").Post(PluginsPath).BodyForm(oauth2Params).Request()
	resp, err := c.Do(req)
	if err != nil {
		s := fmt.Sprintf("Failed to set up oauth2 authentication with error %s.", err.Error())
		defer resp.Body.Close()
		lc.Error(s)
	} else {
		if resp.StatusCode == 200 || resp.StatusCode == 201 || resp.StatusCode == 409 {
			lc.Info("Successful to set up oauth2 authentication.")
		} else {
			s := fmt.Sprintf("Failed to set up oauth2 authentication with errorcode %d.", resp.StatusCode)
			lc.Error(s)
		}
	}

}

func initJWTAuthForRoute(config *tomlConfig, url string, c *http.Client, service string) {
	jwtParams := &KongJWTPlugin{
		Name: config.KongAuth.Name,
	}

	req, err := sling.New().Base(url + "routes/" + service + "/").Post(PluginsPath).BodyForm(jwtParams).Request()
	resp, err := c.Do(req)
	if err != nil {
		s := fmt.Sprintf("Failed to set up jwt authentication.")
		defer resp.Body.Close()
		lc.Error(s)
	} else {
		if resp.StatusCode == 200 || resp.StatusCode == 201 || resp.StatusCode == 409 {
			lc.Info("Successful to set up jwt authentication")
		} else {
			s := fmt.Sprintf("Failed to set up jwt authentication with errorcode %d.", resp.StatusCode)
			lc.Error(s)
		}
	}
}

func initKongRoutes(url string, c *http.Client, r *KongRoute, name string) {
	routesubpath := "services/" + name + "/routes"
	lc.Info(routesubpath)
	req, err := sling.New().Base(url).Post(routesubpath).BodyJSON(r).Request()
	resp, err := c.Do(req)
	if err != nil {
		s := fmt.Sprintf("Failed to set up routes for %s with error %s.", name, err.Error())
		defer resp.Body.Close()
		lc.Error(s)
	} else {
		if resp.StatusCode == 200 || resp.StatusCode == 201 || resp.StatusCode == 409 {
			lc.Info(fmt.Sprintf("Successful to set up route for %s.", name))
		} else {
			s := fmt.Sprintf("Failed to set up route for %s with error %s.", name, resp.Status)
			lc.Error(s)
		}
	}
}
