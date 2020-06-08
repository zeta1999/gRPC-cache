#!/bin/bash


# check namespace if present
{
  NAMESPACE_EXIST=$(kubectl get ns grpc-cache -o jsonpath='{.metadata.name}')
  KEY_EXIST=$(kubectl get secrets grpc-cache.example.com -o jsonpath='{.metadata.name}' -n grpc-cache)
  NGINX_CONTROLLER_EXIST=$(kubectl get svc nginx-ingress-grpc-cache-controller -o jsonpath='{.metadata.name}' -n grpc-cache)
} &> /dev/null

if [ "$NAMESPACE_EXIST" != "grpc-cache" ]; then
  kubectl create ns grpc-cache
fi

# set current context namespace
kubectl config set-context $(kubectl config current-context) --namespace=grpc-cache

if [ "$NGINX_CONTROLLER_EXIST" != "nginx-ingress-grpc-cache-controller" ]; then
  helm install stable/nginx-ingress --name-template=nginx-ingress-grpc-cache
fi

if [ "$KEY_EXIST" != "grpc-cache.example.com" ]; then
  openssl req -x509 -sha256 -nodes -days 365 -newkey rsa:2048 -keyout tls.key -out tls.crt -subj "/CN=grpc-cache.example.com/O=grpc-cache.example.com"

  # create secret
  kubectl create secret tls grpc-cache.example.com --key tls.key --cert tls.crt -n grpc-cache
  rm tls.key tls.crt
fi

# create resources for k8s
helm install grpc-cache ./helm-charts

HOST_INGRESS=$(kubectl get ingress grpc-cache -o jsonpath='{.status.loadBalancer.ingress[0].ip}')

# setup host
while true; do
  if [ "$HOST_INGRESS" != "" ]; then
    echo "Ingress address configured to: $HOST_INGRESS"
    break
  fi
  echo "Waiting for ingress to configure, sleeping for 20 seconds"
  HOST_INGRESS=$(kubectl get ingress grpc-cache -o jsonpath='{.status.loadBalancer.ingress[0].ip}')
  sleep 20
done

ADDRESS_MAP_EXIST=$(grep "$HOST_INGRESS grpc-cache.example.com" /etc/hosts)

if [ "$ADDRESS_MAP_EXIST" == "" ]; then
  echo "Adding ingress host mapping to /etc/hosts"
  sudo -- sh -c -e "echo '$HOST_INGRESS grpc-cache.example.com' >> /etc/hosts";
fi

echo "Started all resources successfully, you can use the application now."
