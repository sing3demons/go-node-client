# go-node-client
golang nodejs fiber&amp;gin&amp;express


```start
kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/controller-v1.8.2/deploy/static/provider/aws/deploy.yaml
kubectl apply -f k8s-ns.yml
kubectl apply -f server/kubernetes.yml
kubectl apply -f client/kubernetes.yml
kubectl apply -f k8s-ingress.yml
```

```delete
kubectl delete -f k8s-ingress.yml
kubectl delete -f client/kubernetes.yml
kubectl delete -f server/kubernetes.yml
kubectl delete -f k8s-ns.yml
```