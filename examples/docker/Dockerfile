FROM ubuntu:14.04
MAINTAINER NP

ADD active-proxy /acproxy/active-proxy
ADD config.yaml /acproxy/config.yaml
RUN chmod u+x /acproxy/active-proxy

ENTRYPOINT ["/acproxy/active-proxy", "--alsologtostderr=true"]
