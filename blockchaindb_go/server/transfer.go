package main

/*
About transfer:
    0. the trasnsaction in longest has flag 0, the one in undecided block has flag 2, pending one has flag 1
    1. Using another thread to store the transaction
    2. Store the new trasaction into a buffer and push it to other server
    3. Verify it with the Miner
    4. Push the transaction to Miner if we get a certain number of transactions
        return True if it's added into the longest branch
        Otherwise return failed
    5. Miner and Transfer shall maitain consensus about all happened transfer
        However I don't want Miner to handle it.
    */

import (
        //"fmt"
        "sync"
        pb "../protobuf/go"
    )


type Transaction struct{
    flag int //state of the transaction, sucess(0), pending(1), or not in longest (2)
    UUID string
    trans *pb.Transaction
}

type TransferManager struct{
    server *Server

    dict map[string] *Transaction
    lock sync.RWMutex

    channel chan *Transaction

    //need something to maintain all transaction with flag = 1
    pending map[string] *Transaction
    pendingLock sync.RWMutex
}


func NewTransferManager()*TransferManager{
    return &TransferManager{}
}

func (T *TransferManager)SetFlag(t *Transaction, flag int){
    //set the flag into pending
    if flag == 1 && t.flag!=1 {
        T.pendingLock.Lock()
        defer T.pendingLock.Unlock()

        T.pending[t.UUID] = t
        t.flag = flag
    }else{
        t.flag = flag
    }
}

func (T *TransferManager)GetPending()*Transaction{
    //return a pending transaction
    //return nil means there is no pending transaction

    //may need carefully design
    T.pendingLock.Lock()
    defer T.pendingLock.Unlock()

    to_delete_list := make([]string,0)

    //I think t == nil at the beginning
    var t *Transaction
    for key, val:=range T.pending{
        if val.flag == 1{
            t = val
        }else{
            to_delete_list = append(to_delete_list, key)
        }
    }
    for _, i:=range to_delete_list{
        delete(T.pending, i)//maybe
    }
    t.flag = 2
    return t
}

func (T *TransferManager)ReadTransaction(UUID string)(*Transaction, bool){
    T.lock.RLock()
    defer T.lock.RUnlock()

    t, ok := T.dict[UUID]
    return t, ok
}

func (T *TransferManager)WriteTransaction(t *Transaction){
    T.lock.Lock()
    defer T.lock.Unlock()

    T.dict[t.UUID] = t
}

func (T *TransferManager)ReadWriteTransaction(t *Transaction)bool{
    //check whether we have not seen this transaction
    T.lock.RLock()
    defer T.lock.RUnlock()
    t, ok := T.dict[t.UUID]
    if ok {
        return false
    }
    T.lock.Lock()
    defer T.lock.Unlock()

    T.SetFlag(t, 1)
    T.dict[t.UUID] = t
    return true
}

func (T *TransferManager)UpdateBlockStatus(block *Block, flag int){
    //add or delete the informations in the block
    T.lock.Lock()
    defer T.lock.Unlock()
    //change the flag
    //add new transactions into the pool

    for _, t :=range block.Transactions{
        //use UUID to mark the flag of transaction
        //avoid the difference version of flag

        //may insert
        T.dict[t.UUID].trans = t
        T.SetFlag(T.dict[t.UUID], flag)
    }
}

func (T* TransferManager)Producer(){
    for {
        transaction := T.server.TRANSFER()
        if T.ReadWriteTransaction(transaction){
            T.channel <- transaction
        }
    }
}

func (T* TransferManager)Customer()*Transaction{
    //add the new  
    transaction := <- T.channel
    return transaction
}

func (T *TransferManager)GetBlock(channel chan *Block){
    //get a new block with certain number of transfers
    block := MakeNewBlock()
    for i:=0;i<50;i++{
        for;;{
            t := T.GetPending()
            if t!=nil{
                block.Transactions = append(block.Transactions, t.trans)
                break
            }
        }
    }
    channel <- block
}