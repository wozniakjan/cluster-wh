#!/bin/bash

set -euo pipefail

# Configure or pass your own overrides
export NAMESPACE=${NAMESPACE:-jw-api-test}
export API_INGRESS_URL=${API_INGRESS_URL:-dmz-apiserver.dev.kubermatic.io}
export CLUSTER_DOMAIN=${CLUSTER_DOMAIN:-cluster.local}

# sync certs for api extension/aggregation communication
if ! kubectl get configmap extension-apiserver-authentication -n $NAMESPACE &> /dev/null; then
kubectl get configmap extension-apiserver-authentication -n kube-system -o yaml | \
    yq '.metadata.namespace = "'$NAMESPACE'"' | \
    kubectl apply -f -
fi

# generate ca bundle and self-signed certs for dmz-api server
mkdir -p k8s-api

# csr for kube-apiserver
cat << EOF > k8s-api/req.cnf
[req]
req_extensions = v3_req
distinguished_name = req_distinguished_name

[req_distinguished_name]
organizationName = jw 
commonName       = jw 

[ v3_req ]
basicConstraints = CA:FALSE
keyUsage = nonRepudiation, digitalSignature, keyEncipherment
subjectAltName = @alt_names

[alt_names]
DNS.1 = dmz-apiserver
DNS.2 = dmz-apiserver.$NAMESPACE.svc
DNS.3 = dmz-apiserver.$NAMESPACE.svc.$CLUSTER_DOMAIN
DNS.4 = $API_INGRESS_URL
IP.1 = 127.0.0.1
EOF

# client csr for cluster-admin
cat << EOF > k8s-api/client-req.cnf
[req]
req_extensions = v3_req
distinguished_name = req_distinguished_name

[req_distinguished_name]
commonName = cluster-admin

[ v3_req ]
basicConstraints = CA:FALSE
keyUsage = nonRepudiation, digitalSignature, keyEncipherment
EOF

openssl genrsa -out k8s-api/ca.key 2048
openssl req -x509 -new -nodes -key k8s-api/ca.key -days 1460 -out k8s-api/ca.crt -subj "/CN=$API_INGRESS_URL"

openssl genrsa -out k8s-api/apiserver.key 2048
openssl req -new -key k8s-api/apiserver.key -out k8s-api/csr.pem -subj "/CN=$API_INGRESS_URL" -config k8s-api/req.cnf
openssl x509 -req -in k8s-api/csr.pem -CA k8s-api/ca.crt -CAkey k8s-api/ca.key -CAcreateserial -out k8s-api/apiserver.crt -days 1460 -extensions v3_req -extfile k8s-api/req.cnf

openssl genrsa -out k8s-api/client-ca.key 2048
openssl req -new -key k8s-api/client-ca.key -out k8s-api/csr.pem -subj "/CN=cluster-admin" -config k8s-api/client-req.cnf
openssl x509 -req -in k8s-api/csr.pem -CA k8s-api/ca.crt -CAkey k8s-api/ca.key -CAcreateserial -out k8s-api/client-ca.crt -days 1460 -extensions v3_req -extfile k8s-api/client-req.cnf

# create secret with ca bundle for the extension api server
if kubectl get secret k8s-api-certs -n $NAMESPACE &> /dev/null; then
    kubectl delete secret k8s-api-certs -n $NAMESPACE
fi
kubectl create secret generic k8s-api-certs -n $NAMESPACE \
    --from-file=k8s-api/apiserver.crt --from-file=k8s-api/apiserver.key \
    --from-file=k8s-api/client-ca.crt --from-file=k8s-api/client-ca.key

# generate ca bundle and self-signed certs for etcd 
mkdir -p etcd

cat << EOF > etcd/req.cnf
[req]
req_extensions = v3_req
distinguished_name = req_distinguished_name

[req_distinguished_name]
organizationName = jw 
commonName       = jw 

[ v3_req ]
basicConstraints = CA:FALSE
keyUsage = nonRepudiation, digitalSignature, keyEncipherment
subjectAltName = @alt_names

[alt_names]
DNS.1 = dmz-apiserver-etcd
DNS.2 = dmz-apiserver-etcd.$NAMESPACE.svc
DNS.3 = dmz-apiserver-etcd.$NAMESPACE.svc.$CLUSTER_DOMAIN
DNS.4 = dmz-apiserver-etcd-0
DNS.5 = dmz-apiserver-etcd-1
DNS.6 = dmz-apiserver-etcd-2
IP.1 = 127.0.0.1
EOF

ETCD_CN=dmz-apiserver-etcd.$NAMESPACE.svc.$CLUSTER_DOMAIN
openssl genrsa -out etcd/ca.key 2048
openssl req -x509 -new -nodes -key etcd/ca.key -days 1460 -out etcd/ca.crt -subj "/CN=$ETCD_CN"

openssl genrsa -out etcd/server.key 2048
openssl req -new -key etcd/server.key -out etcd/csr.pem -subj "/CN=$ETCD_CN" -config etcd/req.cnf
openssl x509 -req -in etcd/csr.pem -CA etcd/ca.crt -CAkey etcd/ca.key -CAcreateserial -out etcd/server.crt -days 1460 -extensions v3_req -extfile etcd/req.cnf

openssl genrsa -out etcd/peer.key 2048
openssl req -new -key etcd/peer.key -out etcd/csr.pem -subj "/CN=$ETCD_CN" -config etcd/req.cnf
openssl x509 -req -in etcd/csr.pem -CA etcd/ca.crt -CAkey etcd/ca.key -CAcreateserial -out etcd/peer.crt -days 1460 -extensions v3_req -extfile etcd/req.cnf

openssl genrsa -out etcd/healthcheck-client.key 2048
openssl req -new -key etcd/healthcheck-client.key -out etcd/csr.pem -subj "/CN=$ETCD_CN" -config etcd/req.cnf
openssl x509 -req -in etcd/csr.pem -CA etcd/ca.crt -CAkey etcd/ca.key -CAcreateserial -out etcd/healthcheck-client.crt -days 1460 -extensions v3_req -extfile etcd/req.cnf

openssl genrsa -out etcd/apiserver-etcd-client.key 2048
openssl req -new -key etcd/apiserver-etcd-client.key -out etcd/csr.pem -subj "/CN=$ETCD_CN" -config etcd/req.cnf
openssl x509 -req -in etcd/csr.pem -CA etcd/ca.crt -CAkey etcd/ca.key -CAcreateserial -out etcd/apiserver-etcd-client.crt -days 1460 -extensions v3_req -extfile etcd/req.cnf

# create secret with ca bundle for the etcd
if kubectl get secret k8s-api-etcd-certs -n $NAMESPACE &> /dev/null; then
    kubectl delete secret k8s-api-etcd-certs -n $NAMESPACE
fi
kubectl create secret generic k8s-api-etcd-certs -n $NAMESPACE \
    --from-file=etcd/ca.crt \
    --from-file=etcd/ca.key \
    --from-file=etcd/server.crt \
    --from-file=etcd/server.key \
    --from-file=etcd/peer.crt \
    --from-file=etcd/peer.key \
    --from-file=etcd/healthcheck-client.crt \
    --from-file=etcd/healthcheck-client.key \
    --from-file=etcd/apiserver-etcd-client.crt \
    --from-file=etcd/apiserver-etcd-client.key 

# create service for dmz-apiserver-etcd
kubectl apply -f resources/dmz-apiserver-etcd.svc.yaml -n $NAMESPACE

# create pod with dmz-apiserver-etcd
kubectl apply -f resources/dmz-apiserver-etcd.pod.yaml -n $NAMESPACE

# create configmap with user-creds
kubectl apply -f resources/user-creds.yaml -n $NAMESPACE

# create service for dmz-apiserver
kubectl apply -f resources/dmz-apiserver.svc.yaml -n $NAMESPACE

# create apiservice apiregistration
caBundle=$(cat k8s-api/apiserver.crt | base64 --wrap=0)
for f in resources/register.apiservice*.yaml; do
    kubectl apply -f <(cat "$f" | \
        yq '.spec.service.namespace = "'$NAMESPACE'"' | \
        yq '.spec.caBundle = "'$caBundle'"' )
done

# create pod with dmz-apiserver
kubectl apply -f resources/dmz-apiserver.pod.yaml -n $NAMESPACE
