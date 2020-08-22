// Package params defines all custom parameter configurations
// for running end to end tests.
package params

//import ()

// Params struct defines the parameters needed for running E2E tests to properly handle test sharding.
type Params struct {
	LogPath string
}

// TestParams is the globally accessible var for getting config elements.
var TestParams *Params

// LogFileName is the file name used for the GoDBLedger logs.
var LogFileName = "godbledgerE2Etest.log"

// Init initializes the E2E config, properly handling test sharding.
func Init() error {

	TestParams = &Params{
		LogPath: "../build/cache",
	}
	return nil
}
