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
    pb "../protobuf/go"
    "os"
    "fmt"
)

type Balance map[string]int

type DatabaseEngine struct {
    balance Balance
    UUID map[string]*Block
    initValue int
    fa *DatabaseEngine
    block *Block //maintain the block
}

func checkKey(userId string)bool{
    return len(userId) == 8
}

func NewDatabaseEngine(fa *DatabaseEngine)*DatabaseEngine{
    //fmt.Println("NewDatabaseEngine")
    D := &DatabaseEngine{balance: make(Balance), initValue: 1000, fa: fa, UUID: make(map[string]*Block)}
    if fa!=nil{
        D.block = D.fa.block
    }
    return D
}

func (db *DatabaseEngine)Add(userId string, delta int)bool{
	if !checkKey(userId){
		return false
	}
    val, ok := db.balance[userId]
    if !ok{
        if db.fa == nil {
            val = db.initValue
        }else{
            val, _ = db.fa.Get(userId)
        }
    }
	val += delta
	db.balance[userId] = val
    if val < 0{
        return false
    }
    return true
}

func (db *DatabaseEngine)Transfer2(from string, to string, value int, value2 int)bool{
    //from pay value
    //to get value2
    ok, ok2 := db.Add(from, -value), db.Add(to, value2)
    if ok && ok2{
        return true
    }
    db.Add(from, value)
    db.Add(to, -value2)
    fmt.Println(db.Get(from))
    return false
}

func (db *DatabaseEngine) GetUUID(UUID string)(*Block, bool){
    b, ok := db.UUID[UUID]
    if ok{
        return b, true
    } else if db.fa!=nil {
        b, ok := db.fa.UUID[UUID]
        if ok{
            db.UUID[UUID] = b
            return b, true
        }
    }
    return nil, false
}

func (db *DatabaseEngine)Transfer(t *pb.Transaction, block *Block, flag int)bool{
    UUID := t.UUID
    if flag >0 {
        b, ok := db.GetUUID(UUID)
        if ok && b!=nil{
            fmt.Println("Invalid UUID", block.BlockID)
            fmt.Println(block, db.block.BlockID)
            os.Exit(1)
            return false
        }
    }
    value, value2 := int(t.Value), int(t.Value - t.MiningFee)
    ok := db.Transfer2(t.FromID, t.ToID, value*flag, value2*flag)
    if ok{
        if flag>0{
            db.UUID[UUID] = block
        } else{
            db.UUID[UUID] = nil
        }
    }
    return ok
}

func (db *DatabaseEngine)Get(userId string)(int, bool){
    val, ok := db.balance[userId]
    if !ok{
        if db.fa == nil{
            val = db.initValue
        } else {
            val, _ = db.fa.Get(userId)
        }
        db.balance[userId] = val
        ok = true
    }
    return val, ok
}

func (db *DatabaseEngine)UpdateBalance(block *Block, flag int)bool{
    fmt.Println(flag)
    //flag is either -1 or 1
    fmt.Println("UpdataBalance")
    num := len(block.Transactions)

    mining_total := 0
    for _, i:=range block.Transactions{
		if int(i.MiningFee)<0{
			return false
		}
        mining_total = mining_total + int(i.MiningFee)
    }

    if flag == -1{
        db.Add(block.MinerID, -mining_total)
    }

    start, end := 0, num
    if flag == -1{
        start, end = num-1, -1
    } 

    for i:=start;i != end;i+=flag{
        ok := db.Transfer(block.Transactions[i], block, flag)
        if !ok{
            for k:=i-flag; k>=0 && k<num; k-=flag{
                db.Transfer(block.Transactions[k], block, -flag)
                if k == start{
                    break
                }
            }
            if flag == -1{
                db.Add(block.MinerID, mining_total)
            }
            return false
        }
    }
    if flag == 1{
        db.Add(block.MinerID, mining_total)
    }
    return true
}

func (db *DatabaseEngine)GetBalance()Balance{
    fmt.Print("")
    return db.balance
}