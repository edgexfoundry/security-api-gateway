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
	"encoding/json"
	"errors"
	"fmt"
	"github.com/dghubble/sling"
	"net/http"
)

type Service struct {
	Connect    Requestor
	CertCfg    CertConfig
	ServiceCfg ServiceConfig
}

type ServiceConfig interface {
	GetProxyAuthMethod() string
	GetProxyAuthTTL() int
	GetProxyAuthResource() string
	GetProxyACLName() string
	GetProxyACLWhiteList() string
	GetSecretSvcSNIS() string
	GetEdgeXSvcs() map[string]service
}

func (s *Service) CheckProxyServiceStatus() error {
	return s.checkServiceStatus(s.Connect.GetProxyBaseURL())
}

func (s *Service) CheckSecretServiceStatus() error {
	return s.checkServiceStatus(s.Connect.GetSecretSvcBaseURL())
}

func (s *Service) checkServiceStatus(path string) error {
	req, err := sling.New().Get(path).Request()
	resp, err := s.Connect.GetHttpClient().Do(req)
	if err != nil {
		e := fmt.Sprintf("the status of service on %s is unknown, the initialization is terminated", path)
		return errors.New(e)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		lc.Info(fmt.Sprintf("the service on %s is up successfully", path))
		return nil
	}

	e := fmt.Sprintf("the service on %s is down", path)
	return errors.New(e)
}

func (s *Service) ResetProxy() error {
	paths := []string{RoutesPath, ServicesPath, ConsumersPath, PluginsPath, CertificatesPath}
	for _, path := range paths {
		d, err := s.getSvcIDs(path)
		if err != nil {
			return err
		}
		for _, c := range d.Section {
			r := &Resource{c.ID, s.Connect}
			err = r.Remove(path)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *Service) Init() error {
	err := s.loadCert()
	if err != nil {
		return err
	}

	for _, service := range s.ServiceCfg.GetEdgeXSvcs() {
		serviceParams := &KongService{
			Name:     service.Name,
			Host:     service.Host,
			Port:     service.Port,
			Protocol: service.Protocol,
		}

		err := s.initKongService(serviceParams)
		if err != nil {
			return err
		}

		routeParams := &KongRoute{
			Paths: []string{"/" + service.Name},
			Name:  service.Name,
		}
		err = s.initKongRoutes(routeParams, service.Name)
		if err != nil {
			return err
		}
	}

	err = s.initAuthmethod(s.ServiceCfg.GetProxyAuthMethod(), s.ServiceCfg.GetProxyAuthTTL())
	if err != nil {
		return err
	}

	err = s.initACL(s.ServiceCfg.GetProxyACLName(), s.ServiceCfg.GetProxyACLWhiteList())
	if err != nil {
		return err
	}

	lc.Info("finishing initialization for reverse proxy")
	return nil
}

func (s *Service) loadCert() error {
	cp, err := s.getCertPair()
	if err != nil {
		return err
	}
	body := &CertInfo{
		Cert: cp.Cert,
		Key:  cp.Key,
		Snis: []string{s.ServiceCfg.GetSecretSvcSNIS()},
	}

	lc.Info("trying to upload cert to proxy server")
	req, err := sling.New().Base(s.Connect.GetProxyBaseURL()).Post(CertificatesPath).BodyJSON(body).Request()
	resp, err := s.Connect.GetHttpClient().Do(req)
	if err != nil {
		lc.Error("failed to upload cert to proxy server with error %s", err.Error())
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusCreated || resp.StatusCode == http.StatusConflict {
		lc.Info("successful to add certificate to the reverse proxy")
		return nil
	}

	return fmt.Errorf("failed to add certificate with errorcode %d", resp.StatusCode)
}

func (s *Service) getCertPair() (*CertPair, error) {
	c := &Certs{s.Connect, s.CertCfg}
	return c.getCertPair()
}

func (s *Service) initKongService(service *KongService) error {
	req, err := sling.New().Base(s.Connect.GetProxyBaseURL()).Post(ServicesPath).BodyForm(service).Request()
	resp, err := s.Connect.GetHttpClient().Do(req)

	if err != nil {
		e := fmt.Sprintf("failed to set up proxy service for %s", service.Name)
		return errors.New(e)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusCreated {
		lc.Info(fmt.Sprintf("successful to set up proxy service for %s", service.Name))

		//serviceObj := KongServiceResponse{}
		//err = json.NewDecoder(resp.Body).Decode(&serviceObj)
		return nil
	} else if resp.StatusCode == http.StatusConflict {
		e := fmt.Sprintf("proxy service for %s has been set up", service.Name)
		lc.Info(e)
		return nil
	} else {
		return fmt.Errorf("failed to set up proxy service for %s with errorcode %d", service.Name, resp.StatusCode)
	}
}

func (s *Service) initKongRoutes(r *KongRoute, name string) error {
	routesubpath := "services/" + name + "/routes"
	req, err := sling.New().Base(s.Connect.GetProxyBaseURL()).Post(routesubpath).BodyJSON(r).Request()
	resp, err := s.Connect.GetHttpClient().Do(req)
	if err != nil {
		e := fmt.Sprintf("failed to set up routes for %s with error %s", name, err.Error())
		lc.Info(e)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusCreated || resp.StatusCode == http.StatusConflict {
		lc.Info(fmt.Sprintf("successful to set up route for %s", name))
		return nil
	}

	e := fmt.Sprintf("failed to set up route for %s with error %s", name, resp.Status)
	lc.Error(e)
	return errors.New(e)
}

func (s *Service) initACL(name string, whitelist string) error {
	aclParams := &KongACLPlugin{
		Name:      name,
		WhiteList: whitelist,
	}
	req, err := sling.New().Base(s.Connect.GetProxyBaseURL()).Post(PluginsPath).BodyForm(aclParams).Request()
	resp, err := s.Connect.GetHttpClient().Do(req)
	if err != nil {
		e := fmt.Sprintf("failed to set up acl")
		lc.Error(e)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusCreated || resp.StatusCode == http.StatusConflict {
		lc.Info("successful to set up acl")
		return nil
	}

	e := fmt.Sprintf("failed to set up acl with errorcode %d", resp.StatusCode)
	lc.Error(e)
	return errors.New(e)
}

func (s *Service) initAuthmethod(name string, ttl int) error {
	lc.Info(fmt.Sprintf("selected auth method as %s.", name))
	if name == "jwt" {
		return s.initJWTAuth()
	} else if name == "oauth2" {
		return s.initOAuth2(ttl)
	}
	return errors.New("unsupported authetication method")
}

func (s *Service) initJWTAuth() error {
	jwtParams := &KongJWTPlugin{
		Name: "jwt",
	}

	req, err := sling.New().Base(s.Connect.GetProxyBaseURL()).Post(PluginsPath).BodyForm(jwtParams).Request()
	resp, err := s.Connect.GetHttpClient().Do(req)
	if err != nil {
		e := fmt.Sprintf("failed to set up jwt authentication")
		lc.Error(e)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusCreated || resp.StatusCode == http.StatusConflict {
		lc.Info("successful to set up jwt authentication")
		return nil
	}

	e := fmt.Sprintf("failed to set up jwt authentication with errorcode %d", resp.StatusCode)
	lc.Error(e)
	return errors.New(e)
}

func (s *Service) initOAuth2(ttl int) error {
	oauth2Params := &KongOAuth2Plugin{
		Name:                    "oauth2",
		Scope:                   OAuth2Scopes,
		MandatoryScope:          "true",
		EnableClientCredentials: "true",
		EnableGlobalCredentials: "true",
		TokenTTL:                ttl,
	}

	req, err := sling.New().Base(s.Connect.GetProxyBaseURL()).Post(PluginsPath).BodyForm(oauth2Params).Request()
	resp, err := s.Connect.GetHttpClient().Do(req)
	if err != nil {
		e := fmt.Sprintf("failed to set up oauth2 authentication with error %s", err.Error())
		lc.Error(e)
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusCreated || resp.StatusCode == http.StatusConflict {
		lc.Info("successful to set up oauth2 authentication")
		return nil
	}

	e := fmt.Sprintf("failed to set up oauth2 authentication with errorcode %d", resp.StatusCode)
	lc.Error(e)
	return errors.New(e)
}

func (s *Service) getSvcIDs(path string) (DataCollect, error) {
	collection := DataCollect{}

	req, err := sling.New().Get(s.Connect.GetProxyBaseURL()).Path(path).Request()
	resp, err := s.Connect.GetHttpClient().Do(req)
	if err != nil {
		e := fmt.Sprintf("failed to get list of %s with error %s", path, err.Error())
		return collection, errors.New(e)
	}
	defer resp.Body.Close()

	json.NewDecoder(resp.Body).Decode(&collection)
	return collection, nil
}
