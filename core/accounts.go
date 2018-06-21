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
	"regexp"

	"github.com/dghubble/sling"
	jwt "github.com/dgrijalva/jwt-go"
)

func isAllowedChars(user string) bool {
	return regexp.MustCompile(`^[a-zA-Z]+$`).MatchString(user)
}

func createConsumer(user string, url string, service string, c *http.Client) error {

	if !isAllowedChars(user) {
		s := "Only a-z and A-Z char are allowed for user name."
		return errors.New(s)
	}
	userNameParams := &KongUser{UserName: user}
	req, err := sling.New().Base(url).Post(ConsumersPath).BodyForm(userNameParams).Request()
	resp, err := c.Do(req)
	if err != nil {
		s := fmt.Sprintf("Failed to create consumer %s for %s service with error %s.", user, service, err.Error())
		return errors.New(s)
	}
	if resp.StatusCode == 200 || resp.StatusCode == 201 || resp.StatusCode == 409 {
		lc.Info(fmt.Sprintf("Successful to create consumer %s for %s service.", user, service))
		return nil
	}
	s := fmt.Sprintf("Failed to create consumer %s for %s service.", user, service)
	lc.Error(s)
	return errors.New(s)
}

func deleteConsumer(user string, url string, c *http.Client) {
	deleteResource(user, url, ConsumersPath, ConsumersPath, c)
}

func createJWTForConsumer(user string, url string, name string, c *http.Client) (string, error) {
	jwtCred := JWTCred{}
	s := sling.New().Set("Content-Type", "application/x-www-form-urlencoded")
	req, err := s.New().Get(url).Post(fmt.Sprintf("consumers/%s/jwt", user)).Request()
	resp, err := c.Do(req)
	if err != nil {
		errString := fmt.Sprintf("Failed to create jwt token for consumer %s with error %s.", user, err.Error())
		return "", errors.New(errString)
	}
	if resp.StatusCode == 200 || resp.StatusCode == 201 || resp.StatusCode == 409 {
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
