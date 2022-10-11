//go:build windows
// +build windows

package link

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/candrewlee14/webman/pkgparse"
	"github.com/candrewlee14/webman/utils"
)

func TestGetLinkPathIfExec(t *testing.T) {

	tmp := t.TempDir()
	utils.Init(tmp)

	exe := filepath.Join(tmp, "executable")
	fmt.Println(exe)
	file, err := os.Create(exe)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("File created successfully")
	defer file.Close()

	err = os.Chmod(exe, 0700)
	if err != nil {
		log.Fatal(err)
	}

	linkPath := GetLinkPathIfExec("executable", []pkgparse.RenameItem{})
	fmt.Printf("----> %v \n", linkPath)
}
