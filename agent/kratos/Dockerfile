FROM gcr.io/up9-docker-hub/mizu-kratos-base/simple-password-policy:latest

USER root

RUN apk add sqlite

RUN mkdir -p /etc/config/kratos

COPY ./kratos.yml /etc/config/kratos/kratos.yml
COPY ./identity.schema.json /etc/config/kratos/identity.schema.json
COPY ./start.sh /opt/start.sh
RUN chmod +x /opt/start.sh

ENTRYPOINT ["/opt/start.sh"]
