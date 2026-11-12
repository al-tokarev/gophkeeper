// // Команда staticlint запускает multichecker для статического анализа Go-кода.
package main

import (
	"github.com/gordonklaus/ineffassign/pkg/ineffassign"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/copylock"
	"golang.org/x/tools/go/analysis/passes/loopclosure"
	"golang.org/x/tools/go/analysis/passes/lostcancel"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/structtag"
	"golang.org/x/tools/go/analysis/passes/unmarshal"
	"golang.org/x/tools/go/analysis/passes/unreachable"
	"honnef.co/go/tools/quickfix"
	"honnef.co/go/tools/staticcheck"
	"honnef.co/go/tools/stylecheck"
)

func main() {
	multichecker.Main(
		getAnalyzers()...,
	)
}

// getAnalyzers собирает список анализаторов для multichecker.
func getAnalyzers() []*analysis.Analyzer {
	analyzers := []*analysis.Analyzer{
		ineffassign.Analyzer,
		copylock.Analyzer,
		loopclosure.Analyzer,
		lostcancel.Analyzer,
		printf.Analyzer,
		structtag.Analyzer,
		unreachable.Analyzer,
		unmarshal.Analyzer,
	}

	for _, a := range staticcheck.Analyzers {
		analyzers = append(analyzers, a.Analyzer)
	}

	checks := map[string]bool{
		"ST1001": true,
		"ST1003": true,
		"QF1002": true,
		"QF1010": true,
	}

	for _, a := range stylecheck.Analyzers {
		if checks[a.Analyzer.Name] {
			analyzers = append(analyzers, a.Analyzer)
		}
	}
	for _, a := range quickfix.Analyzers {
		if checks[a.Analyzer.Name] {
			analyzers = append(analyzers, a.Analyzer)
		}
	}

	return analyzers
}
