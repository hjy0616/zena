package utils

import (
	"os"

	"github.com/urfave/cli/v2"

	"github.com/zenanetwork/go-zenanet/eth"
	"github.com/zenanetwork/go-zenanet/eth/ethconfig"
	"github.com/zenanetwork/go-zenanet/node"
)

var (
	//
	// Zena Specific flags
	//

	// HeimdallURLFlag flag for heimdall url
	HeimdallURLFlag = &cli.StringFlag{
		Name:  "zena.eirene",
		Usage: "URL of Heimdall service",
		Value: "http://localhost:1317",
	}

	// WithoutHeimdallFlag no heimdall (for testing purpose)
	WithoutHeimdallFlag = &cli.BoolFlag{
		Name:  "zena.withoutheimdall",
		Usage: "Run without Heimdall service (for testing purpose)",
	}

	// HeimdallgRPCAddressFlag flag for heimdall gRPC address
	HeimdallgRPCAddressFlag = &cli.StringFlag{
		Name:  "zena.heimdallgRPC",
		Usage: "Address of Heimdall gRPC service",
		Value: "",
	}

	// RunHeimdallFlag flag for running heimdall internally from zena
	RunHeimdallFlag = &cli.BoolFlag{
		Name:  "zena.runheimdall",
		Usage: "Run Heimdall service as a child process",
	}

	RunHeimdallArgsFlag = &cli.StringFlag{
		Name:  "zena.runheimdallargs",
		Usage: "Arguments to pass to Heimdall service",
		Value: "",
	}

	// UseHeimdallApp flag for using internal heimdall app to fetch data
	UseHeimdallAppFlag = &cli.BoolFlag{
		Name:  "zena.useheimdallapp",
		Usage: "Use child heimdall process to fetch data, Only works when zena.runheimdall is true",
	}

	// ZenaFlags all zena related flags
	ZenaFlags = []cli.Flag{
		HeimdallURLFlag,
		WithoutHeimdallFlag,
		HeimdallgRPCAddressFlag,
		RunHeimdallFlag,
		RunHeimdallArgsFlag,
		UseHeimdallAppFlag,
	}
)

// SetZenaConfig sets zena config
func SetZenaConfig(ctx *cli.Context, cfg *eth.Config) {
	cfg.HeimdallURL = ctx.String(HeimdallURLFlag.Name)
	cfg.WithoutHeimdall = ctx.Bool(WithoutHeimdallFlag.Name)
	cfg.HeimdallgRPCAddress = ctx.String(HeimdallgRPCAddressFlag.Name)
	cfg.RunHeimdall = ctx.Bool(RunHeimdallFlag.Name)
	cfg.RunHeimdallArgs = ctx.String(RunHeimdallArgsFlag.Name)
	cfg.UseHeimdallApp = ctx.Bool(UseHeimdallAppFlag.Name)
}

// CreateZenaZenanet Creates zena zenanet object from eth.Config
func CreateZenaZenanet(cfg *ethconfig.Config) *eth.Zenanet {
	workspace, err := os.MkdirTemp("", "zena-command-node-")
	if err != nil {
		Fatalf("Failed to create temporary keystore: %v", err)
	}

	// Create a networkless protocol stack and start an Zenanet service within
	stack, err := node.New(&node.Config{DataDir: workspace, UseLightweightKDF: true, Name: "zena-command-node"})
	if err != nil {
		Fatalf("Failed to create node: %v", err)
	}

	zenanet, err := eth.New(stack, cfg)
	if err != nil {
		Fatalf("Failed to register Zenanet protocol: %v", err)
	}

	// Start the node and assemble the JavaScript console around it
	if err = stack.Start(); err != nil {
		Fatalf("Failed to start stack: %v", err)
	}

	stack.Attach()

	return zenanet
}
