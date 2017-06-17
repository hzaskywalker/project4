package main

import (
    "os"
    "io/ioutil"
)

func WriteJson(hash, Json string){
    w, e := os.Create(hash + ".json")
    defer w.Close()
    if e == nil{
        return
    }
    _, _ = w.Write([]byte(Json))
}

func WriteBlock(block *Block){
    WriteJson(block.GetHash(), block.MarshalToString())
}

func ReadFromDisck(hash string)(*Block, bool){
    str, e:= ioutil.ReadFile(hash + ".json")
    if e!=nil{
        return nil, false
    }
    block := MakeNewBlock()
    e = block.Unmarshal(string(str))
    if e!=nil{
        return nil, false
    }
    return block, true
}

/*
import (
    "os"
    "bufio"
    "fmt"
)

func WriteBalance(balance *map[string]int, hash string, stop chan bool){
    filename := fmt.Sprintf("balance.data.%s", hash)
    w := bufio.NewWriter(filename)
    toWrite := chan string
    go func(){
        toWrite <- hash
        for key, val := range balance{
            toWrite <- key
            toWrite <- fmt.Sprintf("%d", val)
        }
        close(toWrite)
    }
    for ;;{
        select {
            case s := <- toWrite:
                w.WriteLine(s)
            case <-stop:
                return
        }
    }
    w.Flush()
    w.close()
    e := os.Symlink(filename, "balance.data.bk")
    <- stop
}

func ReadBalance(database *DatabaseEngine, stop chan bool){
    //only be called at the begining
    //it must be faster than remote call
    //so I can read it from the begining to the last
    filename := "balance.data.bk"
    r := bufio.NewReader(filename)

}
*/