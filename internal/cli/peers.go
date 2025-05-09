package cli

import (
	"strings"

	"github.com/mitchellh/cli"
)

// PeersCommand is the command to group the peers commands
type PeersCommand struct {
	UI cli.Ui
}

// MarkDown implements cli.MarkDown interface
func (c *PeersCommand) MarkDown() string {
	items := []string{
		"# Peers",
		"The ```peers``` command groups actions to interact with peers:",
		"- [```peers add```](./peers_add.md): Joins the local client to another remote peer.",
		"- [```peers list```](./peers_list.md): Lists the connected peers to the Zena client.",
		"- [```peers remove```](./peers_remove.md): Disconnects the local client from a connected peer if exists.",
		"- [```peers status```](./peers_status.md): Display the status of a peer by its id.",
	}

	return strings.Join(items, "\n\n")
}

// Help implements the cli.Command interface
func (c *PeersCommand) Help() string {
	return `Usage: zena peers <subcommand>

  This command groups actions to interact with peers.
	
  List the connected peers:
  
    $ zena peers list
	
  Add a new peer by enode:
  
    $ zena peers add <enode>

  Remove a connected peer by enode:

    $ zena peers remove <enode>

  Display information about a peer:

    $ zena peers status <peer id>`
}

// Synopsis implements the cli.Command interface
func (c *PeersCommand) Synopsis() string {
	return "Interact with peers"
}

// Run implements the cli.Command interface
func (c *PeersCommand) Run(args []string) int {
	return cli.RunResultHelp
}
