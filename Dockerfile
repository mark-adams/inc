FROM golang:1.7-onbuild

EXPOSE 8080

ENV STATSD_HOST "localhost:8125"
