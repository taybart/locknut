package locknut

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"github.com/taybart/log"
	"go.etcd.io/bbolt"
	"io"
	"os"
	"path/filepath"
)

type boltDB struct {
	*bbolt.DB
}

// The BoltLocknut struct, all fields are not needed to be accessed by other packages
type BoltLocknut struct {
	name      string
	path      string
	fullPath  string
	secret    []byte
	buckets   []string
	batchMode bool
	db        *boltDB
}

// The key error messages generated in the package
var (
	ErrFileNameInvalid = errors.New("invalid file name")
	ErrPathInvalid     = errors.New("invalid path name")
	ErrKeyInvalid      = errors.New("invalid key or key is nil")
)

// NewBoltLocknut The main function to initialize the the DB manager for all DB related operations
// 	name: the db file name, such as mydb.dat, mytest.db
// 	path: the db file's path, can be "" or any other director
// 	secret: the secret value if you want to encrypt the values; if you don't want to encrypt the data, simply put it as ""
// 	batchMode: to control whether to close the db file after each db operation
// 	buckets: the buckets in the db file to be initialized if the db file does not existed
func NewBoltLocknut(name, path string, secret []byte, batchMode bool, buckets []string) (*BoltLocknut, error) {
	var info os.FileInfo
	if path != "" {
		info, err := os.Stat(path)
		if err != nil || !info.Mode().IsDir() {
			return nil, ErrPathInvalid
		}
	}

	fullPath := filepath.Join(path, name)
	info, err := os.Stat(fullPath)
	if err == nil && !info.Mode().IsRegular() {
		return nil, ErrFileNameInvalid
	}

	if err != nil {
		log.Debugf("NewBoltLocknut DB file %s does not exist, will be created", fullPath)
	}

	bl := &BoltLocknut{
		name:      name,
		path:      path,
		fullPath:  fullPath,
		batchMode: batchMode,
		buckets:   buckets,
	}

	bl.SetSecret(secret)

	if err = bl.openDB(); err != nil {
		return nil, err
	}
	defer bl.closeDB()

	return bl, err
}

// SetSecret is to set the AES Cryptor key, if the key is nil, the cryptor is not initialized; otherwise
// the cryptor is initialized, including the key and Cipher block that can be used directly for encrypt and decrypt functions
func (bl *BoltLocknut) SetSecret(secret []byte) {
	bl.secret = secret
	if len(secret) < 32 {
		log.Verbose("Key too short, using hash")
		sh := sha256.Sum256(secret)
		bl.secret = sh[:]
	}
}

// SetBatchMode is to set the batchMode for the boltdb. The boltdb file is always open in the file system unless the Close() is called.
// This cause inconvenience if you want to do some file operation to the db file while the program is running. Thus if the batchMode is
// set to false, the db will be closed after each db operation, this could reduce a certain performance. Thus if you have a lots of db
// operations to execute, you can set the batchMode to be true before those operations.
func (bl *BoltLocknut) SetBatchMode(mode bool) {
	bl.batchMode = mode
	//if the batch mode is turned off, close DB directly
	if !mode {
		bl.closeDB()
	}
}

// This function creates the db file if it doesn't exist, and also initialize the buckets
func (bl *BoltLocknut) openDB() error {
	if bl.batchMode && bl.db != nil {
		return nil
	}

	d, err := bbolt.Open(bl.fullPath, 0600, nil)
	if err != nil {
		return err
	}

	db := &boltDB{d}

	initbuckets := func(tx *bbolt.Tx) error {
		for _, bname := range bl.buckets {
			if _, err := tx.CreateBucketIfNotExists([]byte(bname)); err != nil {
				return err
			}
		}
		return nil
	}

	if err = db.update(initbuckets); err != nil {
		db.Close()
		return err
	}

	bl.db = db
	return nil
}

// The closeDB function closes the db when the bl.db is not nil and the batchmode is false.
// When the bl batchmode is true, please set it to be false in order to close the DB.
func (bl *BoltLocknut) closeDB() {
	if !bl.batchMode && bl.db != nil {
		bl.db.Close()
		bl.db = nil
	}
}

// The view function is to retrieve the records
func (db *boltDB) view(fn func(*bbolt.Tx) error) error {
	wrapper := func(tx *bbolt.Tx) error {
		return fn(tx)
	}
	return db.DB.View(wrapper)
}

// The update function applies changes to the database. There can be only one Update at a time.
func (db *boltDB) update(fn func(*bbolt.Tx) error) error {
	wrapper := func(tx *bbolt.Tx) error {
		return fn(tx)
	}
	return db.DB.Update(wrapper)
}

// GetByPrefix function returns the byte arrays for those records matched with specified Prefix. If the secret is set,
// the function returns the decrypted content.
func (bl *BoltLocknut) GetByPrefix(bucket, prefix string) (map[string][]byte, error) {
	var err error
	if err = bl.openDB(); err != nil {
		return nil, err
	}
	defer bl.closeDB()

	results := make(map[string][]byte)

	seekPrefix := func(tx *bbolt.Tx) error {
		prefixKey := []byte(prefix)

		bkt := tx.Bucket([]byte(bucket))

		if bkt == nil {
			return bbolt.ErrBucketNotFound
		}

		cursor := bkt.Cursor()
		for k, v := cursor.Seek(prefixKey); bytes.HasPrefix(k, prefixKey); k, v = cursor.Next() {
			if len(prefixKey) == 0 && len(k) == 0 && len(v) == 0 { // corner case
				break
			}
			if bl.secret != nil {
				//secret key is set, decrypt the content before return
				content := make([]byte, len(v))
				copy(content, v)

				dec, err := Decrypt(content, bl.secret)
				if err != nil {
					return errors.New("Decrypt error from db " + err.Error())
				}
				results[string(k)] = dec
			} else {
				results[string(k)] = v
			}
		}
		return nil
	}

	if err = bl.db.view(seekPrefix); err != nil {
		log.Error("GetByPrefix return", err)
	}

	return results, err
}

// GetKeyList function returns the string array for keys with specified Prefix.
func (bl *BoltLocknut) GetKeyList(bucket, prefix string) ([]string, error) {
	var err error
	var results []string
	if err = bl.openDB(); err != nil {
		return nil, err
	}
	defer bl.closeDB()

	results = make([]string, 0)

	seekPrefix := func(tx *bbolt.Tx) error {
		prefixKey := []byte(prefix)

		bkt := tx.Bucket([]byte(bucket))

		if bkt == nil {
			return bbolt.ErrBucketNotFound
		}

		cursor := bkt.Cursor()
		for k, _ := cursor.Seek(prefixKey); k != nil && bytes.HasPrefix(k, prefixKey); k, _ = cursor.Next() {
			results = append(results, string(k))
		}
		return nil
	}

	if err = bl.db.view(seekPrefix); err != nil {
		log.Error("GetByPrefix return", err)
	}

	return results, err
}

// GetOne function returns the first record containing the key, If the secret is set,
// the function returns the decrypted content.
func (bl *BoltLocknut) GetOne(bucket, key string) ([]byte, error) {
	var err error
	var result []byte

	if err = bl.openDB(); err != nil {
		return nil, err
	}
	defer bl.closeDB()

	if key == "" {
		return nil, ErrKeyInvalid
	}

	seek := func(tx *bbolt.Tx) error {
		prefixKey := []byte(key)

		bkt := tx.Bucket([]byte(bucket))

		if bkt == nil {
			return bbolt.ErrBucketNotFound
		}

		cursor := bkt.Cursor()
		k, v := cursor.Seek(prefixKey)

		content := make([]byte, len(v))
		copy(content, v)

		if k != nil && bytes.HasPrefix(k, prefixKey) {
			if bl.secret != nil {
				dec, err := Decrypt(content, bl.secret)
				if err != nil {
					return errors.New("Decrypt error from db " + err.Error())
				}
				result = dec
			} else {
				result = content
			}
		}

		return nil
	}

	if err := bl.db.view(seek); err != nil {
		return nil, err
	}

	return result, nil
}

// Save function stores the record into the db file. If the secret value is set, the function
// encrypts the content before storing into the db.
func (bl *BoltLocknut) Save(bucket, key string, data interface{}) error {
	var err error

	if err = bl.openDB(); err != nil {
		return err
	}
	defer bl.closeDB()

	if data == nil {
		return errors.New("data is nil")
	}

	save := func(tx *bbolt.Tx) error {
		var err error
		bkt := tx.Bucket([]byte(bucket))

		value, err := json.Marshal(data)
		if err != nil {
			return err
		}
		if bl.secret != nil {
			//encrypt the content before store in the db
			enc, err := Encrypt(value, bl.secret)
			if err != nil {
				return errors.New("Encrypt error from db " + err.Error())
			}

			if err = bkt.Put([]byte(key), enc); err != nil {
				return err
			}
		} else {
			if err = bkt.Put([]byte(key), value); err != nil {
				return err
			}
		}

		return nil
	}

	return bl.db.update(save)
}

// SaveBytes function stores the record into the db file. If the secret value is set, the function
// encrypts the content before storing into the db.
func (bl *BoltLocknut) SaveBytes(bucket, key string, data []byte) error {
	var err error

	if err = bl.openDB(); err != nil {
		return err
	}
	defer bl.closeDB()

	if data == nil {
		return errors.New("data is nil")
	}

	save := func(tx *bbolt.Tx) error {
		var err error
		bkt := tx.Bucket([]byte(bucket))

		if bl.secret != nil {
			//encrypt the content before store in the db
			enc, err := Encrypt(data, bl.secret)
			if err != nil {
				return errors.New("Encrypt error from db " + err.Error())
			}

			if err = bkt.Put([]byte(key), enc); err != nil {
				return err
			}
		} else {
			if err = bkt.Put([]byte(key), data); err != nil {
				return err
			}
		}

		return nil
	}

	return bl.db.update(save)
}

// Delete function deletes the record specified by the key.
func (bl *BoltLocknut) Delete(bucket, key string) error {
	var err error

	if err = bl.openDB(); err != nil {
		return err
	}
	defer bl.closeDB()

	if key == "" {
		return errors.New("cannot delete, key is nil")
	}

	delete := func(tx *bbolt.Tx) error {
		bkt := tx.Bucket([]byte(bucket))
		if err := bkt.Delete([]byte(key)); err != nil {
			return err
		}
		return nil
	}

	return bl.db.update(delete)
}

// GetDBBytes extracts a byte representation of db
// @TODO: move to exporting directly to stream writer
func (bl BoltLocknut) GetDBBytes() []byte {
	r, w := io.Pipe()
	var buf bytes.Buffer
	go bl.db.View(func(tx *bbolt.Tx) error {
		defer w.Close()
		_, err := tx.WriteTo(w)
		return err
	})
	buf.ReadFrom(r)
	return buf.Bytes()
}
