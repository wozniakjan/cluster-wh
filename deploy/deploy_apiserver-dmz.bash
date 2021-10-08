# create apiservice apiregistration
kubectl apply -f resources/dmz-apiserver.svc.external-name.yaml

kubeclt apply -f resources/register.apiservice.v1.kubermatic.k8s.io.yaml
