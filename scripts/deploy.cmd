eval $(minikube docker-env)
docker build -t go-chat .
kubectl delete -f kubernetes
kubectl apply -f kubernetes