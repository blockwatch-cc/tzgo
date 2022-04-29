// Copyright (c) 2020-2022 Blockwatch Data Inc.
// Author: alex@blockwatch.cc

package contract

var wellKnown = map[string]*TokenMetadata{
    "KT1PWx2mnDueood7fEmfbBDKx1D9BAnnXitn": &TokenMetadata{
        Name:     "tzBTC",
        Symbol:   "tzBTC",
        Decimals: 8,
    },
    "KT1VYsVfmobT7rsMVivvZ4J8i3bPiqz12NaH": &TokenMetadata{
        Name:     "Wrapped Tezos",
        Symbol:   "wXTZ",
        Decimals: 6,
    },
    "KT1LN4LPSqTMS7Sd2CJw4bbDGRkMv2t68Fy9": &TokenMetadata{
        Name:     "USDtez",
        Symbol:   "USDtez",
        Decimals: 6,
    },
    "KT19at7rQUvyjxnZ2fBv7D9zc8rkyG7gAoU8": &TokenMetadata{
        Name:     "ETHtez",
        Symbol:   "ETHtez",
        Decimals: 18,
    },
    "KT1AEfeckNbdEYwaMKkytBwPJPycz7jdSGea": &TokenMetadata{
        Name:     "Staker Governance Token",
        Symbol:   "STKR",
        Decimals: 18,
    },
}
