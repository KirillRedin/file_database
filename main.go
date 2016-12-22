package main

import (
	//	"bufio"
	//	"bytes"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	//	"io"
	"net"
	"os"
	//	"regexp"
	"strings"
	//	"sync"
)

const (
	CONN_HOST = "localhost"
	CONN_PORT = "8888"
	CONN_TYPE = "tcp"
)

type table struct {
	XMLName  xml.Name  `xml:"table"`
	Name     string    `xml:"name"`
	Elements []element `xml:"element"`
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
	// TODO: patterns for queries

	if len(params) >= 2 {
		switch strings.ToLower(params[1]) {
		case "set":
			set_value(conn, c, params)
		case "get":
			get_value(conn, c, params)
		case "delete":
			del_key(conn, c, params)
		case "keys":
			get_keys(conn, c, params)
		default:
			conn.Write([]byte(string("Unknown command\n")))
		}
	} else if len(params) == 1 {
		switch strings.ToLower(params[0]) {
		case "exit":
			exit(conn)
		default:
			conn.Write([]byte(string("Unknown command\n")))
		}
	} else {
		conn.Write([]byte(string("Unknown command\n")))
	}
}

// Handles incoming requests.
func handleConnection(conn net.Conn, c *cache) {
	// Make a buffer to hold incoming data.
	buf := make([]byte, 4096)
	for {
		n, err := conn.Read(buf)
		if (err != nil) || (n == 0) {
			break
		} else {
			go handleRequest(string(buf[0:n]), c, conn)
		}
	}
}

func get_table(c *cache, params []string) *table {

	for _, table := range c.tables {
		if table.Name == params[0] {
			for _, element := range table.Elements {
				if params[1] != "keys" && element.Key == params[2] {
					return &table
				}
			}
		}
	}

	table := parse_xml(params[0])

	if table != nil {
		c.tables = append(c.tables, *table)
	}
	return table
}

func get_keys(conn net.Conn, c *cache, params []string) {
	if len(params) == 2 {
		table := get_table(c, params)
		if table == nil {
			conn.Write([]byte(string("Unknown table\n")))
		} else {
			keys := make([]string, 0, len(table.Elements))
			for _, element := range table.Elements {
				keys = append(keys, element.Key)
			}
			conn.Write([]byte("[" + strings.Join(keys, ", ") + "]" + "\n"))
		}
	} else {
		conn.Write([]byte(string("Unknown command\n")))
	}
}

func get_value(conn net.Conn, c *cache, params []string) {
	if len(params) == 3 {
		table := get_table(c, params)
		if table == nil {
			conn.Write([]byte(string("Unknown table\n")))
		} else {
			for _, element := range table.Elements {
				if element.Key == params[2] {
					conn.Write([]byte(string(element.Value + "\n")))
					return
				}
			}
			conn.Write([]byte(string("key does not exist\n")))
		}
	} else {
		conn.Write([]byte(string("Unknown command\n")))
	}
}

func set_value(conn net.Conn, c *cache, params []string) {

	if len(params) >= 4 {
		table := get_table(c, params)
		if table == nil {
			conn.Write([]byte(string("Unknown table\n")))
		}

		for _, element := range table.Elements {
			if element.Key == params[2] {

				element.Value = strings.Join(params[3:], " ")
				table.save(params[0])
				update_cache(c, params[0])
				conn.Write([]byte(string("OK\n")))
				return
			}
		}
		element := element{XMLName: xml.Name{Local: "element"}, Key: params[2], Value: strings.Join(params[3:], " ")}
		table.Elements = append(table.Elements, element)
		table.save(params[0])
		update_cache(c, params[0])
		conn.Write([]byte(string("OK\n")))
		return
	} else {
		conn.Write([]byte(string("Unknown command\n")))
		return
	}
}

// without saving order
func remove_from_slice(s []element, i int) []element {
	s[len(s)-1], s[i] = s[i], s[len(s)-1]
	return s[:len(s)-1]
}

func update_cache(c *cache, name string) {

	for i, table := range c.tables {
		if table.Name == name {
			//			c.tables = append(c.tables[:i], c.tables[(i+1):])
			if len(c.tables) != i {
				c.tables = c.tables[:i+copy(c.tables[i:], c.tables[i+1:])]
			} else {
				c.tables = c.tables[:len(c.tables)-1]
			}
		}
	}
}

func del_key(conn net.Conn, c *cache, params []string) {

	if len(params) == 3 {
		table := get_table(c, params)
		if table == nil {
			conn.Write([]byte(string("Unknown table\n")))
		} else {

			for i, element := range table.Elements {
				if element.Key == params[2] {
					table.Elements = remove_from_slice(table.Elements, i)
					table.save(params[0])
					update_cache(c, params[0])
					conn.Write([]byte(string("OK\n")))
					return
				}
			}
			conn.Write([]byte(string("key does not exist\n")))
		}
	} else {
		conn.Write([]byte(string("Unknown command\n")))
	}
}

func (table *table) save(name string) {
	data, _ := xml.Marshal(table)
	ioutil.WriteFile(name, data, 0644)
}

func exit(conn net.Conn) {
	conn.Write([]byte(string("Bye\n")))
	conn.Close()
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
