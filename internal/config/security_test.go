package config

import (
	"strings"
	"testing"

	"hashi/internal/security"
)

func TestValidateOutputPath(t *testing.T) {
	tests := []struct {
		name         string
		path         string
		verbose      bool
		shouldErr    bool
		errMsg       string
		verboseErrMsg string // Expected error message in verbose mode
	}{
		// Valid paths
		{
			name:      "empty path allowed",
			path:      "",
			verbose:   false,
			shouldErr: false,
		},
		{
			name:      "txt file allowed",
			path:      "output.txt",
			verbose:   false,
			shouldErr: false,
		},
		{
			name:      "json file allowed",
			path:      "data.json",
			verbose:   false,
			shouldErr: false,
		},
		{
			name:      "csv file allowed",
			path:      "report.csv",
			verbose:   false,
			shouldErr: false,
		},
		{
			name:      "uppercase extension allowed",
			path:      "OUTPUT.TXT",
			verbose:   false,
			shouldErr: false,
		},
		{
			name:      "path with directory allowed",
			path:      "logs/output.txt",
			verbose:   false,
			shouldErr: false,
		},
		{
			name:      "absolute path allowed",
			path:      "/tmp/results.json",
			verbose:   false,
			shouldErr: false,
		},
		
		// Invalid extensions (these should always be specific)
		{
			name:      "shell script blocked",
			path:      "malicious.sh",
			verbose:   false,
			shouldErr: true,
			errMsg:    "output files must have extension",
		},
		{
			name:      "python script blocked",
			path:      "script.py",
			verbose:   false,
			shouldErr: true,
			errMsg:    "output files must have extension",
		},
		{
			name:      "executable blocked",
			path:      "hashi",
			verbose:   false,
			shouldErr: true,
			errMsg:    "output files must have extension",
		},
		{
			name:      "toml file blocked",
			path:      "config.toml",
			verbose:   false,
			shouldErr: true,
			errMsg:    "output files must have extension",
		},
		
		// Default blacklist patterns (security-sensitive - generic in non-verbose)
		{
			name:         "config file blocked - non-verbose",
			path:         "config.txt",
			verbose:      false,
			shouldErr:    true,
			errMsg:       "Unknown write/append error",
		},
		{
			name:         "config file blocked - verbose",
			path:         "config.txt",
			verbose:      true,
			shouldErr:    true,
			verboseErrMsg: "cannot write to file matching security pattern: config",
		},
		{
			name:         "secret file blocked - non-verbose",
			path:         "secret.json",
			verbose:      false,
			shouldErr:    true,
			errMsg:       "Unknown write/append error",
		},
		{
			name:         "secret file blocked - verbose",
			path:         "secret.json",
			verbose:      true,
			shouldErr:    true,
			verboseErrMsg: "cannot write to file matching security pattern: secret",
		},
		
		// Legacy hard-coded patterns (security-sensitive - generic in non-verbose)
		{
			name:         "hashi config blocked - non-verbose",
			path:         ".hashi.toml",
			verbose:      false,
			shouldErr:    true,
			errMsg:       "output files must have extension", // Caught by extension check first
		},
		
		// Config directories blocked (security-sensitive - generic in non-verbose)
		{
			name:         "hashi directory blocked - non-verbose",
			path:         ".hashi/output.txt",
			verbose:      false,
			shouldErr:    true,
			errMsg:       "Unknown write/append error",
		},
		{
			name:         "hashi directory blocked - verbose",
			path:         ".hashi/output.txt",
			verbose:      true,
			shouldErr:    true,
			verboseErrMsg: "cannot write to configuration directory",
		},
		{
			name:         "config hashi directory blocked - non-verbose",
			path:         "/home/user/.config/hashi/output.txt",
			verbose:      false,
			shouldErr:    true,
			errMsg:       "Unknown write/append error",
		},
		{
			name:         "config hashi directory blocked - verbose",
			path:         "/home/user/.config/hashi/output.txt",
			verbose:      true,
			shouldErr:    true,
			verboseErrMsg: "cannot write to configuration directory",
		},
		{
			name:         "case insensitive directory blocked - non-verbose",
			path:         ".HASHI/output.txt",
			verbose:      false,
			shouldErr:    true,
			errMsg:       "Unknown write/append error",
		},
		{
			name:         "case insensitive directory blocked - verbose",
			path:         ".HASHI/output.txt",
			verbose:      true,
			shouldErr:    true,
			verboseErrMsg: "cannot write to configuration directory",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a config with the verbose setting
			cfg := DefaultConfig()
			cfg.Verbose = tt.verbose
			
			err := validateOutputPath(tt.path, cfg)
			
			if tt.shouldErr {
				if err == nil {
					t.Errorf("validateOutputPath(%q, cfg) expected error, got nil", tt.path)
					return
				}
				
				expectedMsg := tt.errMsg
				if tt.verbose && tt.verboseErrMsg != "" {
					expectedMsg = tt.verboseErrMsg
				}
				
				if expectedMsg != "" && !strings.Contains(err.Error(), expectedMsg) {
					t.Errorf("validateOutputPath(%q, cfg) error %q should contain %q", tt.path, err.Error(), expectedMsg)
				}
			} else {
				if err != nil {
					t.Errorf("validateOutputPath(%q, cfg) expected no error, got %v", tt.path, err)
				}
			}
		})
	}
}

func TestValidateConfigWithSecurity(t *testing.T) {
	tests := []struct {
		name       string
		outputFile string
		logFile    string
		logJSON    string
		shouldErr  bool
		errMsg     string
	}{
		{
			name:       "all safe paths",
			outputFile: "results.txt",
			logFile:    "app.txt",
			logJSON:    "debug.json",
			shouldErr:  false,
		},
		{
			name:       "unsafe output file",
			outputFile: ".hashi.toml",
			shouldErr:  true,
			errMsg:     "output file",
		},
		{
			name:       "unsafe log file - default blacklist",
			logFile:    "config.txt",
			shouldErr:  true,
			errMsg:     "log file",
		},
		{
			name:       "unsafe JSON log file",
			logJSON:    ".hashi/debug.json",
			shouldErr:  true,
			errMsg:     "JSON log file",
		},
		{
			name:       "empty paths allowed",
			outputFile: "",
			logFile:    "",
			logJSON:    "",
			shouldErr:  false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := DefaultConfig()
			cfg.OutputFile = tt.outputFile
			cfg.LogFile = tt.logFile
			cfg.LogJSON = tt.logJSON
			
			_, err := ValidateConfig(cfg)
			
			if tt.shouldErr {
				if err == nil {
					t.Errorf("ValidateConfig() expected error, got nil")
					return
				}
				if tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("ValidateConfig() error %q should contain %q", err.Error(), tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("ValidateConfig() expected no error, got %v", err)
				}
			}
		})
	}
}

// TestSecurityValidationFunctions tests the new security validation functions
func TestSecurityValidationFunctions(t *testing.T) {
	t.Run("validateFileName", func(t *testing.T) {
		cfg := DefaultConfig()
		cfg.Verbose = true // Enable verbose mode to see detailed error messages
		
		// Test default blacklist patterns
		tests := []struct {
			filename  string
			shouldErr bool
			errMsg    string
		}{
			{"safe.txt", false, ""},
			{"config.txt", true, "security pattern"},
			{"secret.txt", true, "security pattern"},
			{"password.txt", true, "security pattern"},
			{"key.txt", true, "security pattern"},
			{"credential.txt", true, "security pattern"},
			{"CONFIG.TXT", true, "security pattern"}, // Case insensitive
			{"configfile.txt", true, "security pattern"}, // Starts with "config"
		}
		
				for _, tt := range tests {
		
					opts := security.Options{
		
						Verbose:        cfg.Verbose,
		
						BlacklistFiles: cfg.BlacklistFiles,
		
						WhitelistFiles: cfg.WhitelistFiles,
		
					}
		
					err := security.ValidateFileName(tt.filename, opts)
		
					if tt.shouldErr {
		
						if err == nil {
		
							t.Errorf("ValidateFileName(%q) expected error, got nil", tt.filename)
		
						} else if tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
		
							t.Errorf("ValidateFileName(%q) error %q should contain %q", tt.filename, err.Error(), tt.errMsg)
		
						}
		
					} else {
		
						if err != nil {
		
							t.Errorf("ValidateFileName(%q) expected no error, got %v", tt.filename, err)
		
						}
		
					}
		
				}
		
			})
		
			
		
			t.Run("validateFileName with custom patterns", func(t *testing.T) {
		
				cfg := DefaultConfig()
		
				cfg.BlacklistFiles = []string{"temp*", "draft*"}
		
				cfg.WhitelistFiles = []string{"important_config.txt"}
		
				
		
				tests := []struct {
		
					filename  string
		
					shouldErr bool
		
					errMsg    string
		
				}{
		
					{"temp_file.txt", true, "security pattern"},
		
					{"draft_document.txt", true, "security pattern"},
		
					{"config.txt", true, "security pattern"}, // Default pattern
		
					{"important_config.txt", false, ""}, // Whitelisted
		
					{"normal.txt", false, ""},
		
				}
		
				
		
				for _, tt := range tests {
		
					opts := security.Options{
		
						Verbose:        cfg.Verbose,
		
						BlacklistFiles: cfg.BlacklistFiles,
		
						WhitelistFiles: cfg.WhitelistFiles,
		
					}
		
					err := security.ValidateFileName(tt.filename, opts)
		
					if tt.shouldErr {
		
						if err == nil {
		
							t.Errorf("ValidateFileName(%q) expected error, got nil", tt.filename)
		
						}
		
					} else {
		
						if err != nil {
		
							t.Errorf("ValidateFileName(%q) expected no error, got %v", tt.filename, err)
		
						}
		
					}
		
				}
		
			})
		
			
		
			t.Run("validateDirPath", func(t *testing.T) {
		
				cfg := DefaultConfig()
		
				cfg.Verbose = true // Enable verbose mode to see detailed error messages
		
				
		
				tests := []struct {
		
					path      string
		
					shouldErr bool
		
					errMsg    string
		
				}{
		
					{"safe/output.txt", false, ""},
		
					{"config/output.txt", true, "security pattern"},
		
					{"secret/data/output.txt", true, "security pattern"},
		
					{"logs/password/output.txt", true, "security pattern"},
		
					{"CONFIG/OUTPUT.TXT", true, "security pattern"}, // Case insensitive
		
					{"configdir/output.txt", true, "security pattern"}, // Starts with "config"
		
				}
		
				
		
				for _, tt := range tests {
		
					opts := security.Options{
		
						Verbose:        cfg.Verbose,
		
						BlacklistDirs:  cfg.BlacklistDirs,
		
						WhitelistDirs:  cfg.WhitelistDirs,
		
					}
		
					err := security.ValidateDirPath(tt.path, opts)
		
					if tt.shouldErr {
		
						if err == nil {
		
							t.Errorf("ValidateDirPath(%q) expected error, got nil", tt.path)
		
						} else if tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
		
							t.Errorf("ValidateDirPath(%q) error %q should contain %q", tt.path, err.Error(), tt.errMsg)
		
						}
		
					} else {
		
						if err != nil {
		
							t.Errorf("ValidateDirPath(%q) expected no error, got %v", tt.path, err)
		
						}
		
					}
		
				}
		
			})
		
			
		
			t.Run("validateDirPath with custom patterns and whitelist", func(t *testing.T) {
		
				cfg := DefaultConfig()
		
				cfg.BlacklistDirs = []string{"cache", "tmp*"}
		
				cfg.WhitelistDirs = []string{"important_config"}
		
				
		
				tests := []struct {
		
					path      string
		
					shouldErr bool
		
				}{
		
					{"cache/output.txt", true},
		
					{"tmp_files/output.txt", true},
		
					{"config/output.txt", true}, // Default pattern
		
					{"important_config/output.txt", false}, // Whitelisted
		
					{"normal/output.txt", false},
		
				}
		
				
		
				for _, tt := range tests {
		
					opts := security.Options{
		
						Verbose:        cfg.Verbose,
		
						BlacklistDirs:  cfg.BlacklistDirs,
		
						WhitelistDirs:  cfg.WhitelistDirs,
		
					}
		
					err := security.ValidateDirPath(tt.path, opts)
		
					if tt.shouldErr {
		
						if err == nil {
		
							t.Errorf("ValidateDirPath(%q) expected error, got nil", tt.path)
		
						}
		
					} else {
		
						if err != nil {
		
							t.Errorf("ValidateDirPath(%q) expected no error, got %v", tt.path, err)
		
						}
		
					}
		
				}
		
			})
		
		}
		
		