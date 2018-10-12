# EdgeX Foundry Security API-Gateway Service Implemented with Go
[![license](https://img.shields.io/badge/license-Apache%20v2.0-blue.svg)](LICENSE)

Go implementation of EdgeX security API-Gateway service.

## Features
- Reverse proxy for the existing edgex microservices
- Account creation with optional either OAuth2 or JWT authentication for existing services
- Account creation with arbitrary ACL gourp list

## How to Start the service
The service can be started with 2 different methods listed below:
- start with docker-compose file like other normal EdgeX services do.
- build from source code & start with single docker container.


## Method 1. Run the Security Service with Docker-compose file. Make sure other EdgeX services start as usual (especially volume), then
```
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
make build
make docker
```

### Run the security service
```
docker run -v vault-config:/vault/config --network=edgex-network edgexfoundry/docker-edgex-proxy-go
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
docker run --network=edgex-network edgex/proxy --useradd=<account> --group=<groupname>
docker run --network=edgex-network edgex/proxy --userdel=<account>
```

### Access existing microservice APIs like ping service of command microservice
```
use JWT as query string 
curl -k -v -H "host: edgex" https://kong-container:8443/command/api/v1/ping?jwt= <JWT from account creation>
or use JWT in HEADER
curl -k -v -H "host: edgex" https://kong-container:8443/command/api/v1/ping -H "Authorization: Bearer <JWT from account creation>"
```


 
## Community
- Chat: https://chat.edgexfoundry.org/home
- Mainling lists: https://lists.edgexfoundry.org/mailman/listinfo

## License
[Apache-2.0](LICENSE)