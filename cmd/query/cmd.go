package query

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/spf13/cobra"

	"github.com/aldehir/ut2u/pkg/query"
)

var printMutex sync.Mutex
var printCount int

var timeout int

var queryCommand = &cobra.Command{
	Use:   "query [-t timeout] server [server...]",
	Short: "Query a UT2004 server",
	RunE:  doQuery,

	DisableFlagsInUseLine: true,
}

func EnrichCommand(cmd *cobra.Command) {
	cmd.AddCommand(queryCommand)
}

func init() {
	queryCommand.Flags().IntVarP(&timeout, "timeout", "t", 250, "timeout in milliseconds")
}

func doQuery(cmd *cobra.Command, args []string) error {
	ctx, cancel := context.WithCancel(context.Background())

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt)

	go func() {
		_, ok := <-sigs
		if ok {
			cancel()
		}
	}()

	client, err := query.NewClient()
	if err != nil {
		return err
	}
	defer client.Close()

	var wg sync.WaitGroup

	for _, server := range args {
		wg.Add(1)
		go func(server string) {
			defer wg.Done()

			addr, err := net.ResolveUDPAddr("udp", server)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Could not resolve %s: %v\n", server, err)
				return
			}

			// Use query port
			addr.Port = addr.Port + 1

			opts := []query.QueryOption{
				query.WithRules(),
				query.WithPlayers(),
				query.WithTimeout(time.Duration(timeout) * time.Millisecond),
			}

			details, err := client.Query(ctx, addr, opts...)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to query %s: %v\n", addr.String(), err)
				return
			}

			printDetails(details)
		}(server)
	}

	wg.Wait()

	signal.Stop(sigs)
	close(sigs)

	return nil
}

func printDetails(details query.ServerDetails) {
	printMutex.Lock()
	defer printMutex.Unlock()

	if printCount > 0 {
		fmt.Println()
	}

	fmt.Printf("%s\n", details.Info.ServerName)
	fmt.Printf("  Game: %s\n", details.Info.GameType)
	fmt.Printf("  Map: %s\n", details.Info.MapName)
	fmt.Printf("  Players: %d/%d\n", details.Info.CurrentPlayers, details.Info.MaxPlayers)

	for _, p := range details.Players {
		fmt.Printf("    %s: score=%d ping=%d\n", p.Name.Value, p.Score, p.Ping)
	}

	if len(details.Rules) > 0 {
		fmt.Println("  Rules:")
		for _, rule := range details.Rules {
			fmt.Printf("    %s: %s\n", rule.Key, rule.Value)
		}
	}

	printCount++
}
