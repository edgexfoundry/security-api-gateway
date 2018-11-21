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
 * @version: 0.1.1
 *******************************************************************************/
package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/edgexfoundry/edgex-go/pkg/clients/logging"
)

var lc = CreateLogging()

func CreateLogging() logger.LoggingClient {
	return logger.NewClient(SecurityService, false, fmt.Sprintf("%s-%s.log", SecurityService, time.Now().Format("2006-01-02")), logger.InfoLog)
}

func main() {

	if len(os.Args) < 2 {
		HelpCallback()
	}
	useConsul := flag.Bool("consul", false, "retrieve configuration from consul server")
	insecureSkipVerify := flag.Bool("insureskipverify", true, "skip server side SSL verification, mainly for self-signed cert")
	initNeeded := flag.Bool("init", false, "run init procedure for security service.")
	resetNeeded := flag.Bool("reset", false, "reset reverse proxy by removing all services/routes/consumers")
	userTobeCreated := flag.String("useradd", "", "user that needs to be added to consume the edgex services")
	userofGroup := flag.String("group", "user", "group that the user belongs to. By default it is in user group")
	userTobeDeleted := flag.String("userdel", "", "user that needs to be deleted from the edgex services")
	configFileLocation := flag.String("configfile", "res/configuration.toml", "configuration file")

	flag.Usage = HelpCallback
	flag.Parse()

	config, err := LoadTomlConfig(*configFileLocation)
	if err != nil {
		lc.Error("Failed to retrieve config data from local file. Please make sure res/configuration.toml file exists with correct formats.")
		return
	}

	if *useConsul {
		lc.Info("Retrieving config data from Consul")
		//err := metadata.ConnectToConsul(*config)
		//if err != nil {
		//	lc.Error("Failed to retrieve config from Consul")
		//}
		//lc.Info("Retrieving config data from Consul")
	}

	proxyBaseURL := fmt.Sprintf("http://%s:%s/", config.KongURL.Server, config.KongURL.AdminPort)
	secretServiceBaseURL := fmt.Sprintf("https://%s:%s/", config.SecretService.Server, config.SecretService.Port)

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: *insecureSkipVerify},
	}
	client := &http.Client{Timeout: 10 * time.Second, Transport: tr}

	checkProxyStatus(proxyBaseURL, client)
	//checkSecretServiceStatus(secretServiceBaseURL+config.SecretService.HealthcheckPath, client)

	if *initNeeded == true && *resetNeeded == true {
		lc.Error("can't run initialization and reset at the same time for security service.")
		return
	}

	if *initNeeded == true {
		initSecurityServices(config, proxyBaseURL, secretServiceBaseURL, client)
	}

	if *resetNeeded == true {
		resetProxy(proxyBaseURL, client)
	}

	if *userTobeCreated != "" && *userofGroup != "" {
		err := createConsumer(*userTobeCreated, *userofGroup, proxyBaseURL, EdgeXService, client)
		if err != nil {
			lc.Error(err.Error())
			return
		}

		t, err := createTokenForConsumer(config, *userTobeCreated, proxyBaseURL, EdgeXService, client)
		if err != nil {
			lc.Error(fmt.Sprintf("Failed to create jwt token for edgex service due to error %s.", err.Error()))
		} else {
			fmt.Println(fmt.Sprintf("The JWT for user %s is: %s. Please keep the jwt for accessing edgex services.", *userTobeCreated, t))
		}
	}

	if *userTobeDeleted != "" {
		deleteConsumer(*userTobeDeleted, proxyBaseURL, client)
	}
}
