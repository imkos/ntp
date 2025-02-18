/*
Copyright (c) Facebook, Inc. and its affiliates.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package leapsectz

import (
	"bytes"
	"encoding/binary"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

var tz = []byte{
	'T', 'Z', 'i', 'f', // magic
	0x00, 0x00, 0x00, 0x00, // version
	0x00, 0x00, 0x00, 0x00, // pad
	0x00, 0x00, 0x00, 0x00, // pad
	0x00, 0x00, 0x00, 0x00, // pad
	0x00, 0x00, 0x00, 0x00, // UTC/local
	0x00, 0x00, 0x00, 0x00, // standard/wall
	0x00, 0x00, 0x00, 0x01, // leap
	0x00, 0x00, 0x00, 0x00, // transition
	0x00, 0x00, 0x00, 0x00, // local tz
	0x00, 0x00, 0x00, 0x00, // characters
	0x04, 0xb2, 0x58, 0x00, // leap time
	0x00, 0x00, 0x00, 0x01, // leap count
}

var tzV2 = []byte{
	'T', 'Z', 'i', 'f', // magic
	'2', 0x00, 0x00, 0x00, // version
	0x00, 0x00, 0x00, 0x00, // pad
	0x00, 0x00, 0x00, 0x00, // pad
	0x00, 0x00, 0x00, 0x00, // pad
	0x00, 0x00, 0x00, 0x00, // UTC/local
	0x00, 0x00, 0x00, 0x00, // standard/wall
	0x00, 0x00, 0x00, 0x02, // leap
	0x00, 0x00, 0x00, 0x00, // transition
	0x00, 0x00, 0x00, 0x00, // local tz
	0x00, 0x00, 0x00, 0x00, // characters
	0x04, 0xb2, 0x58, 0x00, // leap time
	0x00, 0x00, 0x00, 0x01, // leap count
	0x05, 0xa4, 0xec, 0x01, // leap time
	0x00, 0x00, 0x00, 0x02, // leap count
	'T', 'Z', 'i', 'f', // magic
	'2', 0x00, 0x00, 0x00, // version
	0x00, 0x00, 0x00, 0x00, // pad
	0x00, 0x00, 0x00, 0x00, // pad
	0x00, 0x00, 0x00, 0x00, // pad
	0x00, 0x00, 0x00, 0x00, // UTC/local
	0x00, 0x00, 0x00, 0x00, // standard/wall
	0x00, 0x00, 0x00, 0x02, // leap
	0x00, 0x00, 0x00, 0x00, // transition
	0x00, 0x00, 0x00, 0x00, // local tz
	0x00, 0x00, 0x00, 0x00, // characters
	0x00, 0x00, 0x00, 0x00, // leap time (first 32 bits)
	0x04, 0xb2, 0x58, 0x00, // leap time (last 32 bits)
	0x00, 0x00, 0x00, 0x01, // leap count
	0x00, 0x00, 0x00, 0x00, // leap time (first 32 bits)
	0x05, 0xa4, 0xec, 0x01, // leap time (last 32 bits)
	0x00, 0x00, 0x00, 0x02, // leap count
	0x00, 0x00, 0x0a, 'U',
	'T', 'C', 0x0a, // 2 bytes of UTC/STD
}

func TestParseV1(t *testing.T) {
	r := bytes.NewReader(tz)

	ls, err := parseVx(r)
	require.NoError(t, err)
	require.Equal(t, 1, len(ls))
	// Saturday, July 1, 1972 12:00:00 AM
	require.Equal(t, uint64(78796800), ls[0].Tleap)
	require.Equal(t, int32(1), ls[0].Nleap)
}

func TestParseV2(t *testing.T) {
	r := bytes.NewReader(tzV2)

	ls, e := parseVx(r)
	if e != nil {
		t.Error(e)
	}

	if len(ls) != 2 {
		t.Errorf("wrong leap second list length")
	}

	// Saturday, July 1, 1972 12:00:00 AM
	require.Equal(t, uint64(78796800), ls[0].Tleap)
	require.Equal(t, int32(1), ls[0].Nleap)
	// January 1, 1973 12:00:00 AM
	require.Equal(t, uint64(94694401), ls[1].Tleap)
	require.Equal(t, int32(2), ls[1].Nleap)
}

func TestParseV2Fail(t *testing.T) {
	tzNoLeapInfo := []byte{
		'T', 'Z', 'i', 'f', // magic
		'2', 0x00, 0x00, 0x00, // version
		0x00, 0x00, 0x00, 0x00, // pad
		0x00, 0x00, 0x00, 0x00, // pad
		0x00, 0x00, 0x00, 0x00, // pad
		0x00, 0x00, 0x00, 0x00, // UTC/local
		0x00, 0x00, 0x00, 0x00, // standard/wall
		0x00, 0x00, 0x00, 0x00, // leap
		0x00, 0x00, 0x00, 0x00, // transition
		0x00, 0x00, 0x00, 0x00, // local tz
		0x00, 0x00, 0x00, 0x00, // characters
		'T', 'Z', 'i', 'f', // magic
		'2', 0x00, 0x00, 0x00, // version
		0x00, 0x00, 0x00, 0x00, // pad
		0x00, 0x00, 0x00, 0x00, // pad
		0x00, 0x00, 0x00, 0x00, // pad
		0x00, 0x00, 0x00, 0x00, // UTC/local
		0x00, 0x00, 0x00, 0x00, // standard/wall
		0x00, 0x00, 0x00, 0x00, // leap
		0x00, 0x00, 0x00, 0x00, // transition
		0x00, 0x00, 0x00, 0x00, // local tz
		0x00, 0x00, 0x00, 0x00, // characters
	}

	r := bytes.NewReader(tzNoLeapInfo)
	ls, err := parseVx(r)
	require.ErrorIs(t, errNoLeapSeconds, err)
	require.Equal(t, 0, len(ls))
}

func TestParse(t *testing.T) {
	expected := []LeapSecond{
		LeapSecond{78796800, 1},
		LeapSecond{94694401, 2},
	}
	f, err := os.CreateTemp(os.TempDir(), "leaptest-")
	require.NoError(t, err)
	defer os.Remove(f.Name())

	_, err = f.Write(tzV2)
	require.NoError(t, err)
	err = f.Close()
	require.NoError(t, err)

	leapFile = f.Name()
	ls, err := Parse("")
	require.NoError(t, err)
	require.ElementsMatch(t, expected, ls)
}

func TestLatest(t *testing.T) {
	expected := &LeapSecond{94694401, 2}
	f, err := os.CreateTemp(os.TempDir(), "leaptest-")
	require.NoError(t, err)
	defer os.Remove(f.Name())

	_, err = f.Write(tzV2)
	require.NoError(t, err)
	err = f.Close()
	require.NoError(t, err)

	ls, err := Latest(f.Name())
	require.NoError(t, err)
	require.Equal(t, expected, ls)
}

func TestLatestFuture(t *testing.T) {
	expected := &LeapSecond{1649346026, 2}

	ls := []LeapSecond{
		{1649346016, 1},
		{1649346026, 2},
		{2649346018, 3},
	}

	f, err := os.CreateTemp(os.TempDir(), "leaptest-")
	require.NoError(t, err)
	defer os.Remove(f.Name())

	err = Write(f, '2', ls, "UTC")
	require.NoError(t, err)

	latest, err := Latest(f.Name())
	require.NoError(t, err)
	require.Equal(t, expected, latest)
}

func TestPrepareHeader(t *testing.T) {
	byteData := []byte{
		'T', 'Z', 'i', 'f', // magic
		'2', 0x00, 0x00, 0x00, // version
		0x00, 0x00, 0x00, 0x00, // pad
		0x00, 0x00, 0x00, 0x00, // pad
		0x00, 0x00, 0x00, 0x00, // pad
		0x00, 0x00, 0x00, 0x01, // UTC/local
		0x00, 0x00, 0x00, 0x01, // standard/wall
		0x00, 0x00, 0x00, 0x01, // leap
		0x00, 0x00, 0x00, 0x00, // transition
		0x00, 0x00, 0x00, 0x01, // local tz
		0x00, 0x00, 0x00, 0x04, // characters
	}

	hdr := prepareHeader('2', 1, "UTC\x00")
	require.Equal(t, byteData, hdr)
}

func TestWritePreData(t *testing.T) {
	byteData := []byte{
		0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 'U', 'T',
		'C', 0x00,
	}
	var b bytes.Buffer
	err := writePreData(&b, "UTC\x00")
	require.NoError(t, err)
	require.Equal(t, byteData, b.Bytes())
}

func TestWritePostData(t *testing.T) {
	byteData := []byte{0x00, 0x00}
	var b bytes.Buffer
	err := writePostData(&b)
	require.NoError(t, err)
	require.Equal(t, byteData, b.Bytes())
}

func TestLeapSecondStructure(t *testing.T) {
	l := LeapSecond{78796800, 1}
	tt := time.Date(1972, time.July, 1, 0, 0, 0, 0, time.UTC)
	require.Equal(t, tt, l.Time().UTC())
}

func TestParserWrongHeaderMagicString(t *testing.T) {
	byteData := []byte{
		'T', 'Z', 'v', '2', // magic
		0x00, 0x00, 0x00, 0x00, // version
		0x00, 0x00, 0x00, 0x00, // pad
		0x00, 0x00, 0x00, 0x00, // pad
		0x00, 0x00, 0x00, 0x00, // pad
		0x00, 0x00, 0x00, 0x00, // UTC/local
		0x00, 0x00, 0x00, 0x00, // standard/wall
		0x00, 0x00, 0x00, 0x01, // leap
		0x00, 0x00, 0x00, 0x00, // transition
		0x00, 0x00, 0x00, 0x00, // local tz
		0x00, 0x00, 0x00, 0x00, // characters
		0x04, 0xb2, 0x58, 0x00, // leap time
		0x00, 0x00, 0x00, 0x01, // leap count
	}

	r := bytes.NewReader(byteData)

	_, err := parseVx(r)
	require.ErrorIs(t, errBadData, err)
}

func TestParserWrongHeaderPadding(t *testing.T) {
	byteData := []byte{
		'T', 'Z', 'i', 'f', // magic
		0x00, 0x00, 0x00, 0x00, // version
		0x00, 0x00, 0x00, 0x00, // pad
		0x00, 0x00, 0x00, 0x00, // pad
	}

	r := bytes.NewReader(byteData)

	_, err := parseVx(r)
	require.ErrorIs(t, errBadData, err)
}

func TestParserWrongHeaderVersion(t *testing.T) {
	byteData := []byte{
		'T', 'Z', 'i', 'f', // magic
		0x02, 0x00, 0x00, 0x00, // version
		0x00, 0x00, 0x00, 0x00, // pad
		0x00, 0x00, 0x00, 0x00, // pad
		0x00, 0x00, 0x00, 0x00, // pad
	}

	r := bytes.NewReader(byteData)

	_, err := parseVx(r)
	require.ErrorIs(t, errUnsupportedVersion, err)
}

func TestReadHeaderStruct(t *testing.T) {
	byteData := []byte{
		0x00, 0x00, 0x00, 0x01, // UTC/local
		0x00, 0x00, 0x00, 0x02, // standard/wall
		0x00, 0x00, 0x00, 0x03, // leap
		0x00, 0x00, 0x00, 0x04, // transition
		0x00, 0x00, 0x00, 0x05, // local tz
		0x00, 0x00, 0x00, 0x06, // characters
		0x04, 0xb2, 0x58, 0x00, // leap time
		0x00, 0x00, 0x00, 0x01, // leap count
	}

	r := bytes.NewReader(byteData)

	var hdr Header
	err := binary.Read(r, binary.BigEndian, &hdr)
	require.NoError(t, err)
	require.Equal(t, uint32(6), hdr.CharCnt, "wrong header - CharCnt")
	require.Equal(t, uint32(5), hdr.TypeCnt, "wrong header - TypeCnt")
	require.Equal(t, uint32(4), hdr.TimeCnt, "wrong header - TimeCnt")
	require.Equal(t, uint32(3), hdr.LeapCnt, "wrong header - LeapCnt")
	require.Equal(t, uint32(2), hdr.IsStdCnt, "wrong header - IsStdCnt")
	require.Equal(t, uint32(1), hdr.IsUtcCnt, "wrong header - IsUtcCnt")
}

func TestWriteV2(t *testing.T) {
	byteData := []byte{
		'T', 'Z', 'i', 'f', // magic
		'2', 0x00, 0x00, 0x00, // version
		0x00, 0x00, 0x00, 0x00, // pad
		0x00, 0x00, 0x00, 0x00, // pad
		0x00, 0x00, 0x00, 0x00, // pad
		0x00, 0x00, 0x00, 0x01, // UTC/local
		0x00, 0x00, 0x00, 0x01, // standard/wall
		0x00, 0x00, 0x00, 0x01, // leap
		0x00, 0x00, 0x00, 0x00, // transition
		0x00, 0x00, 0x00, 0x01, // local tz
		0x00, 0x00, 0x00, 0x04, // characters
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // 6 bytes local tz
		'U', 'T', 'C', 0x00, // timezone chars
		0x04, 0xb2, 0x58, 0x00, // leap time
		0x00, 0x00, 0x00, 0x01, // leap count
		0x00, 0x00, // 2 bytes of UTC/STD
		'T', 'Z', 'i', 'f', // magic
		'2', 0x00, 0x00, 0x00, // version
		0x00, 0x00, 0x00, 0x00, // pad
		0x00, 0x00, 0x00, 0x00, // pad
		0x00, 0x00, 0x00, 0x00, // pad
		0x00, 0x00, 0x00, 0x01, // UTC/local
		0x00, 0x00, 0x00, 0x01, // standard/wall
		0x00, 0x00, 0x00, 0x01, // leap
		0x00, 0x00, 0x00, 0x00, // transition
		0x00, 0x00, 0x00, 0x01, // local tz
		0x00, 0x00, 0x00, 0x04, // characters
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // 6 bytes local tz
		'U', 'T', 'C', 0x00, // timezone chars
		0x00, 0x00, 0x00, 0x00, // leap time (first 32 bits)
		0x04, 0xb2, 0x58, 0x00, // leap time (last 32 bits)
		0x00, 0x00, 0x00, 0x01, // leap count
		0x00, 0x00, 0x0a, 'U',
		'T', 'C', 0x0a, // 2 bytes of UTC/STD
	}

	ls := []LeapSecond{
		{78796800, 1},
	}

	var b bytes.Buffer
	err := Write(&b, '2', ls, "UTC")
	require.NoError(t, err)
	require.Equal(t, byteData, b.Bytes())
}

func TestWriteWrongVersion(t *testing.T) {
	var b bytes.Buffer
	err := Write(&b, '4', []LeapSecond{}, "UTC")
	require.ErrorIs(t, errUnsupportedVersion, err)
}

func FuzzParse(f *testing.F) {
	tzx2 := append(tz, tz...)
	tzV2x2 := append(tzV2, tzV2...)
	for _, seed := range [][]byte{{}, {0}, {9}, tz, tzV2, tzx2, tzV2x2} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, b []byte) {
		r := bytes.NewReader(b)
		ls, err := parseVx(r)
		if err != nil {
			require.Equal(t, 0, len(ls))
		}
	})
}
