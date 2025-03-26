/*
Package bcast provides functionality for broadcasting and receiving messages over a UDP network using a broadcast mechanism. It allows multiple channels to communicate by serializing and deserializing data into JSON format and handling transmission and reception in a structured way.

Key Features:
- Transmitter: Serializes data from multiple channels into JSON, attaches type metadata, and sends it over the network to a broadcast address.
- Receiver: Listens for incoming messages, deserializes them, and sends the data to the appropriate channels based on the type metadata.
- Type safety checks: Ensures all channels passed to the transmitter and receiver are of different element types and compatible with JSON serialization.
- Error handling: Provides checks and panics when the message size exceeds the buffer size or when unsupported types are encountered.

Functions:
- Transmitter: Takes a UDP port and one or more channels as arguments. Listens for updates on the channels, serializes the data, and broadcasts it to all receivers.
- Receiver: Takes a UDP port and one or more channels as arguments. Listens for incoming broadcast messages, deserializes them, and sends the data to the correct channels.
- checkArgs: Ensures the arguments passed to the Transmitter and Receiver are valid, ensuring all channels are of different types and compatible with JSON serialization.

Usage:
- To transmit data from channels, call the Transmitter function with the desired port and channels.
- To receive data, call the Receiver function with the desired port and channels.

Note:
This package uses UDP broadcast to send and receive messages, meaning it requires the network to support UDP broadcasts (e.g., local networks).

Credits: https://github.com/TTK4145/Network-go
*/
package bcast


import (
	"Driver-go/network/conn"
	"encoding/json"
	"fmt"
	"net"
	"reflect"
)

const bufSize = 2048

func Transmitter(port int, chans ...interface{}) {
	checkArgs(chans...)

	typeNames := make([]string, len(chans))

	selectCases := make([]reflect.SelectCase, len(typeNames))
	for i, ch := range chans {
		selectCases[i] = reflect.SelectCase{
			Dir:  reflect.SelectRecv,
			Chan: reflect.ValueOf(ch),
		}
		typeNames[i] = reflect.TypeOf(ch).Elem().String()
	}

	conn := conn.DialBroadcastUDP(port)
	addr, _ := net.ResolveUDPAddr("udp4", fmt.Sprintf("255.255.255.255:%d", port))
	for {
		chosen, value, _ := reflect.Select(selectCases)
		jsonstr, _ := json.Marshal(value.Interface())
		ttj, _ := json.Marshal(typeTaggedJSON{
			TypeId: typeNames[chosen],
			JSON:   jsonstr,
		})

		if len(ttj) > bufSize {
			panic(fmt.Sprintf(
				"Tried to send a message longer than the buffer size (length: %d, buffer size: %d)\n\t'%s'\n"+
					"Either send smaller packets, or go to network/bcast/bcast.go and increase the buffer size",
				len(ttj), bufSize, string(ttj)))
		}
		conn.WriteTo(ttj, addr)

	}
}

func Receiver(port int, chans ...interface{}) {
	checkArgs(chans...)

	chansMap := make(map[string]interface{})
	for _, ch := range chans {
		chansMap[reflect.TypeOf(ch).Elem().String()] = ch
	}

	var buf [bufSize]byte
	conn := conn.DialBroadcastUDP(port)
	for {
		n, _, e := conn.ReadFrom(buf[0:])
		if e != nil {
			fmt.Printf("bcast.Receiver(%d, ...):ReadFrom() failed: \"%+v\"\n", port, e)
		}

		var ttj typeTaggedJSON
		json.Unmarshal(buf[0:n], &ttj)
		ch, ok := chansMap[ttj.TypeId]
		if !ok {
			continue
		}
		v := reflect.New(reflect.TypeOf(ch).Elem())
		json.Unmarshal(ttj.JSON, v.Interface())
		reflect.Select([]reflect.SelectCase{{
			Dir:  reflect.SelectSend,
			Chan: reflect.ValueOf(ch),
			Send: reflect.Indirect(v),
		}})
	}
}

type typeTaggedJSON struct {
	TypeId string
	JSON   []byte
}

func checkArgs(chans ...interface{}) {
	n := 0
	for range chans {
		n++
	}
	elemTypes := make([]reflect.Type, n)

	for i, ch := range chans {
		if reflect.ValueOf(ch).Kind() != reflect.Chan {
			panic(fmt.Sprintf(
				"Argument must be a channel, got '%s' instead (arg# %d)",
				reflect.TypeOf(ch).String(), i+1))
		}

		elemType := reflect.TypeOf(ch).Elem()

		for j, e := range elemTypes {
			if e == elemType {
				panic(fmt.Sprintf(
					"All channels must have mutually different element types, arg# %d and arg# %d both have element type '%s'",
					j+1, i+1, e.String()))
			}
		}
		elemTypes[i] = elemType

		checkTypeRecursive(elemType, []int{i + 1})

	}
}

func checkTypeRecursive(val reflect.Type, offsets []int) {
	switch val.Kind() {
	case reflect.Complex64, reflect.Complex128, reflect.Chan, reflect.Func, reflect.UnsafePointer:
		panic(fmt.Sprintf(
			"Channel element type must be supported by JSON, got '%s' instead (nested arg# %v)",
			val.String(), offsets))
	case reflect.Map:
		if val.Key().Kind() != reflect.String {
			panic(fmt.Sprintf(
				"Channel element type must be supported by JSON, got '%s' instead (map keys must be 'string') (nested arg# %v)",
				val.String(), offsets))
		}
		checkTypeRecursive(val.Elem(), offsets)
	case reflect.Array, reflect.Ptr, reflect.Slice:
		checkTypeRecursive(val.Elem(), offsets)
	case reflect.Struct:
		for idx := 0; idx < val.NumField(); idx++ {
			checkTypeRecursive(val.Field(idx).Type, append(offsets, idx+1))
		}
	}
}
