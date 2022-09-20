package http_server

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/hashicorp/raft"
)

type HttpServer struct {
	Raft *raft.Raft
	Db   *sync.Map
}

// JoinHandler 处理 raft 节点的加入
func (hs HttpServer) JoinHandler(w http.ResponseWriter, r *http.Request) {
	followerId := r.URL.Query().Get("followerId")
	followerAddr := r.URL.Query().Get("followerAddr")
	if hs.Raft.State() != raft.Leader {
		json.NewEncoder(w).Encode(struct {
			Error string `json:"error"`
		}{
			Error: "Not leader",
		})
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	err := hs.Raft.AddVoter(raft.ServerID(followerId), raft.ServerAddress(followerAddr), 0, 0).Error()
	if err != nil {
		log.Printf("Failed to add follower: %v", err)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (hs HttpServer) SetHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	bs, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("[setHandler] ReadAll err: %v", err)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	future := hs.Raft.Apply(bs, 500*time.Microsecond)
	if err := future.Error(); err != nil {
		log.Printf("[setHandler] apply err: %v", err)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	rRes := future.Response()
	if rRes != nil {
		log.Printf("[setHandler] future.Response res: %v", rRes)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (hs HttpServer) GetHandler(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("key")
	value, _ := hs.Db.Load(key)
	if value == nil {
		value = ""
	}
	type GetResp struct {
		Data string `json:"data"`
	}
	resp := GetResp{}
	err := json.NewEncoder(w).Encode(&resp)
	if err != nil {
		log.Printf("[getHandler] Encode: %v", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}

type Config struct {
	Id       string
	HttpPort string
	RaftPort string
}

func GetConfig() Config {
	cfg := Config{}
	for i, arg := range os.Args[1:] {
		if arg == "--node-id" {
			cfg.Id = os.Args[i+2]
			i++
			continue
		}
		if arg == "--http-port" {
			cfg.HttpPort = os.Args[i+2]
			i++
			continue
		}
		if arg == "--raft-port" {
			cfg.RaftPort = os.Args[i+2]
			i++
			continue
		}
	}

	if cfg.Id == "" {
		log.Fatalf("Missing parameter: --node-id")
	}
	if cfg.HttpPort == "" {
		log.Fatalf("Missing parameter: --http-port")
	}
	if cfg.RaftPort == "" {
		log.Fatalf("Missing parameter: --raft-port")
	}

	return cfg
}
