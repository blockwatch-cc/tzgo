// Copyright (c) 2020-2021 Blockwatch Data Inc.
// Author: alex@blockwatch.cc

package rpc

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"blockwatch.cc/tzgo/tezos"
)

// Voter holds information about a vote listing
type Voter struct {
	Delegate tezos.Address `json:"pkh"`
	Rolls    int64         `json:"rolls"`
}

// VoterList contains a list of voters
type VoterList []Voter

// BallotInfo holds information about a vote listing
type BallotInfo struct {
	Delegate tezos.Address    `json:"pkh"`
	Ballot   tezos.BallotVote `json:"ballot"`
}

// BallotList contains a list of voters
type BallotList []BallotInfo

// Ballots holds the current summary of a vote
type BallotSummary struct {
	Yay  int `json:"yay"`
	Nay  int `json:"nay"`
	Pass int `json:"pass"`
}

// Proposal holds information about a vote listing
type Proposal struct {
	Proposal tezos.ProtocolHash
	Upvotes  int64
}

func (p *Proposal) UnmarshalJSON(data []byte) error {
	if len(data) == 0 || bytes.Compare(data, []byte("null")) == 0 || len(data) == 2 {
		return nil
	}
	if data[0] != '[' {
		return fmt.Errorf("rpc: proposal: expected JSON array")
	}
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.UseNumber()
	unpacked := make([]interface{}, 0)
	err := dec.Decode(&unpacked)
	if err != nil {
		return fmt.Errorf("rpc: proposal: %w", err)
	}
	if len(unpacked) != 2 {
		return fmt.Errorf("rpc: proposal: invalid JSON array")
	}
	if err := p.Proposal.UnmarshalText([]byte(unpacked[0].(string))); err != nil {
		return fmt.Errorf("rpc: proposal: %w", err)
	}
	p.Upvotes, err = strconv.ParseInt(unpacked[1].(json.Number).String(), 10, 64)
	if err != nil {
		return fmt.Errorf("rpc: proposal: %w", err)
	}
	return nil
}

// ProposalList contains a list of voters
type ProposalList []Proposal

// ListVoters returns information about all eligible voters for an election
// at block id.
func (c *Client) ListVoters(ctx context.Context, id BlockID) (VoterList, error) {
	voters := make(VoterList, 0)
	u := fmt.Sprintf("chains/main/blocks/%s/votes/listings", id)
	if err := c.Get(ctx, u, &voters); err != nil {
		return nil, err
	}
	return voters, nil
}

// GetVoteQuorum returns information about the current voring quorum at block id.
// Returned value is percent * 10000 i.e. 5820 for 58.20%.
func (c *Client) GetVoteQuorum(ctx context.Context, id BlockID) (int, error) {
	var quorum int
	u := fmt.Sprintf("chains/main/blocks/%s/votes/current_quorum", id)
	if err := c.Get(ctx, u, &quorum); err != nil {
		return 0, err
	}
	return quorum, nil
}

// GetVoteProposal returns the hash of the current voring proposal at block id.
func (c *Client) GetVoteProposal(ctx context.Context, id BlockID) (tezos.ProtocolHash, error) {
	var proposal tezos.ProtocolHash
	u := fmt.Sprintf("chains/main/blocks/%s/votes/current_proposal", id)
	err := c.Get(ctx, u, &proposal)
	return proposal, err
}

// ListBallots returns information about all eligible voters for an election at block id.
func (c *Client) ListBallots(ctx context.Context, id BlockID) (BallotList, error) {
	ballots := make(BallotList, 0)
	u := fmt.Sprintf("chains/main/blocks/%s/votes/ballot_list", id)
	if err := c.Get(ctx, u, &ballots); err != nil {
		return nil, err
	}
	return ballots, nil
}

// GetVoteResult returns a summary of the current voting result at block id.
func (c *Client) GetVoteResult(ctx context.Context, id BlockID) (BallotSummary, error) {
	summary := BallotSummary{}
	u := fmt.Sprintf("chains/main/blocks/%s/votes/ballots", id)
	err := c.Get(ctx, u, &summary)
	return summary, err
}

// ListProposals returns a list of all submitted proposals and their upvote count at block id.
// This call only returns results when block is within a proposal vote period.
func (c *Client) ListProposals(ctx context.Context, id BlockID) (ProposalList, error) {
	proposals := make(ProposalList, 0)
	u := fmt.Sprintf("chains/main/blocks/%s/votes/proposals", id)
	if err := c.Get(ctx, u, &proposals); err != nil {
		return nil, err
	}
	return proposals, nil
}
