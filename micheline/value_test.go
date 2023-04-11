// Copyright (c) 2021 Blockwatch Data Inc.
// Author: alex@blockwatch.cc
//

// Run with tracing enabled
// go test -tags trace ./micheline/

package micheline

import (
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/pmezard/go-difflib/difflib"
)

const testDataRootPathPrefix = "testdata"

var (
	testfiles = make(map[string][]string)
	testmask  string
)

type testcase struct {
	Name      string          `json:"name"`
	NoUnpack  bool            `json:"no_unpack"`
	Type      json.RawMessage `json:"type"`
	Value     json.RawMessage `json:"value"`
	Key       json.RawMessage `json:"key"`
	TypeHex   string          `json:"type_hex"`
	ValueHex  string          `json:"value_hex"`
	KeyHex    string          `json:"key_hex"`
	WantValue json.RawMessage `json:"want_value"`
	WantKey   json.RawMessage `json:"want_key"`
}

func TestMain(m *testing.M) {
	// call flag.Parse() here if TestMain uses flags
	flag.StringVar(&testmask, "only", "", "limit test to contract or op")
	flag.Parse()
	os.Exit(m.Run())
}

func scanTestFiles(t *testing.T, category string) {
	if len(testfiles[category]) > 0 {
		return
	}
	testfiles[category] = make([]string, 0)
	// find all test data directories
	testPaths := make([]string, 0)
	_ = filepath.WalkDir(".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() || path == "." {
			return nil
		}
		if strings.HasPrefix(filepath.Base(path), testDataRootPathPrefix) {
			testPaths = append(testPaths, path)
		}
		return fs.SkipDir
	})
	// load tests from subdirs
	for _, testPath := range testPaths {
		err := filepath.WalkDir(
			filepath.Join(testPath, category),
			func(path string, d fs.DirEntry, err error) error {
				if err != nil {
					return err
				}
				if d.IsDir() || !strings.HasSuffix(path, "json") {
					return nil
				}
				if testmask != "" && !strings.Contains(path, testmask) {
					return nil
				}
				testfiles[category] = append(testfiles[category], path)
				return nil
			})
		if err != nil {
			t.Logf("WARN: loading testfiles from %s: %v", testPath, err)
		}
	}
}

func jsonDiff(t *testing.T, a, b []byte) bool {
	var ja, jb interface{}
	if err := json.Unmarshal(a, &ja); err != nil {
		t.Error(err)
	}
	if err := json.Unmarshal(b, &jb); err != nil {
		t.Error(err)
	}
	if reflect.DeepEqual(ja, jb) {
		return true
	}

	// log line-wise differences
	textA, _ := json.MarshalIndent(ja, "", "  ")
	textB, _ := json.MarshalIndent(jb, "", "  ")
	diff := difflib.UnifiedDiff{
		A:        difflib.SplitLines(string(textA)),
		B:        difflib.SplitLines(string(textB)),
		FromFile: "GOT",
		ToFile:   "WANT",
		Context:  3,
	}
	text, _ := difflib.GetUnifiedDiffString(diff)
	t.Log("DIFF:\n" + text)
	return false
}

func loadNextTestFile(category string, offset int, val interface{}) (int, error) {
	files, ok := testfiles[category]
	if !ok {
		return 0, fmt.Errorf("invalid category %s", category)
	}
	if len(files) <= offset {
		return offset, io.EOF
	}
	buf, err := os.ReadFile(files[offset])
	if err != nil {
		return offset + 1, err
	}
	err = json.Unmarshal(buf, val)
	return offset + 1, err
}

func checkTypeEncoding(t *testing.T, test testcase) Type {
	// decode type (hex & json and compare trees)
	buf, err := hex.DecodeString(test.TypeHex)
	if err != nil {
		t.Errorf("invalid binary type: %v", err)
		t.FailNow()
	}
	typ1 := Type{}
	if err := typ1.UnmarshalBinary(buf); err != nil {
		t.Errorf("invalid binary type: %v", err)
		t.FailNow()
	}
	typ2 := Type{}
	if err := typ2.UnmarshalJSON(test.Type); err != nil {
		t.Errorf("invalid json type: %v", err)
		t.FailNow()
	}
	// compare prim trees
	if !typ1.IsEqualWithAnno(typ2) {
		b1, _ := typ1.MarshalBinary()
		b2, _ := typ2.MarshalBinary()
		t.Errorf("bigmap type decoding mismatch:\n  want=%s %x\n  have=%s %x",
			typ1.Dump(), b1, typ2.Dump(), b2)
	}
	return typ1
}

func checkValueEncoding(t *testing.T, test testcase) Prim {
	// decode value (hex & json and compare trees)
	buf, err := hex.DecodeString(test.ValueHex)
	if err != nil {
		t.Errorf("invalid binary value: %v", err)
		t.FailNow()
	}
	val1 := Prim{}
	if err := val1.UnmarshalBinary(buf); err != nil {
		t.Errorf("invalid binary value: %v", err)
		t.FailNow()
	}
	val2 := Prim{}
	if err := val2.UnmarshalJSON(test.Value); err != nil {
		t.Errorf("invalid json value: %v", err)
		t.FailNow()
	}
	if !val1.IsEqualWithAnno(val2) {
		b1, _ := val1.MarshalBinary()
		b2, _ := val2.MarshalBinary()
		t.Errorf("json/hex value mismatch:\n  A=%s %x\n  B=%s %x",
			val1.Dump(), b1, val2.Dump(), b2)
		t.FailNow()
	}
	return val1
}

func checkKeyEncoding(t *testing.T, test testcase) Prim {
	// decode key (hex & json and compare trees)
	buf, err := hex.DecodeString(test.KeyHex)
	if err != nil {
		t.Errorf("invalid binary key: %v", err)
		t.FailNow()
	}
	key1 := Prim{}
	if err := key1.UnmarshalBinary(buf); err != nil {
		t.Errorf("invalid binary key: %v", err)
		t.FailNow()
	}
	key2 := Prim{}
	if err := key2.UnmarshalJSON(test.Key); err != nil {
		t.Errorf("invalid json key: %v", err)
		t.FailNow()
	}
	// compare prim trees
	if !key1.IsEqualWithAnno(key2) {
		t.Errorf("json/hex key mismatch:\n  A=%s\n  B=%s\n  a=%#v\n  b=%#v",
			key1.Dump(), key2.Dump(), key1, key2)
		t.FailNow()
	}
	return key1
}

func TestBigmapValues(t *testing.T) {
	var (
		next int
		err  error
	)
	scanTestFiles(t, "bigmap")
	trace = t.Logf
	// dbg = t.Logf
	for {
		var tests []testcase
		next, err = loadNextTestFile("bigmap", next, &tests)
		if err != nil {
			if err == io.EOF {
				break
			}
			t.Error(err)
			if len(tests) == 0 {
				break
			}
			continue
		}
		for _, test := range tests {
			t.Run(test.Name, func(t *testing.T) {
				typ1 := checkTypeEncoding(t, test)
				key1 := checkKeyEncoding(t, test)
				val1 := checkValueEncoding(t, test)

				// test bigmap key
				k, err := NewKey(
					typ1.Left(), // from binary (use key type)
					key1,        // from binary
				)
				if err != nil {
					t.Logf("typ: %s", typ1.Left().Dump())
					t.Logf("key: %s", key1.Dump())
					t.Errorf("key render error: %v", err)
				}
				// try unpack
				if k.IsPacked() && !test.NoUnpack {
					up, err := k.Unpack()
					if err != nil {
						t.Errorf("key unpack error: %v", err)
					}
					k = up
				}
				buf, err := k.MarshalJSON()
				if err != nil {
					t.Errorf("value render error: %v", err)
				}
				if !jsonDiff(t, buf, test.WantKey) {
					t.Error("key render mismatch!")
					t.FailNow()
				}

				// test bigmap value
				v := Value{
					Type:   typ1.Right(), // from binary (use value type)
					Value:  val1,         // from binary
					Render: RENDER_TYPE_FAIL,
				}
				if v.IsPackedAny() && !test.NoUnpack {
					up, err := v.UnpackAll()
					if err != nil {
						t.Errorf("value unpack error: %v", err)
					}
					v = up
				}

				buf, err = v.MarshalJSON()
				if err != nil {
					t.Errorf("value render error: %v", err)
				}
				if !jsonDiff(t, buf, test.WantValue) {
					t.Error("value render mismatch!")
					t.FailNow()
				}
			})
		}
	}
}

func TestStorageValues(t *testing.T) {
	var (
		next int
		err  error
	)
	scanTestFiles(t, "storage")
	trace = t.Logf
	// dbg = t.Logf
	for {
		var tests []testcase
		next, err = loadNextTestFile("storage", next, &tests)
		if err != nil {
			if err == io.EOF {
				break
			}
			t.Error(err)
			if len(tests) == 0 {
				break
			}
			continue
		}
		for _, test := range tests {
			t.Run(test.Name, func(t *testing.T) {
				typ1 := checkTypeEncoding(t, test)
				val1 := checkValueEncoding(t, test)

				// test storage value
				v := Value{
					Type:   typ1, // from binary
					Value:  val1, // from binary
					Render: RENDER_TYPE_FAIL,
				}
				if v.IsPackedAny() && !test.NoUnpack {
					up, err := v.UnpackAll()
					if err != nil {
						t.Errorf("value unpack error: %v", err)
						t.FailNow()
					}
					v = up
				}

				buf, err := v.MarshalJSON()
				if err != nil {
					t.Errorf("value render error: %v", err)
					t.FailNow()
				}
				if !jsonDiff(t, buf, test.WantValue) {
					t.Error("value render mismatch, see log for details")
					t.FailNow()
				}
			})
		}
	}
}

func TestParamsValues(t *testing.T) {
	var (
		next int
		err  error
	)
	scanTestFiles(t, "params")
	trace = t.Logf
	// dbg = t.Logf
	for {
		var tests []testcase
		next, err = loadNextTestFile("params", next, &tests)
		if err != nil {
			if err == io.EOF {
				break
			}
			t.Error(err)
			if len(tests) == 0 {
				break
			}
			continue
		}
		for _, test := range tests {
			t.Run(test.Name, func(t *testing.T) {
				typ1 := checkTypeEncoding(t, test)
				val1 := checkValueEncoding(t, test)

				// test storage value
				v := Value{
					Type:   typ1, // from binary
					Value:  val1, // from binary
					Render: RENDER_TYPE_FAIL,
				}
				if v.IsPackedAny() && !test.NoUnpack {
					up, err := v.UnpackAll()
					if err != nil {
						t.Errorf("value unpack error: %v", err)
					}
					v = up
				}

				buf, err := v.MarshalJSON()
				if err != nil {
					t.Errorf("value render error: %v", err)
				}
				if !jsonDiff(t, buf, test.WantValue) {
					t.Error("value render mismatch, see log for details")
					t.FailNow()
				}
			})
		}
	}
}
