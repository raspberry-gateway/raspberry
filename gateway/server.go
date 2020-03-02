package gateway

import "context"

func Start() {
	_, cancel := context.WithCancel(context.Background())
	defer cancel()

}
