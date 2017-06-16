package main
/*
    Test the database and the miner
    Check the balance and block suppose we already have a chain

    Prepare:
        1. Build a fake server with existing block chains
            1. First generate a sequence of transfer 
            2. Build up the blocks
        2. Add some fake transfer across the server 
        3. Calculate the Fake balance sequentially.

    Check:
        1. The balance is correct at any chain
        2. When get Height changed, the processure is correct
    */

import (
    "fmt"
    "os"
    "math/rand"
    pb "../protobuf/go"
)


type MyServer struct{
    //brute force code for balance calculating
    FakeServer //in test_server

    blocks map[string]*Block
    longest *Block
}

func UpdateBalance(balance map[string]int, block *Block){
    totalFee := 0
    for _, t := range block.Transactions{
        balance[t.FromID] -= int(t.Value)
        if t.Value<0 || balance[t.FromID]<0 || t.MiningFee < 0 || t.Value < t.MiningFee{
            fmt.Println("Error on Update Balance, the transaction is")
            fmt.Println(t)
            os.Exit(1)
        }
        balance[t.ToID] += int(t.Value - t.MiningFee)
        totalFee += int(t.MiningFee)
    }
    balance[block.MinerID] += totalFee
}

func ReverseSlice(ans []*Block)[]*Block{
    if len(ans) == 1{
        return ans
    }
    return append(ReverseSlice(ans[1:]), ans[0])
}

func (s *MyServer)GenerateChain(hash string)[]*Block{
    blocks := make([]*Block, 0)

    current, _ := s.blocks[hash]
    blocks = append(blocks, current)

    for ;current.BlockID!=1;{
        current, _ = s.blocks[current.PrevHash]
        blocks = append(blocks, current)
    }

    return ReverseSlice(blocks)
}

func (s *MyServer)CalcBalance(hash string)map[string]int{
    //calculating the balance by brute force 
    balance := make(map[string]int)
    for _, id := range s.people_id{
        //Need check
        balance[id] = 1000
    }
    if hash == InitHash{
        return balance
    }
    blocks := s.GenerateChain(hash)

    for i:=0; i<len(blocks);i++{
        UpdateBalance(balance, blocks[i])
    }
    return balance
}

func CompareBalance(A map[string]int, B map[string]int)bool{
    //check A is the same with B
    //for the element that is not in A, we set it to be 1000
    for key, val := range A{
        if val>0{
            val2, ok := B[key]
            if !ok || val2!=val{
                fmt.Println("B false ", key, val, val2)
                return false
            }
        }
    }
    for key, val := range B{
        if val>0{
            val2, ok := A[key]
            if !ok{
                val2 = 1000
            }
            if val2!=val{
                fmt.Println("false ", key, val, val2)
                return false
            }
        }
    }
    return true
}

func (s *MyServer)init(n int, initilize int, num_trans int){
    s.balance = make(map[string]int)
    s.blocks = make(map[string]*Block)
    for i:=0;i<n;i++{
        s.people_id = append(s.people_id, fmt.Sprintf("%08x", i)) //%x or %d?
    }
    for i:=0;i<n;i++{
        s.balance[s.people_id[i]] = initilize
    }

    blocks := make([]*Block, 0)
    Transactions := make([]*pb.Transaction, 0)
    totalMiningFee := 0

    PrevHash := InitHash

    for i:=0;i<num_trans;i++{
        //genereate trans randomly
        //I can later try to generate in parallel with the miner
        a := s.people_id[rand.Intn(n)]
        b := s.people_id[rand.Intn(n)]

        value := rand.Intn(s.balance[a] + 1)
        mining_fee := rand.Intn(value + 1)
        totalMiningFee += mining_fee

        s.balance[a] -= value
        s.balance[b] += value - mining_fee

        Transactions = append(Transactions, GenTransaction(a, b, value, mining_fee, fmt.Sprintf("%08x", i)).trans)

        if i==num_trans-1 || i%50 == 49{
            block := MakeNewBlock()
            block.Transactions = Transactions

            block.MinerID = s.people_id[rand.Intn(n)]
            s.balance[block.MinerID] += totalMiningFee

            block.PrevHash = PrevHash
            block.BlockID = int32(len(blocks) + 1)
            block.Transactions = Transactions
            block.SolveSync()
            PrevHash = block.GetHash()
            s.blocks[PrevHash] = block

            blocks = append(blocks, block)
            Transactions = make([]*pb.Transaction, 0)
            totalMiningFee = 0
        }
    }
    s.longest = blocks[len(blocks)-1]
    balance := s.CalcBalance(s.longest.GetHash())
    fmt.Println("Check Update Balance int test_databse: ", CompareBalance(balance, s.balance))
}

func TestDatabaseEngine(s *MyServer){
    D := NewDatabaseEngine()

    blocks := s.GenerateChain(s.longest.GetHash())

    for i:=0;i<len(blocks);i++{
        D.UpdateBalance(blocks[i], 1)
    }
    fmt.Println("Check forward result: ", CompareBalance(D.GetBalance(), s.CalcBalance(s.longest.GetHash())))


    now := len(blocks)
    for i:=0;i<10;i++{
        aim := rand.Intn(len(blocks)+1)
        if now>aim{
            for j:=now-1;j>=aim;j--{
                D.UpdateBalance(blocks[j], -1)
            }
        }else{
            for j:=now;j<aim;j++{
                D.UpdateBalance(blocks[j], 1)
            }
        }
        //fmt.Println("block", now, "to", aim)
        now = aim

        var hash string
        if aim!=0{
            hash = blocks[aim-1].GetHash()
        }else{
            hash = InitHash
        }
        res := CompareBalance(D.GetBalance(), s.CalcBalance(hash))
        fmt.Println("Check for update", res)
        if !res{
            os.Exit(1)
        }
    }
}

func TestDatabase(){
    /*
        */
    HashHardness = 2
    fmt.Println("Test databse with HashHardness", HashHardness)
    s := MyServer{}

    //generate 2 blocks
    s.init(4, 1000, 1005)
    TestDatabaseEngine(&s)
}
