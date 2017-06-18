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
    "time"
    pb "../protobuf/go"
)

type MyServer struct{
    //brute force code for balance calculating
    FakeServer //in test_server
    startUUID int

    blocks map[string]*Block
    longest *Block
    sender chan *Block

    TransferSender chan *Transaction
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

func (s *MyServer) SendTransfer(t *Transaction){
    s.TransferSender <- t
}

func (s *MyServer) TRANSFER()*Transaction{
    return <- s.TransferSender
}

func (s *MyServer) GenerateNewBlock(hash string, num_trans int, valid bool)*Block{
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

        toGen := balance[a] + 1
        if toGen <= 0{
            toGen = 1
        }
        value := rand.Intn(toGen)
        if value>balance[a]{
            value = 0
        }

        if !valid && i==3{
            value = balance[a] + 1
        }

        mining_fee := rand.Intn(value + 1)
        totalMiningFee += mining_fee

        balance[a] -= value
        balance[b] += value - mining_fee

        s.startUUID += 1
        t := GenTransaction(a, b, value, mining_fee, fmt.Sprintf("%08x", s.startUUID))
        go s.SendTransfer(t)

        Transactions = append(Transactions, t.trans)

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
    s.sender = make(chan *Block)
    s.balance = make(map[string]int)
    s.blocks = make(map[string]*Block)
    s.TransferSender = make(chan *Transaction)
    for i:=0;i<n;i++{
        s.people_id = append(s.people_id, fmt.Sprintf("%08x", i)) //%x or %d?
    }
    s.longest = s.GenerateNewBlock(InitHash, num_trans, true)
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

func (s *MyServer)PushBlock(block *Block){
	//empty?
}

func TestDatabaseEngine(s *MyServer){
    D := NewDatabaseEngine(nil)

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
    //TODO: check the code for incorrect block
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
    rand.Seed(time.Now().UnixNano())  

    num_before := rand.Intn(1000) + 10
    num_later := rand.Intn(1000) + 10

    s.init(15, 1000, num_before)
    fmt.Println("Init Server")
    fmt.Println("longest", s.longest.BlockID, s.longest.GetHash())

    miner := NewMiner(&s)
    miner.Init()

    pending_size_init := miner.GetTransferManager().GetPendingSize()
    fmt.Println("dict size", miner.GetTransferManager().GetDictSize())
    fmt.Println("pending_size after init ", pending_size_init)
    if pending_size_init>0{
        os.Exit(1)
    }

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
    t := s.GenerateNewBlock(blocks[num_link].GetHash(), num_later, true)

    miner.InsertBlock(t)
    miner.VerifyBlock(t)
    if t.BlockID>s.longest.BlockID{
        s.longest = t
        fmt.Println("switch longest")
    }
    miner.UpdateLongest(t)


    res = CompareBalance(miner.GetBalance(), s.CalcBalance(s.longest.GetHash()))
    if !res{
        fmt.Println("  Miner balance error after switching a longer chain!")
        os.Exit(1)
    }
    fmt.Println("Miner balance correct after switching a longer chain!")

    //TODO: Check for validation
    dict_size := miner.GetTransferManager().GetDictSize()
    pending_size := miner.GetTransferManager().GetPendingSize()
    longest_trans_num := 0
    blocks = s.GenerateChain(s.longest.GetHash())
    for _, b:=range(blocks){
        longest_trans_num += len(b.Transactions)
    }
    pending_gt_num := num_before + num_later - longest_trans_num
    fmt.Println(longest_trans_num, pending_gt_num, num_before + num_later)


    fmt.Println("transfer information", dict_size, pending_size)
    if pending_size != pending_gt_num{
        fmt.Println("pending num is wrong")
        os.Exit(0)
    }

    //invalid block
    invalid := s.GenerateNewBlock(s.longest.GetHash(), 104, false)
    invalid.PrevHash = "xxxx"
    invalid.MyHash = ""
    e := miner.InsertBlock(invalid)
    if e==nil{
        fmt.Println("doesn't detect no father error!")
    }
    invalid = s.GenerateNewBlock(s.longest.GetHash(), 104, false)
    e = miner.InsertBlock(invalid)
    if e != nil{
        fmt.Println("Error on Insert Invalid Block")
    }
    e = miner.VerifyBlock(invalid) 
    if e == nil{
        fmt.Println("Doesn't find the error of Invalid Block")
    } else{
        fmt.Println("Correct detect error ", e)
    }
    res = CompareBalance(miner.GetBalance(), s.CalcBalance(s.longest.GetHash()))
    if !res{
        fmt.Println("  Miner balance error after adding invalid block a longer chain!")
        os.Exit(1)
    }
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
