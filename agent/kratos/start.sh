#!/bin/sh

kratos migrate sql sqlite:///app/data/kratos.sqlite?_fk=true --yes # this initializes the db
kratos serve -c /etc/config/kratos/kratos.yml --dev --watch-courier # start kratos
