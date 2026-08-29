// go-qrcode
// Copyright 2014 Tom Harwood

package qrcode

import (
	"testing"

	bitset "github.com/RiddlerArgentina/go-qrcode/bitset"
)

func TestBuildRegularSymbol(t *testing.T) {
	v := getQRCodeVersion(Low, 1)
	if v == nil {
		t.Fatal("missing version 1")
	}

	data := bitset.New()
	for range 26 {
		data.AppendNumBools(8, false)
	}

	for mask := 0; mask <= 7; mask++ {
		s, err := buildRegularSymbol(*v, mask, data, true)
		if err != nil {
			t.Fatalf("mask %d: %s", mask, err)
		}
		if s.size != 21+8 {
			t.Errorf("mask %d: size = %d, want 29", mask, s.size)
		}
		if s.get(0, 0) != true {
			t.Errorf("mask %d: missing top-left finder pattern", mask)
		}
		if s.numEmptyModules() != 0 {
			t.Errorf("mask %d: %d empty modules", mask, s.numEmptyModules())
		}
	}
}

func TestBuildRegularSymbolNoQuietZone(t *testing.T) {
	v := getQRCodeVersion(Low, 1)
	if v == nil {
		t.Fatal("missing version 1")
	}

	data := bitset.New()
	for range 26 {
		data.AppendNumBools(8, false)
	}

	s, err := buildRegularSymbol(*v, 0, data, false)
	if err != nil {
		t.Fatal(err)
	}
	if s.size != 21 {
		t.Errorf("size = %d, want 21", s.size)
	}
	if s.get(0, 0) != true {
		t.Error("missing top-left finder pattern")
	}
}

func TestSymbolSizeMatchesVersion(t *testing.T) {
	for version := 1; version <= 40; version++ {
		v := getQRCodeVersion(Low, version)
		want := 21 + (version-1)*4
		if v.symbolSize() != want {
			t.Errorf("version %d symbolSize = %d, want %d", version, v.symbolSize(), want)
		}
	}
}
