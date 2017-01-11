package main

import (
	"encoding/json"
	"issue18676/dynamodbstreamsevt"
	"plugin"
)

func main() {
	plugin.Open("plugin.so")

	var x interface{} = (*dynamodbstreamsevt.Event)(nil)
	if _, ok := x.(json.Unmarshaler); !ok {
		println("something")
	}
}
