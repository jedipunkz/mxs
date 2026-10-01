package main

import "testing"

func TestUUIDFormat(t *testing.T) {
	if got, want := floodgateUUID(2535428654845283), "00000000-0000-0000-0009-01f57c534d63"; got != want {
		t.Errorf("floodgateUUID = %s, want %s", got, want)
	}
	if got, err := parseFloodgateUUID("00000000-0000-0000-0009-01f57c534d63"); err != nil || got != 2535428654845283 {
		t.Errorf("parseFloodgateUUID = %d, %v", got, err)
	}
	if _, err := parseFloodgateUUID("069a79f4-44e9-4726-a5be-fca90e38aaf5"); err == nil {
		t.Error("parseFloodgateUUID accepted a non-floodgate uuid")
	}
	for _, s := range []string{"2535414915229641", "901f2496167c9", "901F2496167C9", "0x901f2496167c9", "00000000-0000-0000-0009-01f2496167c9", "0000000000000000000901f2496167c9"} {
		if got, err := parseXUID(s); err != nil || got != 2535414915229641 {
			t.Errorf("parseXUID(%s) = %d, %v", s, got, err)
		}
	}
	if got, want := bedrockInfo("Dream", 2535414915229641), "Gamertag: Dream\nXUID(DEC): 2535414915229641\nXUID(HEX): 901f2496167c9\nFloodgate UUID: 00000000-0000-0000-0009-01f2496167c9"; got != want {
		t.Errorf("bedrockInfo = %q, want %q", got, want)
	}
	for s, want := range map[string]bool{
		"069a79f4-44e9-4726-a5be-fca90e38aaf5": true,
		"069A79F444E94726A5BEFCA90E38AAF5":     true,
		"00000000-0000-0000-0009-01f2496167c9": false, // floodgate
		"901f2496167c9":                        false, // hex xuid
		"2535414915229641":                     false, // dec xuid
	} {
		if _, got := javaUUID(s); got != want {
			t.Errorf("javaUUID(%s) = %v, want %v", s, got, want)
		}
	}
	if got, want := dashed("069a79f444e94726a5befca90e38aaf5"), "069a79f4-44e9-4726-a5be-fca90e38aaf5"; got != want {
		t.Errorf("dashed = %s, want %s", got, want)
	}
}
