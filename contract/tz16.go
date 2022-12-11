// Copyright (c) 2020-2022 Blockwatch Data Inc.
// Author: alex@blockwatch.cc
package contract

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"hash"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"blockwatch.cc/tzgo/micheline"
	"blockwatch.cc/tzgo/rpc"
	"blockwatch.cc/tzgo/tezos"
)

// Represents Tzip16 contract metadata
type Tz16 struct {
	Name        string       `json:"name"`
	Description string       `json:"description,omitempty"`
	Version     string       `json:"version,omitempty"`
	License     *Tz16License `json:"license,omitempty"`
	Authors     []string     `json:"authors,omitempty"`
	Homepage    string       `json:"homepage,omitempty"`
	Source      *Tz16Source  `json:"source,omitempty"`
	Interfaces  []string     `json:"interfaces,omitempty"`
	Errors      []Tz16Error  `json:"errors,omitempty"`
	Views       []Tz16View   `json:"views,omitempty"`
}

type Tz16License struct {
	Name    string `json:"name"`
	Details string `json:"details,omitempty"`
}

func (l *Tz16License) UnmarshalJSON(data []byte) error {
	if len(data) == 0 {
		return nil
	}
	switch data[0] {
	case '"':
		name, err := strconv.Unquote(string(data))
		l.Name = name
		return err
	case '{':
		type alias Tz16License
		return json.Unmarshal(data, (*alias)(l))
	default:
		return fmt.Errorf("invalid license (not string or object)")
	}
}

type Tz16Source struct {
	Tools    []string `json:"tools"`
	Location string   `json:"location,omitempty"`
}

type Tz16Error struct {
	Error     *micheline.Prim `json:"error,omitempty"`
	Expansion *micheline.Prim `json:"expansion,omitempty"`
	Languages []string        `json:"languages,omitempty"`
	View      string          `json:"view,omitempty"`
}

type Tz16View struct {
	Name            string         `json:"name"`
	Description     string         `json:"description,omitempty"`
	Pure            bool           `json:"pure,omitempty"`
	Implementations []Tz16ViewImpl `json:"implementations,omitempty"`
}

type Tz16ViewImpl struct {
	Storage *Tz16StorageView `json:"michelsonStorageView,omitempty"`
	Rest    *Tz16RestView    `json:"restApiQuery,omitempty"`
}

type Tz16StorageView struct {
	ParamType   micheline.Prim       `json:"parameter"`
	ReturnType  micheline.Prim       `json:"returnType"`
	Code        micheline.Prim       `json:"code"`
	Annotations []Tz16CodeAnnotation `json:"annotations,omitempty"`
	Version     string               `json:"version,omitempty"`
}

type Tz16CodeAnnotation struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type Tz16RestView struct {
	SpecUri string `json:"specificationUri"`
	BaseUri string `json:"baseUri"`
	Path    string `json:"path"`
	Method  string `json:"method"`
}

func (t Tz16) Validate() []error {
	// TODO: json schema validator
	return nil
}

func (t Tz16) HasView(name string) bool {
	for _, v := range t.Views {
		if v.Name == name {
			return true
		}
	}
	return false
}

func (t Tz16) GetView(name string) Tz16View {
	for _, v := range t.Views {
		if v.Name == name {
			return v
		}
	}
	return Tz16View{}
}

func (v *Tz16View) Run(ctx context.Context, contract *Contract, args micheline.Prim) (micheline.Prim, error) {
	if len(v.Implementations) == 0 || v.Implementations[0].Storage == nil {
		return micheline.InvalidPrim, fmt.Errorf("missing storage view impl")
	}
	return v.Implementations[0].Storage.Run(ctx, contract, args)
}

// Run executes the TZIP-16 off-chain view using script and storage from contract and
// passed args. Returns the result as primitive which matches the view's return type.
// Note this method does not check or patch the view code to replace illegal instructions
// or inject current context.
func (v *Tz16StorageView) Run(ctx context.Context, contract *Contract, args micheline.Prim) (micheline.Prim, error) {
	// fill empty arguments
	code := v.Code.Clone()
	paramType := v.ParamType
	if !paramType.IsValid() {
		paramType = micheline.NewCode(micheline.T_UNIT)
		if !args.IsValid() {
			args = micheline.NewCode(micheline.D_UNIT)
		}
		code.Args = append(micheline.PrimList{micheline.NewCode(micheline.I_CDR)}, code.Args...)
	}

	// construct request
	req := rpc.RunCodeRequest{
		ChainId: contract.rpc.ChainId,
		Script: micheline.Code{
			Param: micheline.NewCode(
				micheline.K_PARAMETER,
				micheline.NewPairType(paramType, contract.script.Code.Storage.Args[0]),
			),
			Storage: micheline.NewCode(
				micheline.K_STORAGE,
				micheline.NewCode(micheline.T_OPTION, v.ReturnType),
			),
			Code: micheline.NewCode(
				micheline.K_CODE,
				micheline.NewSeq(
					micheline.NewCode(micheline.I_CAR),
					v.Code,
					micheline.NewCode(micheline.I_SOME),
					micheline.NewCode(micheline.I_NIL, micheline.NewCode(micheline.T_OPERATION)),
					micheline.NewCode(micheline.I_PAIR),
				),
			),
		},
		Input:   micheline.NewPair(args, *contract.store),
		Storage: micheline.NewCode(micheline.D_NONE),
		Amount:  tezos.N(0),
		Balance: tezos.N(0),
	}
	var resp rpc.RunCodeResponse
	if err := contract.rpc.RunCode(ctx, rpc.Head, req, &resp); err != nil {
		return micheline.InvalidPrim, err
	}

	// strip the extra D_SOME
	return resp.Storage.Args[0], nil
}

func (c *Contract) ResolveTz16Uri(ctx context.Context, uri string, result interface{}, checksum []byte) error {
	protoIdx := strings.Index(uri, ":")
	if protoIdx < 0 {
		return fmt.Errorf("malformed tzip16 uri %q", uri)
	}

	switch uri[:protoIdx] {
	case "tezos-storage":
		return c.resolveStorageUri(ctx, uri, result, checksum)
	case "http", "https":
		return c.resolveHttpUri(ctx, uri, result, checksum)
	case "sha256":
		parts := strings.Split(strings.TrimPrefix(uri, "sha256://"), "/")
		checksum, err := hex.DecodeString(parts[0][2:])
		if err != nil {
			return fmt.Errorf("invalid sha256 checksum: %v", err)
		}
		if len(parts) < 2 {
			return fmt.Errorf("malformed tzip16 uri %q", uri)
		}
		uri, err = url.QueryUnescape(parts[1])
		if err != nil {
			return fmt.Errorf("malformed tzip16 uri %q: %v", parts[1], err)
		}
		return c.ResolveTz16Uri(ctx, uri, result, checksum)
	case "ipfs":
		return c.resolveIpfsUri(ctx, uri, result, checksum)
	default:
		return fmt.Errorf("unsupported tzip16 protocol %q", uri[:protoIdx])
	}
}

func (c *Contract) resolveStorageUri(ctx context.Context, uri string, result interface{}, checksum []byte) error {
	if !strings.HasPrefix(uri, "tezos-storage:") {
		return fmt.Errorf("invalid tzip16 storage uri prefix: %q", uri)
	}

	// prefix is either `tezos-storage:` or `tezos-storage://`
	uri = strings.TrimPrefix(uri, "tezos-storage://")
	uri = strings.TrimPrefix(uri, "tezos-storage:")
	parts := strings.SplitN(uri, "/", 2)

	// resolve bigmap and key to read
	var (
		key string
		id  int64
		ok  bool
		err error
		con *Contract
	)
	if len(parts) == 1 {
		// same contract
		con = c
		id, ok = c.script.Bigmaps()["metadata"]
		if !ok {
			return fmt.Errorf("%s: missing metadata bigmap", c.addr)
		}
		key = parts[0]
	} else {
		// other contract
		addr, err := tezos.ParseAddress(parts[0])
		if err != nil {
			return fmt.Errorf("malformed tzip16 uri %q: %v", uri, err)
		}
		con = NewContract(addr, c.rpc)
		if err := con.Resolve(ctx); err != nil {
			return fmt.Errorf("cannot resolve %s: %v", addr, err)
		}
		id, ok = con.script.Bigmaps()["metadata"]
		if !ok {
			return fmt.Errorf("%s: missing metadata bigmap", addr)
		}
		key = parts[1]
	}

	// unescape
	key, err = url.QueryUnescape(key)
	if err != nil {
		return fmt.Errorf("malformed tzip16 uri %q: %v", uri, err)
	}
	hash := (micheline.Key{
		Type:      micheline.NewType(micheline.NewPrim(micheline.T_STRING)),
		StringKey: key,
	}).Hash()

	prim, err := con.rpc.GetActiveBigmapValue(ctx, id, hash)
	if err != nil {
		return err
	}
	if !prim.IsValid() || prim.Type != micheline.PrimBytes {
		return fmt.Errorf("Unexpected storage value type %s %q", prim.Type, prim.Dump())
	}

	// unpack JSON data
	if l := len(prim.Bytes); l > 0 && prim.Bytes[0] == '{' && prim.Bytes[l-1] == '}' {
		if checksum != nil {
			hash := sha256.Sum256(prim.Bytes)
			if !bytes.Equal(hash[:], checksum) {
				return fmt.Errorf("checksum mismatch")
			}
		}

		return json.Unmarshal(prim.Bytes, result)
	}

	// try recurse if content looks like another URI
	return con.ResolveTz16Uri(ctx, string(prim.Bytes), result, checksum)
}

func (c *Contract) resolveHttpUri(ctx context.Context, uri string, result interface{}, checksum []byte) error {
	if !strings.HasPrefix(uri, "http") {
		return fmt.Errorf("invalid tzip16 http uri prefix: %q", uri)
	}
	req, err := http.NewRequest(http.MethodGet, uri, nil)
	if err != nil {
		return err
	}
	req = req.WithContext(ctx)
	req.Header.Add("Accept", "text/plain; charset=utf-8")
	req.Header.Add("User-Agent", c.rpc.UserAgent)

	resp, err := c.rpc.Client().Do(req)
	if err != nil {
		return err
	}
	defer func() {
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	}()
	if resp.StatusCode/100 != 2 {
		return fmt.Errorf("GET %s: %d %s", uri, resp.StatusCode, resp.Status)
	}

	var reader io.Reader = resp.Body
	var h hash.Hash
	if checksum != nil {
		h = sha256.New()
		reader = io.TeeReader(reader, h)
	}

	err = json.NewDecoder(reader).Decode(result)
	if err != nil {
		return err
	}
	if checksum != nil && !bytes.Equal(h.Sum(nil), checksum) {
		return fmt.Errorf("checksum mismatch")
	}
	return nil
}

func (c *Contract) resolveIpfsUri(ctx context.Context, uri string, result interface{}, checksum []byte) error {
	if !strings.HasPrefix(uri, "ipfs://") {
		return fmt.Errorf("invalid tzip16 ipfs uri prefix: %q", uri)
	}
	gateway := strings.TrimSuffix(strings.TrimPrefix(c.rpc.IpfsURL.String(), "https://"), "/")
	gateway = "https://" + gateway + "/ipfs/"
	uri = strings.Replace(uri, "ipfs://", gateway, 1)
	return c.resolveHttpUri(ctx, uri, result, checksum)
}
