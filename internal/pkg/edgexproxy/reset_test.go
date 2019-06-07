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
	"testing"
	"fmt"
	//"github.com/stretchr/testify/assert"
 )

 func TestResetProxy( t *testing.T){
	 c, mux, server := testServer()
	 defer server.Close()

	 paths := []string{RoutesPath, ServicesPath, ConsumersPath, PluginsPath, CertificatesPath}
	 basepath := "http://localhost/"		
	 for _, p := range paths {
		 mux.HandleFunc(basepath+p, func(w http.ResponseWriter, r *http.Request) {
			 assertMethod(t, "GET", r)
			 w.Header().Set("Content-Type", "text/plain")
			 fmt.Fprintf(w, "OK")
		 })
	}
	
	ResetProxy(basepath, c)	 
 }
