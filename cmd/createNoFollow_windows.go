package cmd

import "os"

// createNoFollow mirrors the Unix helper. Windows has no O_NOFOLLOW; creating a
// symlink there requires either administrator rights or developer mode, so the
// symlink-planting scenario the flag guards against does not apply in the same
// way. secureDownloadDest still rejects destinations that resolve through a
// symlink outside the download directory.
func createNoFollow(name string) (*os.File, error) {
	return os.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o644)
}
