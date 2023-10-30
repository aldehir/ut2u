package query

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/aldehir/ut2u/pkg/query"
)

var timeout int

var formatterName string
var formatter Formatter

var queryCommand = &cobra.Command{
	Use:   "query [-t timeout] [-f plain|json] server [server...]",
	Short: "Query a UT2004 server",
	RunE:  doQuery,

	DisableFlagsInUseLine: true,
}

func EnrichCommand(cmd *cobra.Command) {
	cmd.AddCommand(queryCommand)
}

func init() {
	queryCommand.Flags().IntVarP(&timeout, "timeout", "t", 250, "timeout in milliseconds")
	queryCommand.Flags().StringVarP(&formatterName, "format", "f", "plain", "format (plain, json)")
}

func doQuery(cmd *cobra.Command, args []string) error {
	ctx, cancel := context.WithCancel(context.Background())

	if strings.EqualFold(formatterName, "json") {
		formatter = &JSONFormatter{}
	} else {
		formatter = &ConsoleFormatter{}
	}

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt)
	defer func() {
		signal.Stop(sigs)
		close(sigs)
	}()

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

	reports := make(chan Server, 10)
	defer close(reports)

	count := 0
	for _, server := range args {
		count += 1

		go func(server string) {
			var rpt Server
			rpt.Address = server

			addr, err := net.ResolveUDPAddr("udp", server)
			if err != nil {
				rpt.Status.Success = false
				rpt.Status.Message = err.Error()
				reports <- rpt
				return
			}

			rpt.IP = addr.IP.String()
			rpt.Port = addr.Port
			rpt.QueryPort = addr.Port + 1

			// Use query port
			addr.Port = addr.Port + 1

			opts := []query.QueryOption{
				query.WithRules(),
				query.WithPlayers(),
				query.WithTimeout(time.Duration(timeout) * time.Millisecond),
			}

			details, err := client.Query(ctx, addr, opts...)
			if err != nil {
				rpt.Status.Success = false
				rpt.Status.Message = fmt.Sprintf("Failed to query %s: %v\n", addr.String(), err)
				reports <- rpt
				return
			}

			rpt.Status.Success = true
			rpt.Status.Message = "success"

			rpt.Info = CreateServerInfo(details.Info)
			rpt.Rules = CreateRules(details.Rules)
			rpt.Players, rpt.Teams = CreatePlayersAndTeams(details.Players, int(details.Info.CurrentPlayers))

			reports <- rpt
		}(server)
	}

	for i := 0; i < count; i++ {
		rpt := <-reports

		err = formatter.Report(rpt)
		if err != nil {
			return err
		}
	}

	err = formatter.Flush()
	if err != nil {
		return err
	}

	return nil
}
