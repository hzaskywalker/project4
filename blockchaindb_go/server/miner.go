package main

import {
    "sync"
    "errors"
    "database"
}

type Miner struct{
    hash2block  *map[string]*Block

    //store the current branches
    block2Height *map[string]int
    database *DatabaseEngine
}

func newMiner()Miner{
    return Miner{
        hash2block: make(map[string]*Block),
        block2Height: make(map[string]int),
        database: NewDatabaseEngin(),
    }
}
