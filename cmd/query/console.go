package query

import (
	"fmt"
	"os"
)

type ConsoleFormatter struct {
	printCount int
}

func (f *ConsoleFormatter) Report(rpt Server) error {
	if f.printCount > 0 {
		fmt.Println()
	}

	defer func() {
		f.printCount++
	}()

	if !rpt.Status.Success {
		fmt.Fprintf(os.Stderr, "Error: %s\n", rpt.Status.Message)
		return nil
	}

	fmt.Printf("%s\n", rpt.Info.Name)
	fmt.Printf("  Game: %s\n", rpt.Info.GameType)
	fmt.Printf("  Map: %s\n", rpt.Info.Map)

	if len(rpt.Teams) > 1 {
		fmt.Println("  Teams:")
		for _, t := range rpt.Teams {
			fmt.Printf("    %s: team=%d score=%d\n", t.Name, t.Index, t.Score)
		}
	}

	fmt.Printf("  Players: %d/%d\n", rpt.Info.Players.Current, rpt.Info.Players.Max)
	for _, p := range rpt.Players {
		fmt.Printf("    %s: index=%d team=%d spec=%v score=%d ping=%d\n", p.Name, p.Index, p.Team, p.Spectator, p.Score, p.Ping)
	}

	if len(rpt.Rules) > 0 {
		fmt.Println("  Rules:")
		for _, rule := range rpt.Rules {
			fmt.Printf("    %s: %s\n", rule.Key, rule.Value)
		}
	}

	return nil
}

func (f *ConsoleFormatter) Flush() error {
	return nil
}
