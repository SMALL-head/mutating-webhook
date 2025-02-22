#!/bin/bash

set -o errexit
set -o nounset
set -o pipefail


#export CA_BUNDLE=$(kubectl config view --raw --flatten -o json | jq -r '.clusters[] | select(.name == "'$(kubectl config current-context)'") | .cluster."certificate-authority-data"')
export CA_BUNDLE=$(kubectl config view --raw --flatten -o json | jq -r '.clusters[] | .cluster."certificate-authority-data"')

if command -v envsubst >/dev/null 2>&1; then
    envsubst < validatingwebhook.yaml > temp.yaml && mv temp.yaml validatingwebhook.yaml
    envsubst < mutatingwebhook.yaml > temp.yaml && mv temp.yaml mutatingwebhook.yaml
else
    sed -e "s|\${CA_BUNDLE}|${CA_BUNDLE}|g"
fi