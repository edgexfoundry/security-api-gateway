FROM alpine:3.7

RUN mkdir -p /edgex/res

WORKDIR /edgex

COPY cmd/edgexproxy/res/configuration-docker.toml res/configuration.toml

ADD cmd/edgexproxy/edgexproxy .

ENTRYPOINT ["./edgexproxy"]

CMD  ["--init=true"]