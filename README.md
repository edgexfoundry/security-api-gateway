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

### Run the security service with docker-compose
```
docker-compose run edgexfoundry/docker-edgex-proxy-go
```

### Other options for security service, E,g, reset the proxy to initial status, create account, delete account
```
docker-compose run edgex-proxy -h
docker-compose run edgex-proxy --reset=true
docker-compose run edgex-proxy --useradd=<account> --group=<groupname>
docker-compose run edgex-proxy --userdel=<account>
```

### Access existing microservice APIs like ping service of command microservice
```
curl -k -v https://{api-gateway-ip}:8443/command/api/v1/ping -H "Authorization: Bearer <access token from account creation>"
```


 
## Community
- Chat: https://edgexfoundry.slack.com/
- Mainling lists: https://lists.edgexfoundry.org/mailman/listinfo

## License
[Apache-2.0](LICENSE)