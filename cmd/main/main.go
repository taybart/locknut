package main

import (
	"github.com/taybart/locknut"
	"github.com/taybart/log"
	"os"
)

type pii struct {
	Name string
}

func main() {
	defer os.Remove("test.db")
	key, err := locknut.GetRandKey()
	if err != nil {
		log.Fatal(err)
	}
	bl, err := locknut.NewBoltLocknut("test.db", ".", key, false, []string{"pii", "jids"})
	if err != nil {
		log.Fatal(err)
	}
	p := pii{Name: "taylor"}
	err = bl.Save("pii", "taylor", p)
	if err != nil {
		log.Fatal(err)
	}
	err = bl.Save("pii", "test", p)
	if err != nil {
		log.Fatal(err)
	}
	err = bl.Save("pii", "asdfasdf", p)
	if err != nil {
		log.Fatal(err)
	}
	keys, err := bl.GetKeyList("pii", "t")
	if err != nil {
		log.Fatal(err)
	}
	log.Info(keys) // taylor, test
	results, err := bl.GetByPrefix("pii", "")
	if err != nil {
		log.Fatal(err)
	}
	for k, v := range results {
		log.Info(k, string(v))
	}
	dec, err := bl.GetOne("pii", "taylor")
	if err != nil {
		log.Fatal(err)
	}
	log.Info(string(dec))
}
