#!/usr/bin/env bash

# https://dev.to/ashokan/sealed-secrets-the-secret-sauce-for-managing-secrets-2hg6
#  head -c 64 /dev/urandom | base64 -w 0
export KUBECONFIG=/home/su/.kube/config_hlab

kubectl create secret generic lcg-secrets -n lcg \
 --from-literal=LCG_SERVER_PASSWORDL= \
 --from-literal=LCG_CSRF_SECRET=\
 --from-literal=LCG_JWT_SECRET=\
 --from-literal=LCG_JWT_TOKEN=\
 --dry-run=client -o yaml | tee secret-cfg.yaml

kubeseal --controller-name=sealed-secrets-controller --controller-namespace=kube-system -o yaml <secret-cfg.yaml | tee sealed-cfg.yaml

rm -f secret-cfg.yaml

kubectl apply -f sealed-cfg.yaml
cp sealed-cfg.yaml ../kustomize/secret.yaml

kubectl get secret lcg-secrets -n lcg -o json | jq ".data | map_values(@base64d)"