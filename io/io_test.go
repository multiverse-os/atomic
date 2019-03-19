package atomicio

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"
)

var testData = "test data for testing the atomic io go library"

func ensureFileContains(name, data string) error {
	fileBytes, err := ioutil.ReadFile(name)
	if err != nil {
		return err
	}
	if string(fileBytes) != data {
		return fmt.Errorf("[test error] wrong data in file: expected %s, got %s", data, string(fileBytes))
	}
	return nil
}

func tempName(count int) string {
	return filepath.Join(os.TempDir(), fmt.Sprintf("scramble-key-%d-%x", count, time.Now().UnixNano()))
}

func testInTempDir() error {
	name := tempName(0)
	defer os.Remove(name)
	f, err := Create(name, 0666)
	if err != nil {
		return err
	}
	if name != f.OriginalName {
		f.Close()
		return fmt.Errorf("[test error] name %q differs from 'OriginalName' attribute: %q", name, f.OriginalName)
	}
	_, err = io.WriteString(f, testData)
	if err != nil {
		f.Close()
		return err
	}
	if err = f.Commit(); err != nil {
		f.Close()
		return err
	}
	//if err = f.Close(); err != nil {
	//	return err
	//}
	return ensureFileContains(name, testData)
}

func TestMakeTempName(t *testing.T) {
	// Make sure temp name is random.
	m := make(map[string]bool)
	for i := 0; i < 100; i++ {
		name, err := makeTempName("/tmp", "temp")
		if err != nil {
			t.Fatal(err)
		}
		if m[name] {
			t.Fatal("[test error] repeated file name")
		}
		m[name] = true
	}
}

func TestFile(t *testing.T) {
	err := testInTempDir()
	if err != nil {
		fmt.Println("[error] in TestFile():", err)
		t.Fatal(err)
	}
}

func TestWriteFile(t *testing.T) {
	name := tempName(1)
	if err := WriteFile(name, []byte(testData), 0666); err != nil {
		t.Fatal(err)
	}
	if err := ensureFileContains(name, testData); err != nil {
		os.Remove(name)
		t.Fatal(err)
	}
	os.Remove(name)
}

func TestAbandon(t *testing.T) {
	name := tempName(2)
	f, err := Create(name, 0666)
	if err != nil {
		t.Fatal(err)
	}
	if err = f.Close(); err != nil {
		t.Fatalf("[test error] abandon failed: %s", err)
	}
	// Make sure temporary file doesn't exist.
	if _, err = os.Stat(f.Name()); err != nil && !os.IsNotExist(err) {
		t.Fatal(err)
	}
}

func TestDoubleCommit(t *testing.T) {
	name := tempName(3)
	f, err := Create(name, 0666)
	if err != nil {
		t.Fatal(err)
	}
	err = f.Commit()
	if err != nil {
		os.Remove(name)
		t.Fatalf("[test error] first commit failed: %s", err)
	}
	err = f.Commit()
	if err != ErrAlreadyCommitted {
		os.Remove(name)
		t.Fatalf("[test error] second commit didn't fail: %s", err)
	}
	err = f.Close()
	//if err != nil {
	//	//os.Remove(name)
	//	t.Fatalf("[test error] closing file failed: %s", err)
	//}
	os.Remove(name)
}

func TestOverwriting(t *testing.T) {
	name := tempName(4)
	defer os.Remove(name)
	olddata := "This is old data"
	if err := ioutil.WriteFile(name, []byte(olddata), 0600); err != nil {
		t.Fatal(err)
	}
	newdata := "This is new data"
	if err := WriteFile(name, []byte(newdata), 0600); err != nil {
		t.Fatal(err)
	}
	if err := ensureFileContains(name, newdata); err != nil {
		t.Fatal(err)
	}
}
