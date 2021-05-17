# go-chat

Proof Of Concept: Horizontally scalable, websocket-based chat application. Uses redis pub/sub to communicate between server nodes.

![Demo](assets/demo.png)

## Running without Docker/Kubernetes/Redis
1. Set `Debug` in `main.go:11` to `true`
2. `go get`
3. Run the code: `go run src/*.go`

## Helpful commands
Build docker image: `docker build -t go-chat .`

Show logs: `kubectl logs -f -l app=chat --all-containers`

Run app in browser: `minikube service chat-service`
