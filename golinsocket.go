package golinsocket

import (
	"fmt"
	"os"
	"time"

	"net/http"
	//"os/signal" // For the example usage
	"strings"

	"github.com/gorilla/websocket"
	"github.com/linenos/easygo"
)

// WebSocketClient represents a WebSocket client
type WebSocketClient struct {
	conn       *websocket.Conn
	headers    http.Header
	serverURL  string
}

// NewWebSocketClient creates a new WebSocket client
func NewWebSocketClient(serverURL string) *WebSocketClient {
	return &WebSocketClient{
		serverURL: serverURL,
	}
}

// Connect establishes a WebSocket connection to the server
func (c *WebSocketClient) Connect(headers http.Header) error {
	dialer := websocket.Dialer{}	
	conn, _, err := dialer.Dial(c.serverURL, headers)
	if err != nil {
		return fmt.Errorf("dial error: %v", err)
	}
	c.conn = conn
	c.headers = headers
	return nil
}

// SendMessage sends a message to the server
func (c *WebSocketClient) SendMessage(message string) error {
	if c.conn == nil {
		return fmt.Errorf("connection not established")
	}
	err := c.conn.WriteMessage(websocket.TextMessage, []byte(message))
	if err != nil {
		return fmt.Errorf("write error: %v", err)
	}
	return nil
}

// Listen listens for messages from the server and calls the provided handler function for each message
func (c *WebSocketClient) Listen(messageHandler func(string)) {
	reconnectAttemps := 0
	maxReconnectAttempts := 50

	go func() {
		for {
			_, message, err := c.conn.ReadMessage()
			if err != nil {
				if (reconnectAttemps >= maxReconnectAttempts) {
					console.Log("Linsocket Crashed: Failed to reconnect after " + easygo.ToString(reconnectAttemps) + " attempts!")
					return;
				}

				// Server closed connection
				if (strings.Contains(easygo.ToString(err), "closed by the remote host")) {
					console.Log("Linsocket: Server closed connection")
					return;
				}
				if (strings.Contains(easygo.ToString(err), "websocket: close")) {
					console.Log("Linsocket: Server closed connection")
					return;
				}

				// ~~~~~~~~~~~~~~~~~~ Reconnecting
				reconnectAttemps ++
				console.Log("[ Attempting to re-connect " + (easygo.ToString(reconnectAttemps) + "/" + easygo.ToString(maxReconnectAttempts)) + " ] Linsocket Crashed:", err)
				time.Sleep(2 * time.Second)

				dialer := websocket.Dialer{}	
				conn, _, err := dialer.Dial(c.serverURL, c.headers)
				if err == nil {
					c.conn = conn
				}
			}
			messageHandler(string(message))
		}
	}()
}

// Close closes the WebSocket connection
func (c *WebSocketClient) Close() error {
	if c.conn == nil {
		return fmt.Errorf("connection not established")
	}
	err := c.conn.Close()
	if err != nil {
		return fmt.Errorf("close error: %v", err)
	}
	return nil
}


var console = easygo.Console{}
var file = easygo.File{}
var eghttp = easygo.Http{}

// Linsocket represents a Linsocket instance
type Linsocket struct {
	Socket *WebSocketClient
	Close func() error

	On     func(string, func(func(int) interface{}))
	Emit   func(method string, content ...interface{})
}

// Connect initializes a Linsocket connection
func Connect(url string, headers ...http.Header) interface{} {
	// ignore 
	easygo.Bun(os.Chdir, os.Interrupt)

	// Defining variables
	htmlURL := strings.Replace(strings.Replace(url, "ws://", "http://", -1), "wss://", "https://", -1)
	wsURL := strings.Replace(strings.Replace(url, "http://", "ws://", -1), "https://", "wss://", -1)

	// URL Parsing
	_, status, _ := eghttp.Get(strings.Replace(htmlURL+"/linsocket.io", "//linsocket.io", "/linsocket.io", -1), &map[string]string{})
	if status != 200 {
		return "Linsocket was not found at " + url
	}

	// Predefined Variables	
	client := NewWebSocketClient(wsURL)
	err := client.Connect(func() http.Header {
		if (len(headers) > 0) {
			return headers[0]
		}
		return http.Header{}
	}())
	if err != nil {
		return "Error connecting to server: " + easygo.ToString(err)
	}

	// Message Handler
	var linsocket *Linsocket;
	var events = map[string]interface{}{}
	var cached = map[string]interface{}{}

	client.Listen(func(message string) {
		// Websocket to Linsocket
		jsonified, err := easygo.JsonToMap(message)
		if err != nil {
			return;
		}
		// ~~~~~~~~~~~~~~~
		var content []interface{}
		method := jsonified["method"]
		if easygo.TypeOf(method) == "string" && easygo.TypeOf(jsonified["content"]) == "[]interface{}" {
			content = jsonified["content"].([]interface{})
			Get := func(index int) interface{} {
				if len(content) >= index {
					return content[index]
				}
				return nil
			}
	
			// Checking for existing $method
			callback, ok := events[method.(string)].(func(func(int) interface{}))
			if !ok {
				cached[method.(string)] = Get
				return
			}

			// Call method
			callback(Get)
		}
	})

	// Linsocket Functions
	linsocket = &Linsocket{
		Socket: client,
		Close: func() error {
			return client.Close()
		},

		On: func(method string, callback func(func(int) interface{})) {
			existing, isType := cached[method].(func(index int) interface{})
			if isType {
				callback(existing)
			}
			events[method] = callback	
		},
		Emit: func(method string, message ...interface{}) {
			//jsonified := string(easygo.MapToByte(message))
			formatted := map[string]interface{}{}
			formatted["method"] = method
			formatted["content"] = message
			client.SendMessage(string(easygo.MapToByte(formatted)))
		},
	}
	return linsocket
}

// Example Usage:
/*
func main() {
	// Setting up details and initializing connection
	server := "http://localhost:4000/customws"
	header := http.Header{}
	header.Add("test", "hello") // Custom Headers ( )

	_content := Connect(server, header)
	if easygo.TypeOf(_content) == "string" {
		console.Log(_content)
		return;
	}

	// Intellisense
	client := _content.(*Linsocket)
	defer client.Close()

	// Go check out the Node.js server: ( OPEN THE readme.md FILE )
	client.On("hello", func(Get func(int) interface{}) {
		arg := Get(0) // "hello"
		arg2 := Get(1) // "buirehbgjieruhbgne"
		arg3 := Get(2) // "haha"
		easygo.Bun(arg, arg2, arg3) // easygo.Bun does nothing, it just removes the  "you must use this variable you defined" error
		// ...etc
	})

	// Control + C to exit
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	<-interrupt
	fmt.Println("Interrupt signal received, closing connection...")
}
*/