package main

import "testing"
import "net"
import "fmt"

import "bufio"
import "time"
import "strings"

func Test(t *testing.T) {
	//	conn, _ := net.Dial("tcp", "127.0.0.1:8888")
	//	c := cache{}
	//	text := "test.xml get key1"
	//	msg := get_value(conn, &c, strings.Fields(text))
	//	fmt.Fprintf(conn, msg)

	go main()
	time.Sleep(100 * time.Millisecond)
	fmt.Println("test starting")
	conn, _ := net.Dial("tcp", "127.0.0.1:8888")
	fmt.Println("send set")
	// read in input from stdin
	text := "test.xml set key1 22\n"
	message, _ := bufio.NewReader(conn).ReadString('\n')
	buf := make([]byte, 4096)
	conn.Read(buf)
	fmt.Println(conn, buf)
	// send to socket
	fmt.Fprintf(conn, text)
	fmt.Println("send select")
	time.Sleep(100 * time.Millisecond)

	text = "test.xml get key1\n"
	// send to socket
	fmt.Fprintf(conn, text)
	fmt.Println("listen")
	time.Sleep(100 * time.Millisecond)

	fmt.Print("Message from server:" + message + "end")
	message = strings.Replace(strings.Replace(message, " ", "", -1), "\n", "", -1)
	fmt.Print("Message from server:" + message + "end")

	if message != "22" {
		t.Error("get return wrong string")

	}
	time.Sleep(100 * time.Millisecond)

	fmt.Fprintf(conn, "test.xml set key2 333\n")
	time.Sleep(100 * time.Millisecond)
	fmt.Fprintf(conn, "test.xml get key2\n")

	message, _ = bufio.NewReader(conn).ReadString('\n')
	message = strings.Replace(strings.Replace(message, " ", "", -1), "\n", "", -1)
	if message != "333" {
		t.Error("update works wrong")

	}
	time.Sleep(100 * time.Millisecond)

	fmt.Fprintf(conn, "test.xml delete key2\n")
	time.Sleep(100 * time.Millisecond)
	fmt.Fprintf(conn, "test.xml get key2\n")

	message, _ = bufio.NewReader(conn).ReadString('\n')
	message = strings.Replace(strings.Replace(message, " ", "", -1), "\n", "", -1)
	if message != "" {
		t.Error("delete works wrong")

	}

	fmt.Fprintf(conn, "exit 1")

	fmt.Println("OK")
}
