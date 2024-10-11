package fmt

import (
	"fmt"

	"github.com/soerenschneider/sc-agent/pkg/api"
)

func PrintReplicationHttpItemsList(items []api.ReplicationHttpItem) {
	for i, item := range items {
		fmt.Printf("HTTP Replication Item Set %d:\n", i+1)
		PrintReplicationHttpItem(item)
		fmt.Println()
	}
}

func PrintReplicationHttpItem(item api.ReplicationHttpItem) {
	expectedChecksum := "N/A"
	if item.ExpectedChecksum != nil {
		expectedChecksum = *item.ExpectedChecksum
	}

	fmt.Printf(`  ID                : %s
  Source            : %s
  Destinations      : %v
  Expected Checksum : %s
  Status            : %s
`,
		item.Id,
		item.Source,
		item.DestUris,
		expectedChecksum,
		item.Status,
	)

	if len(item.PostHooks) > 0 {
		fmt.Println("  Post Hooks:")
		for _, hook := range item.PostHooks {
			fmt.Printf("    - Name: %s, Cmd: %s\n", hook.Name, hook.Cmd)
		}
	} else {
		fmt.Println("  Post Hooks        : N/A")
	}
}
