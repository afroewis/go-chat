# go-chat

Proof of concept: Horizontally scalable, websocket-based chat application. Uses redis pub/sub to communicate between server nodes.


Build docker image: `docker build -t go-chat .`

Show logs: `kubectl logs -f -l app=chat --all-containers`

Run app in browser: `minikube service chat-service`