package main

import (
    "fmt"
    "time"
    "math/rand"
)

func (s *MyServer)GetBlocksByBalance(database *DatabaseEngine, results chan *Block, stop chan int) {
    fmt.Println("GetBlocksByBalance")
    for ;;{
        select{
            case <-stop:
                fmt.Println("Stop GetBlocksByBalance")
                return
            case received := <- s.sender:
                results <- received
            case <- time.After(time.Second):
        }
    }
}

func TestMainLoop(){
    rand.Seed(time.Now().UnixNano())  

    fmt.Println("Test Main Loop")
    HashHardness = 2
    s := MyServer{}
    s.init(15, 1000, 153)
    //fmt.Println(s.CalcBalance(s.longest.GetHash()))

    miner := NewMiner(&s)
    go miner.mainLoop()

    prev := miner.longest
    for ;;{
        block := s.GenerateNewBlock(prev, 54, true)
        //fmt.Println(s.blocks[block.PrevHash].Transactions)
        fmt.Println("sender")
        s.sender <- block
        <- time.After(time.Second)

        res := CompareBalance(miner.GetBalance(), s.CalcBalance(block.GetHash()))
        if !res{
            fmt.Println("Error on calulate balance")
        } else{
            fmt.Println("Success")
        }
        break
    }
    s.longest := s.GenerateNewBlock(s.longest.GetHash(), 203, true)
}
