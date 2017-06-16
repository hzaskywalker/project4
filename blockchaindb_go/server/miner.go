package main

/*
A miner should maintain:

The tree of the blocks, the node of longest chain, and the corresponding balance.

We can:
    1. return the balance of any block:
        find the lca to the longest chain and return the corresponding balance
    2. get the longest block

Modification to the status:
    1. According to getHeight:
        1. add blocks or change the branch
    2. Collect transactions and calculate new blocks starting from the longest chain

Cache and comminication:
    1. Miner can suppose everything is stored on disk, net.go handle the communication between 
       the miner and the (network or disk)

About Transfer:
    1. Reader and writer model
    2. see transfer.go

About mining:
    1. New blocks that haven't been broadcast should not be added into Block 
        => I am still think about it

Some notes:
    1. We use a simple map from string to block so that we can cache all of the blocks that we have.
       Once a block is missing, we use get block to retrive the corresponding blocks.
    2. The most of the computation should be used to solve the hash problem, so we needn't worry too
       much about the other computations.
   */

import (
    "errors"
    "fmt"
    "time"
    "sync"
    "os"
    "../hash"
)

type Miner struct{
    //Cache all block in memory
    //I am not going to solve the bonus
    hash2block  map[string]*Block
    //store the longest chains
    longest *Block

    //database handle the balance of each person
    database *DatabaseEngine

    //transfers handles transactions
    transfers *TransferManager

    server Server
}

func NewMiner(server_ Server) Miner{
    m := Miner{
        hash2block: make(map[string]*Block),
        database: NewDatabaseEngine(),
        server: server_,
        transfers: NewTransferManager(server_)}

    go m.transfers.Producer()
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
    block, ok := m.server.GetBlock(h)
    if !ok{
        return nil, ok
    }
    if hash.GetHashString(block.MarshalToString()) != h{
        fmt.Println("GetBlock's hash is not correct")
        return nil, false
    }
    return block, true
}

func (m *Miner) GetBlock(hash string)(*Block, bool){
    block, ok := m.hash2block[hash]
    if ok && block!=nil{
        return block, ok
    }else if ok && block==nil{
        return nil, false
    }
    block2, ok2 := m.ServerGetBlock(hash)
    if ok2 == false{
        return nil, false
    }
    block2.MyHash = hash
    m.InsertBlock(block2)
    return block2, true
}

func (m *Miner) Findfather(block *Block) (*Block, error){
    fa, ok := m.GetBlock(block.PrevHash)
    if ok == false || (fa!=nil && fa.BlockID+1!=block.BlockID) || fa == nil{
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
    if e!=nil{
        return e
    }

    for ;A!=lca; {
        database.UpdateBalance(A, -1)
        if updateStatus{
            m.transfers.UpdateBlockStatus(A, 1)//should all be pendding
        }
        A, e = m.Findfather(A) //The result should be cached 
    }

    var b []*Block
    for B:=block;B!=lca; {
        b = append(b, B)
        B, e = m.Findfather(B)
    }

    for i:=len(b)-1;i>=0;i--{
        if m.database.UpdateBalance(b[i], 1){
            if updateStatus{
                m.transfers.UpdateBlockStatus(b[i], 0)
                //it's done only when we change longest which has been verified 
            }
        }else{
            for j:=i+1;j<=len(b)-1;j++{
                m.database.UpdateBalance(b[i], -1)
            }
            m.currentDataBase = lca
            return errors.New("Get error on calculating balance")
        }
    }
    database.block = block
    return nil
}

func (m *Miner) UpdateLongest(block *Block)error{
    if m.longest.GetHeight() < block.GetHeight() {
        e := m.UpdateBalance(m.databse, block, true)
        if e!=nil{
            fmt.Println(e)
            return e
        }
        m.longest = block
    }
    return nil
}

func (m *Miner) VerifyBlock(block *Block)error{
    database := NewDatabaseEngine(m.database)
    e := m.UpdateBalance(database, block, false)

    if e == nil{
        return nil
    }else{
        m.hash2block[hash] = nil
        return errors.New("block balance wrong")
    }
}

func (m *Miner) InsertBlock(block *Block)error{
    //Insert block, without veryfy and update
    hash := block.GetHash()
    _, ok := m.Findfather(block)
    if ok == nil {
        m.hash2block[hash] = block
    }else{
        return errors.New("block's father not found")
    }
}

func (m *Miner) GetBalance()map[string]int{
    //lock currentDataBase
    if m.currentDataBase != m.longest{
        m.UpdateBalance(m.longest, false)
    }
    return m.database.GetBalance()
}

func (m *Miner) Init(){
    //I don't 
    go m.transfers.Producer()
    ok := false
    m.hash2block[InitHash] = &Block{MyHash:InitHash}
    m.longest = m.hash2block[InitHash]
    m.currentDataBase = m.hash2block[InitHash]
    m.longest.BlockID = 0

    var newLongest *Block
    for ;!ok;{
        _, newLongest, ok = m.ServerGetHeight()
    }
    e := m.InsertBlock(newLongest)//longest would not be calculated
    if e!=nil{
        fmt.Println(e)
        os.Exit(1)
    }
    e := m.VerifyBlock(newLongest)
    if e!=nil{
        fmt.Println(e)
        os.Exit(1)
    }
    m.UpdateLongest(newLongest)
}

func (m *Miner) AddBlockWithoutCheck(block *Block, finish chan *Block){
    m.InsertBlock(block)
    finish <- block
}

func (m *Miner) mainLoop() error{
    waitBlocks := make(chan *Block, 50)
    stopSelectTrans := make(chan int)

    is_solved := make(chan int, 50)
    stop_solve := make(chan int)

    isAdded := make(chan *block, 50)
    server_query := make(chan int) //stands for all serveer query

    database := NewDatabaseEngine(m.database)
    go m.transfers(database, waitBlocks, stopSelectTrans)

    for ;; {
        var newBlocks *Block
        newBlocks = nil

        select {
            case addedBlock := <- is_add:
                if addedBlock.GetHeight() > m.longest.GetHeight(){
                    e := VerifyBlock(addedBlock)//place where we change the consensus
                    if e == nil{
                        stop_solve <- 1 //stop solving
                        stopSelectTrans <- 1 //so that pending would release lock
                        m.UpdateLongest(addedBlock)

                        go m.transfers.GetBlocksByBalance()
                    }
                }
            case solved := <- is_solved:
                newBlocks = solvingBlock
            case block := <- waitBlocks{
                go block.Solve(stop_solve, is_add)
            }
            case s := <-server_query{
            }
            case <-time.After(time.Second):
                //decide wether to start a new block or any other strategy
                //or do nothing
        }

        if newBlocks{
            go m.AddBlockWithoutCheck(newBlocks, is_add)
        }
    }
}