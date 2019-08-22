# EdgeX Foundry Security API-Gateway Service Implemented with Go
[![license](https://img.shields.io/badge/license-Apache%20v2.0-blue.svg)](LICENSE)

Go implementation of EdgeX security API-Gateway service.

## Features
- Reverse proxy for the existing EdgeX microservices
- Account creation with optional either OAuth2 or JWT authentication for existing services
- Account creation with arbitrary ACL group list

## How to Start the service
The service can be started using the two methods listed below:
- Using a docker-compose file, pull and bring up containers for the EdgeX services as well as the API-Gateway service from Nexus
- Build a docker image from source code and edit your docker-compose file to use this image

## Step 1: Bring up a reverse proxy docker container

### Method 1. Pull a Docker image from Nexus using a docker-compose file.
Using an EdgeX docker-compose file that includes the API-Gateway service, either one from [this repo](deployments) or 
from the EdgeX Foundry [developer-scripts](https://github.com/edgexfoundry/developer-scripts/tree/master/releases) repo,
bring up the EdgeX services following the instructions [here](https://docs.edgexfoundry.org/Ch-GettingStartedUsers.html).

Then, to bring up the security service, use the following commands in order:
 
```
docker-compose up -d vault
docker-compose up -d vault-worker
docker-compose up -d kong-db
docker-compose up -d kong-migrations
docker-compose up -d kong
docker-compose up -d edgex-proxy
```

Then skip ahead to [Step Two](#step-two).

### Method 2. Build Docker image from source
To build a Docker image from source, clone this repo and use the `make docker` command.  This uses this repo's
Dockerfile to create a Docker image.

```
git clone git@github.com:edgexfoundry/security-api-gateway.git
cd security-api-gateway.git
make docker
```

At this point you will have a Docker image on your system for `edgexfoundry/docker-edgex-proxy-go`.
Using the command `docker images`, determine what the image's tag is.  It likely ends in `-dev`.

Download an EdgeX docker-compose file that includes the API-Gateway service, either one from [this repo](deployments) or 
from the EdgeX Foundry [developer-scripts](https://github.com/edgexfoundry/developer-scripts/tree/master/releases) repo.
Find the `edgex-proxy` service and edit the image's tag to match your newly built developer image.

Follow the instructions [here](https://docs.edgexfoundry.org/Ch-GettingStartedUsers.html) to bring up 
the EdgeX services using the docker-compose file, and then use the following commands to bring up the 
security service:

```
docker-compose up -d vault
docker-compose up -d vault-worker
docker-compose up -d kong-db
docker-compose up -d kong-migrations
docker-compose up -d kong
docker-compose up -d edgex-proxy
```

It's important that **the same docker-compose file is used for both the core EdgeX services
and the security service**, as they have common dependencies.  To run the security service
against a development build of EdgeX, change the desired EdgeX service tags in the docker-compose
file to their respective `-dev` tags.  These tags will likely be different from the security `-dev` version.

## Step Two

At this point we need to get an access token to use to authenticate requests to the reverse proxy.

From the directory with your docker-compose file, use the following commands to reset the proxy to its initial status
and create a user and token.

```
docker-compose run edgex-proxy --reset=true --init=false
docker-compose run edgex-proxy --useradd=[username] --group=admin
```

The `useradd` command will print an access token.  You can use this access token to make requests to the reverse proxy, for example:

```
curl -k -v https://127.0.0.1:8443/command/api/v1/ping -H "Authorization: Bearer [[access token goes here]]"
``` 

will make a GET request to the ping endpoint of the EdgeX core-command service.


### Other useful options for security service, E,g, reset the proxy to initial status, create account, delete account
```
docker-compose run edgex-proxy --help
docker-compose run edgex-proxy --userdel=<account>
```
 
## Community
- Chat: https://edgexfoundry.slack.com/
- Mainling lists: https://lists.edgexfoundry.org/mailman/listinfo

## License
[Apache-2.0](LICENSE)