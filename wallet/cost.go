// Copyright (c) 2020-2022 Blockwatch Data Inc.
// Author: alex@blockwatch.cc
//go:build ignore
// +build ignore

package wallet

import (
    "blockwatch.cc/tzgo/rpc"
    "blockwatch.cc/tzgo/tezos"
)

type Cost struct {
    Fee            int64
    Burn           int64
    Gas            int64
    StorageBytes   int64
    StorageBurn    int64
    AllocationBurn int64
}

// var defaultCosts = map[tezos.OpType]Cost{
//     tezos.OpTypeEndorsement:               Cost{},
//     tezos.OpTypeSeedNonceRevelation:       Cost{},
//     tezos.OpTypeDoubleEndorsementEvidence: Cost{},
//     tezos.OpTypeDoubleBakingEvidence:      Cost{},
//     tezos.OpTypeActivateAccount:           Cost{},
//     tezos.OpTypeProposals:                 Cost{},
//     tezos.OpTypeBallot:                    Cost{},
//     tezos.OpTypeFailingNoop:               Cost{},
//     tezos.OpTypeReveal: Cost{
//         Fee:            1000,
//         Burn:           0,
//         Gas:            1000,
//         StorageBytes:   0,
//         StorageBurn:    0,
//         AllocationBurn: 0,
//     },
//     // minimal tx with tx account allocation
//     tezos.OpTypeTransaction: Cost{
//         Fee:  1000,
//         Burn: 0,
//         Gas:  1420, // default gas to tz1/tz2/tz3
//         // Gas:            2078, // default gas to KT1 manager.tz
//         StorageBytes:   257, // bytes for allocation
//         StorageBurn:    0,
//         AllocationBurn: 64250, // bytes * cost_per_byte
//     },
//     // tezos.OpTypeOrigination: Cost{
//     //     Fee:            1000,
//     //     Burn:           0,
//     //     Gas:            1400, // dynamic
//     //     StorageBurn:    0, // dynamic
//     //     AllocationBurn: 64250, // bytes * cost_per_byte
//     // },
//     tezos.OpTypeDelegation: Cost{
//         Fee:            1000,
//         Burn:           0,
//         Gas:            1000,
//         StorageBurn:    0,
//         AllocationBurn: 0,
//     },
//     // tezos.OpTypeRegisterConstant: Cost{
//     //     Fee:            1000,
//     //     Burn:           0,
//     //     Gas:            1234, // dynamic
//     //     StorageBytes:   257, // dynamic
//     //     StorageBurn:    20500, // dynamic, bytes * cost_per_byte
//     //     AllocationBurn: 0,
//     // },
// }

func (x Cost) Add(y Cost) Cost {
    z := x
    z.Fee += y.Fee
    z.Burn += y.Burn
    z.Gas += y.Gas
    z.StorageBurn += y.StorageBurn
    z.AllocationBurn += y.AllocationBurn
    return z
}

func sumOpCost(op rpc.TypedOperation, p *tezos.Params) Cost {
    var c Cost
    switch op.Kind() {

    case tezos.OpTypeReveal:
        o := op.(*rpc.Reveal)
        c.Fee += o.Fee
        c.Gas += o.Metadata.Result.ConsumedGas

    case tezos.OpTypeRegisterConstant:
        o := op.(*rpc.ConstantRegistration)
        c.Fee += o.Fee
        c.Gas += o.Metadata.Result.ConsumedGas
        storageCost := o.Metadata.Result.StorageSize * p.CostPerByte
        c.Burn += storageCost
        c.StorageBurn += storageCost

    case tezos.OpTypeDelegation:
        o := op.(*rpc.Delegation)
        c.Fee += o.Fee
        c.Gas += o.Metadata.Result.ConsumedGas

    case tezos.OpTypeOrigination:
        o := op.(*rpc.Origination)
        c.Fee += o.Fee
        c.Gas += o.Metadata.Result.ConsumedGas
        var burned int64
        for _, v := range o.Metadata.Result.BalanceUpdates {
            switch v.BalanceUpdateKind() {
            case "contract":
                u := v.(*rpc.ContractBalanceUpdate)
                if u.Change < 0 {
                    burned += -u.Change
                } else {
                    burned -= u.Change
                }
            }
        }
        storageCost := o.Metadata.Result.PaidStorageSizeDiff * p.CostPerByte
        c.Burn += burned
        c.StorageBurn += storageCost
        c.AllocationBurn += burned - storageCost

    case tezos.OpTypeTransaction:
        o := op.(*rpc.Transaction)
        c.Fee += o.Fee
        c.Gas += o.Metadata.Result.ConsumedGas
        if o.Metadata.Result.Allocated {
            burned := p.OriginationSize * p.CostPerByte
            c.Burn += burned
            c.AllocationBurn += burned
        }
        storageCost := o.Metadata.Result.PaidStorageSizeDiff * p.CostPerByte
        c.StorageBurn += storageCost
        c.Burn += storageCost

        for _, in := range o.Metadata.InternalResults {
            c.Gas += in.Result.ConsumedGas
            if in.Result.Allocated {
                burned := p.OriginationSize * p.CostPerByte
                c.Burn += burned
                c.AllocationBurn += burned
            }
            storageCost := in.Result.PaidStorageSizeDiff * p.CostPerByte
            c.StorageBurn += storageCost
            c.Burn += storageCost

            // extra burn on internal originations
            if in.OpKind() == tezos.OpTypeOrigination {
                var burned int64
                for _, v := range in.Result.BalanceUpdates {
                    switch v.BalanceUpdateKind() {
                    case "contract":
                        u := v.(*rpc.ContractBalanceUpdate)
                        if u.Change < 0 {
                            burned += -u.Change
                        } else {
                            burned -= u.Change
                        }
                    }
                }
                c.Burn += burned - storageCost
                c.AllocationBurn += burned - storageCost
            }
        }
    }
    return c
}
