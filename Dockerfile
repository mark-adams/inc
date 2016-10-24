FROM golang:1.7-onbuild

EXPOSE 80

ENV PORT 80
ENV MARTINI_ENV "production"
