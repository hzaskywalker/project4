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

    //We need to store the currentDataBase pointer
    //Although in most of the time, currentDataBase == longest
    currentDataBase *Block


    //database handle the balance of each person
    database *DatabaseEngine
    balanceLock sync.RWMutex

    //transfers handles transactions
    transfers *TransferManager

    //server handle the consensus with other servers
    //this is a interface
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

func (m *Miner) UpdateBalance(block *Block)error{
    //Update the balance to this branch

    //Also I need to change the transfer management
    //Add lock here
    m.balanceLock.Lock()
    defer m.balanceLock.Unlock()

    A:=m.currentDataBase
    lca, e := m.LCA(A, block)
    if e!=nil{
        return e
    }
    //fmt.Println(m.currentDataBase.BlockID, block.BlockID)
    for ;A!=lca; {
        m.database.UpdateBalance(A, -1)
        m.transfers.UpdateBlockStatus(A, 1)//should all be pendding
        A, e = m.Findfather(A) //The result should be cached 
    }
    var b []*Block

    for B:=block;B!=lca; {
        b = append(b, B)
        B, e = m.Findfather(B)
    }

    for i:=len(b)-1;i>=0;i--{
        if m.database.UpdateBalance(b[i], 1){
            m.transfers.UpdateBlockStatus(b[i], 0)//should all be success
        }else{
            for j:=i+1;j<=len(b)-1;j++{
                m.database.UpdateBalance(b[i], -1)
            }
            //not for sure
            m.currentDataBase = lca
            //We may need to mask this block to be nil
            //so that we would not calc them again?
            return errors.New("Get error on calculating balance")
        }
    }
    m.currentDataBase = block
    return nil
}

func (m *Miner) UpdateLongest(block *Block)error{
    //Suppose the block is ok and checked
    //m.longest should never be nil
    //error may happen when we try to update the balance
    //we need to make sure it's correct
    if m.longest.GetHeight() < block.GetHeight() {
        e := m.UpdateBalance(block)
        if e!=nil{
            fmt.Println(e)
            return e
        }
        m.longest = block
    }
    return nil
}

func (m *Miner) InsertBlock(block *Block)error{
    //We need verify something
    //should this been paralleled?
    //Here we have checked the block
    //We would only insert the block if it's better than longest
    hash := block.GetHash()
    _, ok := m.Findfather(block)
    if ok == nil {
        m.hash2block[hash] = block
        e := m.UpdateBalance(block)

        if e == nil{
            m.UpdateLongest(block)
            return nil
        }else{
            m.hash2block[hash] = nil
            return errors.New("block balance wrong")
        }
    }else{
        return errors.New("block's father not found")
    }
}

func (m *Miner) GetHeight(hash string) (int, *Block) {
    //return the height and the block of last Block 
    //There should be no error
    return m.longest.GetHeight(), m.longest
}

func (m *Miner) Get(userId string)(int, bool){
    //return the balance information on the last block
    //There should be no error
    return m.database.Get(userId)
}

func (m *Miner) TRANSFER(t *Transaction)bool{
    return false
}

func (m *Miner) Verify(t *Transaction)bool{
    //check
    return false
}

func (m *Miner) GetBalance()map[string]int{
    //lock currentDataBase
    if m.currentDataBase != m.longest{
        m.UpdateBalance(m.longest)
    }
    //the result of balance should never be error
    return m.database.GetBalance()
}

func (m *Miner) GetBalanceString(hash string)(map[string]int, bool){
    block, ok := m.GetBlock(hash)
    if !ok{
        return nil, false
    }
    if block != m.currentDataBase{
        if m.UpdateBalance(block)==nil{
            return nil, false
        }
    }
    return m.database.GetBalance(), true
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
}


func (T *TransferManager)GetPendingByBalance(balance *map[string]int, result chan *Transaction, stop chan int){
    //should stop by stop
    T.PendingLock.Lock()
    defer T.PendingLock.Unlock()


    for ;len(T.Pending)==0;{
        T.PendingNotEmpty.Wait()
    }

    var t *Transaction
    for _, val:=range T.Pending{
        t = val
        val.flag = 2
        break
    }
    if t!=nil{
        delete(T.Pending, t.UUID)
    }
    return t
}

func (m *Miner) mainLoop() error{
    /*
        You have only following two thread:
            1. one for solve hash proble
            2. one for:
                1. receiving other's push, save in a buffer of server
                    if we found it's longer than current, add it and check balance
                    otherwise we ignore it
                2. if we don't have a longer block
                    we keep our balance on longest block
                    verify new transactions and select block 

                However we want parallel the check and select
                    How about maitaining two balance?
                    Or just copy it?

                Right now let's forget it
        */
    get_pending := make(chan *Transaction)
    stop_pending := make(chan int)

    is_solved := make(chan int)

    server_query := make(chan int) //stands for all serveer query

    currentBlock = MakeNewBlockAfter(m.longest, "MYID")
    var solvingBlock *Block
    go m.transfers.GetPendingByBalance(m.GetBalance(), get_pending)

    for ;; {
        /*
        In each round, either
            1. Solve a block
                or Recieve Put a new Block
            2. Get a new transaction 
            */

        //do other things..

        var newBlocks *Block
        newBlocks = nil

        select {
            case is_solved := <- success:
                if is_solved == 1{
                    newBlocks = solvingBlock
                    solvingBlock = nil
                }
            case t := <- get_pending{
                //update balance by t
                go m.transfers.GetPendingByBalance(m.GetBalance(), get_pending, stop_pending)
            }
            case s := <-server_query{
                //anser the server query
            }
            case <-time.After(time.Second):
                //decide wether to start a new block or any other strategy
                //or do nothing
        }

        if newBlocks{
            //add newBlocks
            //return balance to longest
            //then check for new putting block if it's blockid is heigher?
        }
    }
}