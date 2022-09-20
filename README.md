# raft-go
学习 Raft 在项目中的使用

## test
* `go build -o server.exe main.go`
* `./server.exe --node-id node1 --raft-port 2222 --http-port 8222`
* `./server.exe --node-id node2 --raft-port 2223 --http-port 8223`
* `curl -X POST 'localhost:8222/set' -d '{"key": "key1001", "value": "value1001"}' -H 'content-type: application/json'`
* `curl 'localhost:8222/get?key=key1001'`

## ref
* [A minimal distributed key-value database with Hashicorp's Raft library](https://notes.eatonphil.com/minimal-key-value-store-with-hashicorp-raft.html)
* https://raft.github.io/