#! /usr/bin/bash

VERSION=$1
if [ -z "$VERSION" ]; then
    VERSION=latest
fi

docker pull kuznetcovay/lcg:"${VERSION}"

docker run -p 8080:8080 \
  -e LCG_PROVIDER=proxy \
  -e LCG_HOST=https://direct-dev.ru \
  -e LCG_MODEL=GigaChat-2-Max \
  -e LCG_JWT_TOKEN="$(go-ansible-vault --key "$(cat ~/.config/gak)" \
     -i ~/.config/jwt.direct-dev.ru get -m JWT_TOKEN -q)" \
  kuznetcovay/lcg:"${VERSION}"