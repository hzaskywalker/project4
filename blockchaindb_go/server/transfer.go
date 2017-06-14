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

import {
        "fmt"
        "server"
    }

type TransferManager struct{
    server *Server

    dict map[string] *Transaction
    lock RWMutex

    channel chan *Transactions

    //need something to maintain all transaction with flag = 1
    pending map[string] *Transaction
    pendingLock RWMutex
}

func (T *TransferManager)SetFlag(t *Transaction, flag int){
    //set the flag into pending
    if flag == 1 && t.flag!=1 {
        pendingLock.Lock()
        defer pendingLock.Unlock()

        T.pending[t.UUID] = t
        t.flag = flag
    }
    else{
        t.flag = flag
    }
}

func (T *TransferManager)GetPending()*Transactions{
    //return a pending transaction
    //return nil means there is no pending transaction

    //may need carefully design
    T.pendingLock.Lock()
    defer T.pendingLock.Unlock()

    to_delete_list = []string

    //I think t == nil at the beginning
    var t *Transaction
    for key, val:=range T.pending{
        if val.flag == 1{
            t = val
        }
        else{
            append(to_delete_list, key)
        }
    }
    for idx, i:=range to_delete_list{
        delete(T.pending, i)//maybe
    }
    t.flag = 2
    return t
}

func (T *TransferManager)ReadTransaction(UUID string)(*Transaction, bool){
    lock.RLock()
    defer lock.RUnlock()

    t, ok = T.dict[UUID]
    return t, ok
}

func (T *TransferManager)WriteTransaction(t *Transaction){
    lock.WLock()
    defer lock.WUnlock()

    T.dict[UUID] = t
}

func (T *TransferManager)ReadWriteTransaction(t *Transaction)bool{
    //check whether we have not seen this transaction
    lock.RLock()
    defer lock.RUnlock()
    t, ok = T.dict[UUID]
    if ok {
        return false
    }
    lock.WLock()
    defer lock.WUnlock()

    T.SetFlag(t, 1)
    T.dict[UUID] = t
    return true
}

func (T *TransferManager)UpdateBlockStatus(block *Block, status int)error{
    //add or delete the informations in the block
    lock.WLock()
    defer lock.WUnlock()
    //change the flag
    //add new transactions into the pool

    for idx, t :=range Block.transactions{
        //use UUID to mark the flag of transaction
        //avoid the difference version of flag
        T.dict[t.UUID] = t
        T.SetFlag(t, flag)
    }

}

func (T* Transfer)Producer(){
    for {
        transaction := T.server.TRANSFER()
        if T.ReadWriteTransaction(transaction){
            T.channel <- transaction
        }
    }
}

func (T* Transfer)Customer(){
    //add the new  
    transaction := <- T.channel
}

func (T *TransferManager)GetBlock(channel *Block){
    //get a new block with certain number of transfers
    block := MakeNewBlock()
    for i=0;i<50;++i{
        for;;{
            t := T.GetPending()
            if t!=nil{
                break
            }
        }
        append(block.transactions, t)
    }
    channel <- block
}