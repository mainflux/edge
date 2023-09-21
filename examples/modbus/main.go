package main

import (
	"encoding/hex"
	"fmt"
	"log"
	"net/rpc"

	"github.com/mainflux/edge/modbus"
)

func main() {
	client, err := rpc.Dial("tcp", "localhost:8855")
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	var id int
	// configure
	config := modbus.TCPHandlerOptions{
		Address: "localhost:1502",
	}

	err = client.Call("Adapter.ConfigureTCP", config, &id)
	if err != nil {
		fmt.Println("Configure Error:", err)
	} else {
		fmt.Println("Configure Response:", id)
	}

	configRead := modbus.RWOptions{
		Address:   100,
		Quantity:  1,
		DataPoint: modbus.HoldingRegister,
		ID:        id,
	}

	data := make([]byte, 1)

	// Read
	err = client.Call("Adapter.Read", configRead, &data)
	if err != nil {
		fmt.Println("Read Error:", err)
	} else {
		fmt.Println("Read Data:", hex.EncodeToString(data))
	}

	configWrite := modbus.RWOptions{
		Address:   100,
		Quantity:  1,
		Value:     modbus.ValueWrapper{Data: uint16(1)},
		DataPoint: modbus.Register,
		ID:        id,
	}

	// Write
	err = client.Call("Adapter.Write", configWrite, &data)
	if err != nil {
		fmt.Println("Write Error:", err)
	} else {
		fmt.Println("Write Response:", hex.EncodeToString(data))
	}

	// Close

	var closed bool
	err = client.Call("Adapter.Close", id, &closed)
	if err != nil {
		fmt.Println("Close Error:", err)
	} else {
		fmt.Println("Close Response:", closed)
	}

}
