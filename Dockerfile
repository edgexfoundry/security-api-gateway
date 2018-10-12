FROM alpine:3.7

RUN mkdir -p /edgex/res

WORKDIR /edgex

COPY core/res/configuration-docker.toml res/configuration.toml

ADD core/edgexproxy .

ENTRYPOINT ["./edgexproxy"]

CMD  ["--init=true"]