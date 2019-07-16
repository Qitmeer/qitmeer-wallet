package utils

import (
	"testing"
)

func TestOpen(t *testing.T) {
	//Open("www.baidu.com")
	OpenBrowser("http://www.baidu.com")
}

func TestUserDataDir(t *testing.T) {

	t.Log(GetUserDataDir())

	t.Log(MakeDirAll(GetUserDataDir()))
}
