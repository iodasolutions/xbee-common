package main

import (
	"context"
	"github.com/iodasolutions/xbee-common/indus"
	"log"
)

func main() {
	ctx := context.TODO()

	if err := indus.Build(ctx, "indus/build/main", "toto"); err != nil {
		log.Fatal(err)
	}
}
