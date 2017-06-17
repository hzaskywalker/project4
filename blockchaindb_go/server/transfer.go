package main

/*
About transfer:
    1. We have a pool of transaction, with some on longest (type 0) and some on pending (type 1), new (type 2)
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

    dict [3]TransHouse
    lock [3]sync.RWMutex

    //need something to maintain all transaction with flag = 1
    //use map to implement set
    //stopProducer chan bool
    //channel chan *Transaction
}

/*func (T *TransferManager)GetDictSize()int{
    return len(T.dict)
}

func (T *TransferManager)GetPendingSize()int{
    //unsafe, only for debug
    return len(T.Pending)
}*/

func NewTransferManager(_server TransferServer)*TransferManager{
    T := &TransferManager{server: _server}
	for i:=0;i<3;i++{
		T.dict[i] = make(TransHouse)
	}
    //T.PendingNotEmpty = sync.NewCond(&T.PendingLock)

    //go T.Producer() may need add again
    return T
}

func (T *TransferManager)SetFlag(t *Transaction, flag int){
    //set the flag into Pending
	if flag == t.flag{
		return
	}
	//T.lock[t.flag].Lock()
	delete(T.dict[t.flag], t.UUID)
	//T.lock[t.flag].UnLock()
	t.flag = flag
	//T.lock[t.flag].Lock()
	T.dict[t.flag][t.UUID] = t
	//T.lock[t.flag].UnLock()
}

func (T *TransferManager)AddPending(t *Transaction){
	T.lock[2].Lock()
	t.flag = 2
	T.dict[2][t.UUID] = t
	T.lock[2].UnLock()
}

func (T *TransferManager)GetBlocksByBalance(database *DatabaseEngine, result chan *Block, stop chan int){
	for ;;{
		if len(T.dict[2])>0{
			//T.lock[2].Lock()
			//T.lock[1].Lock()
			for _, t := range T.dict[2]{
				t.flag = 1
				T.dict[1][t.UUID] = t
			}
			T.dict[2] = make(TransHouse)
			//T.lock[1].UnLock()
			//T.lock[2].UnLock()
		}
		block := MakeNewBlock()
		mining_total := 0
		for _, t_ := range T.dict[1]{
			select {
				case signal := <- stop:
					if signal == 1{
						return
					}
				default:
					t := t_.trans
					if t.Value > database.Get(t.FromID){
						continue
					}
					database.Transfer(t.FromID, t.ToID, int(t.Value), int(t.Value - t.MiningFee))
					mining_total = mining_total + int(t.MiningFee)
					block.Transactions = append(block.Transactions, t)
					if len(block.Transactions) == 50{
						break
					}
			}
		}
		if len(block.Transactions)>0{
			database.Add(block.MinerID, mining_total)
			return block
		}
	}
}

func (T *TransferManager)GetBlocksByBalance(database *DatabaseEngine, result chan *Block, stop chan int){
    T.server.GetBlocksByBalance(database, result, stop)
}




/*func (T *TransferManager)GetPending()*Transaction{
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

func (T* TransferManager)Producer(){
    for {
        t := T.server.TRANSFER()
        if T.ReadWriteTransaction(t){
        }
    }
}*/

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
