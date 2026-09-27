package main

import (
	"fmt"
	"log"

	"github.com/tanmay21k/goUtils/internal/storeService"
	"github.com/tanmay21k/goUtils/internal/stores/kv"
)

func main() {
	store, err := kv.NewStore(10)
	if err != nil {
		log.Fatal(err)
	}

	appStore := storeService.NewStoreService(store)

	err = appStore.Set("name", "Tanmay")
	if err != nil {
		log.Fatal(err)
	}

	value, err := appStore.Get("name")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(value)

	appStore.Del("name")
}
