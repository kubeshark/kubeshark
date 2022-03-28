#!/bin/sh

keto migrate up -c /etc/config/keto/keto.yml --yes # this initializes the db
keto serve -c /etc/config/keto/keto.yml # start keto
