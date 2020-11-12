// Copyright 2017 Baliance. All rights reserved.

package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	//"strings"

	"baliance.com/gooxml/color"
	"baliance.com/gooxml/document"
	"baliance.com/gooxml/measurement"
	"baliance.com/gooxml/schema/soo/wml"
	"baliance.com/gooxml/spreadsheet"
)

func main() {
	doc := document.New()

	ss, err := spreadsheet.Open("ds.xlsx")
	if err != nil {
		log.Fatalf("error opening spreadsheet file: %s", err)
	}
	fc, _ := ioutil.ReadFile("d.json")
	//fmt.Println(string(fc))
	var dict map[string]Item
	json.Unmarshal(fc, &dict)

	para := doc.AddParagraph()
	for _, s := range ss.Sheets() {
		itemName := s.Name()
		i := []rune(itemName)
		fmt.Println(string(i[:len(i)-1]))
		item := dict[itemName]
		serviceName := item.EnglishName + "Api?wsdl"
		//lastIndex := len(itemName) - 2
		//itemName = itemName[:lastIndex]

		para = doc.AddParagraph()
		para.AddRun().AddText("地方追溯平台" + itemName + "WS接口")

		para = doc.AddParagraph()
		para.AddRun().AddText(itemName + "WS接口")

		para = doc.AddParagraph()
		para.AddRun().AddText("服务调用地址")

		para = doc.AddParagraph()
		para.AddRun().AddText("http://ip:port/sofn-dgap-pre/ws/" + serviceName)

		para = doc.AddParagraph()
		para.AddRun().AddText("接口描述")

		para = doc.AddParagraph()
		para.AddRun().AddText("增加" + itemName)

		para = doc.AddParagraph()
		para.AddRun().AddText("boolean add" + item.EnglishName + "(String token, String id, " + item.EnglishName + " subject);")
		para = doc.AddParagraph()
		para.AddRun().AddText("修改" + itemName)
		para = doc.AddParagraph()
		para.AddRun().AddText("boolean update" + item.EnglishName + "(String token, String id, " + item.EnglishName + " subject);")
		para = doc.AddParagraph()
		para.AddRun().AddText("删除" + itemName)
		para = doc.AddParagraph()
		para.AddRun().AddText("boolean delete" + item.EnglishName + "(String token, String id);")

		para = doc.AddParagraph()
		para.AddRun().AddText(itemName + "定义")
		table := doc.AddTable()
		table.Properties().SetWidthPercent(90)
		table.Properties().SetAlignment(wml.ST_JcTableCenter)
		borders := table.Properties().Borders()
		borders.SetAll(wml.ST_BorderSingle, color.Auto, 1*measurement.Point)
		xml := ""
		for _, r := range s.Rows() {
			row := table.AddRow()
			for _, c := range r.Cells() {
				cell := row.AddCell()
				cell.Properties().SetWidthPercent(20)
				cell.AddParagraph().AddRun().AddText(c.GetString())
			}
		}

		para = doc.AddParagraph()
		para.AddRun().AddText(itemName + "XML描述格式")
		para = doc.AddParagraph()
		para.AddRun().AddText("<" + item.EnglishName + ">")
		for ri, r := range s.Rows() {
			for ci, c := range r.Cells() {
				if ri > 0 && ci == 1 {
					xml = "<" + c.GetString() + ">" + "</" + c.GetString() + ">"
					para = doc.AddParagraph()
					para.AddRun().AddText(xml)
				}
			}
		}
		para = doc.AddParagraph()
		para.AddRun().AddText("</" + item.EnglishName + ">")
	}
	doc.SaveToFile("o.docx")
}
