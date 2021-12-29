FROM oryd/kratos:v0.8.2

USER root

RUN mkdir -p /etc/config/kratos

COPY ./kratos.yml /etc/config/kratos/kratos.yml
COPY ./identity.schema.json /etc/config/kratos/identity.schema.json

ENTRYPOINT ["kratos"]
CMD ["serve"]