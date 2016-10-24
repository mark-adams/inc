FROM golang:1.7-onbuild

EXPOSE 8080

ENV PORT 8080
ENV MARTINI_ENV "production"
ENV STATSD_HOST "localhost:8125"
