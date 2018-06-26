# EdgeX Foundry Security Services Implemented with Go
[![license](https://img.shields.io/badge/license-Apache%20v2.0-blue.svg)](LICENSE)

Go implementation of EdgeX security services.
The security service will need KONG ( https://konghq.com/) and Vault (https://www.vaultproject.io/) to be started first. Make sure they are running and the edgexsecurity will check their status.


## Features
- Reverse proxy for the existing edgex microservices
- Account creation & JWT authentication for existing services


## Run the Security Service with Docker

The repo includes a Dockerfile to dockerize the security service. A docker-compose-proxy.yml file is provided under Docker folder as well to make sure the security service is working with other existing services. They need to be ran in order from the top to bottom.

### Build an image of the security service
```
go get github.com/edgexfoundry/edgexsecurity
cd edgexsecurity
.\build.bat # on Windows
./build.sh # on Linux/Mac
Docker build -t edgex/proxy .
```

### Run the security service
```
docker run -v vault-config:/vault/config --network=edgex-network edgex/proxy
```

Notice here vault-config is the name of the volume that keeps the root_token for Vault service, which can be checked with 
``` 
docker volume ls
docker volume inspect <volume_name>
```
And edgex-network is the name of the private network that edgex containers create, which can be check with 
```
docker network ls
docker network inspect <network_name>
```

### Other options for security service, E,g, reset the proxy to initial status, create account, delete account
```
docker run --network=edgex-network edgex/proxy -h
docker run --network=edgex-network edgex/proxy --reset=true
docker run --network=edgex-network edgex/proxy --useradd=<account>
docker run --network=edgex-network edgex/proxy --userdel=<account>
```

### Access existing microservice APIs like ping service of command microservice
```
use JWT as query string 
curl -k -v -H "host: edgex" https://kong-container:8443/command/api/v1/ping?jwt= <JWT from account creation>
or use JWT in HEADER
curl -k -v -H "host: edgex" https://kong-container:8443/command/api/v1/ping -H "Authorization: Bearer <JWT from account creation>"

```


## Build, Install and Deploy with source files

1. Make sure KONG is up and running. To start KONG with docker-compose file under Docker/ folder, run commands below
```
docker-compse -f docker-compose-proxy.yml up -d kong-db
docker-compse -f docker-compose-proxy.yml up -d kong-migrations
docker-compse -f docker-compose-proxy.yml up -d kong
```
2. Make sure Vault is up and running. It can be done in a similar way with docker-compose file 
```
docker-compse -f docker-compose-proxy.yml up -d vault
docker-compse -f docker-compose-proxy.yml up -d vault-init-unseal
```
3. Build edgexsecurity service with the command below
```
go get github.com/edgexfoundry/edgexsecurity
cd edgexsecurity/core
go build -o edgexsecurity
```
4. Create res folder in the same folder as executable and copy configuration.toml
5. create a vault seed file by running command below, where path-to-res is the path to res folder that is created in step 4
```
Docker cp <vault-container-id>:/vault/config/resp-init.json <path-to-res>/
```
6. Modify the parameters in the configuration.toml file. Make sure the information for the KONG service, Vault service and Edgex microservices are correct
7. Run the edgexsecurity service with the command below
```
./edgexsecurity init=true
```
8. Use command below for more options
```
./edgexsecurity -h
```


### Usage

```
# initialize reverse proxy 
./edgexsecurity init=true

# reset reverse proxy
./edgexsecurity reset=true

# create account and return JWT for the account 
./edgexsecurity userddd=guest

# delete account
./edgexsecurity userdel=guest
```

### Access exisitng microservices APIs like ping service of command microservice
```
use JWT as query string 
curl -k -v -H "host: edgex" https://kong-ip:8443/command/api/v1/ping?jwt= <JWT from account creation>
or use JWT in HEADER
curl -k -v -H "host: edgex" https://kong-ip:8443/command/api/v1/ping -H "Authorization: Bearer <JWT from account creation>"
``` 



 
## Community
- Chat: https://chat.edgexfoundry.org/home
- Mainling lists: https://lists.edgexfoundry.org/mailman/listinfo

## License
[Apache-2.0](LICENSE)