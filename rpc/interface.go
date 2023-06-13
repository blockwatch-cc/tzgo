package rpc

import (
	"context"
	"net/http"
	"time"

	"blockwatch.cc/tzgo/codec"
	"blockwatch.cc/tzgo/micheline"
	"blockwatch.cc/tzgo/tezos"
)

// Ensure Client implements the RpcClient interface
var _ RpcClient = (*Client)(nil)

// RpcClient interface for various clients implementations and mocks generation
type RpcClient interface {
	Init(ctx context.Context) error
	UseIpfsUrl(uri string) error
	Client() *http.Client
	Listen()
	Close()
	ResolveChainConfig(ctx context.Context) error
	Get(ctx context.Context, urlpath string, result interface{}) error
	GetAsync(ctx context.Context, urlpath string, mon Monitor) error
	Put(ctx context.Context, urlpath string, body, result interface{}) error
	Post(ctx context.Context, urlpath string, body, result interface{}) error
	NewRequest(ctx context.Context, method, urlStr string, body interface{}) (*http.Request, error)
	Do(req *http.Request, v interface{}) error
	DoAsync(req *http.Request, mon Monitor) error
	GetBlock(ctx context.Context, id BlockID) (*Block, error)
	GetBlockHeight(ctx context.Context, height int64) (*Block, error)
	GetTips(ctx context.Context, depth int, head tezos.BlockHash) ([][]tezos.BlockHash, error)
	GetHeadBlock(ctx context.Context) (*Block, error)
	GetGenesisBlock(ctx context.Context) (*Block, error)
	GetTipHeader(ctx context.Context) (*BlockHeader, error)
	GetBlockHeader(ctx context.Context, id BlockID) (*BlockHeader, error)
	GetBlockMetadata(ctx context.Context, id BlockID) (*BlockMetadata, error)
	GetBlockHash(ctx context.Context, id BlockID) (hash tezos.BlockHash, err error)
	GetBlockPredHashes(ctx context.Context, hash tezos.BlockHash, count int) ([]tezos.BlockHash, error)
	GetInvalidBlocks(ctx context.Context) ([]*InvalidBlock, error)
	GetInvalidBlock(ctx context.Context, blockID tezos.BlockHash) (*InvalidBlock, error)
	GetChainId(ctx context.Context) (tezos.ChainIdHash, error)
	GetStatus(ctx context.Context) (Status, error)
	GetVersionInfo(ctx context.Context) (VersionInfo, error)
	GetConstants(ctx context.Context, id BlockID) (con Constants, err error)
	GetCustomConstants(ctx context.Context, id BlockID, resp any) error
	GetParams(ctx context.Context, id BlockID) (*tezos.Params, error)
	GetContract(ctx context.Context, addr tezos.Address, id BlockID) (*ContractInfo, error)
	GetContractBalance(ctx context.Context, addr tezos.Address, id BlockID) (tezos.Z, error)
	GetManagerKey(ctx context.Context, addr tezos.Address, id BlockID) (tezos.Key, error)
	GetContractExt(ctx context.Context, addr tezos.Address, id BlockID) (*ContractInfo, error)
	ListContracts(ctx context.Context, id BlockID) (Contracts, error)
	GetContractScript(ctx context.Context, addr tezos.Address) (*micheline.Script, error)
	GetNormalizedScript(ctx context.Context, addr tezos.Address, mode UnparsingMode) (*micheline.Script, error)
	GetContractStorage(ctx context.Context, addr tezos.Address, id BlockID) (micheline.Prim, error)
	GetContractStorageNormalized(ctx context.Context, addr tezos.Address, id BlockID, mode UnparsingMode) (micheline.Prim, error)
	GetContractEntrypoints(ctx context.Context, addr tezos.Address) (map[string]micheline.Type, error)
	ListBigmapKeys(ctx context.Context, bigmap int64, id BlockID) ([]tezos.ExprHash, error)
	ListActiveBigmapKeys(ctx context.Context, bigmap int64) ([]tezos.ExprHash, error)
	GetBigmapValue(ctx context.Context, bigmap int64, hash tezos.ExprHash, id BlockID) (micheline.Prim, error)
	GetActiveBigmapValue(ctx context.Context, bigmap int64, hash tezos.ExprHash) (micheline.Prim, error)
	ListBigmapValues(ctx context.Context, bigmap int64, id BlockID) ([]micheline.Prim, error)
	ListActiveBigmapValues(ctx context.Context, bigmap int64, id BlockID) ([]micheline.Prim, error)
	GetActiveBigmapInfo(ctx context.Context, bigmap int64) (*BigmapInfo, error)
	GetBigmapInfo(ctx context.Context, bigmap int64, id BlockID) (*BigmapInfo, error)
	ListActiveDelegates(ctx context.Context, id BlockID) (DelegateList, error)
	GetDelegate(ctx context.Context, addr tezos.Address, id BlockID) (*Delegate, error)
	GetDelegateBalance(ctx context.Context, addr tezos.Address, id BlockID) (int64, error)
	GetMempool(ctx context.Context) (*Mempool, error)
	MonitorBootstrapped(ctx context.Context, monitor *BootstrapMonitor) error
	MonitorBlockHeader(ctx context.Context, monitor *BlockHeaderMonitor) error
	MonitorMempool(ctx context.Context, monitor *MempoolMonitor) error
	MonitorNetworkPointLog(ctx context.Context, address string, monitor *NetworkPointMonitor) error
	MonitorNetworkPeerLog(ctx context.Context, peerID string, monitor *NetworkPeerMonitor) error
	GetNetworkStats(ctx context.Context) (*NetworkStats, error)
	GetNetworkConnections(ctx context.Context) ([]*NetworkConnection, error)
	GetNetworkPeers(ctx context.Context, filter string) ([]*NetworkPeer, error)
	GetNetworkPeer(ctx context.Context, peerID string) (*NetworkPeer, error)
	BanNetworkPeer(ctx context.Context, peerID string) error
	TrustNetworkPeer(ctx context.Context, peerID string) error
	GetNetworkPeerBanned(ctx context.Context, peerID string) (bool, error)
	GetNetworkPeerLog(ctx context.Context, peerID string) ([]*NetworkPeerLogEntry, error)
	GetNetworkPoints(ctx context.Context, filter string) ([]*NetworkPoint, error)
	GetNetworkPoint(ctx context.Context, address string) (*NetworkPoint, error)
	ConnectToNetworkPoint(ctx context.Context, address string, timeout time.Duration) error
	BanNetworkPoint(ctx context.Context, address string) error
	TrustNetworkPoint(ctx context.Context, address string) error
	GetNetworkPointBanned(ctx context.Context, address string) (bool, error)
	GetNetworkPointLog(ctx context.Context, address string) ([]*NetworkPointLogEntry, error)
	GetBlockOperationHash(ctx context.Context, id BlockID, l, n int) (tezos.OpHash, error)
	GetBlockOperationHashes(ctx context.Context, id BlockID) ([][]tezos.OpHash, error)
	GetBlockOperationListHashes(ctx context.Context, id BlockID, l int) ([]tezos.OpHash, error)
	GetBlockOperation(ctx context.Context, id BlockID, l, n int) (*Operation, error)
	GetBlockOperationList(ctx context.Context, id BlockID, l int) ([]Operation, error)
	GetBlockOperations(ctx context.Context, id BlockID) ([][]Operation, error)
	BroadcastOperation(ctx context.Context, body []byte) (hash tezos.OpHash, err error)
	RunOperation(ctx context.Context, id BlockID, body, resp interface{}) error
	ForgeOperation(ctx context.Context, id BlockID, body, resp interface{}) error
	ListBakingRights(ctx context.Context, id BlockID, max int) ([]BakingRight, error)
	ListBakingRightsCycle(ctx context.Context, id BlockID, cycle int64, max int) ([]BakingRight, error)
	ListEndorsingRights(ctx context.Context, id BlockID) ([]EndorsingRight, error)
	ListEndorsingRightsCycle(ctx context.Context, id BlockID, cycle int64) ([]EndorsingRight, error)
	GetRollSnapshotInfoCycle(ctx context.Context, id BlockID, cycle int64) (*RollSnapshotInfo, error)
	GetStakingSnapshotInfoCycle(ctx context.Context, id BlockID, cycle int64) (*StakingSnapshotInfo, error)
	GetSnapshotIndexCycle(ctx context.Context, id BlockID, cycle int64) (*SnapshotIndex, error)
	ListSnapshotRollOwners(ctx context.Context, id BlockID, cycle, index int64) (*SnapshotOwners, error)
	Complete(ctx context.Context, o *codec.Op, key tezos.Key) error
	Simulate(ctx context.Context, o *codec.Op, opts *CallOptions) (*Receipt, error)
	Validate(ctx context.Context, o *codec.Op) error
	Broadcast(ctx context.Context, o *codec.Op) (tezos.OpHash, error)
	Send(ctx context.Context, op *codec.Op, opts *CallOptions) (*Receipt, error)
	RunCode(ctx context.Context, id BlockID, body, resp interface{}) error
	RunCallback(ctx context.Context, id BlockID, body, resp interface{}) error
	RunView(ctx context.Context, id BlockID, body, resp interface{}) error
	TraceCode(ctx context.Context, id BlockID, body, resp interface{}) error
	ListVoters(ctx context.Context, id BlockID) (VoterList, error)
	GetVoteQuorum(ctx context.Context, id BlockID) (int, error)
	GetVoteProposal(ctx context.Context, id BlockID) (tezos.ProtocolHash, error)
	ListBallots(ctx context.Context, id BlockID) (BallotList, error)
	GetVoteResult(ctx context.Context, id BlockID) (BallotSummary, error)
	ListProposals(ctx context.Context, id BlockID) (ProposalList, error)
}
