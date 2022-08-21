package PlatformUrl

import "testing"

func TestGetUrl(t *testing.T) {
	if _, err := GetPlatformUrl(WODA); err != nil {
		t.Fatalf("%s error", WODA)
	}

	if _, err := GetPlatformUrl(BASE); err != nil {
		t.Fatalf("%s error", BASE)
	}

	if _, err := GetPlatformUrl(JFF); err != nil {
		t.Fatalf("%s error", JFF)
	}

	if _, err := GetPlatformUrl(ZXX); err != nil {
		t.Fatalf("%s error", ZXX)
	}

	if _, err := GetPlatformUrl(BIZ); err != nil {
		t.Fatalf("%s error", BIZ)
	}
}
