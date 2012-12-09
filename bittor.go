package bittor

import(
	"strconv"
	"errors"
	"fmt"
	"io/ioutil"
)


type torData struct {
	data []byte
	pos int
	
	//Gives the name of the torrent thats parsed, used for error messages. 
	tfile string
}

func (t *torData)next() byte {
	b := t.data[t.pos]
	t.pos = t.pos+1
	return b
}

func (t *torData)peek() byte {
	return t.data[t.pos]
}

func (t *torData)prev() {
	t.pos = t.pos-1
}

func intParse(t *torData) (int, error) {
	intStr := ""
	var b byte
	for b= t.next(); b != 'e'; b = t.next() {
		intStr = intStr+string(b)
	}
	integ, err := strconv.Atoi(intStr)
	if err != nil {
		return 0, errors.New(fmt.Sprint("Error in intParse at ", t.pos, ", in: ",t.tfile))
	}
	return integ, nil
}

func stringParse(t *torData) (string, error) {
	t.prev()
	
	stringSize := ""
	for s:=t.next(); s != ':'; s=t.next() {
		stringSize = stringSize+string(s)
	}
	s_size, err := strconv.Atoi(stringSize)
	if err != nil {
		return "", errors.New(fmt.Sprint("Error in stringParse at ", t.pos, ", in: ",t.tfile))
	}
	
	bstring := make([]byte, s_size)
	for i:=0; i<s_size; i++ {
		bstring[i] = t.next()
	}
	return string(bstring), nil
}

func listParse(t *torData) ([]interface{}, error) {
	var itemSlice []interface{}

	//We read until we reach the end 'e' of the list and make this
	//a list item. we peek so we don't fuck it up for nextItem(*Data)
	for t.peek() != 'e' {
		s, err := nextItem(t)
		if err != nil {
			return nil, errors.New(fmt.Sprint("Error in listParse(): ", err))
		}
		itemSlice = append(itemSlice, s)
	}
	t.next() //Throw away the 'e'
	
	return itemSlice, nil
}

func dictParse(t *torData) (map[string]interface{}, error) {
	dictMap := make(map[string]interface{})
	
	//We read until we reach the end 'e' of the dictionary and make this
	//a dictionary item. We peek so we don't fuck it up for nextItem(*Data).
	//We must be able to read two items at a time, otherwise the torrent is faulty
	//formatted
	for t.peek() != 'e' {
		key, err := nextItem(t)
		if err != nil {
			return nil, errors.New(fmt.Sprint("Error in dictParse(): ", err))
		}
		
		value, err := nextItem(t)
		if err != nil {
			return nil, errors.New(fmt.Sprint("Error in dictParse(): ", err))
		}
		
		dictMap[key.(string)] = value
	}
	t.next() //Throw away the 'e'
	
	return dictMap, nil
}

func nextItem(t *torData) (interface{}, error) {
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
			return nil, errors.New(fmt.Sprint("Out of bounds in nextItem() at ", t.pos," in: ", t.tfile, " (Probably badly encoded torrent)"))
	}
	//Unreachable, but needed due to weird controls in go-compiler
	return nil, nil
}

//Reads in the torrentfile f_name and assumes that the torrent just have one
//main dict item (altough the main dict can contain arbitrary many items). 
//The dict itself is assumed to only have strings or ints as keys (Altough both are
//represented as strings) the value can be anything and is returned as an
//interface{}, type is assertion needed to properly access the value. 
func GetMainDict(f_name string) (map[string]interface{}, error) {
	var t torData
	var err error
	
	t.data, err = ioutil.ReadFile(f_name)
	t.tfile = f_name
	if err != nil {
		return nil, err
	}
	
	mainDict, err:= nextItem(&t)
	if err != nil {
		return nil, err
	}
	if len(t.data) > t.pos {
		return nil, errors.New(fmt.Sprint("Torrent isn't bencoded correctly(Has more then an one Main dict item"))
	}
	return mainDict.(map[string]interface{}), nil
}

//Gets the info dict out of a main dict, returns nil if it doesn't exists
func GetInfoDict(m map[string]interface{}) map[string]interface{}{
	for k,v := range m {
		if k == "info" {
			return v.(map[string]interface{})
		}
	}
	return nil
}
