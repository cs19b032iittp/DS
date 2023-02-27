package main

import (
	"fmt"
	"net"
	"net/rpc"
	"sync"
)

type MetaTracker struct {
	FileHash string
	FileName string
	Chunks   [][]string
}

type TrackerService struct {
	Trackers map[string]MetaTracker
}

type AnnounceChunksParams struct {
	FileHash string
	Chunks   []int64
	ID       string
	NickName string
}

type AnnounceChunksResponse struct{}

type AnnounceFileParams struct {
	FileHash    string
	FileName    string
	NumOfChunks int64
	ID          string
	NickName    string
}

type AnnounceFileResponse struct {
	Success bool
	Message string
}

type GetPeersParams struct {
	FileHash   string
	NickName   string
	ChunkIndex int64
}

type GetPeersResponse struct {
	Peers []string
}

func (ts *TrackerService) GetPeers(params *GetPeersParams, resp *GetPeersResponse) error {
	Log("GetPeers", params.NickName)
	fmt.Println(ts.Trackers[params.FileHash].Chunks[params.ChunkIndex])
	resp.Peers = ts.Trackers[params.FileHash].Chunks[params.ChunkIndex]
	return nil
}

func (ts *TrackerService) AnnounceChunks(params *AnnounceChunksParams, resp *AnnounceChunksResponse) error {
	Log("AnnounceChunks", params.NickName)
	m, ok := ts.Trackers[params.FileHash]
	var mutex sync.Mutex
	if ok {
		mutex.Lock()
		for _, val := range params.Chunks {
			m.Chunks[val] = append(m.Chunks[val], params.ID)
		}
		mutex.Unlock()
	}
	return nil
}

func (ts *TrackerService) AnnounceFile(params *AnnounceFileParams, resp *AnnounceFileResponse) error {
	Log("AnnounceFile", params.NickName)
	resp.Success = false
	_, ok := ts.Trackers[params.FileHash]
	if !ok {

		Chunks := make([][]string, params.NumOfChunks)

		var wg sync.WaitGroup
		wg.Add(int(params.NumOfChunks))
		for i := int64(0); i < params.NumOfChunks; i++ {
			go func(chunkIndex int64) {
				defer wg.Done()
				Chunks[chunkIndex] = append(Chunks[chunkIndex], params.ID)
			}(i)
		}
		wg.Wait()

		ts.Trackers[params.FileHash] = MetaTracker{FileHash: params.FileHash, FileName: params.FileName, Chunks: Chunks}

		resp.Success = true
		resp.Message = "File was open to share"
	} else {
		resp.Message = "File already exists"
	}

	return nil
}

func CreateTracker(cfg *config) {

	ipAddr := cfg.listenHost + ":" + cfg.listenPort
	ln, err := net.Listen("tcp", ipAddr)
	if err != nil {
		fmt.Println(err)
		return
	}

	ts := &TrackerService{Trackers: make(map[string]MetaTracker)}
	rpc.RegisterName("Tracker", ts)

	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				fmt.Println(err)
				continue
			}
			go rpc.ServeConn(conn)
		}
	}()
}
