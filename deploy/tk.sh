#!/bin/bash -e

ROOT="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

which jq > /dev/null || (echo "jq is required to run this script. Please install jq and try again." && exit -1)
which tk > /dev/null || (echo "tk is required to run this script. Please install tk and try again." && exit -1)
which kubectl > /dev/null || (echo "kubectl is required to run this script. Please install kubectl and try again." && exit -1)

# retrieving expected server api
expected=`cat ./environments/staging/spec.json | jq -r .spec.apiServer`

# get current server api
actual=`kubectl cluster-info   |awk '/Kubernetes control plane/ {print $7}' | sed -e 's/\x1b\[[0-9;]*m//g'`

if [ "$expected" != "$actual" ]
then
  echo "Error: Cluster '$actual' does not match the expected apiServer '$expected' was found. Please check your KUBECONFIG"
  echo ""
  exit -1
fi


depl=$(kubectl get deployment -n staging  -o json | jq -c '.items[].spec.template | {name: .metadata.labels.name, image: .spec.containers[0].image} | {(.name): .image}')
sts=$(kubectl get statefulset -n staging  -o json | jq -c '.items[].spec.template | {name: .metadata.labels.name, image: .spec.containers[0].image} | {(.name): .image}')
full="${depl} ${sts}"
echo $full | jq -s add > ./environments/staging/images.json
trap "rm `pwd`/environments/staging/images.json" EXIT

tk $@
