package main

/*
A miner should maintain:

The tree of the blocks, the node of longest chain, and the corresponding balance.

We can:
    1. return the balance of any block:
        find the lca to the longest chain and return the corresponding balance
    2. get the longest block

Modification to the status:
    1. According to getHeight:
        1. add blocks or change the branch
    2. Collect transactions and calculate new blocks starting from the longest chain

Some notes:
    1. We use a simple map from string to block so that we can cache all of the blocks that we have.
       Once a block is missing, we use get block to retrive the corresponding blocks.
    2. The most of the computation should be used to solve the hash problem, so we needn't worry too
       much about the other computations.
   */

import {
    "sync"
    "errors"
    "database"
}

type Miner struct{
    hash2block  *map[string]*Block

    //store the current branches string with their height
    block2Height *map[string]int

    //store the longest chains
    longest *block
    longest_hash *block

    database *DatabaseEngine
}

func newMiner()Miner{
    return Miner{
        hash2block: make(map[string]*Block),
        block2Height: make(map[string]int),
        database: NewDatabaseEngin(),
    }
}

func (m *Miner) mainLoop(){
}