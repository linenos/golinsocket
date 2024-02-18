// Created by Linen#3485
// Go made easier
package easygo

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"io/ioutil"
	"net/http"
	"reflect"
	"runtime"

	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"os/exec"
	"path"
)

var Credits string = "Linen#3485 on Discord"

// Functions ->
func Bun(...any) { // Pass any variable to bypass the 'variable not used' or 'you must use this variable' error
	return
}

func TypeOf(i any) string {
	return strings.Replace(ToString(reflect.TypeOf(i)), " ", "", -1)
}

func IndexExists(slice []string, index int) any { // Check if index exist within a []string
  if (index >= 0 && index < len(slice)) {
    return slice[index]
  }
  return nil
}

func GenerateUniqueFileName(dir, baseName, ext string, deleteExisting bool) string { // Loops through directory and checks for a free file name, how it works: "${baseName}.${ext}" | "testname1.exe"
	// Initialize the new file name with the base name and extension
	newFileName := baseName + ext
	tempFileName := baseName + ext
	
	//console := Console{}
	file := File{}

	// Loop through files in the directory to find a non-existing filename
	if (deleteExisting) {
		for i := 1; i < 500; i++ {
			exist := file.Exists(path.Join(tempFileName))
			if (exist) {
				file.Delete(path.Join(newFileName))
			}
			tempFileName = fmt.Sprintf("%s%d%s", baseName, i, ext)
		}
	}

	for i := 1; ; i++ {
		exist := file.Exists(path.Join(newFileName))
		if !exist {
			break
		}
		newFileName = fmt.Sprintf("%s%d%s", baseName, i, ext)
	}

	return newFileName
}

func MapToByte(smap interface{}, args ...string) []byte {
	prefix := ""
	indent := "    "

	if value := IndexExists(args, 0); value != nil {
		prefix = args[0]
	}

	if value := IndexExists(args, 1); value != nil {
		indent = args[1]
	}

	responseBodyJSON, err := json.MarshalIndent(smap, prefix, indent)
	if err != nil {
		return []byte("{\n}")
	}
	return responseBodyJSON
}

func ToString(value interface{}) string {
	return fmt.Sprintf("%v", value)
}

func IsNil(value interface{}) bool {
	stringified := ToString(value)
	if stringified == "<nil>" {
		return true
	}
	return false
}

func StoreMapToFile(filename string, data map[string]interface{}) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ") // Optional: to make the output JSON more readable

	if err := encoder.Encode(data); err != nil {
		return err
	}
	return nil
}

func JsonToMap(jsonStr string) (map[string]interface{}, error) {
	var result map[string]interface{}

	// Unmarshal JSON string into map
	err := json.Unmarshal([]byte(jsonStr), &result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func ReadMapFromFile(filename string) (map[string]interface{}, error) {
	// Open the file
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Decode the JSON data
	var data map[string]interface{}
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&data)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func EasyExit(content string, extra ...any) {
	fmt.Printf(content, extra...)
	fmt.Println("\n[ Press Enter to exit... ]")
	Console{}.WaitForEnter()
	os.Exit(1)
	panic("")
}

// encrypt string to base64 crypto using AES
func KeyEncrypt(keyStr string, cryptoText string) string {
	keyBytes := sha256.Sum256([]byte(keyStr))
	return Encrypt(keyBytes[:], cryptoText)
}

func Encrypt(key []byte, text string) string {
	plaintext := []byte(text)

	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err)
	}

	// The IV needs to be unique, but not secure. Therefore it's common to
	// include it at the beginning of the ciphertext.
	ciphertext := make([]byte, aes.BlockSize+len(plaintext))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		panic(err)
	}

	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(ciphertext[aes.BlockSize:], plaintext)

	return base64.StdEncoding.EncodeToString(ciphertext)
}

// decrypt from base64 to decrypted string
func KeyDecrypt(keyStr string, cryptoText string) string {
	keyBytes := sha256.Sum256([]byte(keyStr))
	return Decrypt(keyBytes[:], cryptoText)
}

func Decrypt(key []byte, cryptoText string) string {
	ciphertext, err := base64.StdEncoding.DecodeString(cryptoText)
	if err != nil {
		panic(err)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err)
	}

	// The IV needs to be unique, but not secure. Therefore it's common to
	// include it at the beginning of the ciphertext.
	if len(ciphertext) < aes.BlockSize {
		panic("ciphertext too short")
	}
	iv := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]
	stream := cipher.NewCFBDecrypter(block, iv)

	// XORKeyStream can work in-place if the two arguments are the same.
	stream.XORKeyStream(ciphertext, ciphertext)
	return fmt.Sprintf("%s", ciphertext)
}

// ---------------------------------------------- \\

// Colors ->
func (c ConsoleColors) Color(text string, code string) string {
	return fmt.Sprintf("\033[%sm%s\033[0m", code, text)
}

func (c ConsoleColors) Blue(text string) string {
	return "\033[34m" + text + "\033[0m"
}

func (c ConsoleColors) Green(text string) string {
	return "\033[32m" + text + "\033[0m"
}

func (c ConsoleColors) Red(text string) string {
	return "\033[31m" + text + "\033[0m"
}

func (c ConsoleColors) Yellow(text string) string {
	return "\033[33m" + text + "\033[0m"
}

func (c ConsoleColors) Bold(text string) string {
	return "\033[1m" + text + "\033[0m"
}

// Console ->
func (c Console) LogC(content string, extra ...any) {
	fmt.Printf(content, extra...)
}

func (c Console) Log(content ...any) {
	fmt.Println(content...)
}

func (c Console) Clear() {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "linux", "darwin": // Unix-like systems
		cmd = exec.Command("clear")
	case "windows": // Windows
		cmd = exec.Command("cmd", "/c", "cls")
	default:
		return
	}

	cmd.Stdout = os.Stdout
	cmd.Run()
}

func (c Console) GetInput(customLog ...string) string {
	if len(customLog) > 0 {
		c.LogC(customLog[0]) // Print customLog if provided
	}

	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()

	// Retrieve the text that was entered
	return scanner.Text()
}

func (c Console) WaitForEnter() {
	bufio.NewReader(os.Stdin).ReadBytes('\n')
}

type ConsoleColors struct{}
type Console struct{}

// ---------------------------------------------- \\

// Maps ->
func NewObservableMap(callback func(key string, value interface{}, alldata map[string]interface{})) *ObservableMap {
	return &ObservableMap{
		data:     make(map[string]interface{}),
		callback: callback,
	}
}

func (m *ObservableMap) GetAll() map[string]interface{} {
	return m.data
}

func (m *ObservableMap) Get(keys ...string) interface{} {
	predata := m.data
	return GetInMap(predata, keys...)
}

func (m *ObservableMap) SetAll(data map[string]interface{}) {
	m.data = data
}

func (m *ObservableMap) Set(keys []string, value interface{}) {
	m.data = SetInMap(m.data, keys, value)
	if m.callback != nil {
		lastKey := keys[len(keys)-1]
		m.callback(lastKey, value, m.GetAll())
	}
}

func SetInMap(data map[string]interface{}, keys []string, value interface{}) map[string]interface{} {
	if len(keys) == 0 {
		return data
	}

	key := keys[0]
	if len(keys) == 1 {
		data[key] = value
	} else {
		if _, ok := data[key]; !ok {
			data[key] = make(map[string]interface{})
		}
		data[key] = SetInMap(data[key].(map[string]interface{}), keys[1:], value)
	}
	return data
}

func GetInMap(data map[string]interface{}, keys ...string) interface{} {
	var current interface{} = data

	// Iterate through each key in the keys slice
	for _, key := range keys {
		// Check if the current value is a map
		if nestedMap, ok := current.(map[string]interface{}); ok {
			// Retrieve the value associated with the current key
			if val, found := nestedMap[key]; found {
				// Update the current value to the retrieved value
				current = val
			} else {
				// If the key is not found, return nil
				return nil
			}
		} else {
			// If the current value is not a map, return nil
			return nil
		}
	}
	// Return the final retrieved value
	return current
}

type ObservableMap struct {
	data     map[string]interface{}
	callback func(key string, value interface{}, alldata map[string]interface{})
}

// ---------------------------------------------- \\

// HTTP(S) ->
func (h Http) GetRaw(url string, headers *map[string]string, body ...map[string]interface{}) (interface{}, error) {
	// Convert body to JSON if provided
	var bodyBytes []byte
	if len(body) > 0 {
		var err error
		bodyBytes, err = json.Marshal(body[0])
		if err != nil {
			return "", err
		}
	}

	// Create a new HTTP client
	client := &http.Client{}

	// Create a new request
	request, err := http.NewRequest("GET", url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return "", err
	}

	// Set custom headers if provided
	if headers != nil {
		for key, value := range *headers {
			request.Header.Set(key, value)
		}
	}

	// Send the request
	response, err := client.Do(request)
	if err != nil {
		return "", err
	}

	// Return raw body
	return response, err
}

func (h Http) Get(url string, headers *map[string]string, body ...map[string]interface{}) (string, int, error) {
	// Convert body to JSON if provided
	var bodyBytes []byte
	if len(body) > 0 {
		var err error
		bodyBytes, err = json.Marshal(body[0])
		if err != nil {
			return "", 404, err
		}
	}

	// Create a new HTTP client
	client := &http.Client{}

	// Create a new request
	request, err := http.NewRequest("GET", url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return "", 404, err
	}

	// Set custom headers if provided
	if headers != nil {
		for key, value := range *headers {
			request.Header.Set(key, value)
		}
	}

	// Send the request
	response, err := client.Do(request)
	if err != nil {
		return "", 404, err
	}
	defer response.Body.Close()

	// Read the response body
	responseBody, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return "", 404, err
	}

	// Convert response body to string and return
	return string(responseBody), response.StatusCode, nil
}

func (h Http) Post(url string, headers *map[string]string, body ...map[string]interface{}) (string, int, error) {
	// Convert body to JSON if provided
	var bodyBytes []byte
	if len(body) > 0 {
		var err error
		bodyBytes, err = json.Marshal(body[0])
		if err != nil {
			return "", 404, err
		}
	}

	// Create a new HTTP client
	client := &http.Client{}

	// Create a new request
	request, err := http.NewRequest("POST", url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return "", 404, err
	}

	// Set custom headers if provided
	if headers != nil {
		for key, value := range *headers {
			request.Header.Set(key, value)
		}
	}

	// Send the request
	response, err := client.Do(request)
	if err != nil {
		return "", 404, err
	}
	defer response.Body.Close()

	// Read the response body
	responseBody, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return "", 404, err
	}

	// Convert response body to string and return
	return string(responseBody), response.StatusCode, nil
}

func (f File) Delete(filepath string) bool {
	err := os.Remove(filepath)
	if err != nil {
		return false
	}
	return true
}

func (f File) ReadFileToString(filename string) (string, error) {
    // Read the entire file
    content, err := ioutil.ReadFile(filename)
    if err != nil {
        return "", err
    }
    // Convert content to string and return
    return string(content), nil
}

func (f File) Exists(filePath string) bool {
	info, err := os.Stat(filePath)
	if err == nil {
		return !info.IsDir()
	}
	if errors.Is(err, os.ErrNotExist) {
		return false
	}
	return false
}

func (f File) GetSHA256(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}

type File struct {
}
type Http struct {
}

// ---------------------------------------------- \\

// Pointer Spoofs ->
func (p PointerMap) Interface(typ interface{}) *interface{} {
	return &typ
}

func (p PointerMap) String(typ string) *string {
	return &typ
}

func (p PointerMap) Int(typ int) *int {
	return &typ
}

func (p PointerMap) Int8(typ int8) *int8 {
	return &typ
}

func (p PointerMap) Int16(typ int16) *int16 {
	return &typ
}

func (p PointerMap) Int32(typ int32) *int32 {
	return &typ
}

func (p PointerMap) Int64(typ int64) *int64 {
	return &typ
}

func (p PointerMap) Float64(typ float64) *float64 {
	return &typ
}

func (p PointerMap) Bool(typ bool) *bool {
	return &typ
}

type PointerMap struct {
}
