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
	"errors"
	"io/ioutil"
	"encoding/json"
	"crypto/tls"
	"time"
	"github.com/dghubble/sling"
	jwt "github.com/dgrijalva/jwt-go"
	logger "github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	model "github.com/edgexfoundry/go-mod-core-contracts/models"
)

var lc = CreateLogging()

func CreateLogging() logger.LoggingClient {
	return logger.NewClient(SecurityService, false, fmt.Sprintf("%s-%s.log", SecurityService, time.Now().Format("2006-01-02")), model.InfoLog)
}


 type Consumer struct {
	 Name string	 
	 BaseURL string
	 Client *http.Client
 }

 type acctParams struct {
	Group string `url:"group"`
}

 func (c *Consumer) Delete() error {
	 return deleteResource(c.Name, c.BaseURL, ConsumersPath, ConsumersPath, c.Client)
 }

 func (c *Consumer) Create(service string) error {	 
	path := fmt.Sprintf("%s%s", ConsumersPath, c.Name)
	req, err := sling.New().Base(c.BaseURL).Put(path).Request()
	resp, err := c.Client.Do(req)
	if err != nil {
		s := fmt.Sprintf("Failed to create consumer %s for %s service with error %s.", c.Name, service, err.Error())
		return errors.New(s)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusCreated || resp.StatusCode == http.StatusConflict {
		lc.Info(fmt.Sprintf("Successful to create consumer %s for %s service.", c.Name, service))
		return nil 
	}
	s := fmt.Sprintf("Failed to create consumer %s for %s service.", c.Name, service)	
	return errors.New(s)
 }

 func (c *Consumer) AssociateWithGroup(g string) error {
		
	acc := acctParams{g}
	path := fmt.Sprintf("%s%s/acls", ConsumersPath, c.Name)
	req, err := sling.New().Base(c.BaseURL).Post(path).BodyForm(acc).Request()
	resp, err := c.Client.Do(req)
	if err != nil {
		s := fmt.Sprintf("Failed to associate consumer %s for with group %s with error %s.", c.Name, g, err.Error())
		return errors.New(s)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusCreated || resp.StatusCode == http.StatusConflict {
		lc.Info(fmt.Sprintf("Successful to associate consumer %s with group %s.", c.Name, g))
		return nil
	}
	b, _ := ioutil.ReadAll(resp.Body)
	s := fmt.Sprintf("Failed to associate consumer %s with group %s with error %s,%s.", c.Name, g, resp.Status, string(b))
	lc.Error(s)
	return errors.New(s)
 }

 func (c *Consumer) CreateToken(config *tomlConfig) (string, error) {
	if config.KongAuth.Name == "jwt" {
		lc.Info("autheticate the user with jwt authentication.")
		return c.createJWTToken()	
	} else if config.KongAuth.Name == "oauth2" {
		lc.Info("authenticate the user with oauth2 authentication.")
		return c.createOAuth2Token(config)		
	}
	return "", errors.New("unknown authentication method provided")
 }

 func (c *Consumer) createJWTToken() (string, error) {
	jwtCred := JWTCred{}
	s := sling.New().Set("Content-Type", "application/x-www-form-urlencoded")
	req, err := s.New().Get(c.BaseURL).Post(fmt.Sprintf("consumers/%s/jwt", c.Name)).Request()
	resp, err := c.Client.Do(req)
	if err != nil {
		errString := fmt.Sprintf("Failed to create jwt token for consumer %s with error %s.", c.Name, err.Error())
		return "", errors.New(errString)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusCreated || resp.StatusCode == http.StatusConflict {		
		json.NewDecoder(resp.Body).Decode(&jwtCred)
		lc.Info(fmt.Sprintf("successful on retrieving JWT credential for consumer %s.", c.Name))

		// Create the Claims
		claims := KongJWTClaims{
			jwtCred.Key,
			c.Name,
			jwt.StandardClaims{
				Issuer: EdgeXService,
			},
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		return token.SignedString([]byte(jwtCred.Secret))
	}
	errString := fmt.Sprintf("Failed to create JWT for consumer %s with errorCode %d.", c.Name, resp.StatusCode)
	return "", errors.New(errString)
 }

//curl -X POST "http://localhost:8001/consumers/user123/oauth2" -d "name=edgex.com" --data "client_id=user123" -d "client_secret=user123"  -d "redirect_uri=http://www.edgex.com/"
//curl -k -v https://localhost:8443/{service}/oauth2/token -d "client_id=user123" -d "grant_type=client_credentials" -d "client_secret=user123" -d "scope=email"
 func (c *Consumer) createOAuth2Token(config *tomlConfig) (string, error) {
	
	url := fmt.Sprintf("http://%s:%s/", config.KongURL.Server, config.KongURL.AdminPort)

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	client := &http.Client{Timeout: 10 * time.Second, Transport: tr}

	token := KongOauth2Token{}
	ko := &KongConsumerOauth2{
		Name:         EdgeXService,
		ClientID:     c.Name,
		ClientSecret: c.Name,
		RedirectURIS: "http://" + EdgeXService,
	}

	req, err := sling.New().Base(url).Post(fmt.Sprintf("consumers/%s/oauth2", c.Name)).BodyForm(ko).Request()
	resp, err := client.Do(req)
	if err != nil {
		lc.Error(fmt.Sprintf("Failed to enable oauth2 authentication for consumer %s with error %s.", c.Name, err.Error()))
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusCreated || resp.StatusCode == http.StatusConflict {
		lc.Info(fmt.Sprintf("successful on enabling oauth2 for consumer %s.", c.Name))

		// obtain token
		tokenreq := &KongOuath2TokenRequest{
			ClientId:     c.Name,
			ClientSecret: c.Name,
			GrantType:    OAuth2GrantType,
			Scope:        OAuth2Scopes,
		}

		url = fmt.Sprintf("https://%s:%s/", config.KongURL.Server, config.KongURL.ApplicationPortSSL)
		path := fmt.Sprintf("%s/oauth2/token", config.KongAuth.Resource)
		lc.Info(fmt.Sprintf("Creating token on the endpoint of %s", path))
		req, err := sling.New().Base(url).Post(path).BodyForm(tokenreq).Request()
		resp, err := client.Do(req)
		if err != nil {
			lc.Error(fmt.Sprintf("Failed to create oauth2 token for client_id %s with error %s.", c.Name, err.Error()))
			return "", err
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusCreated {
			defer resp.Body.Close()
			json.NewDecoder(resp.Body).Decode(&token)
			lc.Info(fmt.Sprintf("successful on retrieving bearer credential for consumer %s.", c.Name))
			return token.AccessToken, nil
		}
		b, _ := ioutil.ReadAll(resp.Body)
		errString := fmt.Sprintf("Failed to create bearer token for oauth authentication at endpoint oauth2/token with error %s,%s.", resp.Status, string(b))
		return "", errors.New(errString)
	}

	errString := fmt.Sprintf("Failed to enable oauth2 for consumer %s with error code %d.", c.Name, resp.StatusCode)
	return "", errors.New(errString)
 }