#! /usr/bin/bash

LCG_PROVIDER=ollama LCG_HOST=http://192.168.87.108:11434/ \
LCG_MODEL=hf.co/yandex/YandexGPT-5-Lite-8B-instruct-GGUF:Q4_K_M \
go run . $1 $2 $3 $4 $5 $6 $7 $8 $9

