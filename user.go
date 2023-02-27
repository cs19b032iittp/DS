package main

import (
	"crypto/subtle"
	"fmt"
	"math"
	"math/rand"
	"net"
	"net/rpc"
	"os"
	"sync"
)

func CreateUser(cfg *config) {

	ipAddr := cfg.listenHost + ":" + cfg.listenPort
	ln, err := net.Listen("tcp", ipAddr)
	if err != nil {
		fmt.Println(err)
		return
	}

	fs := &FileService{Files: make(map[string]Meta)}
	rpc.RegisterName("File", fs)

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

	for {
		var option string
		fmt.Println("\x1b[32mChoose an option,")
		fmt.Println("1. Upload a file")
		fmt.Println("2. Download a file")
		fmt.Println("3. Shutdown\x1b[0m")
		fmt.Print("> ")
		fmt.Scanln(&option)

		switch option {
		case "1":
			var filePath string
			fmt.Println("\x1b[32mEnter the path of the file u want to seed,\x1b[0m")
			fmt.Print("> ")
			fmt.Scanln(&filePath)

			m := Meta{}
			m.URL = "127.0.0.1:4000"
			GenerateMeta(filePath, &m)

			client, err := rpc.Dial("tcp", m.URL)
			if err != nil {
				fmt.Println(err)
				return
			}

			params := &AnnounceFileParams{
				FileHash:    m.Hash,
				ID:          ipAddr,
				FileName:    m.Name,
				NumOfChunks: m.NumOfChunks,
				NickName:    cfg.nick,
			}
			resp := &AnnounceFileResponse{}

			err = client.Call("Tracker.AnnounceFile", params, resp)
			if err != nil {
				fmt.Println(err)
				return
			}

			var wg sync.WaitGroup
			wg.Add(int(m.NumOfChunks))
			for i := int64(0); i < m.NumOfChunks; i++ {
				go func(chunkIndex int64) {
					defer wg.Done()
					m.Chunks[chunkIndex].Downloaded = true
				}(i)
			}
			wg.Wait()

			m.FilePath = filePath
			fs.Files[m.Hash] = m

			if resp.Success {
				fmt.Println(resp.Message)
			} else {
				fmt.Println(resp.Message)

			}
		case "2":
			var filePath string
			fmt.Println("\x1b[32mEnter the path of the meta file,\x1b[0m")
			fmt.Print("> ")
			fmt.Scanln(&filePath)

			m := &Meta{}
			DecodeMeta(filePath, m)
			m.FilePath = cfg.nick + "/" + m.Name

			fs.Files[m.Hash] = *m

			outputFile, err := os.Create(m.FilePath)
			if err != nil {
				panic(err)
			}

			fileSize := m.FileSize
			err = outputFile.Truncate(fileSize)
			if err != nil {
				panic(err)
			}

			var mutex = &sync.Mutex{}
			// var wg sync.WaitGroup
			// wg.Add(int(m.NumOfChunks))
			for i := int64(0); i < m.NumOfChunks; i++ {
				func(idx int64) {
					// defer wg.Done()
					flag := true
					for flag {
						resp := &GetPeersResponse{Peers: make([]string, 0)}
						params := &GetPeersParams{
							FileHash:   m.Hash,
							NickName:   cfg.nick,
							ChunkIndex: idx,
						}

						client, err := rpc.Dial("tcp", m.URL)
						if err != nil {
							fmt.Println(err)
							return
						}

						err = client.Call("Tracker.GetPeers", params, &resp)
						if err != nil {
							panic(err)
						}

						ChunkOffset := idx * m.ChunkSize
						ChunkSize := m.ChunkSize
						NumOfPieces := m.ChunkSize / pieceSize
						if idx == m.NumOfChunks-1 {
							ChunkSize = m.FileSize - ((m.NumOfChunks - 1) * m.ChunkSize)
							NumOfPieces = int64(math.Ceil(float64(float64(ChunkSize) / pieceSize)))
						}

						chunk := make([]byte, ChunkSize)

						for j := int64(0); j < NumOfPieces; j++ {

							fmt.Println("Started downloading piece, ", j, "for chunk: ", idx)
							k := rand.Intn(len(resp.Peers))
							PeerID := resp.Peers[k]

							PieceOffset := j * pieceSize
							PieceSize := pieceSize

							if j == NumOfPieces-1 {
								PieceSize = int(ChunkSize) - ((int(NumOfPieces) - 1) * PieceSize)
							}
							params := &GetPieceParams{
								FileHash:   m.Hash,
								ChunkIndex: idx,
								Size:       int64(PieceSize),
								Offset:     ChunkOffset + PieceOffset,
								NickName:   cfg.nick,
							}
							resp1 := &GetPieceResponse{}

							client, err := rpc.Dial("tcp", PeerID)
							if err != nil {
								fmt.Println(err)
								return
							}
							err = client.Call("File.GetPiece", params, resp1)

							fmt.Println(resp1, PieceSize, j, PieceOffset)

							for x := int64(0); x < int64(PieceSize); x++ {
								chunk[PieceOffset+x] = resp1.Data[x]
							}
						}

						hash1 := []byte(m.Chunks[int(idx)].Hash)
						hash2 := []byte(hashContent(chunk))

						if subtle.ConstantTimeCompare(hash1[:], hash2[:]) == 1 {

							m.Chunks[idx].Downloaded = true

							mutex.Lock()
							_, err = outputFile.Seek(ChunkOffset, 0)
							if err != nil {
								panic(err)
							}

							// Write the data to the file
							_, err = outputFile.Write(chunk)
							if err != nil {
								panic(err)
							}
							flag = false
							mutex.Unlock()
						}

					}
				}(i)
			}
			// wg.Wait()
			outputFile.Close()

			// client, err := rpc.Dial("tcp", m.URL)
			// if err != nil {
			// 	fmt.Println(err)
			// 	return
			// }

			// err = client.Call("Tracker.GetPeers", params, &resp)
			// if err != nil {
			// 	panic(err)
			// }

			// if len(resp.Peers) == 0 {
			// 	fmt.Println("Unable to connect to peers, Please try again after sometime!")
			// 	continue
			// }

			// chunkMap := make(map[int][]string)

			// for i := int64(0); i < m.NumOfChunks; i++ {
			// 	peerlist := make([]string, 0)
			// 	chunkMap[int(i)] = peerlist
			// }

			// for PeerID, ChunksList := range resp.Peers {
			// 	for idx, val := range ChunksList {
			// 		if val == 1 {
			// 			chunkMap[idx] = append(chunkMap[idx], PeerID)
			// 		}
			// 	}
			// }

			// outputFile, err := os.Create(m.FilePath)
			// if err != nil {
			// 	panic(err)
			// }

			// // Set the file size
			// fileSize := m.FileSize
			// err = outputFile.Truncate(fileSize)
			// if err != nil {
			// 	panic(err)
			// }

			// var start = 0

			// go func() {

			// 	for {

			// 		switch start {
			// 		case 0:
			// 			continue
			// 		case 1:
			// 			Chunks := make([]int64, m.NumOfChunks)
			// 			for i := int64(0); i < m.NumOfChunks; i++ {
			// 				if m.Chunks[i].Downloaded {
			// 					Chunks[i] = 1
			// 				} else {
			// 					Chunks[i] = 0
			// 				}
			// 			}

			// 			params := &AnnounceChunksParams{
			// 				FileHash: m.Hash,
			// 				ID:       ipAddr,
			// 				Chunks:   Chunks,
			// 				NickName: cfg.nick,
			// 			}
			// 			resp := &AnnounceChunksResponse{}

			// 			client, err := rpc.Dial("tcp", m.URL)
			// 			if err != nil {
			// 				fmt.Println(err)
			// 				return
			// 			}
			// 			err = client.Call("Tracker.AnnounceChunks", params, resp)
			// 			if err != nil {
			// 				panic(err)
			// 			}

			// 		case 2:
			// 			break
			// 		}

			// 		time.Sleep(time.Millisecond * 2000)
			// 	}

			// }()

			// var wg sync.WaitGroup
			// wg.Add(int(m.NumOfChunks))
			// var mutex = &sync.Mutex{}
			// for idx := int64(0); idx < m.NumOfChunks; idx++ {
			// 	go func(i int64) {
			// 		defer wg.Done()

			// 		t := time.Second * time.Duration(rand.Intn(int(m.NumOfChunks-1)+3))
			// 		time.Sleep(t)

			// 		flag := true
			// 		for flag {

			// 			index := rand.Intn(len(chunkMap[int(i)]))
			// 			ID := chunkMap[int(i)][index]
			// 			offset := i * m.ChunkSize
			// 			size := m.ChunkSize

			// 			if i == m.NumOfChunks-1 {
			// 				size = m.FileSize - offset
			// 			}
			// 			params := &GetChunkParams{
			// 				FileHash:   m.Hash,
			// 				ChunkIndex: i,
			// 				Size:       size,
			// 				Offset:     offset,
			// 				NickName:   cfg.nick,
			// 			}
			// 			resp := &GetChunkResponse{}

			// 			client, err := rpc.Dial("tcp", ID)
			// 			if err != nil {
			// 				fmt.Println(err)
			// 				return
			// 			}
			// 			err = client.Call("File.GetChunk", params, resp)
			// 			if err == nil {

			// 				hash1 := []byte(m.Chunks[int(i)].Hash)
			// 				hash2 := []byte(hashContent(resp.Data))

			// 				if subtle.ConstantTimeCompare(hash1[:], hash2[:]) == 1 {

			// 					start = 1

			// 					mutex.Lock()
			// 					_, err = outputFile.Seek(offset, 0)
			// 					if err != nil {
			// 						panic(err)
			// 					}

			// 					// Write the data to the file
			// 					_, err := outputFile.Write(resp.Data)
			// 					if err != nil {
			// 						panic(err)
			// 					}

			// 					mutex.Unlock()

			// 					m.Chunks[i].Downloaded = true
			// 					flag = false
			// 					fmt.Println("Downloaded Chunk: ", i)
			// 				} else {
			// 					flag = true
			// 				}

			// 			}

			// 		}
			// 	}(idx)
			// }

			// wg.Wait()
			// start = 2

			// outputFile.Close()
			// fmt.Println(m.Name, "was downloaded")

		case "3":
			os.Exit(0)
		}
	}
}
