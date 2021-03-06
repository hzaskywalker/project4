package main

import (
    "fmt"
    "time"
    //"math/rand"
    //"../hash"
    "os"
    pb "../protobuf/go"
)

func (s *MyServer)GetBlocksByBalance(database *DatabaseEngine, results chan *Block, stop chan int) {
    fmt.Println("GetBlocksByBalance")
    for ;;{
        select{
            case <-stop:
                //fmt.Println("Stop GetBlocksByBalance")
                return
            case received := <- s.sender:
                results <- received
            case <- time.After(time.Second):
        }
    }
}

func TestMainLoop(){
    //rand.Seed(time.Now().UnixNano())  

    fmt.Println("Test Main Loop")
    HashHardness = 2
    waitTime := time.Duration(1)
    s := MyServer{}
    s.init(15, 1000, 153)
    //fmt.Println(s.CalcBalance(s.longest.GetHash()))

    miner := NewMiner(&s)

    service := NewService()
    go miner.mainLoop(service)

    for ;miner.longest == nil;{
        <- time.After(time.Second)
    }
    prev := miner.longest

    hash1 := GetHashString(prev.MarshalToString())
    tmp := MakeNewBlock()
    tmp.Unmarshal(prev.MarshalToString())
    hash2 := GetHashString(tmp.MarshalToString())
    if hash1 != hash2{
        fmt.Println(hash1)
        fmt.Println(hash2)
        fmt.Println("hash unequal")
        os.Exit(1)
    }
    for ;;{
        block := s.GenerateNewBlock(prev.GetHash(), 30, true)
        //fmt.Println(s.blocks[block.PrevHash].Transactions)
        fmt.Println("sender")
        s.sender <- block
        <- time.After(time.Second * waitTime)

        res := CompareBalance(miner.GetBalance(), s.CalcBalance(block.GetHash()))
        if !res{
            fmt.Println("Error on calulate balance")
        } else{
            fmt.Println("Success")
        }
        break
    }
    tmp = s.GenerateNewBlock(s.longest.GetHash(), 60, true)
    s.sender <- tmp

    s.longest = s.GenerateNewBlock(s.longest.GetHash(), 203, true)
    service.PushBlock(&pb.JsonBlockString{Json:s.longest.MarshalToString()})
    <- time.After(time.Second * waitTime)

    res := CompareBalance(miner.GetBalance(), s.CalcBalance( s.longest.GetHash()) )
    if !res{
        fmt.Println("Wrong balance after push block")
        os.Exit(1)
    } else{
        fmt.Println("Success Push Block")
    }
    for ;;{
        //golang has trouble for infinite loop
        service.Hello <- true
        <- time.After(time.Second * 5)
    }
}
