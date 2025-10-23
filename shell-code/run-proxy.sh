#! /usr/bin/bash

LCG_PROVIDER=proxy LCG_HOST=https://direct-dev.ru \
LCG_MODEL=GigaChat-2 \
LCG_JWT_TOKEN=$(go-ansible-vault --key $(cat ~/.config/gak) -i ~/.config/jwt.direct-dev.ru get -m JWT_TOKEN -q) \
go run . $1 $2 $3 $4 $5 $6 $7 $8 $9

