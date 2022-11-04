package module

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/gorilla/mux"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/spf13/cobra"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/osmosis-labs/osmosis/v12/simulation/simtypes"
	"github.com/osmosis-labs/osmosis/v12/x/gamm/pool-models/balancer"
	"github.com/osmosis-labs/osmosis/v12/x/swaprouter"
	swaprouterclient "github.com/osmosis-labs/osmosis/v12/x/swaprouter/client"
	"github.com/osmosis-labs/osmosis/v12/x/swaprouter/client/cli"
	"github.com/osmosis-labs/osmosis/v12/x/swaprouter/client/grpc"
	"github.com/osmosis-labs/osmosis/v12/x/swaprouter/client/queryproto"
	swaproutersimulation "github.com/osmosis-labs/osmosis/v12/x/swaprouter/simulation"
	"github.com/osmosis-labs/osmosis/v12/x/swaprouter/types"
)

var (
	_ module.AppModule      = AppModule{}
	_ module.AppModuleBasic = AppModuleBasic{}
)

type AppModuleBasic struct{}

func (AppModuleBasic) Name() string { return types.ModuleName }

func (AppModuleBasic) RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	types.RegisterLegacyAminoCodec(cdc)
}

func (AppModuleBasic) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
	return cdc.MustMarshalJSON(types.DefaultGenesis())
}

// ValidateGenesis performs genesis state validation for the swaprouter module.
func (AppModuleBasic) ValidateGenesis(cdc codec.JSONCodec, config client.TxEncodingConfig, bz json.RawMessage) error {
	var genState types.GenesisState
	if err := cdc.UnmarshalJSON(bz, &genState); err != nil {
		return fmt.Errorf("failed to unmarshal %s genesis state: %w", types.ModuleName, err)
	}
	return genState.Validate()
}

// ---------------------------------------
// Interfaces.
func (b AppModuleBasic) RegisterRESTRoutes(ctx client.Context, r *mux.Router) {
}

func (b AppModuleBasic) RegisterGRPCGatewayRoutes(clientCtx client.Context, mux *runtime.ServeMux) {
	if err := queryproto.RegisterQueryHandlerClient(context.Background(), mux, queryproto.NewQueryClient(clientCtx)); err != nil {
		panic(err)
	}
}

func (b AppModuleBasic) GetTxCmd() *cobra.Command {
	return cli.NewTxCmd()
}

func (b AppModuleBasic) GetQueryCmd() *cobra.Command {
	return cli.GetQueryCmd()
}

// RegisterInterfaces registers interfaces and implementations of the gamm module.
func (AppModuleBasic) RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	types.RegisterInterfaces(registry)
}

type AppModule struct {
	AppModuleBasic

	k          swaprouter.Keeper
	gammKeeper types.GammKeeper
}

func (am AppModule) RegisterServices(cfg module.Configurator) {
	types.RegisterMsgServer(cfg.MsgServer(), swaprouter.NewMsgServerImpl(&am.k))
	balancer.RegisterMsgServer(cfg.MsgServer(), swaprouter.NewBalancerMsgServerImpl(&am.k))
	queryproto.RegisterQueryServer(cfg.QueryServer(), grpc.Querier{Q: swaprouterclient.Querier{K: am.k}})
}

func NewAppModule(swaprouterKeeper swaprouter.Keeper, gammKeeper types.GammKeeper) AppModule {
	return AppModule{
		AppModuleBasic: AppModuleBasic{},
		k:              swaprouterKeeper,
		gammKeeper:     gammKeeper,
	}
}

func (am AppModule) RegisterInvariants(ir sdk.InvariantRegistry) {
}

func (am AppModule) Route() sdk.Route {
	return sdk.Route{}
}

// QuerierRoute returns the gamm module's querier route name.
func (AppModule) QuerierRoute() string { return types.RouterKey }

// LegacyQuerierHandler returns the x/gamm module's sdk.Querier.
func (am AppModule) LegacyQuerierHandler(legacyQuerierCdc *codec.LegacyAmino) sdk.Querier {
	return func(sdk.Context, []string, abci.RequestQuery) ([]byte, error) {
		return nil, fmt.Errorf("legacy querier not supported for the x/%s module", types.ModuleName)
	}
}

// InitGenesis performs genesis initialization for the swaprouter module.
// no validator updates.
func (am AppModule) InitGenesis(ctx sdk.Context, cdc codec.JSONCodec, gs json.RawMessage) []abci.ValidatorUpdate {
	var genesisState types.GenesisState

	cdc.MustUnmarshalJSON(gs, &genesisState)

	am.k.InitGenesis(ctx, &genesisState)
	return []abci.ValidatorUpdate{}
}

// ExportGenesis returns the exported genesis state as raw bytes for the swaprouter.
// module.
func (am AppModule) ExportGenesis(ctx sdk.Context, cdc codec.JSONCodec) json.RawMessage {
	genState := am.k.ExportGenesis(ctx)
	return cdc.MustMarshalJSON(genState)
}

// BeginBlock performs a no-op.
func (AppModule) BeginBlock(_ sdk.Context, _ abci.RequestBeginBlock) {}

// EndBlock performs a no-op.
func (am AppModule) EndBlock(ctx sdk.Context, _ abci.RequestEndBlock) []abci.ValidatorUpdate {
	return []abci.ValidatorUpdate{}
}

// ConsensusVersion implements AppModule/ConsensusVersion.
func (AppModule) ConsensusVersion() uint64 { return 1 }

// **** simulation implementation ****
// GenerateGenesisState creates a randomized GenState of the swaprouter module.
// **** simulation implementation ****
// GenerateGenesisState creates a randomized GenState of the gamm module.
func (am AppModule) SimulatorGenesisState(simState *module.SimulationState, s *simtypes.SimCtx) {
	swaprouterGen := types.DefaultGenesis()
	// change the pool creation fee denom from uosmo to stake
	swaprouterGen.Params.PoolCreationFee = sdk.NewCoins(swaproutersimulation.PoolCreationFee)
	DefaultGenJson := simState.Cdc.MustMarshalJSON(swaprouterGen)
	simState.GenState[types.ModuleName] = DefaultGenJson
}

func (am AppModule) Actions() []simtypes.Action {
	return swaproutersimulation.DefaultActions(am.k, am.gammKeeper)
}
