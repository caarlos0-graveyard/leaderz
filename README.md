# leaderz

An example project for http://carlosbecker.dev/posts/k8s-leader-election

## Applying

```shell script
kubectl apply -f kube/
```

## Remove

```shell script
kubectl delete -f kube/
kubectl delete lease my-lock
```

## Building

```shell script
GOOS=linux GOARCH=amd64 go build -o leaderz .
docker build -t caarlos0/leaderz .
docker push caarlos0/leaderz
```
