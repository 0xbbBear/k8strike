package plugin

import (
	"fmt"
	"os"
	"sort"
	"text/tabwriter"
)

type ExploitInterface interface {
	Desc() string
	Run(args []string) bool
	GetExploitType() string
}

var Exploits map[string]ExploitInterface

func init() {
	Exploits = make(map[string]ExploitInterface)
}

func ListAllExploit() {

	writer := tabwriter.NewWriter(os.Stdout, 1, 1, 1, ' ', tabwriter.AlignRight|tabwriter.Debug)

	type kv struct {
		Name        string
		ExploitType string
		Desc        string
	}

	sortedExploits := make([]kv, 0)

	for name, plugin := range Exploits {
		sortedExploits = append(sortedExploits, kv{name, plugin.GetExploitType(), plugin.Desc()})
	}

	sort.Slice(sortedExploits, func(i, j int) bool {
		return sortedExploits[i].ExploitType < sortedExploits[j].ExploitType
	})

	fmt.Fprintln(writer, "TYPE \t NAME \t DESC")

	for _, kv := range sortedExploits {
		str := fmt.Sprintf("%s \t %s \t %s", kv.ExploitType, kv.Name, kv.Desc)
		fmt.Fprintln(writer, str)
	}

	writer.Flush()
}

func RunSingleExploit(name string, args []string) {
	Exploits[name].Run(args)
}

func RegisterExploit(name string, exploit ExploitInterface) {
	Exploits[name] = exploit
}
