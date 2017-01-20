FROM ubuntu:14.04
MAINTAINER NP

ADD active-proxy /acproxy/active-proxy
ADD bootstrap.sh /acproxy/bootstrap.sh
ADD config.yaml /acproxy/config.yaml
RUN chmod u+x /acproxy/active-proxy && \
		chmod u+x /acproxy/bootstrap.sh
ENV PATH $PATH:/acproxy

WORKDIR /acproxy
