// Copyright 2017 Baliance. All rights reserved.

package main

import (
	"encoding/json"
	"fmt"
	"log"

	"baliance.com/gooxml/spreadsheet"
)

/*
 */
func main() {

	ss, err := spreadsheet.Open("ds.xlsx")
	if err != nil {
		log.Fatalf("error opening spreadsheet file: %s", err)
	}

	data := make(map[string]Item)
	for _, s := range ss.Sheets() {
		itemName := s.Name()
		var item Item
		item.Tabname = itemName
		data[itemName] = item
	}
	str, err := json.Marshal(data)
	if err != nil {
		panic(err.Error())
	}
	fmt.Println(string(str))

}
