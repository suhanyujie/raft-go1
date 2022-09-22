# raft-go
学习 Raft 在项目中的使用

## test
* `go build -o kv-server.exe main.go`
* `./kv-server.exe --node-id node1 --raft-port 2221 --http-port 8221`
* `./kv-server.exe --node-id node2 --raft-port 2222 --http-port 8222`
* `curl 'localhost:8221/join?followerAddr=localhost:2222&followerId=node2'`
* `curl -X POST 'localhost:8221/set' -d '{"key": "key1001", "value": "value1001"}' -H 'content-type: application/json'`
* `curl 'localhost:8221/get?key=key1001'`

### 测试数据的清理
* `rm -rf raft-go-data/raft-node*/*`

## ref
* [A minimal distributed key-value database with Hashicorp's Raft library](https://notes.eatonphil.com/minimal-key-value-store-with-hashicorp-raft.html)
* https://raft.github.io/