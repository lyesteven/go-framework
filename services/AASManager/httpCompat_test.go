package main

import "testing"

func TestParseJson(t *testing.T) {
	var json_list = []string{
		`{}`,
		`{"AppKey":"JinJin","Data":"[]","Sign":"d1ef817326852f59990f3d9ed96ee0ec","TimeStamp":1573807776}`,
		`{"AppKey":"JinJin","Data":"[]","Sign":"d1ef817326852f59990f3d9ed96ee0ec","TimeStamp":"1573807776"}`,
		`{"AppKey":"JinJin","Data":"[]","Sign":"d1ef817326852f59990f3d9ed96ee0ec","TimeStamp":"1573807776","UnknownKey":"Unknown"}`,
	}

	for _, j := range json_list {
		_, _, err := parseHttpFormatCompat([]byte(j))
		if err != nil {
			t.Fatal(err)
		}
	}
}
