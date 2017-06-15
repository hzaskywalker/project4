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

import {
    "sync"
    "errors"
    "database"
    "server"
}

type Miner struct{
    //Cache all block in memory
    //I am not going to solve the bonus
    hash2block  *map[string]*Block
    //store the longest chains
    longest *block


    //database handle the balance of each person
    database *DatabaseEngine

    //transfers handles transactions
    tansfers *TransferManager

    //server handle the consensus with other servers
    server *Server
}

func newMiner(server_ *Server) Miner{
    return Miner{
        hash2block: make(map[string]*Block),
        block2Height: make(map[string]int),
        database: NewDatabaseEngin(),
        server: server_
    }
}

func (m *Miner) ServerGetHeight()(int *Block, bool){
    //need check validity here or anything before server
    return m.server.GetHeight(hash)
}

func (m *Miner) ServerGetBlock(hash string)(*Block, bool){
    //need check validity here or anything before server
    return m.server.Get(hash)
}

func (m *Miner) GetBlock(hash string)(*Block, bool){
    block, ok = m.hash2block[hash]
    if ok{
        return block, ok
    }
    block, ok2 = m.ServerGetBlock(hash)
    if ok2==nil || ok2 == false{
        return nil, false
    }
    block.MyHash = hash
    m.InsertBlock(block)
    return block, true
}

func (m *Miner) Findfather(block *Block) (*Block, error){
    fa, ok := GetBlock(block.PrevHash)
    if ok == false{
        return nil, errors.New("No father here")
    }
    return fa, nil
}

func (m *Miner) LCA(a *Block, b *Block)(*Block, error){
    var e, error
    for ;a.GetHeight()!=b.GetHeight(); {
        if a.GetHeight() > b.GetHeight(){
            a, e = m.Findfather(a)
        }
        else{
            b, e = m.Findfather(b)
        }
        if e!=nil{
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

func (m *Miner) UpdateBalance(block *Block)error{
    //Update the balance to this branch

    //Also I need to change the transfer management
    A:=m.longest
    lca, e := m.LCA(A, block)
    if e!=nil{
        return e
    }
    for ;A!=lca; {
        m.database.UpdateBalance(A, -1)
        m.transfers.UpdateBlockStatus(A, 1)//should all be pendding
        A = m.Findfather(A) //The result should be cached 
    }
    var b = []*Block
    for ;B!=lca; {
        append(b, block)
        block = block.Findfather(block)
    }
    for i:=len(b)-1;i>=0;--i{
        m.transfers.UpdateBlockStatus(A, 0)//should all be success
        m.database.UpdateBalance(b[i], 1)
    }
}

func (m *Miner) UpdateLongest(block *Block)error{
    //Suppose the block is ok and checked
    //m.longest should never be nil
    //error may happen when we try to update the balance
    //we need to make sure it's correct
    if m.longest.GetHeight() < block.GetHeight() {
        e = m.UpdateBalance(block)
        if e!=nil{
            return e
        }
        m.longest = block
    }
}

func (m *Miner) InsertBlock(block *Block)error{
    //should this been paralleled?
    hash, e := block.GetHash()
    if e!=nil{
        return e
    }
    fa, ok := Findfather(block)
    if ok {
        m.hash2block[hash] = block

        block.SetHeight(fa.GetHeight()+1)
        m.UpdateLongest(block)
        return nil
    }
    else{
        return errors.New("block's father not found")
    }
}

func (m *Miner) GetHeight(hash string) (int, *Block) {
    //return the height and the block of last Block 
    //There should be no error
    return m.longest.GetHeight(), m.longest
}

func (m *Miner) Get(userId string)(int){
    //return the balance information on the last block
    //There should be no error
    return m.database.Get(userId)
}

func (m *Miner) TRANSFER(t *Transaction)bool{
}

func (m *Miner) Verify(t *Transaction)bool{
    //check
}

func (m *Miner) AddNewBlock(){
    //communicate with the transfer server 
}

func SolveBlock(block* Block){
}

func (m *Miner) mainLoop() error{
    for ;; {
        block := m.Transfer.GetBlock()
        var stop, sucess chan int
        go block.Solve(stop, sucess)

        //do other things..
        for ;;{
            //listen to the server
            select {
                case is_solved<-sucess:
                    if is_solved == 1{
                        if block.GetHeight()>longest.GetHeight(){
                            //check whether the block is after longest
                            //push the block
                            //I don't know when it is successful
                            m.InsertBlock(block)
                            m.UpdateLongest(block)
                        }
                    }
                    else{
                        break
                    }
                case <-time.After(time.Second * 0.0001):
                    fmt.Println("timeout 2")
            }
        }
    }
}