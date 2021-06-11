o.docx: main.go item.go d.json
	go run $< item.go

dict.json: dict.go item.go ds.xlsx
	go run $< item.go 2>/dev/null >dict.json

#d.json: dict.json
#	cat $< | jq '.' >$@

test:
	go test -v -run Test1 main_test.go main.go item.go


clean:
	rm -rf dict.json o.docx
