FROM golang:latest
MAINTAINER MP

RUN mkdir /go/scripts

ADD active-proxy /go/scripts/active-proxy
ADD bootstrap.sh /go/scripts/bootstrap.sh
ADD config.yaml /go/scripts/config.yaml
RUN chmod u+x /go/scripts/active-proxy
RUN chmod u+x /go/scripts/bootstrap.sh
ENV PATH $PATH:/go/scripts

