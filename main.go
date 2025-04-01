package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

var (
	ConfigPath string

	Denom         string
	ListenAddress string
	NodeAddress   string
	TendermintRPC string
	LogLevel      string
	JsonOutput    bool
	Limit         uint64

	Prefix                    string
	AccountPrefix             string
	AccountPubkeyPrefix       string
	ValidatorPrefix           string
	ValidatorPubkeyPrefix     string
	ConsensusNodePrefix       string
	ConsensusNodePubkeyPrefix string

	ChainID          string
	ConstLabels      map[string]string
	DenomCoefficient float64
	DenomExponent    uint64
)

var log = zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout}).With().Timestamp().Logger()

var rootCmd = &cobra.Command{
	Use:  "cosmos-exporter",
	Long: "Scrape the data about the validators set, specific validators or wallets in the Cosmos network.",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if ConfigPath == "" {
			setBechPrefixes(cmd)
			return nil
		}

		viper.SetConfigFile(ConfigPath)
		if err := viper.ReadInConfig(); err != nil {
			if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
				log.Info().Err(err).Msg("Error reading config file")
				return err
			}
		}

		cmd.Flags().VisitAll(func(f *pflag.Flag) {
			if !f.Changed && viper.IsSet(f.Name) {
				val := viper.Get(f.Name)
				if err := cmd.Flags().Set(f.Name, fmt.Sprintf("%v", val)); err != nil {
					log.Fatal().Err(err).Msg("Could not set flag")
				}
			}
		})

		setBechPrefixes(cmd)
		return nil
	},
	Run: Execute,
}

func setBechPrefixes(cmd *cobra.Command) {
	if flag, err := cmd.Flags().GetString("bech-account-prefix"); flag != "" && err == nil {
		AccountPrefix = flag
	} else {
		AccountPrefix = Prefix
	}

	if flag, err := cmd.Flags().GetString("bech-account-pubkey-prefix"); flag != "" && err == nil {
		AccountPubkeyPrefix = flag
	} else {
		AccountPubkeyPrefix = Prefix + "pub"
	}

	if flag, err := cmd.Flags().GetString("bech-validator-prefix"); flag != "" && err == nil {
		ValidatorPrefix = flag
	} else {
		ValidatorPrefix = Prefix + "valoper"
	}

	if flag, err := cmd.Flags().GetString("bech-validator-pubkey-prefix"); flag != "" && err == nil {
		ValidatorPubkeyPrefix = flag
	} else {
		ValidatorPubkeyPrefix = Prefix + "valoperpub"
	}

	if flag, err := cmd.Flags().GetString("bech-consensus-node-prefix"); flag != "" && err == nil {
		ConsensusNodePrefix = flag
	} else {
		ConsensusNodePrefix = Prefix + "valcons"
	}

	if flag, err := cmd.Flags().GetString("bech-consensus-node-pubkey-prefix"); flag != "" && err == nil {
		ConsensusNodePubkeyPrefix = flag
	} else {
		ConsensusNodePubkeyPrefix = Prefix + "valconspub"
	}
}

func Execute(cmd *cobra.Command, args []string) {
	logLevel, err := zerolog.ParseLevel(LogLevel)
	if err != nil {
		log.Fatal().Err(err).Msg("Could not parse log level")
	}

	if JsonOutput {
		log = zerolog.New(os.Stdout).With().Timestamp().Logger()
	}

	zerolog.SetGlobalLevel(logLevel)

	log.Info().
		Str("--bech-account-prefix", AccountPrefix).
		Str("--bech-account-pubkey-prefix", AccountPubkeyPrefix).
		Str("--bech-validator-prefix", ValidatorPrefix).
		Str("--bech-validator-pubkey-prefix", ValidatorPubkeyPrefix).
		Str("--bech-consensus-node-prefix", ConsensusNodePrefix).
		Str("--bech-consensus-node-pubkey-prefix", ConsensusNodePubkeyPrefix).
		Str("--denom", Denom).
		Str("--denom-coefficient", fmt.Sprintf("%f", DenomCoefficient)).
		Str("--denom-exponent", fmt.Sprintf("%d", DenomExponent)).
		Str("--listen-address", ListenAddress).
		Str("--node", NodeAddress).
		Str("--log-level", LogLevel).
		Msg("Started with following parameters")

	config := sdk.GetConfig()
	config.SetBech32PrefixForAccount(AccountPrefix, AccountPubkeyPrefix)
	config.SetBech32PrefixForValidator(ValidatorPrefix, ValidatorPubkeyPrefix)
	config.SetBech32PrefixForConsensusNode(ConsensusNodePrefix, ConsensusNodePubkeyPrefix)
	config.Seal()

	grpcConn, err := grpc.Dial(
		NodeAddress,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatal().Err(err).Msg("Could not connect to gRPC node")
	}
	defer grpcConn.Close()

	setChainID()
	setDenom(grpcConn)

	http.HandleFunc("/metrics/wallet", func(w http.ResponseWriter, r *http.Request) {
		WalletHandler(w, r, grpcConn)
	})

	http.HandleFunc("/metrics/validator", func(w http.ResponseWriter, r *http.Request) {
		ValidatorHandler(w, r, grpcConn)
	})

	http.HandleFunc("/metrics/validators", func(w http.ResponseWriter, r *http.Request) {
		ValidatorsHandler(w, r, grpcConn)
	})

	http.HandleFunc("/metrics/params", func(w http.ResponseWriter, r *http.Request) {
		ParamsHandler(w, r, grpcConn)
	})

	http.HandleFunc("/metrics/general", func(w http.ResponseWriter, r *http.Request) {
		GeneralHandler(w, r, grpcConn)
	})

	log.Info().Str("address", ListenAddress).Msg("Listening")
	err = http.ListenAndServe(ListenAddress, nil)
	if err != nil {
		log.Fatal().Err(err).Msg("Could not start application")
	}
}

func setChainID() {
	// Создаем HTTP клиент
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// Формируем URL для запроса статуса через Tendermint RPC
	url := fmt.Sprintf("%s/status", TendermintRPC)

	// Выполняем HTTP запрос
	resp, err := client.Get(url)
	if err != nil {
		log.Warn().Err(err).Msg("Could not query node status, using default chain ID")
		ChainID = "union"
		return
	}
	defer resp.Body.Close()

	// Читаем тело ответа
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Warn().Err(err).Msg("Could not read response body, using default chain ID")
		ChainID = "union"
		return
	}

	// Парсим JSON ответ
	var result struct {
		Result struct {
			NodeInfo struct {
				Network string `json:"network"`
			} `json:"node_info"`
		} `json:"result"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		log.Warn().Err(err).Msg("Could not parse response JSON, using default chain ID")
		ChainID = "union"
		return
	}

	// Получаем chain_id из поля network
	if result.Result.NodeInfo.Network != "" {
		ChainID = result.Result.NodeInfo.Network
		log.Info().Str("chain_id", ChainID).Msg("Got chain ID from node_info.network")
	} else {
		log.Warn().Msg("Chain ID not found in node_info.network, using default")
		ChainID = "union"
	}

	// Обновляем ConstLabels с новым chain_id
	ConstLabels = prometheus.Labels{
		"chain_id": ChainID,
	}
}

func setDenom(grpcConn *grpc.ClientConn) {
	if isUserProvidedAndHandled := checkAndHandleDenomInfoProvidedByUser(); isUserProvidedAndHandled {
		return
	}

	bankClient := banktypes.NewQueryClient(grpcConn)
	denoms, err := bankClient.DenomsMetadata(
		context.Background(),
		&banktypes.QueryDenomsMetadataRequest{},
	)
	if err != nil {
		log.Fatal().Err(err).Msg("Error querying denom")
	}

	if len(denoms.Metadatas) == 0 {
		log.Fatal().Msg("No denom infos. Try running the binary with --denom and --denom-coefficient to set them manually.")
	}

	metadata := denoms.Metadatas[0]
	if Denom == "" {
		Denom = metadata.Display
	}

	for _, unit := range metadata.DenomUnits {
		log.Debug().
			Str("denom", unit.Denom).
			Uint32("exponent", unit.Exponent).
			Msg("Denom info")
		if unit.Denom == Denom {
			DenomCoefficient = math.Pow10(int(unit.Exponent))
			log.Info().
				Str("denom", Denom).
				Float64("coefficient", DenomCoefficient).
				Msg("Got denom info")
			return
		}
	}

	log.Fatal().Msg("Could not find the denom info")
}

func checkAndHandleDenomInfoProvidedByUser() bool {
	if Denom != "" {
		if DenomCoefficient != 1 && DenomExponent != 0 {
			log.Fatal().Msg("denom-coefficient and denom-exponent are both provided. Must provide only one")
		}

		if DenomCoefficient != 1 {
			log.Info().
				Str("denom", Denom).
				Float64("coefficient", DenomCoefficient).
				Msg("Using provided denom and coefficient.")
			return true
		}

		if DenomExponent != 0 {
			DenomCoefficient = math.Pow10(int(DenomExponent))
			log.Info().
				Str("denom", Denom).
				Uint64("exponent", DenomExponent).
				Float64("calculated coefficient", DenomCoefficient).
				Msg("Using provided denom and denom exponent and calculating coefficient.")
			return true
		}

		return false
	}

	return false
}

func main() {
	rootCmd.PersistentFlags().StringVar(&ConfigPath, "config", "", "Config file path")
	rootCmd.PersistentFlags().StringVar(&Denom, "denom", "", "Cosmos coin denom")
	rootCmd.PersistentFlags().Float64Var(&DenomCoefficient, "denom-coefficient", 1, "Denom coefficient")
	rootCmd.PersistentFlags().Uint64Var(&DenomExponent, "denom-exponent", 0, "Denom exponent")
	rootCmd.PersistentFlags().StringVar(&ListenAddress, "listen-address", ":9300", "The address this exporter would listen on")
	rootCmd.PersistentFlags().StringVar(&NodeAddress, "node", "localhost:9090", "RPC node address")
	rootCmd.PersistentFlags().StringVar(&LogLevel, "log-level", "info", "Logging level")
	rootCmd.PersistentFlags().Uint64Var(&Limit, "limit", 1000, "Pagination limit for gRPC requests")
	rootCmd.PersistentFlags().StringVar(&TendermintRPC, "tendermint-rpc", "http://localhost:26657", "Tendermint RPC address")
	rootCmd.PersistentFlags().BoolVar(&JsonOutput, "json", false, "Output logs as JSON")

	rootCmd.PersistentFlags().StringVar(&Prefix, "bech-prefix", "persistence", "Bech32 global prefix")
	rootCmd.PersistentFlags().StringVar(&AccountPrefix, "bech-account-prefix", "", "Bech32 account prefix")
	rootCmd.PersistentFlags().StringVar(&AccountPubkeyPrefix, "bech-account-pubkey-prefix", "", "Bech32 pubkey account prefix")
	rootCmd.PersistentFlags().StringVar(&ValidatorPrefix, "bech-validator-prefix", "", "Bech32 validator prefix")
	rootCmd.PersistentFlags().StringVar(&ValidatorPubkeyPrefix, "bech-validator-pubkey-prefix", "", "Bech32 pubkey validator prefix")
	rootCmd.PersistentFlags().StringVar(&ConsensusNodePrefix, "bech-consensus-node-prefix", "", "Bech32 consensus node prefix")
	rootCmd.PersistentFlags().StringVar(&ConsensusNodePubkeyPrefix, "bech-consensus-node-pubkey-prefix", "", "Bech32 pubkey consensus node prefix")

	if err := rootCmd.Execute(); err != nil {
		log.Fatal().Err(err).Msg("Could not start application")
	}
}
