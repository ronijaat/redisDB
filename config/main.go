package config

var HOST string = "0.0.0.0"
var PORT int = 7379
var KeysLimit int = 5

var EvictionStrategy string = "simple-first"
var AOFFile string = "./redis-master.aof"
