FROM powerman/alpine-runit-volume

ENV VOLUME_DIR=/data \
    SYSLOG_DIR=/data/syslog
WORKDIR /app

RUN set -x -e -o pipefail; \
    build_pkgs="curl"; \
    apk add --no-cache $build_pkgs; \
    ### consul
    CONSUL="1.0.2"; \
    curl -s -o /tmp/consul.zip https://releases.hashicorp.com/consul/${CONSUL}/consul_${CONSUL}_linux_amd64.zip; \
    unzip -d /usr/local/bin/ /tmp/consul.zip; \
    ### Cleanup
    apk del $build_pkgs; \
    rm -rf /tmp/*

EXPOSE 8080

COPY service /app/service
RUN set -x -e -o pipefail; \
    ln -nsf /app/service/* /etc/service/

COPY bin /usr/local/bin/
