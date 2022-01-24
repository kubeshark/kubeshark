FROM oryd/keto:v0.7.0-alpha.1-sqlite

USER root

RUN apk add sqlite

RUN mkdir -p /etc/config/keto

COPY ./keto.yml /etc/config/keto/keto.yml
COPY ./start.sh /opt/start.sh
RUN chmod +x /opt/start.sh

ENTRYPOINT ["/opt/start.sh"]
