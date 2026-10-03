package main

var example = [][]string{
	{"GET", "key1"},
	{"SET", "key1", "value1"},
	{"DEL", "key1"},
	{"RENAME", "oldKey", "newKey"},
	{"POP", "key1"},
	{"EXISTS", "key1"},
	{"SCAN", "*"},
	{"SETNX", "key2", "value2"},
	{"SETXX", "key2", "value3"},
}
