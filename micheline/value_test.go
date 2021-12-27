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
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/pmezard/go-difflib/difflib"
)

const testDataRootPathPrefix = "testdata"

var (
	testcats  = []string{"bigmap", "storage", "params"}
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
	err := filepath.WalkDir(".", func(path string, d fs.DirEntry, err error) error {
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
		err = filepath.WalkDir(
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
			t.Fatalf("loading testfiles from %s: %v", testPath, err)
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
	buf, err := ioutil.ReadFile(files[offset])
	if err != nil {
		return offset + 1, err
	}
	err = json.Unmarshal(buf, val)
	return offset + 1, err
}

func checkTypeEncoding(T *testing.T, test testcase) Type {
	// decode type (hex & json and compare trees)
	buf, err := hex.DecodeString(test.TypeHex)
	if err != nil {
		T.Errorf("invalid binary type: %v", err)
		T.FailNow()
	}
	typ1 := Type{}
	if err := typ1.UnmarshalBinary(buf); err != nil {
		T.Errorf("invalid binary type: %v", err)
		T.FailNow()
	}
	typ2 := Type{}
	if err := typ2.UnmarshalJSON(test.Type); err != nil {
		T.Errorf("invalid json type: %v", err)
		T.FailNow()
	}
	// compare prim trees
	if !typ1.IsEqualWithAnno(typ2) {
		T.Errorf("bigmap type decoding mismatch:\n  want=%s\n  have=%s", typ1.Dump(), typ2.Dump())
	}
	return typ1
}

func checkValueEncoding(T *testing.T, test testcase) Prim {
	// decode value (hex & json and compare trees)
	buf, err := hex.DecodeString(test.ValueHex)
	if err != nil {
		T.Errorf("invalid binary value: %v", err)
		T.FailNow()
	}
	val1 := Prim{}
	if err := val1.UnmarshalBinary(buf); err != nil {
		T.Errorf("invalid binary value: %v", err)
		T.FailNow()
	}
	val2 := Prim{}
	if err := val2.UnmarshalJSON(test.Value); err != nil {
		T.Errorf("invalid json value: %v", err)
		T.FailNow()
	}
	if !val1.IsEqualWithAnno(val2) {
		T.Errorf("json/hex value mismatch:\n  A=%s\n  B=%s", val1.Dump(), val2.Dump())
		T.FailNow()
	}
	return val1
}

func checkKeyEncoding(T *testing.T, test testcase) Prim {
	// decode key (hex & json and compare trees)
	buf, err := hex.DecodeString(test.KeyHex)
	if err != nil {
		T.Errorf("invalid binary key: %v", err)
		T.FailNow()
	}
	key1 := Prim{}
	if err := key1.UnmarshalBinary(buf); err != nil {
		T.Errorf("invalid binary key: %v", err)
		T.FailNow()
	}
	key2 := Prim{}
	if err := key2.UnmarshalJSON(test.Key); err != nil {
		T.Errorf("invalid json key: %v", err)
		T.FailNow()
	}
	// compare prim trees
	if !key1.IsEqualWithAnno(key2) {
		T.Errorf("json/hex key mismatch:\n  A=%s\n  B=%s\n  a=%#v\n  b=%#v",
			key1.Dump(), key2.Dump(), key1, key2)
		T.FailNow()
	}
	return key1
}

func TestBigmapValues(t *testing.T) {
	var (
		next int
		err  error
	)
	UseTrace(t.Logf)
	scanTestFiles(t, "bigmap")
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
			t.Run(test.Name, func(T *testing.T) {
				typ1 := checkTypeEncoding(T, test)
				key1 := checkKeyEncoding(T, test)
				val1 := checkValueEncoding(T, test)

				// test bigmap key
				k, err := NewKey(
					typ1.Left(), // from binary (use key type)
					key1,        // from binary
				)
				if err != nil {
					T.Errorf("key render error: %v", err)
				}
				// try unpack
				if k.IsPacked() && !test.NoUnpack {
					up, err := k.Unpack()
					if err != nil {
						T.Errorf("key unpack error: %v", err)
					}
					k = up
				}
				buf, err := k.MarshalJSON()
				if err != nil {
					T.Errorf("value render error: %v", err)
				}
				if !jsonDiff(t, buf, test.WantKey) {
					T.Error("key render mismatch, see log for details")
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
						T.Errorf("value unpack error: %v", err)
					}
					v = up
				}

				buf, err = v.MarshalJSON()
				if err != nil {
					T.Errorf("value render error: %v", err)
				}
				if !jsonDiff(t, buf, test.WantValue) {
					T.Error("value render mismatch, see log for details")
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
	UseTrace(t.Logf)
	scanTestFiles(t, "storage")
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
			t.Run(test.Name, func(T *testing.T) {
				typ1 := checkTypeEncoding(T, test)
				val1 := checkValueEncoding(T, test)

				// test storage value
				v := Value{
					Type:   typ1, // from binary
					Value:  val1, // from binary
					Render: RENDER_TYPE_FAIL,
				}
				if v.IsPackedAny() && !test.NoUnpack {
					up, err := v.UnpackAll()
					if err != nil {
						T.Errorf("value unpack error: %v", err)
						t.FailNow()
					}
					v = up
				}

				buf, err := v.MarshalJSON()
				if err != nil {
					T.Errorf("value render error: %v", err)
					t.FailNow()
				}
				if !jsonDiff(t, buf, test.WantValue) {
					T.Error("value render mismatch, see log for details")
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
	UseTrace(t.Logf)
	scanTestFiles(t, "params")
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
			t.Run(test.Name, func(T *testing.T) {
				typ1 := checkTypeEncoding(T, test)
				val1 := checkValueEncoding(T, test)

				// test storage value
				v := Value{
					Type:   typ1, // from binary
					Value:  val1, // from binary
					Render: RENDER_TYPE_FAIL,
				}
				if v.IsPackedAny() && !test.NoUnpack {
					up, err := v.UnpackAll()
					if err != nil {
						T.Errorf("value unpack error: %v", err)
					}
					v = up
				}

				buf, err := v.MarshalJSON()
				if err != nil {
					T.Errorf("value render error: %v", err)
				}
				if !jsonDiff(t, buf, test.WantValue) {
					T.Error("value render mismatch, see log for details")
					t.FailNow()
				}
			})
		}
	}
}
