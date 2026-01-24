// Package main provides the entry point for the hashi CLI tool.
//
// hashi is a command-line hash comparison tool that computes and compares
// cryptographic hashes. It follows industry-standard CLI design guidelines
// for human-first design, composability, and robustness.
//
// Usage:
//
//	// hashi [flags] [files...]
//
// When run with no arguments, hashi processes all non-hidden files in the
// current directory.
package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/Les-El/hashi/internal/color"
	"github.com/Les-El/hashi/internal/config"
	"github.com/Les-El/hashi/internal/conflict"
	"github.com/Les-El/hashi/internal/console"
	"github.com/Les-El/hashi/internal/errors"
	"github.com/Les-El/hashi/internal/hash"
	"github.com/Les-El/hashi/internal/output"
	"github.com/Les-El/hashi/internal/progress"
	"github.com/Les-El/hashi/internal/signals"
)

func main() {
	os.Exit(run())
}

func run() int {
	// Set up signal handling
	sigHandler := signals.NewSignalHandler(nil)
	sigHandler.Start()
	defer sigHandler.Stop()

	colorHandler := color.NewColorHandler()
	errHandler := errors.NewErrorHandler(colorHandler)

	// 1. Parse arguments
	cfg, warnings, err := config.ParseArgs(os.Args[1:])
	if err != nil {
		fmt.Fprintln(os.Stderr, errHandler.FormatError(err))
		return config.ExitInvalidArgs
	}

	// 2. Initialize streams
	streams, cleanup, err := console.InitStreams(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing I/O: %v\n", err)
		return config.ExitInvalidArgs
	}
	defer cleanup()

	if len(warnings) > 0 {
		fmt.Fprint(streams.Err, conflict.FormatAllWarnings(warnings))
	}

	// 3. Handle basic flags
	if cfg.ShowHelp {
		fmt.Fprintln(streams.Out, config.HelpText())
		return config.ExitSuccess
	}
	if cfg.ShowVersion {
		fmt.Fprintln(streams.Out, config.VersionText())
		return config.ExitSuccess
	}

	// 4. Discover and validate files
	if err := prepareFiles(cfg, errHandler, streams); err != nil {
		return errors.DetermineDiscoveryExitCode(err)
	}

	// 5. Select and execute mode
	return executeMode(cfg, colorHandler, streams, errHandler)
}

func prepareFiles(cfg *config.Config, errHandler *errors.Handler, streams *console.Streams) error {
	if cfg.HasStdinMarker() {
		cfg.Files = expandStdinFiles(cfg.Files)
	}

	if len(cfg.Files) > 0 || len(cfg.Hashes) == 0 {
		discOpts := hash.DiscoveryOptions{
			Recursive:      cfg.Recursive,
			Hidden:         cfg.Hidden,
			Include:        cfg.Include,
			Exclude:        cfg.Exclude,
			MinSize:        cfg.MinSize,
			MaxSize:        cfg.MaxSize,
			ModifiedAfter:  cfg.ModifiedAfter,
			ModifiedBefore: cfg.ModifiedBefore,
		}
		discovered, err := hash.DiscoverFiles(cfg.Files, discOpts)
		if err != nil {
			fmt.Fprintln(streams.Err, errHandler.FormatError(err))
			return err
		}
		cfg.Files = discovered
	}
	return nil
}

func executeMode(cfg *config.Config, colorHandler *color.Handler, streams *console.Streams, errHandler *errors.Handler) int {
	// Edge case validation
	if len(cfg.Hashes) > 0 {
		if len(cfg.Files) > 1 {
			fmt.Fprintln(streams.Err, errHandler.FormatError(fmt.Errorf("Cannot compare multiple files with hash strings")))
			return config.ExitInvalidArgs
		}
		if cfg.HasStdinMarker() {
			fmt.Fprintln(streams.Err, errHandler.FormatError(fmt.Errorf("Cannot use stdin input with hash comparison")))
			return config.ExitInvalidArgs
		}
	}

	if len(cfg.Files) == 0 && len(cfg.Hashes) > 0 {
		return runHashValidationMode(cfg, colorHandler, streams)
	}
	if len(cfg.Files) == 1 && len(cfg.Hashes) == 1 {
		return runFileHashComparisonMode(cfg, colorHandler, streams)
	}
	if len(cfg.Files) > 0 {
		return runStandardHashingMode(cfg, colorHandler, streams, errHandler)
	}

	return config.ExitSuccess
}

// runStandardHashingMode processes multiple files, computing hashes and formatting output.
func runStandardHashingMode(cfg *config.Config, colorHandler *color.Handler, streams *console.Streams, errHandler *errors.Handler) int {
	computer, err := hash.NewComputer(cfg.Algorithm)
	if err != nil {
		fmt.Fprintln(streams.Err, errHandler.FormatError(err))
		return config.ExitInvalidArgs
	}

	results := &hash.Result{Entries: make([]hash.Entry, 0, len(cfg.Files))}
	bar := setupProgressBar(cfg, streams)
	if bar != nil {
		defer bar.Finish()
	}

	start := time.Now()
	for _, path := range cfg.Files {
		processFile(path, computer, results, bar, cfg, streams, errHandler)
	}
	results.Duration = time.Since(start)

	results.Matches, results.Unmatched = groupResults(results.Entries)
	outputResults(results, cfg, streams)

	return errors.DetermineExitCode(cfg, results)
}

func setupProgressBar(cfg *config.Config, streams *console.Streams) *progress.Bar {
	if !cfg.Quiet && !cfg.Bool {
		return progress.NewBar(&progress.Options{
			Total:       int64(len(cfg.Files)),
			Description: "Hashing files...",
			Writer:      streams.Err,
		})
	}
	return nil
}

func processFile(path string, computer *hash.Computer, results *hash.Result, bar *progress.Bar, cfg *config.Config, streams *console.Streams, errHandler *errors.Handler) {
	entry, err := computer.ComputeFile(path)
	if err != nil {
		results.Errors = append(results.Errors, err)
		results.Entries = append(results.Entries, hash.Entry{Original: path, Error: err})
		if !cfg.Quiet {
			if bar != nil && bar.IsEnabled() {
				fmt.Fprint(streams.Err, "\r\033[K")
			}
			fmt.Fprintln(streams.Err, errHandler.FormatError(err))
		}
	} else {
		results.Entries = append(results.Entries, *entry)
		results.FilesProcessed++
		results.BytesProcessed += entry.Size
	}
	if bar != nil {
		bar.Increment()
	}
}

func outputResults(results *hash.Result, cfg *config.Config, streams *console.Streams) {
	if cfg.Bool {
		success := isSuccess(results, cfg)
		fmt.Fprintln(streams.Out, success)
	} else if !cfg.Quiet {
		formatter := output.NewFormatter(cfg.OutputFormat, cfg.PreserveOrder)
		fmt.Fprintln(streams.Out, formatter.Format(results))
	}
}

func isSuccess(results *hash.Result, cfg *config.Config) bool {
	if cfg.MatchRequired {
		return len(results.Matches) > 0
	}
	if len(results.Entries) == 1 && len(results.Errors) == 0 {
		return true
	}
	return len(results.Matches) == 1 && len(results.Unmatched) == 0
}

// groupResults categorizes entries into matches and unique hashes.
func groupResults(entries []hash.Entry) ([]hash.MatchGroup, []hash.Entry) {
	groups := make(map[string][]hash.Entry)
	for _, e := range entries {
		if e.Error == nil {
			groups[e.Hash] = append(groups[e.Hash], e)
		}
	}

	var matches []hash.MatchGroup
	var unmatched []hash.Entry

	for h, groupEntries := range groups {
		if len(groupEntries) > 1 {
			matches = append(matches, hash.MatchGroup{
				Hash:    h,
				Entries: groupEntries,
				Count:   len(groupEntries),
			})
		} else {
			unmatched = append(unmatched, groupEntries[0])
		}
	}

	return matches, unmatched
}

// runHashValidationMode validates hash strings and displays results.
// This mode is triggered when no files are provided, only hash strings.
// Requirements: 24.1, 24.2, 24.3
func runHashValidationMode(cfg *config.Config, colorHandler *color.Handler, streams *console.Streams) int {
	allValid := true
	for _, hashStr := range cfg.Hashes {
		if !validateHash(hashStr, cfg, colorHandler, streams) {
			allValid = false
		}
	}

	if allValid {
		return config.ExitSuccess
	}
	return config.ExitInvalidArgs
}

func validateHash(hashStr string, cfg *config.Config, colorHandler *color.Handler, streams *console.Streams) bool {
	algorithms := hash.DetectHashAlgorithm(hashStr)
	if len(algorithms) == 0 {
		reportInvalidHash(hashStr, cfg, colorHandler, streams)
		return false
	}

	reportValidHash(hashStr, algorithms, cfg, colorHandler, streams)
	return true
}

func reportInvalidHash(hashStr string, cfg *config.Config, colorHandler *color.Handler, streams *console.Streams) {
	if !cfg.Quiet {
		fmt.Fprintf(streams.Err, "%s %s - Invalid hash format\n", colorHandler.Red("✗"), hashStr)
		fmt.Fprintf(streams.Err, "  Hash strings must contain only hexadecimal characters and have a valid length.\n")
	}
}

func reportValidHash(hashStr string, algorithms []string, cfg *config.Config, colorHandler *color.Handler, streams *console.Streams) {
	if !cfg.Quiet {
		fmt.Fprintf(streams.Err, "%s %s - Valid hash\n", colorHandler.Green("✓"), hashStr)
		if len(algorithms) == 1 {
			fmt.Fprintf(streams.Err, "  Algorithm: %s\n", algorithms[0])
		} else {
			fmt.Fprintf(streams.Err, "  Possible algorithms: %s\n", formatAlgorithmList(algorithms))
		}
	}
}

// formatAlgorithmList formats a list of algorithms for display.
func formatAlgorithmList(algorithms []string) string {
	if len(algorithms) == 0 {
		return ""
	}
	if len(algorithms) == 1 {
		return algorithms[0]
	}
	
	result := ""
	for i, alg := range algorithms {
		if i > 0 {
			if i == len(algorithms)-1 {
				result += " or "
			} else {
				result += ", "
			}
		}
		result += alg
	}
	return result
}

// runFileHashComparisonMode compares a file's hash against a provided hash string.
// This mode is triggered when exactly one file and one hash string are provided.
// Requirements: 25.1, 25.2, 25.3
func runFileHashComparisonMode(cfg *config.Config, colorHandler *color.Handler, streams *console.Streams) int {
	filePath := cfg.Files[0]
	expectedHash := cfg.Hashes[0]

	computer, err := hash.NewComputer(cfg.Algorithm)
	if err != nil {
		handleComparisonError(err, "Failed to initialize hash computer", cfg, colorHandler, streams)
		return config.ExitInvalidArgs
	}

	entry, err := computer.ComputeFile(filePath)
	if err != nil {
		handleComparisonError(err, fmt.Sprintf("Failed to compute hash for %s", filePath), cfg, colorHandler, streams)
		return errors.DetermineDiscoveryExitCode(err)
	}

	match := strings.EqualFold(entry.Hash, expectedHash)
	outputComparisonResult(match, filePath, expectedHash, entry.Hash, cfg, colorHandler, streams)

	if match {
		return config.ExitSuccess
	}
	return config.ExitNoMatches
}

func handleComparisonError(err error, msg string, cfg *config.Config, colorHandler *color.Handler, streams *console.Streams) {
	if !cfg.Quiet {
		fmt.Fprintf(streams.Err, "%s %s: %v\n", colorHandler.Red("✗"), msg, err)
	}
}

func outputComparisonResult(match bool, filePath, expected, computed string, cfg *config.Config, colorHandler *color.Handler, streams *console.Streams) {
	if cfg.Bool {
		fmt.Fprintln(streams.Out, match)
		return
	}
	if cfg.Quiet {
		return
	}

	if match {
		fmt.Fprintf(streams.Out, "%s %s\n", colorHandler.Green("PASS"), filePath)
	} else {
		fmt.Fprintf(streams.Out, "%s %s\n", colorHandler.Red("FAIL"), filePath)
		fmt.Fprintf(streams.Out, "  Expected: %s\n", expected)
		fmt.Fprintf(streams.Out, "  Computed: %s\n", computed)
	}
}
		
		// expandStdinFiles reads file paths from stdin and adds them to the file list.
		func expandStdinFiles(files []string) []string {
			var result []string
			
			// Remove the "-" marker
			for _, f := range files {
				if f != "-" {
					result = append(result, f)
				}
			}
			
			// Read from stdin
			scanner := bufio.NewScanner(os.Stdin)
			for scanner.Scan() {
				path := strings.TrimSpace(scanner.Text())
				if path != "" {
					result = append(result, path)
				}
			}
			
			return result
		}
		