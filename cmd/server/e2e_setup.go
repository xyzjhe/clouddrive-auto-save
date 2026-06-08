//go:build e2e

package main

import "github.com/zcq/clouddrive-auto-save/internal/core"

func setupE2EMock() {
	core.SetupE2EHTTPMock()
}
