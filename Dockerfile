FROM golang:1.19

RUN groupadd -r inc && useradd --no-log-init -r -g inc inc

COPY . /src
WORKDIR /src

RUN CGO_ENABLED=0 go build -tags netgo -a -o /bin/inc

FROM scratch

EXPOSE 8080
ENV STATSD_HOST "localhost:8125"

COPY --from=0 /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=0 /etc/passwd /etc/group /etc/
USER inc

COPY --from=0 /bin/inc /bin/inc


CMD ["/bin/inc"]
