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
    //fmt.Println("before update", balance)

    totalFee := 0
    for _, t := range block.Transactions{
        balance[t.FromID] -= int(t.Value)
        if t.Value<0 || balance[t.FromID]<0 || t.MiningFee < 0 || t.Value < t.MiningFee{
            fmt.Println("Error on Update Balance(", block.BlockID, "), the transaction is")
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
    //fmt.Println(balance)
    //fmt.Println("   Length of block chain", len(blocks), blocks[0], blocks[0].GetHash(), hash)

    for i:=0; i<len(blocks);i++{
        //fmt.Println(i, blocks[i].GetHash())
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

func (s *MyServer) GenerateNewBlock(hash string, num_trans int, valid bool, startUUID int)*Block{
    /*
        Geerate NewBlocks after the blocks[hash]
    */
    balance := s.CalcBalance(hash)

    blocks := make([]*Block, 0)
    Transactions := make([]*pb.Transaction, 0)
    totalMiningFee := 0

    PrevHash := hash
    n := len(s.people_id)

    for i:=0;i<num_trans;i++{
        //genereate trans randomly
        //I can later try to generate in parallel with the miner
        a := s.people_id[rand.Intn(n)]
        b := s.people_id[rand.Intn(n)]

        value := rand.Intn(balance[a] + 1)
        mining_fee := rand.Intn(value + 1)
        totalMiningFee += mining_fee

        balance[a] -= value
        balance[b] += value - mining_fee

        Transactions = append(Transactions, GenTransaction(a, b, value, mining_fee, fmt.Sprintf("%08x", i + startUUID)).trans)

        if i==num_trans-1 || i%50 == 49{
            block := MakeNewBlock()
            block.Transactions = Transactions

            block.MinerID = s.people_id[rand.Intn(n)]
            balance[block.MinerID] += totalMiningFee

            block.PrevHash = PrevHash
            if PrevHash == InitHash{
                block.BlockID = 1
            }else{
                block.BlockID = s.blocks[PrevHash].BlockID + 1
            }
            block.Transactions = Transactions
            block.SolveSync()
            PrevHash = block.GetHash()
            s.blocks[PrevHash] = block

            blocks = append(blocks, block)
            Transactions = make([]*pb.Transaction, 0)
            totalMiningFee = 0
        }
    }
    s.balance = balance
    return blocks[len(blocks) - 1]
}

func (s *MyServer)init(n int, initilize int, num_trans int){
    s.balance = make(map[string]int)
    s.blocks = make(map[string]*Block)
    for i:=0;i<n;i++{
        s.people_id = append(s.people_id, fmt.Sprintf("%08x", i)) //%x or %d?
    }
    s.longest = s.GenerateNewBlock(InitHash, num_trans, true, 0)
    balance := s.CalcBalance(s.longest.GetHash())

    res := CompareBalance(balance, s.balance)
    fmt.Println("Check Update Balance int test_databse: ", res)
    if !res{
        os.Exit(1)
    }
}

func (s *MyServer)GetBlock(hash string)(*Block, bool){
    block, ok := s.blocks[hash]
    if !ok{
        return nil, ok
    }
    return block, ok
}

func (s *MyServer)GetHeight()(int, *Block, bool){
    return int(s.longest.BlockID), s.longest, true
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
    //TODO: check the code
}

func TestMiner(){
    /*
        First, let's test the sync code:
            maitaining the longest chain 

            fist we have a certain number of chain
            then add a longer chain, test the balance
        */
    HashHardness = 2
    s := MyServer{}
    fmt.Println("Test Miner")

    num_before := rand.Intn(1000) + 10
    num_later := rand.Intn(1000) + 10

    s.init(15, 1000, num_before)
    fmt.Println("Init Server")

    miner := NewMiner(&s)
    miner.Init()

    //ttt := s.GenerateChain(s.longest.GetHash())
    //fmt.Println(ttt[0])

    res := CompareBalance(miner.GetBalance(), s.CalcBalance(s.longest.GetHash()))
    if !res{
        fmt.Println("Miner init balance error!")
        os.Exit(1)
    }
    fmt.Println("Miner init balance correct!")

    blocks := s.GenerateChain(s.longest.GetHash())

    num_link := rand.Intn(len(blocks))

    //must break the tie to make s.longest
    t := s.GenerateNewBlock(blocks[num_link].GetHash(), num_later, true, num_before + 10)
    miner.InsertBlock(t)
    if t.BlockID>s.longest.BlockID{
        s.longest = t
    }


    res = CompareBalance(miner.GetBalance(), s.CalcBalance(s.longest.GetHash()))
    if !res{
        fmt.Println("  Miner balance error after switching a longer chain!")
        os.Exit(1)
    }
    fmt.Println("Miner balance correct after switching a longer chain!")
}

func TestDatabase(){
    /*
        */
    HashHardness = 2
    fmt.Println("Test databse with HashHardness", HashHardness)
    s := MyServer{}

    //generate 2 blocks
    s.init(15, 1000, 1004)
    TestDatabaseEngine(&s)
}
