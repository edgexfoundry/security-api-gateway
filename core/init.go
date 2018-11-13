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
 * @version: 0.1.0
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

		// create the kong service first so that we can get the
		// service ID to use when adding the route
		serviceObject, err := initKongService(baseURL, client, serviceParams)
		if err != nil {
			lc.Error(err.Error())
			return
		}

		// create the route using the Host as the same thing as the configured sni
		routeParams := &KongRoute{
			Paths: []string{"/" + service.Name},
			Hosts: []string{config.SecretService.SNIS},
			// only capture the ID from the service response
			Service: &KongServiceResponse{
				ID: serviceObject.ID,
			},
		}
		initKongRoutes(baseURL, client, routeParams, RoutesPath, service.Name)
		//pluginPath := fmt.Sprintf("%s%s/%s", ServicesPath, service.Name, PluginsPath)
		//initAuthmethodForService(config, baseURL, client, pluginPath, service.Name)
	}

	initAuthmethodForService(config, baseURL, client)
	initACLForServices(config, baseURL, client)
	//initKongAdminInterface(config, baseURL, client)

	lc.Info("Finishing initialization for reverse proxy.")
}

func initKongService(url string, c *http.Client, service *KongService) (*KongServiceResponse, error) {
	req, err := sling.New().Base(url).Post(ServicesPath).BodyForm(service).Request()
	resp, err := c.Do(req)
	if err != nil {
		s := fmt.Sprintf("Failed to set up proxy service for %s.", service.Name)
		return nil, errors.New(s)
	} else {
		if resp.StatusCode == 201 || resp.StatusCode == 409 {
			lc.Info(fmt.Sprintf("Successful to set up proxy service for %s.", service.Name))
			serviceObj := KongServiceResponse{}
			err = json.NewDecoder(resp.Body).Decode(&serviceObj)
			if err != nil {
				return nil, err
			}
			return &serviceObj, nil
		} else {
			return nil, fmt.Errorf("failed to set up proxy service for %s", service.Name)
		}
	}
}

func initACLForServices(config *tomlConfig, url string, c *http.Client) {
	lc.Info("Enabling ACL for api gateway service.")
	aclParams := &KongACLPlugin{
		Name:      config.KongACL.Name,
		WhiteList: config.KongACL.WhiteList,
	}
	req, err := sling.New().Base(url).Post(PluginsPath).BodyForm(aclParams).Request()
	resp, err := c.Do(req)
	if err != nil {
		s := fmt.Sprintf("Failed to set up acl.")
		lc.Error(s)
	} else {
		if resp.StatusCode == 200 || resp.StatusCode == 201 || resp.StatusCode == 409 {
			lc.Info("Successful to set up acl.")
		} else {
			s := fmt.Sprintf("Failed to set up acl with errorcode %d.", resp.StatusCode)
			lc.Error(s)
		}
	}
}

func initAuthmethodForService(config *tomlConfig, url string, c *http.Client) {
	lc.Info(fmt.Sprintf("selected auth method as %s.", config.KongAuth.Name))
	if config.KongAuth.Name == "jwt" {
		initJWTAuthForService(config, url, c)
	} else if config.KongAuth.Name == "oauth2" {
		initOauth2ForService(config, url, c)
	}
}

func initOauth2ForService(config *tomlConfig, url string, c *http.Client) {
	oauth2Params := &KongOAuth2Plugin{
		Name:                    config.KongAuth.Name,
		Scope:                   config.KongAuth.Scopes,
		MandatoryScope:          config.KongAuth.MandatoryScope,
		EnableClientCredentials: config.KongAuth.EnableClientCredentials,
	}

	req, err := sling.New().Base(url).Post(PluginsPath).BodyForm(oauth2Params).Request()
	resp, err := c.Do(req)
	if err != nil {
		s := fmt.Sprintf("Failed to set up oauth2 authentication with error %s.", err.Error())
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
func initJWTAuthForService(config *tomlConfig, url string, c *http.Client) {
	jwtParams := &KongJWTPlugin{
		Name: config.KongAuth.Name,
	}

	req, err := sling.New().Base(url).Post(PluginsPath).BodyForm(jwtParams).Request()
	resp, err := c.Do(req)
	if err != nil {
		s := fmt.Sprintf("Failed to set up jwt authentication.")
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

func initKongRoutes(url string, c *http.Client, r *KongRoute, path string, name string) {
	req, err := sling.New().Base(url).Post(path).BodyJSON(r).Request()
	resp, err := c.Do(req)
	if err != nil {
		s := fmt.Sprintf("Failed to set up routes for %s with error %s.", name, err.Error())
		lc.Error(s)
	} else {
		if resp.StatusCode == 200 || resp.StatusCode == 201 || resp.StatusCode == 409 {
			lc.Info(fmt.Sprintf("Successful to set up route for %s.", name))
		} else {
			s := fmt.Sprintf("Failed to set up route for %s with errorcode %d.", name, resp.StatusCode)
			lc.Error(s)
		}
	}
}

//redirect request for 8001 to an admin service of 8000, and add authentication
func initKongAdminInterface(config *tomlConfig, url string, c *http.Client) {
	adminServiceParams := &KongService{
		Name:     "admin",
		Host:     config.KongURL.Server,
		Port:     config.KongURL.AdminPort,
		Protocol: "http",
	}
	req, err := sling.New().Base(url).Post(ServicesPath).BodyForm(adminServiceParams).Request()
	resp, err := c.Do(req)
	if err != nil {
		s := fmt.Sprintf("Failed to set up service for admin loopback with error %s.", err.Error())
		lc.Error(s)
	} else {
		if resp.StatusCode == 200 || resp.StatusCode == 201 || resp.StatusCode == 409 {
			lc.Info("Successful to set up admin loopback.")
		} else {
			lc.Error("Failed to set up admin loopback.")
		}
	}

	adminRouteParams := &KongRoute{Paths: []string{"/admin"}}
	adminRoutePath := fmt.Sprintf("%sadmin/routes", ServicesPath)
	req, err = sling.New().Base(url).Post(adminRoutePath).BodyJSON(adminRouteParams).Request()
	resp, err = c.Do(req)
	if err != nil {
		lc.Error("Failed to set up admin service route.")
	} else {
		if resp.StatusCode == 200 || resp.StatusCode == 201 || resp.StatusCode == 409 {
			lc.Info("Successful to set up admin service routes.")
		} else {
			lc.Error("Failed to set up admin service routes.")
		}
	}
}
