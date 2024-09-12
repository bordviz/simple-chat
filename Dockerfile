FROM golang:1.22.1

RUN mkdir /auth
WORKDIR /auth

COPY . .
RUN chmod a+x docker/*.sh