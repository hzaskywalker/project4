package main

import {
    "sync"
    "errors"
}

type DatabaseEngine struct {
    balance map[string]int
    sync.RWMutex
}

func NewDatabaseEngin(_config *DatabaseConfig, _logger *Logger){
    return DatabaseConfig{balance:make(map[string]int), config:_config, logger:_logger}
}

func checkKey(userId string){
    return len(string) == 8
}

func (db *DatabaseEngine)Transfer(from string, to string, delta int)(int, int, error){
    if delta < 0{
        a, b, c := Transfer(to, from, -delta)
        return b, a, c
    }
    db.Lock()
    db.Unlock()
    from_val, from_ok := db.balance[from]
    to_val, to_ok := db.balance[to]
    if !from_ok {
        from_val = 0
    }
    if !to_ok {
        to_val = 0
    }
    if from_val < delta{
        return from_val, to_val, errors.New("Transfer: Not enough money")
    }
    from_val = from_val - delta
    to_val = to_val + delta
    db.balance[from] = from_val
    db.balance[to] = to_val
    return from_val, to_val, nil
}