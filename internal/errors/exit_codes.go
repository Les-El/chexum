// Package errors provides error handling and exit code logic.
package errors

import (
	"hashi/internal/config"
	"hashi/internal/hash"
)

// ExitCode constants are defined in internal/config/config.go to avoid circular dependencies
// and because they are part of the application's configuration contract.

// DetermineExitCode determines the final exit code based on the operation results and configuration.
func DetermineExitCode(cfg *config.Config, result *hash.Result) int {
	// 1. Check for interruptions (handled by signals package, but here for completeness)
	// if result.Interrupted { return config.ExitInterrupted }

	// 2. Check for hard failures (any processing error)
	if len(result.Errors) > 0 {
		// If all files failed, return the most specific error if possible
		if len(result.Errors) == len(result.Entries) && len(result.Entries) > 0 {
			// Find the most severe error type
			groups := GroupErrors(result.Errors)
			if _, ok := groups[ErrorTypePermission]; ok {
				return config.ExitPermissionErr
			}
			if _, ok := groups[ErrorTypeFileNotFound]; ok {
				return config.ExitFileNotFound
			}
			if _, ok := groups[ErrorTypeIntegrity]; ok {
				return config.ExitIntegrityFail
			}
		}
		// Otherwise, it's a partial failure
		return config.ExitPartialFailure
	}

	// 3. Match Required Logic
	// If --match-required is set, we only return 0 if at least one match was found.
	if cfg.MatchRequired {
		if len(result.Matches) > 0 {
			return config.ExitSuccess
		}
		return config.ExitNoMatches
	}

	// 4. Default Success
	return config.ExitSuccess
}
