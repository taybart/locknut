# BoltDB Locknut
A boltdb wrapper to encrypt and decrypt the values stored in the boltdb via AES Cryptor, and also provides common 
db operations such as GetOne, GetByPrefix, GetKeyList, Save, Delete and etc. 

Boltdb file is always open in the file system unless the DB.Close() is called, which cause inconvenience 
if you want to do some file operations to the db file while the program is running. This package provides the parameter: batchMode to 
control whether to close the db after each db operation, this has performance impact but could be a useful feature.

### Usage Example

```golang
import (
	"github.com/taybart/locknut"
	"github.com/taybart/log"
)

type pii struct {
	Name string
}

func main() {
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
	keys, err := bl.GetKeyList("pii", "t")
	if err != nil {
		log.Fatal(err)
	}
	log.Info(keys)
	dec, err := bl.GetOne("pii", "taylor")
	if err != nil {
		log.Fatal(err)
	}
	log.Info(string(dec))
}
```
