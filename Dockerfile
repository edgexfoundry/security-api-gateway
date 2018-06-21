FROM alpine:3.7

RUN mkdir -p /edgex/res

WORKDIR /edgex

COPY Docker/res/configuration.toml res/

ADD core/edgexproxy .

ENTRYPOINT ["./edgexproxy"]

CMD  ["--init=true"]