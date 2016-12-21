package main

import (
	"bufio"
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"net"
	"os"
	"regexp"
	"strings"
	"sync"
)

const (
	CONN_HOST = "localhost"
	CONN_PORT = "8888"
	CONN_TYPE = "tcp"
)

type table struct {
	XMLName  xml.Name `xml:"table"`
	Name     string   `xml:"name"`
	Elements elements `xml:"elements"`
}

type elements struct {
	XMLName xml.Name  `xml:"elements"`
	Element []element `xml:"element"`
}

type element struct {
	XMLName xml.Name `xml:"element"`
	Key     string   `xml:"key"`
	Value   string   `xml:"value"`
}

func parse_xml(name string) *table {
	data, err := ioutil.ReadFile(name)
	if err != nil {
		return nil
	}

	table := &table{}

	err = xml.Unmarshal(data, table)
	if err != nil {
		fmt.Print("Error")
	}
	return table
}

type cache struct {
	elements []element
}

func handleRequest(command string, c *cache, table_chan chan<- table, conn net.Conn) {
	fmt.Println("handle request: " + command)

	params := strings.Fields(command)
	// remove empty entries and remove whitespaces
	fmt.Println(params)
	fmt.Println(params[0])
	// remove empty entries and remove whitespaces
	switch strings.ToLower(params[0]) {
	case "exit":
		os.Exit(0)
	case: ""

		// TODO: patterns for queries
	if len(params) >= 2 {
		switch strings.ToLower(query_split[1]) {
			case "set":
				setVal(c, table_chan, tables, query_split)
			case "get":
				getVal(c, tables, query_split)
			case "delete":
				delKey(c, table_chan, tables, query_split)
			case "keys":
				getKeys(c, tables, query_split)
			default:
				c.Write([]byte(string("Unknown command\n")))
		}
	} else if len(query_split) == 1 {
		switch strings.ToLower(query_split[0]) {
			case "exit":
				exit(c)
			case "help":
				help(c)
			default:
				c.Write([]byte(string("Unknown command\n")))
			}
	} else {
		c.Write([]byte(string("Unknown command\n")))
	}
}

// Handles incoming requests.
func handleConnection(conn net.Conn, table_chan chan<- table, c *cache) {
	// Make a buffer to hold incoming data.
	var buf [512]byte
	empty_string := ""
	for {
		copy(buf[:], empty_string) //make  buffer empty
		_, err := conn.Read(buf[0:])
		if err != nil {
			return
		}
		n := bytes.Index(buf[:], []byte{0})
		s := string(buf[:n])
		go handleRequest(s, c, table_chan, conn)
	}
}

func (table *table) get_value(key string) string {
	data := ""
	for _, element := range table.Elements.Element {
		if element.Key == key {
			data = element.Value
		}
	}
	return data
}

func (table *table) set_value(key string, val string, name string) {
	for _, element := range table.Elements.Element {
		if element.Key == key {
			element.Value = val
			table.save(name)
			return
		}
	}
	element := element{XMLName: table.Elements.Element[0].XMLName, Key: key, Value: val}
	table.Elements.Element = append(table.Elements.Element, element)

	table.save(name)
}

func (table *table) save(name string) {
	data, _ := xml.Marshal(table)
	ioutil.WriteFile(name, data, 0644)
}

func main() {
	// Listen for incoming connections.
	c := cache{}
	table_chan := make(chan table)
	
	l, err := net.Listen(CONN_TYPE, CONN_HOST+":"+CONN_PORT)
	if err != nil {
		fmt.Println("Error listening:", err.Error())
		os.Exit(1)
	}
	// Close the listener when the application closes.
	defer l.Close()
	fmt.Println("Listening on " + CONN_HOST + ":" + CONN_PORT)
	for {
		// Listen for an incoming connection.
		conn, err := l.Accept()
		fmt.Println("accept")

		if err != nil {
			fmt.Println("Error accepting: ", err.Error())
			os.Exit(1)
		}
		// Handle connections in a new goroutine.
		go handleConnection(conn, table_chan, &c)
	}
}
