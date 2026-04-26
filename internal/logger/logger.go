package logger

import (
	"fmt"
	"os"
)

var verbose bool

// SetVerbose controls whether Result shows detailed error messages.
func SetVerbose(v bool) { verbose = v }

// Result prints a k8s-style status line for a resource operation.
// If the operation succeeded (err is nil), prints "resource action".
// If it failed and verbose is on, prints the full error detail.
// IDEA: logger.log.info
func Info(resource, action string, err error) {
	if err == nil {
		fmt.Printf("%s %s\n", resource, action)
		return
	}
	if verbose {
		fmt.Printf("%s %s failed: %v\n", resource, action, err)
	} else {
		// fmt.Printf("%s %s failed: %v\n", resource, action, err)
		fmt.Printf("%s %s failed\n", resource, action)
	}
}

// Errorf prints a colored error message to stderr.
func Errorf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "\033[31merror: \033[0m"+format+"\n", args...)
}

// Warnf prints a colored warning message to stderr.
func Warnf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "\033[33mwarning:\033[0m "+format+"\n", args...)
}

// Fatalf prints a colored error message to stderr and exits.
func Fatalf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "\033[31mfatal:\033[0m "+format+"\n", args...)
	os.Exit(1)
}
