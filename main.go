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
	tables []table
}

func handleRequest(command string, c *cache, conn net.Conn) {
	fmt.Println("handle request: " + command)

	params := strings.Fields(command)
	// remove empty entries and remove whitespaces
	fmt.Println(params)
	fmt.Println(params[0])
	// remove empty entries and remove whitespaces
	switch strings.ToLower(params[0]) {
	case "exit":
		os.Exit(0)

		// TODO: patterns for queries

	default:
		conn.Write([]byte(string("Your command didn't match the pattern\n")))
	}
}

// Handles incoming requests.
func handleConnection(conn net.Conn, c *cache) {
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
		go handleRequest(s, c, conn)
	}
}

func (table *table) get_value(key string) string {
	for _, element := range table.Elements.Element {
		if element.Key == key {
			return element.Value
		}
	}
	return ""
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

// without saving order
func remove_from_slice(s []element, i int) []element {
	s[len(s)-1], s[i] = s[i], s[len(s)-1]
	return s[:len(s)-1]
}

func (table *table) del_key(key string, file_name string) bool {
	for i, element := range table.Elements.Element {
		if element.Key == key {
			table.Elements.Element = remove_from_slice(table.Elements.Element, i)
			table.save(file_name)
			return true
		}
	}
	return false
}

func (table *table) save(name string) {
	data, _ := xml.Marshal(table)
	ioutil.WriteFile(name, data, 0644)
}

func main() {
	// Listen for incoming connections.
	c := cache{}
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
		go handleConnection(conn, &c)
	}
}
