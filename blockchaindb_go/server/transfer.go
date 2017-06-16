package main

/*
About transfer:
    0. the trasnsaction in longest has flag 0, the one in undecided block has flag 2, Pending one has flag 1
    1. Using another thread to store the transaction
    2. Store the new trasaction into a buffer and push it to other server
    3. Verify it with the Miner
    4. Push the transaction to Miner if we get a certain number of transactions
        return True if it's added into the longest branch
        Otherwise return failed
    5. Miner and Transfer shall maitain consensus about all happened transfer
        However I don't want Miner to handle it.
    6. We needn't store the transanctions on disk. Suppose program crashed, we can rebuild the Transfer system easily by blocks
        Just forget anything about the pending one.
    */

import (
        "fmt"
        "sync"
        pb "../protobuf/go"
    )

type TransferServer interface{
    TRANSFER()*Transaction
}


type Transaction struct{
    flag int //state of the transaction, sucess(0), Pending(1), or not in longest (2)
    UUID string
    trans *pb.Transaction
}

func NewTransaction()*Transaction{
    return &Transaction{}
}

func (t *Transaction) String()string{
    return fmt.Sprintf("UUID: %s: From %s, To %s, Value: %d", t.UUID, t.trans.FromID, t.trans.ToID, int(t.trans.Value))
}



type TransHouse map[string]*Transaction

type TransferManager struct{
    server TransferServer

    dict TransHouse
    lock sync.RWMutex

    //need something to maintain all transaction with flag = 1
    //use map to implement set
    Pending TransHouse
    PendingLock sync.RWMutex

    PendingNotEmpty *sync.Cond
}

func (T *TransferManager)GetDictSize()int{
    return len(T.dict)
}

func (T *TransferManager)GetPendingSize()int{
    //unsafe, only for debug
    return len(T.Pending)
}

func NewTransferManager(_server TransferServer)*TransferManager{
    T := &TransferManager{server: _server, 
        Pending: make(TransHouse),
        dict: make(TransHouse)}
    T.PendingNotEmpty = sync.NewCond(&T.PendingLock)

    //go T.Producer() may need add again
    return T
}

func (T *TransferManager)SetFlag(t *Transaction, flag int){
    //set the flag into Pending
    if flag == 1 && t.flag!=1 {
        T.PendingLock.Lock()
        defer T.PendingLock.Unlock()

        T.Pending[t.UUID] = t
        t.flag = flag

        T.PendingNotEmpty.Signal()
    }else{
        if t.flag == 1 && flag!=1{
            T.PendingLock.Lock()
            defer T.PendingLock.Unlock()

            delete(T.Pending, t.UUID)
        }
        t.flag = flag
    }
}

func (T *TransferManager)GetPending()*Transaction{
    //flag 0 in longest chain
    //flag 1 in Pending list
    //flag in some blocks that not on the longest chain

    //return a Pending transaction
    //wait if the result is nil

    //may need carefully design
    T.PendingLock.Lock()
    defer T.PendingLock.Unlock()


    for ;len(T.Pending)==0;{
        //fmt.Println("Pending")
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
    //fmt.Println("get something in Pending list: ", t)
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
    //seems that Go doesn't support two phase lock..
    T.lock.Lock()
    defer T.lock.Unlock()

    _, ok := T.dict[t.UUID]
    if ok {
        return false
    }
    //T.lock.Lock()
    //defer T.lock.Unlock()

    T.dict[t.UUID] = t
    T.SetFlag(t, 1)
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

        _, ok := T.dict[t.UUID]
        if !ok{
            T.dict[t.UUID] = new(Transaction)
            T.dict[t.UUID].flag = flag
        }
        T.dict[t.UUID].trans = t
        //fmt.Println(t.UUID, flag, len(T.Pending))
        T.SetFlag(T.dict[t.UUID], flag)
        if flag == 3{
            //delete: for debug
            delete(T.dict, t.UUID)
        }
    }
}

func (T* TransferManager)Producer(){
    for {
        t := T.server.TRANSFER()
        T.ReadWriteTransaction(t)
    }
}

func (T *TransferManager)GetBlock(channel chan *Block){
    //get a new block with certain number of transfers
    block := MakeNewBlock()
    for i:=0;i<50;i++{
        t := T.GetPending()
        block.Transactions = append(block.Transactions, t.trans)
    }
    channel <- block
}

func (T *TransferManager)GetBlockSync()*Block{
    channel := make(chan *Block)
    go T.GetBlock(channel)
    return <- channel
}