package main

/*
A database for transfering money

Question:
    Should we use some persistent data structure for thread safety?
    Now we only mining one block at the same time which means that there is no multi thread problem.
*/


type Balance map[string]int

import {
    "sync"
    "errors"
    "block"
}

type DatabaseEngine struct {
    balance Balance
    sync.RWMutex
}

func NewDatabaseEngin(_config *DatabaseConfig, _logger *Logger){
    return DatabaseConfig{balance:make(map[string]int), config:_config, logger:_logger}
}

func checkKey(userId string){
    return len(string) == 8
}

func (db *DatabaseEngine)Transfer(from string, to string, delta int)(int, int, error){
    if delta < 0{
        a, b, c := Transfer(to, from, -delta)
        return b, a, c
    }
    //db.Lock()
    //defer db.Unlock()

    from_val, from_ok := db.balance[from]
    to_val, to_ok := db.balance[to]
    if !from_ok {
        from_val = 0
    }
    if !to_ok {
        to_val = 0
    }
    if from_val < delta{
        return from_val, to_val, errors.New("Transfer: Not enough money")
    }
    from_val = from_val - delta
    to_val = to_val + delta
    db.balance[from] = from_val
    db.balance[to] = to_val
    return from_val, to_val, nil
}

func (db *DatabaseEngine)Get(userId string)(int, error){
    return db.Balance[userId]
}

func (db *DatabaseEngine)Add(userId string, value int)(int, error){
    val, ok = db.Balance[userId]
    if !ok{
        return 0, errors.New("No user in Add, should we add the account?")
    }
    val += value
    db.Balance[userId] = val
    return val, nil
}

func (db *DatabaseEngine)UpdateBalance(block *Block, int flag)error{
    //flag is either -1 or 1
    miner_id := Block.MinerID
    num := len(block.Transactions)

    mining_total := 0 
    for idx, i:=range Block.transactions{
        mining_total = mining_total + i.MiningFee
    }

    //If minerId not in balance, what would happen?
    //If the user not in balance, what would happen?
    //chech Minder here?

    if flag == -1{
        db.Add(Block.MinerID, -mining_total)
    }
    for i:=0;i<num;i++{
        j := i
        if flag<0{
            j = num-i-1
        }
        transaction := *Block.transactions[j]
        a,b,ok := db.Transfer(transaction.FromID, transactions.ToID, transactions.value * flag)

        if ok != nil{
            //restore the transaction before
            for k:=0;k<i;k++{
                j := k
                if flag<0 {
                    j=num-k-1
                }
                db.Transfer(transactions.FromID, transactions.ToID, transactions.value * -flag)
            }
            if flag == -1{
                //This shouldn't happend because flag==-1 if and only if the transaction has succeed before. 
                db.Add(Block.MinerID, mining_total)
            }
            return errors.New("Block failed, nothing happend.")
        }
    }
    if flag == 1{
        db.Add(Block.MinerID, mining_total)
    }
    return nil
}