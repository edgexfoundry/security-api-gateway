# EdgeX Foundry Security API-Gateway Service Implemented with Go
[![license](https://img.shields.io/badge/license-Apache%20v2.0-blue.svg)](LICENSE)

Go implementation of EdgeX security API-Gateway service.

## Features
- Reverse proxy for the existing edgex microservices
- Account creation & JWT authentication for existing services

## How to Start the service
The service can be started with 3 different methods listed below:
- start with docker-compose file like other normal EdgeX services do.
- start with single docker container.
- start from command line by building from the source file.

## Method 1. Run the Security Service with Docker-compose file. Make sure other EdgeX services start as usual (especially volume), then
```
cd Docker
docker-compose up -d vault
docker-compose up -d vault-worker
docker-compose up -d kong-db
docker-compose up -d kong-migrations
docker-compose up -d kong
docker-compose up -d edgex-proxy
```

## Method 2. Build Docker image and Run Api-gateway Security Service
The repo includes a Dockerfile to dockerize the security service. Run the docker-componse commands same as in method one except the last step, then

### Build an image of the security service
```
go get github.com/edgexfoundry/security-api-gateway
cd security-api-gateway
.\build.bat # on Windows
./build.sh # on Linux/Mac
Docker build -t edgex/proxy .
```

### Run the security service
```
docker run -v vault-file:/vault/file --network=edgex-network edgex/proxy
```

Notice here vault-file is the name of the volume that keeps the root_token for Vault service, which can be checked with 
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

## Method 3. Build, Install and Deploy with source files

1. Run the docker-componse commands same as in method 1 except the last step.
2. Build edgexsecurity service with the command below
```
go get github.com/edgexfoundry/security-api-gateway
cd security-api-gateway/core
go build -o edgexsecurity
```
3. Create res folder in the same folder as executable and copy configuration.toml
4. create a vault seed file by running command below, where path-to-res is the path to res folder that is created in step 4
```
Docker cp <vault-container-id>:/vault/file/resp-init.json <path-to-res>/
```
5. Modify the parameters in the configuration.toml file. Make sure the information for the KONG service, Vault service and Edgex microservices are correct
6. Run the edgexsecurity service with the command below
```
./edgexsecurity init=true
```
7. Use command below for more options
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