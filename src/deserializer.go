package main

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

func main() {
	// checks that a nosj filename was actually provided
	argLength := len(os.Args[1:])
	if argLength < 1 {
		handleError(errors.New("no input filename provided"))
	}

	fileName := os.Args[1]

	// tries to find this nosj file in the current directory
	content, err := os.ReadFile(fileName)
	if err != nil {
		handleError(errors.New("failed to open file"))
	}

	// builds the map from the file
	m := buildMap(string(content))
	// outputs the map according to the specification
	printMap(m)
}

// turns a nosj string into a map
// - if at any point it is determined
func buildMap(content string) map[string]interface{} {
	content = strings.TrimSpace(content)
	if len(content) < 1 {
		handleError(errors.New("malformed file"))
	}
	if content[0] != '<' || content[len(content)-1] != '>' {
		handleError(errors.New("malformed file"))
	}

	outerMap := make(map[string]interface{})
	parts := strings.Split(content[1:len(content)-1], ",")
	for _, part := range parts {
		if part == "" {
			continue
		}

		key, value := splitKeyAndValue(part)
		if key == "" || value == "" {
			handleError(errors.New("malformed file"))
		}

		if !isValidKey(key) {
			handleError(errors.New("malformed file"))
		} else if outerMap[key] != nil {
			handleError(errors.New("malformed file"))
		}

		if value[0] == '<' {
			outerMap[key] = buildMap(value)
		} else if strings.Contains(value, "%") {
			if !isValidComplexString(value) {
				handleError(errors.New("malformed file"))
			}

			outerMap[key] = getComplexString(value)
		} else if value[len(value)-1] == 's' {
			if !isValidSimpleString(value) {
				handleError(errors.New("malformed file"))
			}

			outerMap[key] = getSimpleString(value)
		} else if value[0] == 'i' {
			if !isValidInteger(value) {
				handleError(errors.New("malformed file"))
			}

			outerMap[key] = getInteger(value)
		} else {
			handleError(errors.New("malformed file"))
		}
	}

	return outerMap
}

// outputs the map according to the specification
func printMap(data map[string]interface{}) {
	fmt.Println("begin-map")

	for key, value := range data {
		reflectType := reflect.TypeOf(value).Kind()

		if reflectType == reflect.Map {
			fmt.Println(key, "-- map --")
			printMap(value.(map[string]interface{}))
		} else {
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

func splitKeyAndValue(text string) (string, string) {
	splitIndex := strings.Index(text, ":")
	return text[0:splitIndex], text[splitIndex+1:]
}

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

// determines whether a map key k is valid, that is:
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

// tests whether a string s is a valid complex string according to nosj spec, that is:
func isValidComplexString(text string) bool {
	percentSplit := strings.Split(text, "%")
	for i := 1; i < len(percentSplit); i++ {
		if len(percentSplit[i]) < 2 {
			return false
		} else if !isAlphaNumeric(percentSplit[i]) {
			return false
		}

		asciiCode, error := strconv.ParseUint(percentSplit[i][0:2], 16, 64)
		if error != nil || len(string(rune(asciiCode))) != 1 {
			return false
		}
	}

	return true
}

func getComplexString(text string) string {
	percentSplit := strings.Split(text, "%")
	for i := 1; i < len(percentSplit); i++ {
		asciiCode, error := strconv.ParseUint(percentSplit[i][0:2], 16, 64)
		if error != nil {
			handleError(errors.New("malformed file"))
		}

		percentSplit[i] = string(rune(asciiCode)) + percentSplit[i][2:]
	}

	return strings.Join(percentSplit, "")
}

func isValidSimpleString(text string) bool {
	return isAlphaNumericOrWhiteSpace(text)
}

func getSimpleString(text string) string {
	return text[:len(text)-1]
}

func isValidInteger(text string) bool {
	if len(text) < 2 {
		return false
	} else if text[1] == '-' {
		if len(text) < 3 || !isNumeric(text[2:]) {
			return false
		}
	} else if !isNumeric(text[1:]) {
		return false
	}

	return true
}

func getInteger(text string) int {
	isNegative := text[1] == '-'

	firstDigitIndex := 1
	if isNegative {
		firstDigitIndex = 2
	}

	num, err := strconv.ParseUint(text[firstDigitIndex:], 10, 64)
	if err != nil {
		handleError(errors.New("malformed file"))
	}

	if isNegative {
		return -int(num)
	} else {
		return int(num)
	}
}

// handles any errors that would cause the program to terminate prematurely
// 1. filename not provided
// 2. file not found
// 3. malformed file
func handleError(err error) {
	const prefix = "ERROR --"
	fmt.Println(prefix, err.Error())
	os.Exit(66)
}