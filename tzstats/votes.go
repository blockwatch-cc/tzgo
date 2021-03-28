// Copyright (c) 2020-2021 Blockwatch Data Inc.
// Author: alex@blockwatch.cc

package tzstats

import (
	"context"
	"fmt"
	"time"

	"blockwatch.cc/tzgo/tezos"
)

type Election struct {
	Id                int                    `json:"election_id"`
	NumPeriods        int                    `json:"num_periods"`
	NumProposals      int                    `json:"num_proposals"`
	StartTime         time.Time              `json:"start_time"`
	EndTime           time.Time              `json:"end_time"`
	StartHeight       int64                  `json:"start_height"`
	EndHeight         int64                  `json:"end_height"`
	IsEmpty           bool                   `json:"is_empty"`
	IsOpen            bool                   `json:"is_open"`
	IsFailed          bool                   `json:"is_failed"`
	NoQuorum          bool                   `json:"no_quorum"`
	NoMajority        bool                   `json:"no_majority"`
	NoProposal        bool                   `json:"no_proposal"`
	VotingPeriodKind  tezos.VotingPeriodKind `json:"voting_period"`
	ProposalPeriod    *Vote                  `json:"proposal"`
	ExplorationPeriod *Vote                  `json:"exploration"`
	CooldownPeriod    *Vote                  `json:"cooldown"`
	PromotionPeriod   *Vote                  `json:"promotion"`
	AdoptionPeriod    *Vote                  `json:"adoption"`
}

func (e Election) Period(p tezos.VotingPeriodKind) *Vote {
	switch p {
	case tezos.VotingPeriodProposal:
		return e.ProposalPeriod
	case tezos.VotingPeriodExploration:
		return e.ExplorationPeriod
	case tezos.VotingPeriodCooldown:
		return e.CooldownPeriod
	case tezos.VotingPeriodPromotion:
		return e.PromotionPeriod
	case tezos.VotingPeriodAdoption:
		return e.AdoptionPeriod
	default:
		return nil
	}
}

type Vote struct {
	VotingPeriod     int64                  `json:"voting_period"`
	VotingPeriodKind tezos.VotingPeriodKind `json:"voting_period_kind"`
	StartTime        time.Time              `json:"period_start_time"`
	EndTime          time.Time              `json:"period_end_time"`
	StartHeight      int64                  `json:"period_start_block"`
	EndHeight        int64                  `json:"period_end_block"`
	EligibleRolls    int                    `json:"eligible_rolls"`
	EligibleVoters   int                    `json:"eligible_voters"`
	QuorumPct        int                    `json:"quorum_pct"`
	QuorumRolls      int                    `json:"quorum_rolls"`
	TurnoutRolls     int                    `json:"turnout_rolls"`
	TurnoutVoters    int                    `json:"turnout_voters"`
	TurnoutPct       int                    `json:"turnout_pct"`
	TurnoutEma       int                    `json:"turnout_ema"`
	YayRolls         int                    `json:"yay_rolls"`
	YayVoters        int                    `json:"yay_voters"`
	NayRolls         int                    `json:"nay_rolls"`
	NayVoters        int                    `json:"nay_voters"`
	PassRolls        int                    `json:"pass_rolls"`
	PassVoters       int                    `json:"pass_voters"`
	IsOpen           bool                   `json:"is_open"`
	IsFailed         bool                   `json:"is_failed"`
	IsDraw           bool                   `json:"is_draw"`
	NoProposal       bool                   `json:"no_proposal"`
	NoQuorum         bool                   `json:"no_quorum"`
	NoMajority       bool                   `json:"no_majority"`
	Proposals        []*Proposal            `json:"proposals"`
}

type Proposal struct {
	Hash          tezos.ProtocolHash `json:"hash"`
	SourceAddress tezos.Address      `json:"source"`
	BlockHash     tezos.BlockHash    `json:"block_hash"`
	OpHash        tezos.OpHash       `json:"op_hash"`
	Height        int64              `json:"height"`
	Time          time.Time          `json:"time"`
	Rolls         int64              `json:"rolls"`
	Voters        int64              `json:"voters"`
}

type Ballot struct {
	RowId            uint64                 `json:"row_id"`
	Height           int64                  `json:"height"`
	Timestamp        time.Time              `json:"time"`
	ElectionId       int                    `json:"election_id"`
	VotingPeriod     int64                  `json:"voting_period"`
	VotingPeriodKind tezos.VotingPeriodKind `json:"voting_period_kind"`
	Proposal         tezos.ProtocolHash     `json:"proposal"`
	OpHash           tezos.OpHash           `json:"op"`
	Ballot           tezos.BallotVote       `json:"ballot"`
	Rolls            int64                  `json:"rolls"`
	Sender           tezos.Address          `json:"sender"`
}

type Voter struct {
	RowId     uint64               `json:"row_id"`
	Address   tezos.Address        `json:"address"`
	Rolls     int64                `json:"rolls"`
	Stake     int64                `json:"stake"`
	Ballot    tezos.BallotVote     `json:"ballot"`
	HasVoted  bool                 `json:"has_voted"`
	Proposals []tezos.ProtocolHash `json:"proposals"`
}

func (c *Client) GetElection(ctx context.Context, id int) (*Election, error) {
	e := &Election{}
	u := fmt.Sprintf("/explorer/election/%d", id)
	if err := c.get(ctx, u, nil, e); err != nil {
		return nil, err
	}
	return e, nil
}

func (c *Client) ListVoters(ctx context.Context, id int, stage int) ([]Voter, error) {
	voters := make([]Voter, 0)
	u := fmt.Sprintf("/explorer/election/%d/%d/voters?limit=5000", id, stage)
	if err := c.get(ctx, u, nil, &voters); err != nil {
		return nil, err
	}
	return voters, nil
}

func (c *Client) ListBallots(ctx context.Context, id int, stage int) ([]Ballot, error) {
	ballots := make([]Ballot, 0)
	u := fmt.Sprintf("/explorer/election/%d/%d/ballots?limit=5000", id, stage)
	if err := c.get(ctx, u, nil, &ballots); err != nil {
		return nil, err
	}
	return ballots, nil
}

func (c *Client) ListVoterBallots(ctx context.Context, addr string) ([]Ballot, error) {
	ballots := make([]Ballot, 0)
	u := fmt.Sprintf("/explorer/account/%s/ballots?limit=5000", addr)
	if err := c.get(ctx, u, nil, &ballots); err != nil {
		return nil, err
	}
	return ballots, nil
}
