package main

/*
About transfer:
    1. We have a pool of transaction, with some on longest (type 0) and some on pending (type 1)
    2. type 0 would change only when changing longest, new comer become type 1

    3. Miner would have a thread that receive all of the things of type 1 and the new comer.
        1. We have 2 producer for miner: previous pending and new comers 
        2. Miner can store the transactions now and decide later
        3. But we need to guarantee that new comer is not the same with the previous one
    */

import (
        "fmt"
        "sync"
        pb "../protobuf/go"
    )

type TransferServer interface{
    TRANSFER()*Transaction
    GetBlocksByBalance(*DatabaseEngine, chan *Block, chan int)
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
    stopProducer chan bool

    PendingNotEmpty *sync.Cond
    channel chan *Transaction
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
    T.lock.Lock()
    defer T.lock.Unlock()

    _, ok := T.dict[t.UUID]
    if ok {
        return false
    }

    T.dict[t.UUID] = t
    T.SetFlag(t, 1)
    return true
}

func (T *TransferManager)UpdateBlockStatus(block *Block, flag int){
    T.lock.Lock()
    defer T.lock.Unlock()

    for _, t :=range block.Transactions{
        _, ok := T.dict[t.UUID]
        if !ok{
            T.dict[t.UUID] = new(Transaction)
            T.dict[t.UUID].flag = flag
        }
        T.dict[t.UUID].trans = t
        T.SetFlag(T.dict[t.UUID], flag)
        if flag == 3{
            delete(T.dict, t.UUID)
        }
    }
}

/*
func (T* TransferManager)Producer(){
    for {
        select {
            case <- self.stopProducer:
                return
            default:
                t := T.server.TRANSFER()
                if T.ReadWriteTransaction(t){
                    T.channel <- t
                }
        }
    }
}

func (T *TransferManager)GetBlock(channel chan *Block){
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

func (T *TransferManager) ProducerByPendingList(Pending TransHouse, channel chan *Transaction){
    for _, t:=range Pending{
        channel <- t
    }
}

func (T *TransferManager)GetBlocksByBalance(database *DatabaseEngine, result chan *Block, stop chan int){
    //Maybe we need to verify the previous result first
    T.stopProducer <- true
    T.channel = make(chan *Transactions, 500)
    go T.Producer()

    T.PendingLock.Lock()
    go ProducerByPendingList(T.Pending, T.channel)
    T.Pending = make(TransHouse) 
    T.PendingLock.Unlock()

    block := MakeNewBlock()
    for ;;{
        select {
            case t := <- T.channel:
                //I also need to check UUID
                if t.Value > database.Get(t.FromID){
                    continue
                }

                m.database.Transfer(t.FromID, t.ToID, int(t.Value), int(t.Value - t.MiningFee))
                block.Transactions = append(block.Transactions, t)

                if len(block.Transactions) == 50{ //Or stop
                    for _, tmp:=range block.Transactions{
                        database.Add(block.MinerID, int(tmp.MiningFee))
                    }
                    result <- block
                    block = MakeNewBlock()
                }
            case signal := <- stop:
                if signal == 1{
                }
                return
        }
    }
}
*/
func (T *TransferManager)GetBlocksByBalance(database *DatabaseEngine, result chan *Block, stop chan int){
    T.server.GetBlocksByBalance(database, result, stop)
}
