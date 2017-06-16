package main

/*
A database for transfering money
I maintain the balance here.
It's not for block.

We keep the result of the balance always non-negative after every function

need lock?

Question:
    Should we use some persistent data structure for thread safety?
    Now we only mining one block at the same time which means that there is no multi thread problem.
*/


import (
    "sync"
    "fmt"
    //"errors"
)

type Balance map[string]int

type DatabaseEngine struct {
    balance Balance
    initValue int
    sync.RWMutex
}

func checkKey(userId string)bool{
    return len(userId) == 8
}

func NewDatabaseEngine()*DatabaseEngine{
    fmt.Println("NewDatabaseEngine")
    return &DatabaseEngine{balance: make(Balance), initValue: 1000}
}

func (db *DatabaseEngine)Add(userId string, delta int)bool{
	if !checkKey(userId){
		return false
	}
    val, ok := db.balance[userId]
    if !ok{
        val = db.initValue
    }
	val += delta
	db.balance[userId] = val
    if val < 0{
        return false
    }
    return true
}

func (db *DatabaseEngine)Transfer(from string, to string, value int, value2 int)bool{
    //from pay value
    //to get value2
    ok, ok2 := db.Add(from, -value), db.Add(to, value2)
    if ok && ok2{
        return true
    }
    db.Add(from, value)
    db.Add(to, -value2)
    return false
}

func (db *DatabaseEngine)Get(userId string)(int, bool){
	if !checkKey(userId){
		return 0, false
	}
    val, ok := db.balance[userId]
    if !ok{
        val = db.initValue
        db.balance[userId] = val
		ok = true
    }
    return val, ok
}

func (db *DatabaseEngine)UpdateBalance(block *Block, flag int)bool{
    //flag is either -1 or 1
    num := len(block.Transactions)

    mining_total := 0
    for _, i:=range block.Transactions{
        //suppose i.MiningFee >= 0
		if int(i.MiningFee)<0{
			return false
		}
        mining_total = mining_total + int(i.MiningFee)
    }

    /*if flag == -1{
        //When flag == -1, I don't need to check whether it's correct
        db.Add(block.MinerID, -mining_total)
    }*/

    start, end := 0, num
    if flag == -1{
        start, end = num-1, -1
    } 
    //fmt.Println("start-end:", start, end)
    for i:=start;i != end;i+=flag{
        //fmt.Println(i, flag)
        transaction := block.Transactions[i]

        value, value2 := int(transaction.Value), int(transaction.Value - transaction.MiningFee)
        ok := db.Transfer(transaction.FromID, transaction.ToID, value*flag, value2*flag)

        if !ok{
            //restore the transaction before
            for k:=i-flag;;k-=flag{
                transaction := block.Transactions[k]
                value, value2 := int(transaction.Value), int(transaction.Value - transaction.MiningFee)
                db.Transfer(transaction.FromID, transaction.ToID, -value*flag, -value2*flag)
                if k == start{
                    break
                }
            }
            return false
        }
    }
    db.Add(block.MinerID, mining_total*flag)
    return true
}

func (db *DatabaseEngine)GetBalance()Balance{
    return db.balance
}