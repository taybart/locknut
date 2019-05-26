package locknut

import (
	"encoding/json"
	"os"
	"testing"
)

type Article struct {
	ID    string `json:"id"`
	Title string `json:"title"`
}

func ExampleNewBoltLocknut() {
	var err error
	bucketName := "article"

	bl, err := NewBoltLocknut("test.db", ".", []byte("secret"), false, []string{bucketName})
	if err != nil {
		// Handle the error
	}

	var bytes []byte
	if bytes, err = bl.GetOne(bucketName, "data.ID"); err != nil {
		// handle the error
	}

	resNew := new(Article)
	if err = json.Unmarshal(bytes, resNew); err != nil {
		// handle the error
	}
}

func TestBLCreate(t *testing.T) {
	var err error
	bucketName := "article"

	bl, err := NewBoltLocknut("test.db", ".", []byte("secret"), false, []string{bucketName})
	if err != nil {
		t.Errorf("NewBoltLocknut: %s", err)
	}

	data := Article{
		ID:    "ID-0001",
		Title: "input with more than 16 characters",
	}

	if err = bl.Save(bucketName, data.ID, data); err != nil {
		t.Errorf("TestBLCreate save data return err: %s", err)
	}

	var bytes []byte
	if bytes, err = bl.GetOne(bucketName, data.ID); err != nil {
		t.Errorf("TestBLCreate GetOne return err: %s", err)
	}

	resNew := new(Article)
	if err = json.Unmarshal(bytes, resNew); err != nil {
		t.Errorf("json.Unmarshal return err: %s", err)
	}

	if resNew.Title != data.Title {
		t.Errorf("returned Title is not equal: new:%s, org: %s", resNew.Title, data.Title)
	}

	if err = bl.Delete(bucketName, data.ID); err != nil {
		t.Errorf("TestDBMCreate Delete return err: %s", err)
	}

	os.Remove("test.db")
	return
}

func BenchmarkDBMOps(b *testing.B) {
	var err error
	bucketName := "article"

	bl, err := NewBoltLocknut("test.db", ".", []byte("secret"), false, []string{bucketName})
	if err != nil {
		b.Errorf("BenchmarkDBMOps: %s", err)
	}

	data := Article{
		ID:    "ID-0001",
		Title: "input with more than 16 characters",
	}

	for n := 0; n < b.N; n++ {
		if err = bl.Save(bucketName, data.ID, data); err != nil {
			b.Errorf("TestDBMCreate save data return err: %s", err)
		}

		var bytes []byte
		if bytes, err = bl.GetOne(bucketName, data.ID); err != nil {
			b.Errorf("TestDBMCreate GetOne return err: %s", err)
		}

		resNew := new(Article)
		if err = json.Unmarshal(bytes, resNew); err != nil {
			b.Errorf("json.Unmarshal return err: %s", err)
		}

		if resNew.Title != data.Title {
			b.Errorf("returned Title is not equal: new:%s, org: %s", resNew.Title, data.Title)
		}

		if err = bl.Delete(bucketName, data.ID); err != nil {
			b.Errorf("TestDBMCreate Delete return err: %s", err)
		}
	}
	os.Remove("test.db")
}

func BenchmarkDBMOpsBatchMode(b *testing.B) {
	var err error
	bucketName := "article"

	// dbm, err := NewDBManager("test.db", ".", []byte(""), true, []string{bucketName})
	bl, err := NewBoltLocknut("test.db", ".", []byte(""), true, []string{bucketName})
	if err != nil {
		b.Errorf("BenchmarkDBMOps: %s", err)
	}

	data := Article{
		ID:    "ID-0001",
		Title: "input with more than 16 characters",
	}

	for n := 0; n < b.N; n++ {
		if err = bl.Save(bucketName, data.ID, data); err != nil {
			b.Errorf("BenchmarkDBMOpsBatchMode save data return err: %s", err)
		}

		var bytes []byte
		if bytes, err = bl.GetOne(bucketName, data.ID); err != nil {
			b.Errorf("BenchmarkDBMOpsBatchMode GetOne return err: %s", err)
		}

		resNew := new(Article)
		if err = json.Unmarshal(bytes, resNew); err != nil {
			b.Errorf("json.Unmarshal return err: %s", err)
		}

		if resNew.Title != data.Title {
			b.Errorf("returned Title is not equal: new:%s, org: %s", resNew.Title, data.Title)
		}

		if err = bl.Delete(bucketName, data.ID); err != nil {
			b.Errorf("BenchmarkDBMOpsBatchMode Delete return err: %s", err)
		}
	}

	bl.SetBatchMode(false)
	os.Remove("test.db")
}

func BenchmarkDBMOpsNoEncryption(b *testing.B) {
	var err error
	bucketName := "article"

	// dbm, err := NewDBManager("test.db", ".", []byte(""), false, []string{bucketName})
	bl, err := NewBoltLocknut("test.db", ".", []byte(""), false, []string{bucketName})
	if err != nil {
		b.Errorf("BenchmarkDBMOpsNoEncryption: %s", err)
	}

	data := Article{
		ID:    "ID-0001",
		Title: "input with more than 16 characters",
	}

	for n := 0; n < b.N; n++ {
		if err = bl.Save(bucketName, data.ID, data); err != nil {
			b.Errorf("BenchmarkDBMOpsNoEncryption save data return err: %s", err)
		}

		var bytes []byte
		if bytes, err = bl.GetOne(bucketName, data.ID); err != nil {
			b.Errorf("BenchmarkDBMOpsNoEncryption GetOne return err: %s", err)
		}

		resNew := new(Article)
		if err = json.Unmarshal(bytes, resNew); err != nil {
			b.Errorf("json.Unmarshal return err: %s", err)
		}

		if resNew.Title != data.Title {
			b.Errorf("returned Title is not equal: new:%s, org: %s", resNew.Title, data.Title)
		}

		if err = bl.Delete(bucketName, data.ID); err != nil {
			b.Errorf("BenchmarkDBMOpsNoEncryption Delete return err: %s", err)
		}
	}
	os.Remove("test.db")
}
