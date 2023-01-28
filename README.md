# ACCTER
A simple radius accounting server which can be used to transfer accounting data (in json format) to a endpoint.

## Getting Started

### Installation

```bash
go get github.com/dinifarb/accter
```
### Usage
This server just prints the received accounting data to stdout. You can use it as a template to implement your own accounting server.
```go
package main

import (
	"fmt"

	"github.com/dinifarb/accter/pkg/accter"
)

func main() {
	handler := func(packet *accter.JsonPacket) error {
		fmt.Println(packet.Code)
		fmt.Println(packet.Id)
		fmt.Println(packet.Authenticator)
		fmt.Println(packet.Key)
		fmt.Println(packet.RemoteAddr)
		for _, attr := range packet.Attributes {
			fmt.Println(attr.Name, "=", attr.Value)
		}
		// return an error if no accounting data should be sent to the origin server
		return nil
	}
	server := accter.PacketServer{
		Port:          1813,
		Secret:        "secret",
		HandleRequest: handler,
	}
	if err := server.Serve(); err != nil {
		panic(err)
	}
}
```

