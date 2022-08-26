package main

import (
	"errors"
	"fmt"
	"log"
	"net/url"
	"os"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

func main() {
	// setup logging style for stderr
	log.SetPrefix("ERROR -- ")
	log.SetFlags(0)

	// validates that a nosj filepath was actually provided as an argument
	argLength := len(os.Args[1:])
	if argLength < 1 {
		raiseError("no input filename provided")
	}

	fileName := os.Args[1]

	// tries to find this nosj file in the current directory
	content, err := os.ReadFile(fileName)
	if err != nil {
		raiseError("failed to open file")
	}

	// builds the map from the file, and outputs it
	m := buildMap(string(content))
	printMap(m)
}

// turns a nosj string into a map
// - if at any point it is determined to be malformed, it will terminate with an error
func buildMap(content string) *OrderedMap[string, interface{}] {
	// whitespace around a root map doesn't matter
	content = strings.TrimSpace(content)
	// make sure the file has content and has a root map <>
	if len(content) < 1 {
		raiseError("malformed file")
	}
	if content[0] != '<' || content[len(content)-1] != '>' {
		raiseError("malformed file")
	}

	// this map is what will be returned
	outerMap := NewOrderedMap[string, interface{}]()

	// split the inside of the map into parts, where each part is a key-value pair
	parts := strings.Split(content[1:len(content)-1], ",")
	for _, part := range parts {
		if part == "" {
			continue
		}

		// both the key and the value need some content inside them
		if !isValidMapEntry(part) {
			raiseError("malformed file")
		}
		key, value := splitMapEntry(part)

		// make sure key is valid and has not already been used
		if !isValidKey(key) {
			raiseError("malformed file")
		} else if outerMap.Has(key) {
			raiseError("malformed file")
		}

		// based on the contents, build a value which can be one of 4 types
		// 1. map
		// 2. integer
		// 3. simple string
		// 4. complex string
		var newValue interface{} = nil
		
		// map validation takes place inside buildMap function
		// the rest will validate before building the value
		if value[0] == '<' {
			newValue = buildMap(value)
		} else if strings.Contains(value, "%") {
			if !isValidComplexString(value) {
				raiseError("malformed file")
			}

			newValue = getComplexString(value)
		} else if value[len(value)-1] == 's' {
			if !isValidSimpleString(value) {
				raiseError("malformed file")
			}

			newValue = getSimpleString(value)
		} else if value[0] == 'i' {
			if !isValidInteger(value) {
				raiseError("malformed file")
			}

			newValue = getInteger(value)
		} else {
			// it conforms to nothing else, so it must be a malformed file
			raiseError("malformed file")
		}

		// if we get here, we have a valid key-value pair, so add it to the map
		outerMap.Set(key, newValue)
	}

	return outerMap
}

// outputs the map according to the specification
func printMap(omap *OrderedMap[string, interface{}]) {
	fmt.Println("begin-map")

	// iterate over the map keys
	iterator := omap.Iterator()

	for {
		i, k, value := iterator()
		if i == nil {
			break
		}

		key := *k
		// basically have to use reflection to get the type of the value
		// because we saved it as a map of interface{}
		reflectType := reflect.TypeOf(value).Kind()

		// if the value is a map, we need to recurse into it
		if reflectType == reflect.Ptr {
			fmt.Println(key, "-- map -- ")
			printMap(value.(*OrderedMap[string, interface{}]))
		} else {
			// otherwise, we can just print the value
			valueType := ""
			switch reflectType {
			case reflect.String:
				valueType = "string"
			case reflect.Int:
				valueType = "integer"
			}

			fmt.Println(key, "--", valueType, "--", value)
		}
	}

	fmt.Println("end-map")
}

// for my purposes, i am defining a valid map entry
// as a string of the form a:b where the lengths of
// substrings a and b are at least 1
func isValidMapEntry(text string) bool {
	if len(text) < 3 {
		return false
	}

	// can't have an entry of the form :c or c:
	colonIndex := strings.Index(text, ":")
	if colonIndex < 1 || colonIndex > len(text) - 2 {
		return false
	}

	return true
}

// takes a map entry and splits it into a key and value
func splitMapEntry(text string) (string, string) {
	splitIndex := strings.Index(text, ":")
	return text[0:splitIndex], text[splitIndex+1:]
}

// determines whether a string k is a valid nosj key, that is:
// 1. k is non-empty
// 2. k is alphanumeric
func isValidKey(key string) bool {
	if len(key) < 1 {
		return false
	} else if !isAlphaNumeric(key) {
		return false
	}

	return true
}

// region NOSJ PRIMITIVE SPEC VALIDATION

// tests whether a string i is a valid nosj integer, that is:
//  1. i is at least two characters long
//  2. i[0] is 'i'
//  3. if i[1] is '-', then i must be at least 3 characters long and all
//     but the first 2 characters must be numeric
//
// 4. if i[1] is not '-', then all but the first character must be numeric
func isValidInteger(text string) bool {
	if len(text) < 2 {
		return false
	} else if text[0] != 'i' {
		return false
	}

	if text[1] == '-' {
		// validation for negative integers
		if len(text) < 3 || !isNumeric(text[2:]) {
			return false
		}
	} else {
		// validation for positive integers
		if !isNumeric(text[1:]) {
			return false
		}
	}

	return true
}

// tests whether a string s is a valid nosj simple string, that is:
// 1. s is non-empty
// 2. s ends with 's'
// 3. s is alphanumeric or whitespace
func isValidSimpleString(text string) bool {
	if len(text) < 1 || text[len(text)-1] != 's' {
		return false
	} else if !isAlphaNumericOrWhiteSpace(text) {
		return false
	}

	return true
}

// tests whether a string s is a valid nosj complex string, that is:
func isValidComplexString(text string) bool {
	if !strings.Contains(text, "%") {
		return false
	} else if !isAlphaNumericOrPercents(text) {
		return false
	}

	_, error := url.QueryUnescape(text)
	if error != nil {
		raiseError("malformed file")
	}

	return true
}

// endregion

// region NOSJ PRIMITIVE CONVERSIONS

// converts a nosj integer string to an integer
func getInteger(text string) int {
	isNegative := text[1] == '-'

	// indicates which index is where the numeric portion starts
	firstDigitIndex := 1
	if isNegative {
		firstDigitIndex = 2
	}

	num, err := strconv.ParseUint(text[firstDigitIndex:], 10, 64)
	if err != nil {
		raiseError("malformed file")
	}

	if isNegative {
		return -int(num)
	} else {
		return int(num)
	}
}

// converts a nosj simple string to a string
func getSimpleString(text string) string {
	return text[:len(text)-1]
}

// converts a nosj complex string to a string
func getComplexString(text string) string {
	decoded, error := url.QueryUnescape(text)
	if error != nil {
		raiseError("malformed file")
	}

	return decoded
}

// endregion

// region ERROR HANDLING

// handles any errors that would cause the program to terminate prematurely
// 1. filename not provided
// 2. file not found
// 3. malformed file
func raiseError(errorMessage string) {
	error := errors.New(errorMessage)

	log.Println(error.Error())
	os.Exit(66)
}

// handles errors concerned with malformed files, that is
// files that specified and exist, but are not formatted
// according to the nosj specification
func raiseMalformedFileError() {
	raiseError("malformed file")
}


// endregion

// region STRING HELPERS

// tests whether a string contains only numerical digits
func isNumeric(text string) bool {
	return regexp.MustCompile(`^[0-9]*$`).MatchString(text)
}

// tests whether a string contains only numerical digits or lower/upper case letters
func isAlphaNumeric(text string) bool {
	return regexp.MustCompile(`^[a-zA-Z0-9]*$`).MatchString(text)
}

// tests whether a string contains only numerical digits or lower/upper case letters or whitespace
func isAlphaNumericOrWhiteSpace(text string) bool {
	return regexp.MustCompile(`^[a-zA-Z0-9\s]*$`).MatchString(text)
}

// tests whether a string contains only numerical digits or lower/upper case letters or percent signs
func isAlphaNumericOrPercents(text string) bool {
	return regexp.MustCompile(`^[a-zA-Z0-9%]*$`).MatchString(text)
}

// endregion
