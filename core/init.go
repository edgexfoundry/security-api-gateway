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
	"fmt"
	"net/http"

	"github.com/dghubble/sling"
)

func initSecurityServices(config *tomlConfig, baseURL string, secretBaseURL string, client *http.Client) {
	for _, service := range config.EdgexServices {
		serviceParams := &KongService{
			Name:     service.Name,
			Host:     service.Host,
			Port:     service.Port,
			Protocol: service.Protocol,
		}

		initKongService(baseURL, client, serviceParams)
		jwtServicePath := fmt.Sprintf("%s%s/%s", ServicesPath, service.Name, PluginsPath)
		initJWTAuthForService(baseURL, client, jwtServicePath, service.Name)
	}

	for _, service := range config.EdgexServices {
		routeParams := &KongRoute{
			Paths: []string{"/" + service.Name},
			Hosts: []string{EdgeXService},
		}
		routePath := fmt.Sprintf("%s%s/%s", ServicesPath, service.Name, RoutesPath)
		initKongRoutes(baseURL, client, routeParams, routePath, service.Name)
	}

	initKongAdminInterface(config, baseURL, client)
	err := loadKongCerts(config, baseURL, secretBaseURL, client)
	if err != nil {
		lc.Error(err.Error())
	}
	lc.Info("Finishing initialization for reverse proxy.")
}

func initKongService(url string, c *http.Client, service *KongService) {
	req, err := sling.New().Base(url).Post(ServicesPath).BodyForm(service).Request()
	resp, err := c.Do(req)
	if err != nil {
		s := fmt.Sprintf("Failed to set up proxy service for %s.", service.Name)
		lc.Error(s)
	} else {
		if resp.StatusCode == 201 || resp.StatusCode == 409 {
			lc.Info(fmt.Sprintf("Successful to set up proxy service for %s.", service.Name))
		} else {
			lc.Error(fmt.Sprintf("Failed to set up proxy service for %s.", service.Name))
		}
	}
}

func initJWTAuthForService(url string, c *http.Client, path string, name string) {
	jwtParams := &KongPlugin{
		Name: "jwt",
	}

	req, err := sling.New().Base(url).Post(path).BodyForm(jwtParams).Request()
	resp, err := c.Do(req)
	if err != nil {
		s := fmt.Sprintf("Failed to set up jwt authentication for service %s with error %s.", name, err.Error())
		lc.Error(s)
	} else {
		if resp.StatusCode == 200 || resp.StatusCode == 201 || resp.StatusCode == 409 {
			lc.Info(fmt.Sprintf("Successful to set up jwt authentication for service %s.", name))
		} else {
			s := fmt.Sprintf("Failed to set up jwt authentication for service %s with errorcode %d.", name, resp.StatusCode)
			lc.Error(s)
		}
	}
}

func initKongRoutes(url string, c *http.Client, r *KongRoute, path string, name string) {
	req, err := sling.New().Base(url).Post(path).BodyForm(r).Request()
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
	req, err = sling.New().Base(url).Post(adminRoutePath).BodyForm(adminRouteParams).Request()
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

	jwtAdminServicePath := "services/admin/plugins"
	initJWTAuthForService(url, c, jwtAdminServicePath, "admin")
}
