package bittor

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"crypto/sha1"
	"hash"
)

type torData struct {
	data []byte
	pos  int
}

func (t *torData) next() byte {
	b := t.data[t.pos]
	t.pos = t.pos + 1
	return b
}

func (t *torData) peek() byte {
	return t.data[t.pos]
}

func (t *torData) prev() {
	t.pos = t.pos - 1
}

func intParse(t *torData) (int, error) {
	intStr := ""
	var b byte
	for b = t.next(); b != 'e'; b = t.next() {
		intStr = intStr + string(b)
	}
	integ, err := strconv.Atoi(intStr)
	if err != nil {
		return 0, errors.New(fmt.Sprint("Error in intParse at ", t.pos))
	}
	return integ, nil
}

func stringParse(t *torData) (string, error) {
	t.prev()

	stringSize := ""
	for s := t.next(); s != ':'; s = t.next() {
		stringSize = stringSize + string(s)
	}
	sSize, err := strconv.Atoi(stringSize)
	if err != nil {
		return "", errors.New(fmt.Sprint("Error in stringParse at ", t.pos))
	}

	bstring := make([]byte, sSize)
	for i := 0; i < sSize; i++ {
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
		return nil, errors.New(fmt.Sprint("Out of bounds in nextItem() at ", t.pos, " (Probably badly encoded torrent)"))
	}
	//Unreachable, but needed due to weird controls in go-compiler
	return nil, nil
}

//Reads in the torrentfile f_name and assumes that the torrent just have one
//main dict item (altough the main dict can contain arbitrary many items).
//The dict itself is assumed to only have strings or ints as keys (Altough both are
//represented as strings) the value can be anything and is returned as an
//interface{}, type assertion is needed to properly access the value.
func GetMainDict(tData []byte) (map[string]interface{}, error) {
	var t torData
	var err error

	t.data = tData

	mainDict, err := nextItem(&t)
	if err != nil {
		return nil, err
	}
	if len(t.data) > t.pos {
		return nil, errors.New(fmt.Sprint("Torrent isn't bencoded correctly(Has more then an one Main dict item"))
	}
	return mainDict.(map[string]interface{}), nil
}




//Gets the info dict out of a main dict, returns nil if it doesn't exists
func GetInfoDict(m map[string]interface{}) (map[string]interface{}, error) {
	for k, v := range m {
		if k == "info" {
			return v.(map[string]interface{}), nil
		}
	}
	return nil, errors.New(fmt.Sprint("No info dict found"))
}

//Gets a string list value from a dictionary (or map[string]Interface{}) that
//has the key value "key". These reflection things are subtle and recommended
//reading is
//
//http://blog.golang.org/2011/09/laws-of-reflection.html,
//http://research.swtch.com/interfaces and the
//The "fmt" source code.
//
//If there is diffrent list-values in a list you'll have to implement your own
//corresponding function. The string list value is so common that this function is
//included ;)
func GetStringListFromDict(key string, dict map[string]interface{}) ([]string, error) {

	iVal, ok := dict[key]

	iList, ok := iVal.([]interface{})
	if !ok {
		return nil, errors.New(fmt.Sprint("Not a list value corresponding to the key \"", key, "\" in the dict"))
	}

	list := make([]string, len(iList))
	for i, it := range iList {
		v := reflect.ValueOf(it)

		//There's some special conditions for these types of lists and they have
		//to be tested in the correct order so the program don't panic if a faulty
		// list-value is corresponding to a key.
		if v.Kind() == reflect.Slice && v.Index(0).IsValid() && v.Len() == 1 &&
			v.Index(0).Elem().Kind() == reflect.String {
			list[i] = v.Index(0).Elem().Interface().(string) //Love reflection! xD
		} else {
			return nil, errors.New(fmt.Sprint("An incorrect list value corresponding to the key \"", key, "\" in the dict"))
		}
	}
	return list, nil

}

func GetInfoHash(tData []byte) (hash.Hash, error) {
	t := new(torData)
	
	t.data = tData
	if t.next() != 'd' {
		return nil, errors.New("Not a torrent file! Atleast not correctly bencoded file")
	}
	
	info, err := infoByteValue(t)
	if err != nil {
		return nil, errors.New(fmt.Sprint("Error GetInfoHash: ", err))
	}
	
	th := sha1.New()
	_,err = th.Write(info)
	if err != nil {
		return nil, errors.New(fmt.Sprint("Error GetInfoHash: ", err))
	}
	
	return th, nil
}

func infoByteValue(t *torData) ([]byte, error) {

	//We read until we reach the end 'e' of the dictionary and make this
	//a dictionary item. We peek so we don't fuck it up for nextItem().
	//We must be able to read two items at a time, otherwise the torrent is faulty
	//formatted
	for t.peek() != 'e' {
		key, err := nextItem(t)
		if err != nil {
			return nil, errors.New(fmt.Sprint("Error in dictParse(): ", err))
		}
		
		if key.(string) != "info" {
			_, err = nextItem(t)
		if err != nil {
			return nil,errors.New(fmt.Sprint("Error in dictParse(): ", err))
		}
			continue
		}
		
		s := t.pos
		_, err = nextItem(t)
		if err != nil {
			return nil,errors.New(fmt.Sprint("Error in dictParse(): ", err))
		}
		return t.data[s:t.pos-1], nil
		
	}
	return nil, errors.New("SHOULD NOT REACH THIS IN INFOBYTEVALUE!")
}
