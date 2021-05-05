//
// Testcase Generator
//
// fetches contract script and storage/operations and prepares
// testcases for bigmaps, storage data and call parameters
//

package main

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"blockwatch.cc/tzgo/micheline"
	"blockwatch.cc/tzgo/tezos"
	"blockwatch.cc/tzstats-go"
	"github.com/echa/log"
)

var (
	testPath = "micheline/testdata" // + net / bigmap, storage, params
	rootPath = ""                   // root path to write files to (`testpath-$net`)
	flags    = flag.NewFlagSet("qa", flag.ContinueOnError)
	verbose  bool
	force    bool
	nounpack bool
	index    string
)

func init() {
	flags.Usage = func() {}
	flags.BoolVar(&verbose, "v", false, "be verbose")
	flags.BoolVar(&force, "f", false, "force-overwrite files")
	flags.BoolVar(&nounpack, "no-unpack", false, "disable value unpacking")
	flags.StringVar(&index, "index", "https://api.tzstats.com", "tzwatch url")
}

var knownContracts = map[string][]string{
	"Mainnet": []string{
		"KT1QuofAgnsWffHzLA7D78rxytJruGHDe7XG", // vesting           storage
		"KT1ETPG89SUW4qnuR7WpjcNju9wyjWcjY2W7", // tezos team        storage
		"KT1VG2WtYdSWz5E7chTeAdDPZNy2MpP8pTfL", // atomex            storage/bigmap
		"KT1ChNsEFxwyCbJyWGSL3KdjeXE28AY1Kaog", // baker registry    storage/bigmap
		"KT1PWx2mnDueood7fEmfbBDKx1D9BAnnXitn", // tzbtc             bigmap (packed)
		"KT1EctCuorV2NfVb1XTQgvzJ88MQtWP8cMMv", // staker            storage/bigmap
		"KT1WRUe3csC1jiThN9KUtaji2bd412upfn1E", // early             bigmap
		"KT1E8Qzgx3C5AAE4iGuXvqSQjdd21LK2aXAk", // pair-key          bigmap
		"KT1UvfyLytrt71jh63YV4Yex5SmbNXpWHxtg", // smartpy game      storage/bigmap
		"KT1Tr2eG3eVmPRbymrbU2UppUmKjFPXomGG9", // dexter usdtz      storage/bigmap
		"KT1RJ6PbjHpwc3M5rw5s2Nbmefwbuwbdxton", // hic nft           storage/bigmaps
		"KT1AFA2mwNUMNd4SsujE1YYp29vd8BZejyKW", // hdao fa2          storage/bigmaps
		"KT1EpGgjQs73QfFJs9z7m1Mxm5MTnpC2tqse", // kalamint          storage/bigmaps
		"KT1K9gCRgaLRFKTErYt1wVxA3Frb9FjasjTV", // kolibri fa12      storage/bigmaps
		"KT1LN4LPSqTMS7Sd2CJw4bbDGRkMv2t68Fy9", // usdtz fa12        storage/bigmaps
		"KT1Ndt3iD7WkH9BqpNVTLjGBww7aTpfbYXmE", // openminter        storage/bigmaps
		"KT1NQfJvo9v8hXmEgqos8NP7sS8V4qaEfvRF", // DS token          storage/bigmap
		"KT1LGCVzePeHZB4jHTdAdrieMPvgahd2x9Qz", // list(struct)      storage
		"KT1RA48D7YPmS1bcpfhZKsN6DpZbC4oAxpVW", // map(map)          storage
		"KT1PzYG3eUARDqJoK7kfUgz3R5LyTufXTPCf", // lists             storage
	},
	"Edonet2": []string{ // scanned at block 140,700, errors from v1 engine
		"KT195Mo1aBBLvzMVGQhiSWRDngp5PoQANavy", // Tezos Mandala V2
		"KT19fRB3uiUyqXfwdVQB4DG1DATN16zBuSe9", // exchange? domain?
		"KT1A3i2oEwDQn3V95ptLDGX77wXReMVLdip3", // SmartPy oracle, tzip16
		"KT1AEtyfnWjfGM4ryEeSqmhSax9hkhbAhjP8", // fa1.2 lunartez
		"KT1B4i9KFmHxuzPxV9et1shHHjoNveEDzEM9", // permatacos: nested or
		"KT1BBdbw3nnX6QrhcrT5QFAG4K2oYzQ6GvQh", // fa.12 with dividends
		"KT1BeEf292EPCsCXW8ge6wa2ciewx2xezapo", // fa2 Water Recharge Certificates
		"KT1BoGD6nLYq63D8YzrFMjaAJpiqL9JBx4AK", // fa2, salsa
		"KT1DeaG92tJuYXeQ77fkToNnGucNUFdPyntf", // bls12_381
		"KT1K36wjZ3VwQ54sKctXZYvkQ8mvF8jmZJY9", // baseDAO
		"KT1M7keBVNkvRoc8kGaAQ3cLGWKqqcKDXiTi", // oracle ??
		"KT1RRKf2Ljb2Y5X7kbuBnpyc6Mc94ZC42Xsv", // dex
		"KT1PRfJARc2o5pUoKd1zcVD8nEoecjUWWu1e", // ticket (bigmap)
		"KT1JrX3LMNmN2Sy5DdW4TyjvWzww49rjgnVK", // packed lambdas
		"KT1C8pXFQaRhMnGXtXwG6KqQ4aHZQ8B4QyaF", // good comb testcase
	},
}

var knownOps = map[string][]string{
	"Mainnet": []string{
		"opaMeiEwG1ccuLHWRNzMRSdCqhsVLT2tf2qRqbkaJuj7qMvbzLk", // openminter mint
		"opAh17oT2tV9bxguwpXzG5Mhvm33ZYkk2Jhmq4bGXeUqNMfE6V8", // equisafe mint
		"opVW5n7gca2w9myUreVQTTDKLUaPdAVAVz3WAksSatjCL7iLj6u", // hdao fa2 tx
	},
	"Edonet2": []string{
		"oogELfU8ZxZtGE8EU4j5b1aj9V29qmDZ24UzSkREH8m2uS3i3tK", // mandala update
	},
}

func main() {
	if err := flags.Parse(os.Args[1:]); err != nil {
		if err == flag.ErrHelp {
			fmt.Println("TzGo Testcase Generator")
			flags.PrintDefaults()
			os.Exit(0)
		}
		log.Fatal("Error:", err)
	}

	if verbose {
		log.SetLevel(log.LevelDebug)
		tzstats.UseLogger(log.Log)
	}

	if err := run(); err != nil {
		log.Fatal("Error:", err)
	}
}

func run() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if flags.NArg() == 0 {
		return fmt.Errorf("expected contract or operation hash")
	}

	c, err := tzstats.NewClient(index, nil)
	if err != nil {
		return err
	}

	// get index tip
	tip, err := c.GetTip(ctx)
	if err != nil {
		return err
	}
	log.Infof("Running on %s %s block %d", tip.Name, tip.Network, tip.Height)

	// make sure target path exists
	rootPath = fmt.Sprintf("%s-%s", testPath, strings.ToLower(tip.Network))
	if err := os.MkdirAll(rootPath, 0755); err != nil {
		return err
	}

	h := flags.Arg(0)
	switch tezos.ParseHashType(h) {
	case tezos.HashTypePkhNocurve:
		return fetchContractData(ctx, c, tip.Network, h)
	case tezos.HashTypeOperation:
		return fetchOperationData(ctx, c, tip.Network, h)
	default:
		if h != "all" {
			return fmt.Errorf("expected contract or operation hash")
		}
		cc, _ := knownContracts[tip.Network]
		for _, v := range cc {
			if err := fetchContractData(ctx, c, tip.Network, v); err != nil {
				log.Errorf("%s: %v", v, err)
			}
		}
		oo, _ := knownOps[tip.Network]
		for _, v := range oo {
			if err := fetchOperationData(ctx, c, tip.Network, v); err != nil {
				log.Errorf("%s: %v", v, err)
			}
		}
		return nil
	}
}

type Testcase struct {
	Name      string          `json:"name"`
	NoUnpack  bool            `json:"no_unpack,omitempty"`
	Type      json.RawMessage `json:"type"`
	Value     json.RawMessage `json:"value"`
	Key       json.RawMessage `json:"key,omitempty"`      // bigmap only
	TypeHex   string          `json:"type_hex,omitempty"` // bigmap only
	ValueHex  string          `json:"value_hex"`
	KeyHex    string          `json:"key_hex,omitempty"`
	WantValue json.RawMessage `json:"want_value"`
	WantKey   json.RawMessage `json:"want_key,omitempty"`
}

func writeFile(name string, content interface{}) error {
	name += ".json"
	if _, err := os.Stat(name); !os.IsNotExist(err) && !force {
		return fmt.Errorf("file %s exists", name)
	}
	buf, err := json.MarshalIndent(content, "", "  ")
	if err != nil {
		return err
	}
	return ioutil.WriteFile(name, buf, 0644)
}

func fetchContractData(ctx context.Context, c *tzstats.Client, net, hash string) error {
	// fetch data
	log.Infof("Fetching contract %s", hash)
	params := tzstats.NewContractParams().WithPrim()
	script, err := c.GetContractScript(ctx, hash, params)
	if err != nil {
		return err
	}
	storage, err := c.GetContractStorage(ctx, hash, params)
	if err != nil {
		return err
	}
	contract, err := c.GetContract(ctx, hash, params)
	if err != nil {
		return err
	}
	bmids := make([]int64, 0)
	for _, id := range contract.BigMaps {
		bmids = append(bmids, id)
	}
	sort.Slice(bmids, func(i, j int) bool { return bmids[i] < bmids[j] })

	// construct testcase data
	tc := Testcase{
		Name:     net + "/" + hash,
		NoUnpack: nounpack,
	}
	if buf, err := script.Script.StorageType().Prim.MarshalJSON(); err == nil {
		tc.Type = json.RawMessage(buf)
	} else {
		return err
	}
	if buf, err := script.Script.StorageType().Prim.MarshalBinary(); err == nil {
		tc.TypeHex = hex.EncodeToString(buf)
	} else {
		return err
	}
	if buf, err := storage.Prim.MarshalJSON(); err == nil {
		tc.Value = json.RawMessage(buf)
	} else {
		return err
	}
	if buf, err := storage.Prim.MarshalBinary(); err == nil {
		tc.ValueHex = hex.EncodeToString(buf)
	} else {
		return err
	}
	if buf, err := json.Marshal(storage.Value); err == nil {
		tc.WantValue = json.RawMessage(buf)
	} else {
		return err
	}

	// write to file
	if err := writeFile(filepath.Join(rootPath, "storage", hash), []Testcase{tc}); err != nil {
		return err
	}

	// fetch bigmaps (if any)
	for idx, id := range bmids {
		bigmap := make([]Testcase, 0)
		bm, err := c.GetBigmap(ctx, id, params)
		if err != nil {
			return err
		}
		if bm.NKeys == 0 {
			continue
		}
		log.Infof("> bigmap %d", id)
		p := params.WithLimit(500)
		if !nounpack {
			p = p.WithUnpack()
		}
		vals, err := c.GetBigmapValues(ctx, id, p)
		if err != nil {
			return err
		}
		for _, val := range vals {
			tc := Testcase{
				Name:     fmt.Sprintf("%s/%s-%d-%s", net, hash, id, val.KeyHash),
				NoUnpack: nounpack,
			}
			bmtype, _ := script.Script.Code.Storage.FindOpCodes(micheline.T_BIG_MAP)
			if buf, err := bmtype[idx].MarshalJSON(); err == nil {
				tc.Type = json.RawMessage(buf)
			} else {
				return err
			}
			if buf, err := bmtype[idx].MarshalBinary(); err == nil {
				tc.TypeHex = hex.EncodeToString(buf)
			} else {
				return err
			}
			if buf, err := val.ValuePrim.MarshalJSON(); err == nil {
				tc.Value = json.RawMessage(buf)
			} else {
				return err
			}
			if buf, err := val.ValuePrim.MarshalBinary(); err == nil {
				tc.ValueHex = hex.EncodeToString(buf)
			} else {
				return err
			}
			if buf, err := val.KeyPrim.MarshalJSON(); err == nil {
				tc.Key = json.RawMessage(buf)
			} else {
				return err
			}
			if buf, err := val.KeyPrim.MarshalBinary(); err == nil {
				tc.KeyHex = hex.EncodeToString(buf)
			} else {
				return err
			}
			// check for embedded value error
			if _, ok := val.GetString("error.message"); ok {
				log.Errorf("> value error in %s", tc.Name)
			}
			if buf, err := json.Marshal(val.Value); err == nil {
				tc.WantValue = json.RawMessage(buf)
			} else {
				return err
			}
			// check for embedded value error
			if _, ok := val.Key.GetString("error.message"); ok {
				log.Errorf("> key error in %s", tc.Name)
			}
			if buf, err := json.Marshal(val.Key); err == nil {
				tc.WantKey = json.RawMessage(buf)
			} else {
				return err
			}
			bigmap = append(bigmap, tc)
		}
		if len(bigmap) > 0 {
			// bigmaps/:contract-:id.json
			if err := writeFile(filepath.Join(rootPath, "bigmap", hash+"-"+strconv.FormatInt(id, 10)), &bigmap); err != nil {
				return err
			}
		}
	}
	return nil
}

func fetchOperationData(ctx context.Context, c *tzstats.Client, net, hash string) error {
	log.Infof("Fetching op %s", hash)
	p := tzstats.NewOpParams().WithPrim().WithMeta()
	if !nounpack {
		p = p.WithUnpack()
	}
	ops, err := c.GetOp(ctx, hash, p)
	if err != nil {
		return err
	}

	params := make([]Testcase, 0)
	bigmap := make([]Testcase, 0)

	for _, v := range ops {
		// skip reveals and internal delegations/originations
		if v.Type != tezos.OpTypeTransaction {
			continue
		}
		// skip non-contract tx
		if !v.IsContract {
			continue
		}

		// fetch target contract
		log.Infof("> receiver %s", v.Receiver)
		contract, err := c.GetContract(ctx, v.Receiver.String(), tzstats.NewContractParams().WithPrim())
		if err != nil {
			return err
		}
		bmids := make([]int64, 0)
		for _, id := range contract.BigMaps {
			bmids = append(bmids, id)
		}
		sort.Slice(bmids, func(i, j int) bool { return bmids[i] < bmids[j] })

		// fetch target contract script
		script, err := c.GetContractScript(ctx, v.Receiver.String(), tzstats.NewContractParams().WithPrim())
		if err != nil {
			return err
		}

		// handle call parameters
		tc := Testcase{
			Name:     fmt.Sprintf("%s/%s/%d/%d", net, hash, v.OpC, v.OpI),
			NoUnpack: nounpack,
		}
		if buf, err := script.Entrypoints[v.Parameters.Call].Prim.MarshalJSON(); err == nil {
			tc.Type = json.RawMessage(buf)
		} else {
			return err
		}
		if buf, err := script.Entrypoints[v.Parameters.Call].Prim.MarshalBinary(); err == nil {
			tc.TypeHex = hex.EncodeToString(buf)
		} else {
			return err
		}
		if buf, err := v.Parameters.Prim.MarshalJSON(); err == nil {
			tc.Value = json.RawMessage(buf)
		} else {
			return err
		}
		if buf, err := v.Parameters.Prim.MarshalBinary(); err == nil {
			tc.ValueHex = hex.EncodeToString(buf)
		} else {
			return err
		}
		// check for embedded value error
		if _, ok := v.Parameters.GetString("error.message"); ok {
			log.Errorf("> param error in %s", tc.Name)
		}
		if buf, err := json.Marshal(v.Parameters.Value); err == nil {
			tc.WantValue = json.RawMessage(buf)
		} else {
			return err
		}
		params = append(params, tc)

		// handle bigmap updates
		for i, bmd := range v.BigmapDiff {
			// skip non-update actions
			if bmd.Action != micheline.DiffActionUpdate {
				continue
			}
			// reverse-lookup bigmap position
			idx := sort.Search(len(bmids), func(i int) bool { return bmids[i] >= bmd.Meta.BigMapId })

			tc := Testcase{
				Name:     fmt.Sprintf("%s/%s/%d/%d/%d-%d", net, hash, v.OpC, v.OpI, i, bmd.Meta.BigMapId),
				NoUnpack: nounpack,
			}
			bmtype, _ := script.Script.Code.Storage.FindOpCodes(micheline.T_BIG_MAP)
			if buf, err := bmtype[idx].MarshalJSON(); err == nil {
				tc.Type = json.RawMessage(buf)
			} else {
				return err
			}
			if buf, err := bmtype[idx].MarshalBinary(); err == nil {
				tc.TypeHex = hex.EncodeToString(buf)
			} else {
				return err
			}
			if buf, err := bmd.ValuePrim.MarshalJSON(); err == nil {
				tc.Value = json.RawMessage(buf)
			} else {
				return err
			}
			if buf, err := bmd.ValuePrim.MarshalBinary(); err == nil {
				tc.ValueHex = hex.EncodeToString(buf)
			} else {
				return err
			}
			if buf, err := bmd.KeyPrim.MarshalJSON(); err == nil {
				tc.Key = json.RawMessage(buf)
			} else {
				return err
			}
			if buf, err := bmd.KeyPrim.MarshalBinary(); err == nil {
				tc.KeyHex = hex.EncodeToString(buf)
			} else {
				return err
			}
			// check for embedded value error
			if _, ok := bmd.GetString("error.message"); ok {
				log.Errorf("> value error in %s", tc.Name)
			}
			if buf, err := json.Marshal(bmd.Value); err == nil {
				tc.WantValue = json.RawMessage(buf)
			} else {
				return err
			}
			// check for embedded value error
			if _, ok := bmd.Key.GetString("error.message"); ok {
				log.Errorf("> key error in %s", tc.Name)
			}
			if buf, err := json.Marshal(bmd.Key); err == nil {
				tc.WantKey = json.RawMessage(buf)
			} else {
				return err
			}
			bigmap = append(bigmap, tc)
		}
	}

	// write params to file
	if len(params) > 0 {
		if err := writeFile(filepath.Join(rootPath, "params", hash), &params); err != nil {
			return err
		}
	}
	if len(bigmap) > 0 {
		if err := writeFile(filepath.Join(rootPath, "bigmap", hash), &bigmap); err != nil {
			return err
		}
	}

	return nil
}
