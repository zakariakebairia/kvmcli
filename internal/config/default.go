package config

// This file defines the main configuration for kvmcli.
//
// The configuration is planned to be stored in TOML format (subject to validation).
// Its primary purpose is to centralize default values used by kvmcli, including but
// not limited to:
//   - Default CPU and memory allocations
//   - MAC address pattern
//   - Naming conventions for networks and virtual machines
//
// Centralizing these defaults avoids hardcoding values across the codebase and
// improves maintainability and flexibility.
//
// Several VM-related parameters currently hardcoded in the domain XML definitions
// are strong candidates to be moved into this configuration layer.
//
// In the future, this configuration may evolve into a standalone SQLite database
// if it proves to be a better fit for managing defaults and metadata. This decision
// will be made after further experimentation and validation.
