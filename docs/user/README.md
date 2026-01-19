# User Documentation Plan

This directory contains user-facing documentation for hashi. The documentation is organized to provide both quick reference and comprehensive guidance.

## Current Documentation

- **[examples.md](examples.md)** - Usage examples and common patterns
- **[scripting.md](scripting.md)** - Scripting and automation guidance  
- **[error-handling.md](error-handling.md)** - Error messages and troubleshooting

## Planned Documentation

### Core Usage Documentation
- **getting-started.md** - Installation, first steps, basic concepts
- **command-reference.md** - Complete flag and option reference
- **output-formats.md** - Understanding different output formats (default, verbose, json, plain)
- **algorithms.md** - Supported hash algorithms and when to use them

### Advanced Usage Documentation  
- **configuration.md** - Configuration files, environment variables, precedence
- **filtering.md** - File filtering options (size, date, patterns)
- **performance.md** - Performance tips for large file sets
- **integration.md** - Integrating with CI/CD, build systems, other tools

### Troubleshooting and Reference
- **error-handling.md** ✅ - Error messages and troubleshooting (completed)
- **faq.md** - Frequently asked questions
- **migration.md** - Migrating from other hash tools
- **security.md** - Security considerations for users

## Documentation Principles

### User-Focused Approach
- Start with what users want to accomplish
- Provide working examples first, explanation second
- Use real-world scenarios and file names
- Include both simple and complex use cases

### Error Handling Documentation Strategy
The error handling documentation follows these principles:

1. **Transparency with Security** - Explain that generic errors exist for security reasons
2. **Actionable Guidance** - Always provide next steps (use --verbose, check permissions, etc.)
3. **Progressive Disclosure** - Start with common solutions, provide detailed troubleshooting
4. **Empowerment** - Teach users how to diagnose issues themselves

### Error Message Coverage
The documentation covers three categories of errors:

1. **Validation Errors** - Always specific, immediate feedback
   - Invalid extensions, algorithms, formats
   - Syntax errors, missing arguments
   - File not found, invalid paths

2. **Generic Errors** - Security and system issues
   - Configuration file protection
   - Permission denied, disk full
   - Network issues, path problems
   - Explain why these are generic and how to get details

3. **System Errors** - OS and environment issues
   - File system errors, network problems
   - Resource limitations, platform differences
   - Integration with other tools

## Content Guidelines

### Writing Style
- **Conversational but precise** - Match hashi's friendly, discoverable nature
- **Example-driven** - Show, don't just tell
- **Assumption-aware** - Don't assume deep technical knowledge
- **Solution-oriented** - Focus on helping users succeed

### Error Documentation Specifics
- **Never reveal security internals** - Don't explain the obfuscation strategy
- **Focus on user benefit** - "This helps protect your system"
- **Provide clear escalation paths** - Always explain how to get more info
- **Include realistic examples** - Use actual error messages users will see

### Code Examples
- Use realistic file names and paths
- Show both successful and error cases
- Include expected output
- Demonstrate the --verbose pattern for troubleshooting

## Implementation Priority

### Phase 1: Essential Documentation
1. **getting-started.md** - New user onboarding
2. **command-reference.md** - Complete flag reference
3. **error-handling.md** ✅ - Troubleshooting guide (completed)

### Phase 2: Advanced Usage
1. **configuration.md** - Config files and environment
2. **output-formats.md** - Format options and use cases
3. **algorithms.md** - Hash algorithm guidance

### Phase 3: Integration and Reference
1. **scripting.md** - Enhanced automation guidance
2. **integration.md** - CI/CD and toolchain integration
3. **faq.md** - Common questions and edge cases

## Maintenance Strategy

### Regular Updates
- Update examples when new features are added
- Refresh error message examples when error handling changes
- Add new troubleshooting scenarios based on user feedback

### User Feedback Integration
- Monitor support requests for documentation gaps
- Track which error messages cause the most confusion
- Update based on real user scenarios

### Cross-Reference Maintenance
- Keep error messages in sync with actual implementation
- Ensure examples work with current version
- Update when security model changes

## Success Metrics

### User Experience Goals
- Users can resolve common issues without external help
- Error messages lead to successful resolution
- New users can get started quickly
- Advanced users can find detailed information

### Documentation Quality Indicators
- Reduced support requests for covered topics
- Positive feedback on troubleshooting effectiveness
- Successful user onboarding without confusion
- Clear escalation paths for complex issues