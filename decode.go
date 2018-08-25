package govid

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
)

type TRecord struct {
	Field []TField
	Index int
}
func (r TRecord) Print() {
	fmt.Printf("%d: %d fields\n", r.Index, len(r.Field))
	for _, fi := range r.Field {
		fmt.Printf("  code: %s, name: %s\n", fi.Code, fi.Name)
		for _, ct := range fi.Content {
			fmt.Printf("    - %s\n", ct)
		}
	}
}

type TField struct {
	Content []string
	Code string
	Name string
}
func (f *TField) add(t string) { f.Content = append(f.Content, t) }
func (f *TField) addAll(data []string) { f.Content = append(f.Content, data...) }

func ParseXML(reader io.Reader) ([]TRecord, error) {
	xmlDecoder := xml.NewDecoder(reader)
	var results []TRecord
	for {
		tok, err := xmlDecoder.Token()

		if err == io.EOF { break }
		if err != nil { return nil, err }
		if se, ok := tok.(xml.StartElement); ok && se.Name.Local == "record" {
			rec := readRecord(xmlDecoder)
			if idx, err := getRecordIndex(se); err == nil {
				rec.Index = idx
			}
			results = append(results, rec)
		}
	}
	return results, nil
}

func getAttrValue(se xml.StartElement, name string) (string, error) {
	for _, at := range se.Attr {
		if at.Name.Local != name { continue }
		return at.Value, nil
	}
	return "", errors.New("not found")
}

func getRecordIndex(se xml.StartElement) (int, error) {
	val, err := getAttrValue(se, "index")
	if err != nil { return -1, err }

	return strconv.Atoi(strings.Trim(val, " ."))
}

func readData(decoder *xml.Decoder) []string {
	var data []string

	for {
		tok, _ := decoder.Token()

		if se, ok := tok.(xml.StartElement); ok && se.Name.Local == "T" {
			txt := readText(decoder)
			if len(txt) > 0 { data = append(data, txt) }
		}
		if cd, ok := tok.(xml.CharData); ok { data = append(data, string(cd)) }
		if ee, ok := tok.(xml.EndElement); ok && ee.Name.Local == "D" { break /* Data is done */ }
	}

	return data
}

func readField(decoder *xml.Decoder) TField {
	var field TField

	for {
		tok, _ := decoder.Token()

		if se, ok := tok.(xml.StartElement); ok {
			if se.Name.Local == "T" {
				txt := readText(decoder)
				if len(txt) > 0 { field.add(txt) }
			}
			if se.Name.Local == "D" {
				data := readData(decoder)
				if len(data) > 0 { field.addAll(data) }
			}
		}

		if ee, ok := tok.(xml.EndElement); ok && ee.Name.Local == "F" { break /* Field is done */ }
	}

	return field
}

func readRecord(decoder *xml.Decoder) TRecord {
	var record TRecord

	for {
		tok, _ := decoder.Token()

		if se, ok := tok.(xml.StartElement); ok && se.Name.Local == "F" {
			field := readField(decoder)
			if code, err := getAttrValue(se, "C"); err == nil { field.Code = code }
			if name, err := getAttrValue(se, "L"); err == nil { field.Name = name }
			record.Field = append(record.Field, field)
		}

		if ee, ok := tok.(xml.EndElement); ok && ee.Name.Local == "record" { break /* Record is Done */ }
	}

	return record
}

func readText(decoder *xml.Decoder) string {
	var buffer bytes.Buffer

	for {
		tok, _ := decoder.Token()

		if ee, ok := tok.(xml.EndElement); ok {
			if ee.Name.Local == "BR" { buffer.WriteString("\n") }
			if ee.Name.Local == "T" { break /* End of Text */ }
		}

		if _, ok := tok.(xml.StartElement); ok { /* Ignored start of BR */ }

		if cd, ok := tok.(xml.CharData); ok { buffer.WriteString(string(cd)) }
	}

	return string(bytes.Trim(buffer.Bytes(), " \n.();,")) /* Removing Punctuation */
}