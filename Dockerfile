FROM golang:1.12-alpine AS builder

ENV GO111MODULE=on
WORKDIR /go/src/github.com/edgexfoundry/security-api-gateway

RUN sed -e 's/dl-cdn[.]alpinelinux.org/nl.alpinelinux.org/g' -i~ /etc/apk/repositories

RUN apk update && apk add make git

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

RUN make build

FROM scratch

WORKDIR /

COPY --from=builder /go/src/github.com/edgexfoundry/security-api-gateway/cmd/edgexproxy/res/configuration-docker.toml /res/configuration.toml

COPY --from=builder /go/src/github.com/edgexfoundry/security-api-gateway/cmd/edgexproxy/edgexproxy .

ENTRYPOINT ["./edgexproxy","--init=true"]