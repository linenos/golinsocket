// Added a middle man for intercepting request
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
	cache map[interface{}]interface{}
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
	c.cache = map[interface{}]interface{}{}
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

// Handle Cache
func (c * WebSocketClient) AddCache(index interface{}, value interface{}) {
	c.cache[index] = value
}

// Listen listens for messages from the server and calls the provided handler function for each message
func (c *WebSocketClient) OnClose(reason string) {
	close, itype := c.cache["OnClose"].(func(string))
	if itype {
		close(reason)
	}
}
func (c *WebSocketClient) Listen(messageHandler func(string)) {
	//reconnectAttemps := 0
	//maxReconnectAttempts := 50

	go func() {
		for {
			_, message, err := c.conn.ReadMessage()
			if err != nil {
				err := err.Error()
				// Server closed connection
				if (strings.Contains(err, "closed by the remote host")) {
					c.OnClose("Server closed connection")
					return;
				}
				if (strings.Contains(err, "websocket: close")) {
					c.OnClose("Server closed connection")
					return;
				}
				c.OnClose(err)
				return;
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
	OnClose func(func(reason string))

	MiddleMan func(request []interface{}) []interface{}

	RemoveOnEvent func(method string)
	On     func(string, func(Get func(int) interface{}, content []interface{}))
	Emit   func(method string, content ...interface{})
}

// Connect initializes a Linsocket connection
func Connect(url string, params string, headers ...http.Header) interface{} {
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
	client := NewWebSocketClient(wsURL + params)
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
			content = linsocket.MiddleMan(jsonified["content"].([]interface{}))
			Get := func(index int) interface{} {
				if len(content) > index {
					return content[index]
				}
				return nil
			}
	
			// Checking for existing $method
			callback, ok := events[method.(string)].(func(func(int) interface{}, []interface{}))
			if !ok {
				cached[method.(string)] = Get
				return
			}

			// Call method
			callback(Get, content)
		}
	})

	// Linsocket Functions
	linsocket = &Linsocket{
		Socket: client,

		// Middle man for intercepting contents of a request [ not the method, only contents ]
		MiddleMan: func(request []interface{}) []interface{} {
			return request
		},

		OnClose: func(closeEvent func(reason string)) {
			client.AddCache("OnClose", closeEvent)
		},

		Close: func() error {
			return client.Close()
		},

		RemoveOnEvent: func(method string) {
			events[method] = nil
		},

		On: func(method string, callback func(Get func(int) interface{}, content []interface{})) {
			existing, isType := cached[method].(func(index int) interface{})
			existing2, isType2 := cached[method + "_content"].([]interface{})
			if isType && isType2 {
				callback(existing, existing2)
				cached[method] = nil
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
package main
import (
        "fmt"
	"os/signal"
	"github.com/linenos/golinsocket"
)

func main() {
	// Setting up details and initializing connection
	server := "http://localhost:4000/customws"
	header := http.Header{}
	header.Add("test", "hello") // Custom Headers ( )

	_content := Connect(server, header)
	if easygo.TypeOf(_content) == "string" {
		fmt.Println(_content)
		return;
	}

	// Intellisense
	client := _content.(*Linsocket)
	defer client.Close()

    // When the websocket connection is closed
	client.OnClose(func(reason string) {
		fmt.Println(reason)
	})

	// Go check out the Node.js server: ( OPEN THE readme.md FILE )
	client.On("hello", func(Get func(int) interface{}, content []interface{}) {
		arg := Get(0) // "hello"
		arg2 := Get(1) // "buirehbgjieruhbgne"
		arg3 := Get(2) // "haha"

		// content is the contents that the function 'Get' is indexing through
		args := content[1:] // This gets all the arguments above argument 1, basically Get(1) and Get(2) and Get(...) -- argument 1 and over

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
