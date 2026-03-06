package plugins

import (
	zipa "archive/zip"
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// makeZipEntry builds an in-memory zip with a single file entry using the
// exact name provided (no normalization), then returns the *zipa.File so it
// can be passed directly to unzipFile.
func makeZipEntry(t *testing.T, name, content string) *zipa.File {
	t.Helper()

	var buf bytes.Buffer
	w := zipa.NewWriter(&buf)

	fh := &zipa.FileHeader{Name: name, Method: zipa.Deflate}
	fh.SetMode(0o644)

	fw, err := w.CreateHeader(fh)
	require.NoError(t, err)
	_, err = fw.Write([]byte(content))
	require.NoError(t, err)
	require.NoError(t, w.Close())

	r, err := zipa.NewReader(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	require.NoError(t, err)
	require.Len(t, r.File, 1)
	return r.File[0]
}

// makeZipDirEntry builds a directory entry (name ends with /).
func makeZipDirEntry(t *testing.T, name string) *zipa.File {
	t.Helper()

	var buf bytes.Buffer
	w := zipa.NewWriter(&buf)

	fh := &zipa.FileHeader{Name: name, Method: zipa.Store}
	fh.SetMode(0o755 | os.ModeDir)

	_, err := w.CreateHeader(fh)
	require.NoError(t, err)
	require.NoError(t, w.Close())

	r, err := zipa.NewReader(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	require.NoError(t, err)
	require.Len(t, r.File, 1)
	return r.File[0]
}

func Test_unzipFile(t *testing.T) {
	t.Run("valid file is extracted into dest", func(t *testing.T) {
		dest := t.TempDir()
		f := makeZipEntry(t, "prefix/subdir/file.txt", "hello")

		err := unzipFile(f, dest)
		require.NoError(t, err)

		got, err := os.ReadFile(filepath.Join(dest, "subdir", "file.txt"))
		require.NoError(t, err)
		assert.Equal(t, "hello", string(got))
	})

	t.Run("valid directory entry is created", func(t *testing.T) {
		dest := t.TempDir()
		f := makeZipDirEntry(t, "prefix/mydir/")

		err := unzipFile(f, dest)
		require.NoError(t, err)

		info, err := os.Stat(filepath.Join(dest, "mydir"))
		require.NoError(t, err)
		assert.True(t, info.IsDir())
	})

	t.Run("entry with no slash returns error (no root directory)", func(t *testing.T) {
		dest := t.TempDir()
		f := makeZipEntry(t, "justfilename.txt", "data")

		err := unzipFile(f, dest)
		assert.ErrorContains(t, err, "no root directory")
	})

	t.Run("dot-dot traversal is rejected", func(t *testing.T) {
		dest := t.TempDir()
		f := makeZipEntry(t, "prefix/../../../etc/passwd", "evil")

		err := unzipFile(f, dest)
		assert.Error(t, err)

		_, statErr := os.Stat(filepath.Join(dest, "..", "..", "etc", "passwd"))
		assert.True(t, os.IsNotExist(statErr), "file must not have been written")
	})

	t.Run("dot-dot in nested segment is rejected", func(t *testing.T) {
		dest := t.TempDir()
		f := makeZipEntry(t, "prefix/a/b/../../../../../../etc/shadow", "evil")

		err := unzipFile(f, dest)
		assert.Error(t, err)
	})

	// Advisory claim: prefix//etc/passwd → cleanName="/etc/passwd"
	// filepath.Join(dest, "/etc/passwd") returns "/etc/passwd" in Go.
	// The advisory says this bypasses the check and writes to /etc/passwd.
	// This test demonstrates whether the strings.HasPrefix guard on line 279
	// catches it before any file is written.
	t.Run("absolute path in second segment is caught by HasPrefix guard (advisory scenario)", func(t *testing.T) {
		dest := t.TempDir()

		// Entry name: "prefix//etc/passwd"  →  pathParts[1] = "/etc/passwd"
		f := makeZipEntry(t, "prefix//etc/passwd", "injected")

		err := unzipFile(f, dest)
		assert.Error(t, err, "absolute path must be rejected")

		// Confirm nothing was written to /etc/passwd.
		content, readErr := os.ReadFile("/etc/passwd")
		if readErr == nil {
			assert.NotContains(t, string(content), "injected",
				"/etc/passwd must not have been overwritten")
		}
	})

	// The real gap: strings.HasPrefix does not enforce a separator boundary.
	// If the crafted absolute path starts with absDest (no separator check),
	// the guard passes and the file is written to a sibling directory.
	// This test constructs the path dynamically so it matches the prefix.
	t.Run("sibling directory shares prefix — HasPrefix guard passes (real gap)", func(t *testing.T) {
		// Use a parent temp dir so we control the directory name.
		parent := t.TempDir()
		dest := filepath.Join(parent, "plugin")
		require.NoError(t, os.MkdirAll(dest, 0o755))

		absDest, err := filepath.Abs(dest)
		require.NoError(t, err)

		// Craft an absolute path that starts with absDest but is a sibling:
		//   /tmp/.../plugin_sibling/evil.txt
		siblingPath := absDest + "_sibling"
		evilFile := filepath.Join(siblingPath, "evil.txt")

		// Entry name: "prefix" + "/" + evilFile  (evilFile is absolute)
		entryName := "prefix/" + evilFile
		f := makeZipEntry(t, entryName, "sibling-escape")

		err = unzipFile(f, dest)

		if err == nil {
			// The guard passed — file was written outside dest.
			_, statErr := os.Stat(evilFile)
			if statErr == nil {
				content, _ := os.ReadFile(evilFile)
				t.Logf("WRITTEN outside dest → %s: %q", evilFile, string(content))
				// Clean up.
				_ = os.RemoveAll(siblingPath)
			}
			t.Errorf("expected error: file escaped dest via sibling prefix bypass")
		} else {
			// Guard caught it — behaviour is safe here.
			t.Logf("correctly rejected with: %v", err)
			assert.True(t,
				strings.Contains(err.Error(), "escapes") ||
					strings.Contains(err.Error(), "invalid"),
				"unexpected error message: %v", err)
		}
	})

	t.Run("path that equals dest itself is not extracted as file", func(t *testing.T) {
		dest := t.TempDir()
		absDest, err := filepath.Abs(dest)
		require.NoError(t, err)

		// Entry name resolves exactly to dest (no trailing content).
		entryName := "prefix/" + absDest
		f := makeZipEntry(t, entryName, "data")

		err = unzipFile(f, dest)
		// Behaviour: either an error or a no-op — but must not overwrite dest itself.
		info, statErr := os.Stat(dest)
		require.NoError(t, statErr)
		assert.True(t, info.IsDir(), "dest directory must still be a directory")
	})

	t.Run("entry with only prefix slash produces dot cleanName", func(t *testing.T) {
		dest := t.TempDir()
		// "prefix/" → pathParts[1] = "" → filepath.Clean("") = "."
		f := makeZipEntry(t, "prefix/", "data")

		// Should not panic; behaviour (error or no-op) is acceptable.
		_ = unzipFile(f, dest)
	})

	t.Run("deeply nested valid path is extracted correctly", func(t *testing.T) {
		dest := t.TempDir()
		f := makeZipEntry(t, "prefix/a/b/c/d/e/deep.txt", "deep content")

		err := unzipFile(f, dest)
		require.NoError(t, err)

		got, err := os.ReadFile(filepath.Join(dest, "a", "b", "c", "d", "e", "deep.txt"))
		require.NoError(t, err)
		assert.Equal(t, "deep content", string(got))
	})

	t.Run("file content is written exactly", func(t *testing.T) {
		dest := t.TempDir()
		content := strings.Repeat("x", 4096)
		f := makeZipEntry(t, "prefix/big.bin", content)

		err := unzipFile(f, dest)
		require.NoError(t, err)

		got, err := os.ReadFile(filepath.Join(dest, "big.bin"))
		require.NoError(t, err)
		assert.Equal(t, content, string(got))
	})
}
