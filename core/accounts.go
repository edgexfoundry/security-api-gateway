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
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"time"

	"github.com/dghubble/sling"
	jwt "github.com/dgrijalva/jwt-go"
)

func isAllowedChars(user string) bool {
	return regexp.MustCompile(`^[a-zA-Z]+$`).MatchString(user)
}

func createConsumer(user string, group string, url string, service string, c *http.Client) error {

	if !isAllowedChars(user) {
		s := "Only a-z and A-Z char are allowed for user name."
		return errors.New(s)
	}

	path := fmt.Sprintf("%s%s", ConsumersPath, user)
	req, err := sling.New().Base(url).Put(path).Request()
	resp, err := c.Do(req)
	if err != nil {
		s := fmt.Sprintf("Failed to create consumer %s for %s service with error %s.", user, service, err.Error())
		return errors.New(s)
	}
	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusCreated || resp.StatusCode == http.StatusConflict {
		lc.Info(fmt.Sprintf("Successful to create consumer %s for %s service.", user, service))

		err = associateConsumerWithGroup(user, group, url, c)
		if err == nil {
			return nil
		}
	}
	s := fmt.Sprintf("Failed to create consumer %s for %s service.", user, service)
	lc.Error(s)
	return errors.New(s)
}

func associateConsumerWithGroup(user string, g string, url string, c *http.Client) error {

	type acctParams struct {
		Group string `url:"group"`
	}

	acc := acctParams{g}
	path := fmt.Sprintf("%s%s/acls", ConsumersPath, user)
	req, err := sling.New().Base(url).Post(path).BodyForm(acc).Request()
	resp, err := c.Do(req)
	if err != nil {
		s := fmt.Sprintf("Failed to associate consumer %s for with group %s with error %s.", user, g, err.Error())
		return errors.New(s)
	}
	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusCreated || resp.StatusCode == http.StatusConflict {
		lc.Info(fmt.Sprintf("Successful to associate consumer %s with group %s.", user, g))
		return nil
	}
	b, _ := ioutil.ReadAll(resp.Body)
	s := fmt.Sprintf("Failed to associate consumer %s with group %s with error %s,%s.", user, g, resp.Status, string(b))
	lc.Error(s)
	return errors.New(s)
}

func deleteConsumer(user string, url string, c *http.Client) {
	deleteResource(user, url, ConsumersPath, ConsumersPath, c)
}

func createTokenForConsumer(config *tomlConfig, user string, url string, name string, c *http.Client) (string, error) {
	if config.KongAuth.Name == "jwt" {
		lc.Info("autheticate the user with jwt authentication.")
		t, err := createTokenWithJWT(user, url, name, c)
		return t, err
	} else if config.KongAuth.Name == "oauth2" {
		lc.Info("authenticate the user with oauth2 authentication.")
		t, err := createTokenWithOauth2(config, user, url)
		return t, err
	}
	return "", errors.New("unknown authentication method provided")
}

func createTokenWithJWT(user string, url string, name string, c *http.Client) (string, error) {
	jwtCred := JWTCred{}
	s := sling.New().Set("Content-Type", "application/x-www-form-urlencoded")
	req, err := s.New().Get(url).Post(fmt.Sprintf("consumers/%s/jwt", user)).Request()
	resp, err := c.Do(req)
	if err != nil {
		errString := fmt.Sprintf("Failed to create jwt token for consumer %s with error %s.", user, err.Error())
		return "", errors.New(errString)
	}
	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusCreated || resp.StatusCode == http.StatusConflict {
		defer resp.Body.Close()
		json.NewDecoder(resp.Body).Decode(&jwtCred)
		lc.Info(fmt.Sprintf("successful on retrieving JWT credential for consumer %s.", user))

		// Create the Claims
		claims := KongJWTClaims{
			jwtCred.Key,
			user,
			jwt.StandardClaims{
				Issuer: EdgeXService,
			},
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		return token.SignedString([]byte(jwtCred.Secret))
	}
	errString := fmt.Sprintf("Failed to create JWT for consumer %s with errorCode %d.", user, resp.StatusCode)
	return "", errors.New(errString)
}

func createTokenWithOauth2(config *tomlConfig, user string, url string) (string, error) {
	//curl -X POST "http://localhost:8001/consumers/user123/oauth2" -d "name=edgex.com" --data "client_id=user123" -d "client_secret=user123"  -d "redirect_uri=http://www.edgex.com/"
	//curl -k -v https://localhost:8443/{service}/oauth2/token -d "client_id=user123" -d "grant_type=client_credentials" -d "client_secret=user123" -d "scope=email"

	url = fmt.Sprintf("http://%s:%s/", config.KongURL.Server, config.KongURL.AdminPort)

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	client := &http.Client{Timeout: 10 * time.Second, Transport: tr}

	token := KongOauth2Token{}
	ko := &KongConsumerOauth2{
		Name:         EdgeXService,
		ClientID:     user,
		ClientSecret: user,
		RedirectURIS: "http://" + EdgeXService,
	}

	req, err := sling.New().Base(url).Post(fmt.Sprintf("consumers/%s/oauth2", user)).BodyForm(ko).Request()
	resp, err := client.Do(req)
	if err != nil {
		lc.Error(fmt.Sprintf("Failed to enable oauth2 authentication for consumer %s with error %s.", user, err.Error()))
		return "", err
	}
	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusCreated || resp.StatusCode == http.StatusConflict {
		defer resp.Body.Close()
		lc.Info(fmt.Sprintf("successful on enabling oauth2 for consumer %s.", user))

		// obtain token
		tokenreq := &KongOuath2TokenRequest{
			ClientId:     user,
			ClientSecret: user,
			GrantType:    OAuth2GrantType,
			Scope:        OAuth2Scopes,
		}

		url = fmt.Sprintf("https://%s:%s/", config.KongURL.Server, config.KongURL.ApplicationPortSSL)
		path := fmt.Sprintf("%s/oauth2/token", config.KongAuth.Resource)
		lc.Info(fmt.Sprintf("Creating token on the endpoint of %s", path))
		req, err := sling.New().Base(url).Post(path).BodyForm(tokenreq).Request()
		resp, err := client.Do(req)
		if err != nil {
			lc.Error(fmt.Sprintf("Failed to create oauth2 token for client_id %s with error %s.", user, err.Error()))
			return "", err
		}
		if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusCreated {
			defer resp.Body.Close()
			json.NewDecoder(resp.Body).Decode(&token)
			lc.Info(fmt.Sprintf("successful on retrieving bearer credential for consumer %s.", user))
			return token.AccessToken, nil
		}
		b, _ := ioutil.ReadAll(resp.Body)
		errString := fmt.Sprintf("Failed to create bearer token for oauth authentication at endpoint oauth2/token with error %s,%s.", resp.Status, string(b))
		return "", errors.New(errString)
	}

	errString := fmt.Sprintf("Failed to enable oauth2 for consumer %s with error code %d.", user, resp.StatusCode)
	return "", errors.New(errString)
}
