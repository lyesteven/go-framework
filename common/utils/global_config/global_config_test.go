package global_config

import "testing"

func TestGetString(t *testing.T) {
	Init()
	Init()
	var r, err = GetString("HotMall", "General","LogPath")
	if err != nil {
		t.Fatal(err)
	}

	if r != "/data/logs" {
		t.Fatal(r)
	}
}

func TestGetInt(t *testing.T) {
	Init()
	Init()
	var r, err = GetInt("HotMall", "mysql","port")
	if err != nil {
		t.Fatal(err)
	}

	if r != 3306 {
		t.Fatal(r)
	}
}
