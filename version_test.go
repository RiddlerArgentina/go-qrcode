// go-qrcode
// Copyright 2014 Tom Harwood

package qrcode

import (
	"testing"

	bitset "github.com/RiddlerArgentina/go-qrcode/bitset"
)

func TestFormatInfo(t *testing.T) {
	tests := []struct {
		level       RecoveryLevel
		maskPattern int

		expected uint32
	}{
		{ // L=01 M=00 Q=11 H=10
			Low,
			1,
			0x72f3,
		},
		{
			Medium,
			2,
			0x5e7c,
		},
		{
			High,
			3,
			0x3a06,
		},
		{
			Highest,
			4,
			0x0762,
		},
		{
			Low,
			5,
			0x6318,
		},
		{
			Medium,
			6,
			0x4f97,
		},
		{
			High,
			7,
			0x2bed,
		},
	}

	for i, test := range tests {
		v := getQRCodeVersion(test.level, 1)

		result := v.formatInfo(test.maskPattern)

		expected := bitset.New()
		expected.AppendUint32(test.expected, formatInfoLengthBits)

		if !expected.Equals(result) {
			t.Errorf("formatInfo test #%d got %s, expected %s", i, result.String(),
				expected.String())
		}
	}
}

func TestVersionInfo(t *testing.T) {
	tests := []struct {
		version  int
		expected uint32
	}{
		{
			7,
			0x007c94,
		},
		{
			10,
			0x00a4d3,
		},
		{
			20,
			0x0149a6,
		},
		{
			30,
			0x01ed75,
		},
		{
			40,
			0x028c69,
		},
	}

	for i, test := range tests {
		var v *qrCodeVersion

		v = getQRCodeVersion(Low, test.version)

		result := v.versionInfo()

		expected := bitset.New()
		expected.AppendUint32(test.expected, versionInfoLengthBits)

		if !expected.Equals(result) {
			t.Errorf("versionInfo test #%d got %s, expected %s", i, result.String(),
				expected.String())
		}
	}
}

func TestGetQRCodeVersion(t *testing.T) {
	if getQRCodeVersion(Low, 0) != nil {
		t.Error("version 0 should be undefined")
	}
	if getQRCodeVersion(Low, 41) != nil {
		t.Error("version 41 should be undefined")
	}
	if getQRCodeVersion(RecoveryLevel(99), 1) != nil {
		t.Error("invalid recovery level should be undefined")
	}

	for version := 1; version <= 40; version++ {
		for _, level := range []RecoveryLevel{Low, Medium, High, Highest} {
			v := getQRCodeVersion(level, version)
			if v == nil {
				t.Errorf("missing version %d level %v", version, level)
				continue
			}
			if v.version != version || v.level != level {
				t.Errorf("got version=%d level=%v, want %d %v", v.version, v.level, version, level)
			}
			if v.numDataBits() <= 0 {
				t.Errorf("version %d level %v has no data bits", version, level)
			}
			if v.numBlocks() <= 0 {
				t.Errorf("version %d level %v has no blocks", version, level)
			}
		}
	}
}

func TestVersionInfoNilBelowVersion7(t *testing.T) {
	v := getQRCodeVersion(Low, 6)
	if v.versionInfo() != nil {
		t.Error("version 6 should not have version information")
	}
	v = getQRCodeVersion(Low, 7)
	if v.versionInfo() == nil {
		t.Error("version 7 should have version information")
	}
}

func TestNumBitsToPadToCodeoword(t *testing.T) {
	tests := []struct {
		level   RecoveryLevel
		version int

		numDataBits int
		expected    int
	}{
		{
			Low,
			1,
			0,
			0,
		}, {
			Low,
			1,
			1,
			7,
		}, {
			Low,
			1,
			7,
			1,
		}, {
			Low,
			1,
			8,
			0,
		},
	}

	for i, test := range tests {
		var v *qrCodeVersion

		v = getQRCodeVersion(test.level, test.version)

		result := v.numBitsToPadToCodeword(test.numDataBits)

		if result != test.expected {
			t.Errorf("numBitsToPadToCodeword test %d (version=%d numDataBits=%d), got %d, expected %d",
				i, test.version, test.numDataBits, result, test.expected)
		}
	}
}
