package main
/*
    It's very clear to know what I am doing by looking at the mainloop
    */

import (
    "errors"
    "fmt"
    "time"
    "sync"
    "os"
    //"../hash"
)

type Miner struct{
    hash2block  map[string]*Block
    longest *Block

    databaseLongest *DatabaseEngine
    transfers *TransferManager
    server Server

    mapLock sync.RWMutex
    longestLock sync.RWMutex

    cached bool
	MinerID string
	dataPath string
}

func NewMiner(server_ Server) Miner{
    m := Miner{
        hash2block: make(map[string]*Block),
        databaseLongest: NewDatabaseEngine(nil),
        server: server_,
        transfers: NewTransferManager(server_)}
    return m
}

func (m *Miner) GetTransferManager()*TransferManager{
    return m.transfers
}

func (m *Miner) ServerGetHeight()(int, *Block, bool){
    //need check validity here or anything before server
    return m.server.GetHeight()
}

func (m *Miner) ServerGetBlock(h string)(*Block, bool){
    //need check validity here or anything before server

    //verify that the hash value equal to its real value

    if m.cached {
        block, ok := ReadFromDisck(h, m.dataPath)
        if ok{
            return block, ok
        }
    }

    block, ok := m.server.GetBlock(h)
    if !ok{
        return nil, ok
    }
    Json := block.MarshalToString()
    /*if GetHashString(Json)!= h{
        fmt.Println("GetBlock's hash is not correct")
        return nil, false
    }*/
    if m.cached{
        go WriteJson(h, Json, m.dataPath)
    }
    return block, true
}

func (m *Miner) GetBlock(hash string)(*Block, bool){
	fmt.Println("func GetBlock", hash)
    m.mapLock.RLock()
    block, ok := m.hash2block[hash]
    m.mapLock.RUnlock()
	fmt.Println("block, ok=", block, ok)
    if ok && block!=nil{
        return block, ok
    }else if ok && block==nil{
		fmt.Println("-------error--------")
        return nil, false
    }
	fmt.Println("func GetBlock ask server", hash)
    block2, ok2 := m.ServerGetBlock(hash)
	fmt.Println("func GetBlock ask server end", hash)
    if ok2 == false{
        return nil, false
    }
    block2.MyHash = hash
    m.InsertBlock(block2)
    return block2, true
}

func (m *Miner) Findfather(block *Block) (*Block, error){
	//if block.PrevHash == InitHash{
	fmt.Println("Findfather", block.GetHash())
	if block.BlockID <= 0{
		return nil, nil  //errors.New("is root")
	}
    fa, ok := m.GetBlock(block.PrevHash)
    if ok == false || (fa!=nil && fa.BlockID+1!=block.BlockID) || fa == nil{
		fmt.Println("-------find father error--------")
        return nil, errors.New("No father here")
    }
    return fa, nil
}

func (m *Miner) LCA(a *Block, b *Block)(*Block, error){
    var e error
    //fmt.Println("start")
    for ;a.GetHeight()!=b.GetHeight(); {
        //_, ok:=m.hash2block[b.PrevHash]
        //fmt.Println("LCA k", a.BlockID, b.BlockID, ok)
        if a.GetHeight() > b.GetHeight(){
            a, e = m.Findfather(a)
        }else{
            b, e = m.Findfather(b)
        }
        if e!=nil{
            //fmt.Println("error")
            return nil, errors.New("Get error when finding father in lca")
        }
    }
    for ;a!=b; {
        a, e = m.Findfather(a)
        if e!=nil {
            break
        }
        b, e = m.Findfather(b)
        if e!=nil{
            break
        }
    }
    if a!=b {
        return nil, errors.New("No lca")
    }
    return a, nil
}

func (m *Miner) UpdateBalance(database *DatabaseEngine, block *Block, updateStatus bool)error{

    A := database.block
    if A == block {
        return nil
    }
    lca, e := m.LCA(A, block)
	for A_:=A;A_!=lca; {
		fmt.Println("A_ LCA", A_.GetHash())
		A_, e = m.Findfather(A_)
	}
	for B_:=block;B_!=lca; {
		fmt.Println("B_ LCA", B_.GetHash())
		B_, e = m.Findfather(B_)
	}
	fmt.Println("LCA", lca.GetHash())
	
    if e!=nil{
        return e
    }

    for ;A!=lca; {
		fmt.Println("A LCA", A.GetHash())
        database.UpdateBalance(A, -1)
        if updateStatus{
            m.transfers.UpdateBlockStatus(A, 1)//should all be pendding
        }
        A, e = m.Findfather(A) //The result should be cached 
    }

    var b []*Block
    for B:=block;B!=lca; {
		fmt.Println("B LCA", B.GetHash())
        b = append(b, B)
        B, e = m.Findfather(B)
    }
	fmt.Println("end LCA")
    for i:=len(b)-1;i>=0;i--{
        if database.UpdateBalance(b[i], 1){
            if updateStatus{
                m.transfers.UpdateBlockStatus(b[i], 0)
                //it's done only when we change longest which has been verified 
            }
        }else{
            for j:=i+1;j<=len(b)-1;j++{
                database.UpdateBalance(b[j], -1)
            }
            return errors.New("Get error on calculating balance")
        }
    }
    database.block = block
    return nil
}

func (m *Miner) UpdateLongest(block *Block)error{
    if m.longest.GetHeight() < block.GetHeight() || m.longest.GetHeight() == block.GetHeight() && block.GetHash() < m.longest.GetHash(){
        e := m.UpdateBalance(m.databaseLongest, block, true)
        if e!=nil{
            fmt.Println(e)
            return e
        }
        m.longest = block
    }
    return nil
}

func (m *Miner) VerifyBlock(block *Block)error{
    //fmt.Println(m.databaseLongest)
    database := NewDatabaseEngine(m.databaseLongest)
	fmt.Println("func verify")
    e := m.UpdateBalance(database, block, false)
	fmt.Println("func verify end")
    if e == nil{
        return nil
    } else{
		//fmt.Println("verify func start", block.BlockID)
        m.mapLock.Lock()
        //m.hash2block[block.GetHash()] = nil
		delete(m.hash2block, block.GetHash())
        m.mapLock.Unlock()
		//fmt.Println("end verify func")
		fmt.Println("------error------")
		//os.Exit(1)
        return errors.New("block balance wrong")
    }
}

func (m *Miner) InsertBlock(block *Block)error{
    //Insert block, without verify and update
    hash := block.GetHash()
    _, ok := m.Findfather(block)
    if ok == nil {
        m.mapLock.Lock()
        defer m.mapLock.Unlock()
        m.hash2block[hash] = block
    }else{
        return errors.New("block's father not found")
    }
    return nil
}

func (m *Miner) GetBalance()map[string]int{
    //lock currentDataBase
    if m.databaseLongest.block != m.longest{
        fmt.Println("In miner:GetBalance() m.databaseLongest.block != m.longest")
        os.Exit(1)
    }
    return m.databaseLongest.GetBalance()
}

func (m *Miner) Init(){
    //I don't 
    m.cached = true
    ok := false
    m.hash2block[InitHash] = &Block{MyHash:InitHash}
    m.longest = m.hash2block[InitHash]
    m.longest.BlockID = 0
    m.databaseLongest.block = m.longest

    var newLongest *Block
    //for ;!ok;{
	for i:=0;i<10 && !ok;i++{
        _, newLongest, ok = m.ServerGetHeight()
    }
	if !ok{
		//newLongest = m.longest
		return
	}
	
    e := m.InsertBlock(newLongest)//longest would not be calculated
    if e!=nil{
        fmt.Println(e)
        os.Exit(1)
    }
    e = m.VerifyBlock(newLongest)
    if e!=nil{
        fmt.Println(e)
        os.Exit(1)
    }
    m.UpdateLongest(newLongest)
}

func (m *Miner) AddBlockWithoutCheck(block *Block, finish chan *Block){
    m.InsertBlock(block)
    fmt.Println("finish")
    finish <- block
}

func (m *Miner)GetBlocksByBalance(database *DatabaseEngine, result chan *Block, stop chan int){
    stop_ := make(chan int, 1)
    res_ := make(chan *Block)
    worker := func(res_ chan *Block, stop_ chan int){
        m.transfers.GetBlocksByBalance(database, res_, stop_)
    }
    go worker(res_, stop_)
    for ;;{
        select{
            case <- stop:
                stop_ <- 1
                break
            case res := <- res_:
                result <- res
                stop_ := make(chan int, 1)
                go worker(res_, stop_)
            case <- time.After(time.Second * 10):
        }
    }
    <-stop
}

func (m *Miner) mainLoop(service *Service) error{
    m.Init()

    waitBlocks := make(chan *Block, 50)
    stopSelectTrans := make(chan int, 1)  //should be 1

    is_solved := make(chan *Block)
    var stop_solve chan int

    //toSolve := make([]*Block, 0)
    isAdded := make(chan *Block, 50)

    database := NewDatabaseEngine(m.databaseLongest)
    go m.transfers.GetBlocksByBalance(database, waitBlocks, stopSelectTrans)


    var newBlocks *Block
    for ;; {
        newBlocks = nil
        //fmt.Println(isAdded, is_solved, waitBlocks, service.GetRequest, service.VerifyRequest, service.PushBlockRequest)
        //fmt.Println(service.GetBlockRequest, service.GetHeightRequest)
        fmt.Println("====== main loop =======")

        select {
            case addedBlock := <- isAdded:
                fmt.Println("getNew")
				fmt.Println(addedBlock.GetHeight(), m.longest.GetHeight())
                if addedBlock.GetHeight() > m.longest.GetHeight() || addedBlock.GetHeight() == m.longest.GetHeight() && addedBlock.GetHash() < m.longest.GetHash(){
					fmt.Println("start verify")
					fmt.Println(addedBlock.GetHash(), m.longest.GetHash())
                    e := m.VerifyBlock(addedBlock) //It's better to build a verify list
					fmt.Println("end verify", e)
                    //place where we change the consensus
                    if e == nil{
                        //if true{
						fmt.Println("zxc")
						if stop_solve != nil{
							stop_solve <- 1 //stop solving
							stop_solve = nil
						}
						fmt.Println("asd")
						stopSelectTrans <- 1 //so that pending would release lock
						fmt.Println("qwe")
                        //}

                        //we need stop other verifier, otherwise their databse would be wrong
                        fmt.Println("Update longest", addedBlock.GetHeight())
                        m.UpdateLongest(addedBlock)
						stopSelectTrans = make(chan int, 1)  //should be 1
						database := NewDatabaseEngine(m.databaseLongest)
                        go m.transfers.GetBlocksByBalance(database, waitBlocks, stopSelectTrans)
                    }
                }
				fmt.Println("end getNew")
            case solved := <- is_solved:
				fmt.Println("enter solved")
                //fmt.Println("In Solved")
                newBlocks = solved
				stop_solve = nil
				go WriteBlock(newBlocks, m.dataPath)
				go m.server.PushBlock(newBlocks)
                //stop_solve <- 1
                //stop_solve = nil
                /*if len(toSolve) > 0{
                    stop_solve = make(chan int)
                    go toSolve[0].Solve(stop_solve, is_solved)
                    toSolve = toSolve[1:]
                }*/
				fmt.Println("end solved")

            case block := <- waitBlocks:
                fmt.Println("In waitBlock")
				block.PrevHash = m.longest.GetHash()
				block.BlockID = m.longest.BlockID + 1
				block.MinerID = m.MinerID
                //block.MinerID = "xxxx"
                //toSolve = append(toSolve, block)
                //if stop_solve == nil{
				stop_solve = make(chan int, 1)
				fmt.Println("start solve")
				go block.Solve(stop_solve, is_solved)
                    //toSolve = toSolve[1:]
                //}
            case UserID := <- service.GetRequest:
                //fmt.Println("In GetRequest")
				fmt.Println("get3")
                val, _ := m.databaseLongest.Get(UserID)
				fmt.Println("get4")
                service.GetResponse <- val
            case UUID := <- service.VerifyRequest:
                //m.longest dabase
				fmt.Println("enter verify")
                block, ok := m.databaseLongest.GetUUID(UUID)
                if !ok || block == nil{
                    service.VerifyResponse <- &MyVerifyResponse{t: 3, hash: "?"}
                } else{
                    flag := true
                    t := m.longest
                    for i:=0;i<6;i++{
						if t == nil{
							flag = false
                            break
						}
                        if t.GetHash() == block.GetHash(){
                            flag = false
                            break
                        }
                        t, _ = m.Findfather(t)
                    }
                    if flag{
                        service.VerifyResponse <- &MyVerifyResponse{t:1, hash:block.GetHash()}
                    } else{
                        service.VerifyResponse <- &MyVerifyResponse{t:2, hash:block.GetHash()}
                    }
                }
				fmt.Println("end verify")
            case PushedBlock := <- service.PushBlockRequest:
			fmt.Println("enter PushedBlock")
                //fmt.Println("In push block")
                service.PushBlockResponse <- true
                newBlocks = PushedBlock
				fmt.Println("end PushedBlock")
            case GetBlockHash := <- service.GetBlockRequest:
                //fmt.Println("In GetBlock")
				fmt.Println("enter GetBlockHash")
				m.mapLock.RLock()
                block, ok := m.hash2block[GetBlockHash]
				m.mapLock.RUnlock()
                if !ok{
                    block = nil
                }
                service.GetBlockReponse <- block
				fmt.Println("end GetBlockHash")
            case <-service.GetHeightRequest:
				fmt.Println("enter GetHeightRequest")
                //fmt.Println("In GetHeight")
                service.GetHeightResponse <- m.longest
				fmt.Println("end GetHeightRequest")
            case <- time.After(time.Second * 2):
                //decide wether to start a new block or any other strategy
                //or do nothing
            case <- service.Hello:
                fmt.Println("Hello world")
        }

        if newBlocks!=nil {
            fmt.Println("is added", isAdded)
            go m.AddBlockWithoutCheck(newBlocks, isAdded)
        }
    }
    fmt.Println("End mainloop")
    return nil
}