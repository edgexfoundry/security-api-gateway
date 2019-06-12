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
	"fmt"
	"encoding/json"
	"errors"
	"github.com/dghubble/sling"
 )

 type Service struct {
	 ProxyServiceURL string
	 SecretServiceURL string
	 Client *http.Client
 }

func (s *Service) CheckProxyStatus() error {
	req, err := sling.New().Get(s.ProxyServiceURL).Request()
	resp, err := s.Client.Do(req)
	if err != nil {
		lc.Error(fmt.Sprintf("The status of reverse proxy is unknown with error %s, the initialization is terminated.", err.Error()))
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		lc.Info("Reverse proxy is up successfully.")
		return nil
	} 

	e := fmt.Sprintf("The status of reverse proxy is unknown with error code %d, the initialization is terminated.", resp.StatusCode)
	return errors.New(e)
}

func (s *Service) CheckSecretServiceStatus() error {
	req, err := sling.New().Get(s.SecretServiceURL).Request()
	resp, err := s.Client.Do(req)
	if err != nil {
		lc.Error("The status of secret service is unknown, the initialization is terminated.")
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		lc.Info("Secret management service is up successfully.")
		return nil
	} else {
		e:=fmt.Sprintf("Secret management service is down. Please check the status of secret service with endpoint %s.", s.SecretServiceURL)
		return errors.New(e)
	}
}


func (s *Service) ResetProxy() {
	paths := []string{RoutesPath, ServicesPath, ConsumersPath, PluginsPath, CertificatesPath}
	for _, p := range paths {
		d, err := s.getIDListFromEndpoint(p)
		if err != nil {
			lc.Error(err.Error())
		} else {
			for _, c := range d.Section {
				err = deleteResource(c.ID, s.ProxyServiceURL, p, p, s.Client)
				if err != nil {
					lc.Error(err.Error())
				}
			}
		}
	}
}

func (s *Service) Init(config *tomlConfig) {	
	err := s.loadCert(config)
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
		serviceObject, err := s.initKongService(serviceParams)
		if err != nil {
			lc.Info(err.Error())
			continue
		}

		lc.Info(fmt.Sprintf("kong service object ID is %s", serviceObject.ID))

		// create the route using the Host as the same thing as the configured sni
		routeParams := &KongRoute{
			Paths: []string{"/" + service.Name},
			Name:  service.Name,
		}
		s.initKongRoutes(routeParams, service.Name)
	}

	s.initAuthmethod(config)
	s.initACL(config)

	lc.Info("Finishing initialization for reverse proxy.")
}

func (s *Service) loadCert(config *tomlConfig) error {
	cp := &CertPair{"", "", &Requestor{s.ProxyServiceURL, s.SecretServiceURL, s.Client}}		
	cert, key, err := cp.init(config)
	if err != nil {
		return err
	}
	body := &CertInfo{
		Cert: cert,
		Key:  key,
		Snis: []string{config.SecretService.SNIS},
	}

	lc.Info("Trying to upload cert to proxy server.")
	req, err := sling.New().Base(s.ProxyServiceURL).Post(CertificatesPath).BodyForm(body).Request()
	resp, err := s.Client.Do(req)
	if err != nil {
		lc.Error("Failed to upload cert to proxy server with error %s", err.Error())
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusCreated || resp.StatusCode == http.StatusConflict {
		lc.Info("Successful to add certificate to the reverse proxy.")
		return nil
	}

	return fmt.Errorf("failed to add certificate with errorcode %d", resp.StatusCode)
}

func (s *Service) initKongService(service *KongService) (*KongServiceResponse, error) {
	req, err := sling.New().Base(s.ProxyServiceURL).Post(ServicesPath).BodyForm(service).Request()
	resp, err := s.Client.Do(req)

	if err != nil {
		s := fmt.Sprintf("Failed to set up proxy service for %s.", service.Name)
		return nil, errors.New(s)
	} 
	defer resp.Body.Close()
	
	if resp.StatusCode == http.StatusCreated {
			lc.Info(fmt.Sprintf("Successful to set up proxy service for %s.", service.Name))

			serviceObj := KongServiceResponse{}
			err = json.NewDecoder(resp.Body).Decode(&serviceObj)
			if err != nil {
				return nil, err
			}
			return &serviceObj, nil
		} else if resp.StatusCode == http.StatusConflict {
			return nil, fmt.Errorf("proxy service for %s has been set up", service.Name)
		} else {
			return nil, fmt.Errorf("failed to set up proxy service for %s", service.Name)
	}
}


func (s *Service) initKongRoutes(r *KongRoute, name string) {
	routesubpath := "services/" + name + "/routes"
	req, err := sling.New().Base(s.ProxyServiceURL).Post(routesubpath).BodyJSON(r).Request()
	resp, err := s.Client.Do(req)
	if err != nil {
		s := fmt.Sprintf("Failed to set up routes for %s with error %s.", name, err.Error())
		lc.Error(s)		
	}
	defer resp.Body.Close()
	
	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusCreated || resp.StatusCode == http.StatusConflict {
		lc.Info(fmt.Sprintf("Successful to set up route for %s.", name))
	} else {
		s := fmt.Sprintf("Failed to set up route for %s with error %s.", name, resp.Status)
		lc.Error(s)	
	}
}

func (s *Service) initACL(config *tomlConfig) {
	lc.Info("Enabling global ACL for api gateway route.")
	aclParams := &KongACLPlugin{
		Name:      config.KongACL.Name,
		WhiteList: config.KongACL.WhiteList,
	}
	req, err := sling.New().Base(s.ProxyServiceURL).Post(PluginsPath).BodyForm(aclParams).Request()
	resp, err := s.Client.Do(req)
	if err != nil {
		s := fmt.Sprintf("Failed to set up acl.")
		lc.Error(s)
	} 
	defer resp.Body.Close()
	
	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusCreated || resp.StatusCode == http.StatusConflict {
		lc.Info("Successful to set up acl.")
	} else {
		s := fmt.Sprintf("Failed to set up acl with errorcode %d.", resp.StatusCode)
		lc.Error(s)
	}
}

func (s *Service) initAuthmethod(config *tomlConfig) {
	lc.Info(fmt.Sprintf("selected auth method as %s.", config.KongAuth.Name))
	if config.KongAuth.Name == "jwt" {
		s.initJWTAuth(config)
	} else if config.KongAuth.Name == "oauth2" {
		s.initOAuth2(config)
	}
}

func (s *Service) initJWTAuth(config *tomlConfig) {
	jwtParams := &KongJWTPlugin{
		Name: config.KongAuth.Name,
	}

	req, err := sling.New().Base(s.ProxyServiceURL).Post(PluginsPath).BodyForm(jwtParams).Request()
	resp, err := s.Client.Do(req)
	if err != nil {
		s := fmt.Sprintf("Failed to set up jwt authentication.")
		lc.Error(s)
	} 
	defer resp.Body.Close()
	
	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusCreated || resp.StatusCode == http.StatusConflict {
		lc.Info("Successful to set up jwt authentication")
	} else {
		s := fmt.Sprintf("Failed to set up jwt authentication with errorcode %d.", resp.StatusCode)
		lc.Error(s)
	}
}


func (s *Service) initOAuth2(config *tomlConfig) {
	oauth2Params := &KongOAuth2Plugin{
		Name:                    config.KongAuth.Name,
		Scope:                   OAuth2Scopes,
		MandatoryScope:          "true",
		EnableClientCredentials: "true",
		EnableGlobalCredentials: "true",
		TokenTTL:                config.KongAuth.TokenTTL,
	}

	req, err := sling.New().Base(s.ProxyServiceURL).Post(PluginsPath).BodyForm(oauth2Params).Request()
	resp, err := s.Client.Do(req)
	if err != nil {
		s := fmt.Sprintf("Failed to set up oauth2 authentication with error %s.", err.Error())
		lc.Error(s)
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusCreated || resp.StatusCode == http.StatusConflict {
			lc.Info("Successful to set up oauth2 authentication.")
	} else {
		s := fmt.Sprintf("Failed to set up oauth2 authentication with errorcode %d.", resp.StatusCode)
		lc.Error(s)
	}
}

func (s *Service) getIDListFromEndpoint(path string) (DataCollect, error) {
	collection := DataCollect{}

	req, err := sling.New().Get(s.ProxyServiceURL).Path(path).Request()
	resp, err := s.Client.Do(req)
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