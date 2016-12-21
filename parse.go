// parse
package main

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
)

type db struct {
	XMLName xml.Name `xml:"db"`
	Nm      string   `xml:"name"`
	Tbls    tables   `xml:"tables"`
}

type tables struct {
	XMLName xml.Name `xml:"tables"`
	Tb      []table  `xml:"table"`
}

type table struct {
	XMLName xml.Name `xml:"table"`
	Nm      string   `xml:"name"`
	Elms    elements `xml:"elements"`
}

type elements struct {
	XMLName xml.Name  `xml:"elements"`
	Elem    []element `xml:"element"`
}

type element struct {
	XMLName xml.Name `xml:"element"`
	Key     string   `xml:"key"`
	Val     string   `xml:"value"`
}

func parse_xml(name string) *db {
	data, err := ioutil.ReadFile(name)
	if err != nil {
		return nil
	}

	db := &db{}

	err = xml.Unmarshal(data, db)
	if err != nil {
		fmt.Print("Error")
	}
	return db
}

func main() {
	p := parse_xml("my.xml")
	fmt.Println(p)
}
