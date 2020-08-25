package osfs

import (
	"os"
)

// Options are the file system options.
type Options struct {
	// RootDirectory is the directory at which the store would base upon.
	RootDirectory string
	// DefaultDirectory (optional) is the default name for the directory (if not provided).
	DefaultDirectory string
	// DefaultBucket (optional) is the default bucket name for the store (if not provided).
	DefaultBucket string
	// DirectoryPermissions are the permissions for the directories created by this store.
	DirectoryPermissions os.FileMode
	// // FileVersions is the flag that allows files to have their own versions.
	FileVersions bool
}

// Option is a function that changes options.
type Option func(o *Options)

// WithRootDirectory sets the RootDirectory option.
func WithRootDirectory(option string) Option {
	return func(o *Options) {
		o.RootDirectory = option
	}
}

// WithDefaultDirectory sets the DefaultDirectory option.
func WithDefaultDirectory(option string) Option {
	return func(o *Options) {
		o.DefaultDirectory = option
	}
}

// WithDefaultBucket sets the DefaultBucket option.
func WithDefaultBucket(option string) Option {
	return func(o *Options) {
		o.DefaultBucket = option
	}
}

// WithDirectoryPermissions sets the DirectoryPermissions option.
func WithDirectoryPermissions(option os.FileMode) Option {
	return func(o *Options) {
		o.DirectoryPermissions = option
	}
}

// WithFileVersions allow to create file with versions.
func WithFileVersions(allowVersions bool) Option {
	return func(o *Options) {
		o.FileVersions = allowVersions
	}
}
