package osfs

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVersioned(t *testing.T) {
	o, err := New(
		WithFileVersions(true),
		WithDefaultBucket("test_bucket"),
		WithDefaultDirectory("test_directory"),
	)
	require.NoError(t, err)

	defer func() {
		os.RemoveAll("./test_bucket")
	}()

	ctx := context.Background()
	f, err := o.NewFile(ctx, "test_file.txt")
	require.NoError(t, err)

	_, err = f.Write([]byte("test"))
	require.NoError(t, err)

	err = o.PutFile(ctx, f)
	require.NoError(t, err)
	firstV := f.Version()
	assert.NotZero(t, firstV)

	_, err = f.Write([]byte("second write"))
	require.NoError(t, err)

	err = o.PutFile(ctx, f)
	require.NoError(t, err)

	secondV := f.Version()
	assert.NotZero(t, secondV)
	assert.NotEqual(t, firstV, secondV)

	err = o.DeleteFile(ctx, f)
	require.NoError(t, err)

	rs, err := o.GetFile(ctx, f.Name())
	require.NoError(t, err)

	assert.Equal(t, rs.Version(), firstV)

	fileList, err := o.ListFiles(ctx, "")
	require.NoError(t, err)
	assert.Len(t, fileList, 1)

	nf, err := o.NewFile(ctx, f.Name())
	require.NoError(t, err)

	_, err = nf.Write([]byte("third write"))
	require.NoError(t, err)

	err = o.PutFile(ctx, nf)
	require.NoError(t, err)

	// Delete non latest version.
	err = o.DeleteFile(ctx, rs)
	require.NoError(t, err)

	rs, err = o.GetFile(ctx, f.Name())
	require.NoError(t, err)

	assert.Equal(t, rs.Version(), nf.Version())
}
