package cli

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/FactomProject/factom"
	walletlib "github.com/FactomProject/wallet/v2"
)

// Run executes the wallet subcommands handled directly by anchorchaind.
func Run(args []string) error {
	if len(args) == 0 {
		return errors.New("wallet command requires a subcommand (import, addresses, balance)")
	}

	switch args[0] {
	case "import":
		return runImport(args[1:])
	case "addresses":
		return runAddresses(args[1:])
	case "balance":
		return runBalance(args[1:])
	default:
		return fmt.Errorf("unknown wallet subcommand: %s", args[0])
	}
}

func runImport(args []string) error {
	fs := flag.NewFlagSet("wallet import", flag.ContinueOnError)
	mnemonic := fs.String("mnemonic", "", "BIP39 mnemonic phrase to seed the wallet")
	path := fs.String("path", defaultWalletPath(), "wallet database path")
	overwrite := fs.Bool("overwrite", false, "replace the existing wallet if present")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if strings.TrimSpace(*mnemonic) == "" {
		return errors.New("--mnemonic is required")
	}

	if _, err := os.Stat(*path); err == nil {
		if !*overwrite {
			return fmt.Errorf("wallet already exists at %s (use --overwrite to replace)", *path)
		}
		if err := os.Remove(*path); err != nil {
			return fmt.Errorf("failed to remove existing wallet: %w", err)
		}
	}

	w, err := walletlib.ImportWalletFromMnemonic(*mnemonic, *path)
	if err != nil {
		return err
	}
	defer w.Close()

	fmt.Printf("Imported wallet seed into %s\n", *path)
	return nil
}

func runAddresses(args []string) error {
	fs := flag.NewFlagSet("wallet addresses", flag.ContinueOnError)
	path := fs.String("path", defaultWalletPath(), "wallet database path")
	if err := fs.Parse(args); err != nil {
		return err
	}

	w, err := walletlib.NewOrOpenBoltDBWallet(*path)
	if err != nil {
		return err
	}
	defer w.Close()

	factoidAddrs, ecAddrs, err := w.GetAllAddresses()
	if err != nil {
		return err
	}

	fmt.Printf("Wallet: %s\n", *path)
	fmt.Println("Factoid addresses:")
	if len(factoidAddrs) == 0 {
		fmt.Println("  (none)")
	}
	for _, addr := range factoidAddrs {
		fmt.Printf("  %s\n", addr.String())
	}

	fmt.Println("Entry credit addresses:")
	if len(ecAddrs) == 0 {
		fmt.Println("  (none)")
	}
	for _, addr := range ecAddrs {
		fmt.Printf("  %s\n", addr.String())
	}

	return nil
}

func runBalance(args []string) error {
	fs := flag.NewFlagSet("wallet balance", flag.ContinueOnError)
	address := fs.String("address", "", "Address to query (FA... or EC...)")
	factomdServer := fs.String("factomd", "localhost:8088", "factomd API host:port")
	if err := fs.Parse(args); err != nil {
		return err
	}
	addr := strings.TrimSpace(*address)
	if addr == "" {
		return errors.New("--address is required")
	}

	factom.SetFactomdServer(*factomdServer)

	switch {
	case strings.HasPrefix(addr, "FA"):
		bal, err := factom.GetFactoidBalance(addr)
		if err != nil {
			return err
		}
		fmt.Printf("Factoid balance for %s: %d\n", addr, bal)
	case strings.HasPrefix(addr, "EC"):
		bal, err := factom.GetECBalance(addr)
		if err != nil {
			return err
		}
		fmt.Printf("Entry credit balance for %s: %d\n", addr, bal)
	default:
		return fmt.Errorf("unrecognized address prefix for %s (expected FA... or EC...)", addr)
	}

	return nil
}

func defaultWalletPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "anchor_wallet.db"
	}
	anchorPath := filepath.Join(home, ".anchorchain", "wallet", "anchor_wallet.db")
	legacyPath := filepath.Join(home, ".factom", "wallet", "factom_wallet.db")
	if _, err := os.Stat(anchorPath); err == nil {
		return anchorPath
	}
	if _, err := os.Stat(legacyPath); err == nil {
		return legacyPath
	}
	return anchorPath
}
