package locknut

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/taybart/log"
	"testing"
)

func TestAES(t *testing.T) {

	obj := Cryptor{
		plain: []byte("I like seafood."),
	}

	Key, err := GetRandKey()
	assert.NoError(t, err)

	log.Testf("AES:Encryption...")
	err = obj.Encrypt(Key)
	assert.NoError(t, err)

	fmt.Printf("%spass ✔%s\n", log.Green, log.Rtd)

	log.Testf("AES:Decryption...")
	dec, err := obj.Decrypt(Key)
	assert.NoError(t, err)

	fmt.Printf("%spass ✔%s\n", log.Green, log.Rtd)

	assert.Equal(t, obj.plain, dec, "The two words should be the same.")

}
