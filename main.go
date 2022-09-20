package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"path"
	"raftGo1/http_server"
	"raftGo1/kv"
	"sync"
	"time"

	"github.com/hashicorp/raft"
	raftboltdb "github.com/hashicorp/raft-boltdb"
)

func main() {
	cfg := http_server.GetConfig()
	db := &sync.Map{}
	kf := &kv.KvFsm{
		db,
	}

	dataDir := "raft-go-data"
	err := os.MkdirAll(dataDir, os.ModePerm)
	if err != nil {
		log.Fatalf("[main] MkdirAll err: %v", err)
	}

	raftAddr := fmt.Sprintf("localhost:%s", cfg.RaftPort)
	fmt.Printf("[main] raft start at: %s", raftAddr)
	ra, err := setupRaft(path.Join(dataDir, "raft-"+cfg.Id), cfg.Id, raftAddr, kf)
	if err != nil {
		log.Fatalf("[main] setupRaft err: %v", err)
	}

	hs := http_server.HttpServer{Raft: ra, Db: db}
	http.HandleFunc("/set", hs.SetHandler)
	http.HandleFunc("/get", hs.GetHandler)
	http.HandleFunc("/join", hs.JoinHandler)
	httpAddr := fmt.Sprintf(":%s", cfg.HttpPort)
	fmt.Printf("[main] http server start at: %v", httpAddr)
	http.ListenAndServe(httpAddr, nil)
}

func setupRaft(dir, nodeId, raftAddress string, kf *kv.KvFsm) (*raft.Raft, error) {
	store, err := raftboltdb.NewBoltStore(path.Join(dir, "bolt"))
	if err != nil {
		return nil, fmt.Errorf("[setupRaft] NewBoltStore err: %v", err)
	}
	snapshots, err := raft.NewFileSnapshotStore(path.Join(dir, "snapshot"), 2, os.Stderr)
	if err != nil {
		return nil, fmt.Errorf("[setupRaft] NewFileSnapshotStore err: %v", err)
	}
	tcpAddr, err := net.ResolveTCPAddr("tcp", raftAddress)
	if err != nil {
		return nil, fmt.Errorf("[setupRaft] ResolveTCPAddr err: %v", err)
	}
	transport, err := raft.NewTCPTransport(raftAddress, tcpAddr, 10, time.Second*10, os.Stderr)
	if err != nil {
		return nil, fmt.Errorf("[setupRaft] ResolveTCPAddr err: %v", err)
	}
	raftCfg := raft.DefaultConfig()
	raftCfg.LocalID = raft.ServerID(nodeId)
	r, err := raft.NewRaft(raftCfg, kf, store, store, snapshots, transport)
	if err != nil {
		return nil, fmt.Errorf("[setupRaft] NewRaft err: %v", err)
	}

	r.BootstrapCluster(raft.Configuration{
		Servers: []raft.Server{
			{
				ID:      raft.ServerID(nodeId),
				Address: transport.LocalAddr(),
			},
		},
	})

	return r, nil
}
