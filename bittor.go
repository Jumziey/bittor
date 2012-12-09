package bittor

import(
	"log"
	"strconv"
)

type TorData struct {
	Data []byte
	pos int
}

func (t *Data)next() byte {
	b := t.data[t.pos]
	t.pos = t.pos+1
	return b
}

func (t *Data)peek() byte {
	return t.data[t.pos]
}

func (t *Data)prev() {
	t.pos = t.pos-1
}

func intParse(t *Data) int {
	intStr := ""
	var b byte
	for b= t.next(); b != 'e'; b = t.next() {
		intStr = intStr+string(b)
	}
	integ, err := strconv.Atoi(intStr)
	if err != nil {
		log.Fatalln("Error in intParse: ", err)
	}
	return integ
}

func stringParse(t *Data) string {
	t.prev()
	
	stringSize := ""
	for s:=t.next(); s != ':'; s=t.next() {
		stringSize = stringSize+string(s)
	}
	s_size, err := strconv.Atoi(stringSize)
	if err != nil {
		log.Fatalln("Error in stringParse: ", err)
	}
	
	bstring := make([]byte, s_size)
	for i:=0; i<s_size; i++ {
		bstring[i] = t.next()
	}
	return string(bstring)
}

func listParse(t *Data) []interface{} {
	var itemSlice []interface{}

	//We read until we reach the end 'e' of the list and make this
	//a list item. we peek so we don't fuck it up for nextItem(*Data)
	for t.peek() != 'e' {
		s := nextItem(t)
		itemSlice = append(itemSlice, s)
	}
	t.next() //Throw away the 'e'
	
	return itemSlice
}

func dictParse(t *Data) map[string]interface{} {
	dictMap := make(map[string]interface{})
	
	//We read until we reach the end 'e' of the dictionary and make this
	//a dictionary item. We peek so we don't fuck it up for nextItem(*Data).
	//We must be able to read two items at a time, otherwise the torrent is faulty
	//formatted
	for t.peek() != 'e' {
		key := nextItem(t)
		value := nextItem(t)
		dictMap[key.(string)] = value
	}
	t.next() //Throw away the 'e'
	
	return dictMap
}

func nextItem(t *Data) interface{} {
	switch t.next() {
		case 'd':
			return dictParse(t)
		case 'l':
			return listParse(t)
		case 'i':
			return intParse(t)
		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
			return stringParse(t)
		default:
			log.Fatalln("Out of bonds in nextItem")
	}
	//Unreachable, but needed due to weird controls in go-compiler
	return nil
}

func GetMainDict(t *Data) map[string]interface{} {
	
	infoDict := nextItem(t)
	if len(t.data) > t.pos {
		log.Fatalln("Torrent isn't bencoded correctly(Has more then an info dict")
	}
	return infoDict.(map[string]interface{})
}

func GetInfoDict(m map[string]interface{}) map[string]interface{}{
	for k,v := range m {
		if k == "info" {
			return v.(map[string]interface{})
		}
	}
	return nil
}
