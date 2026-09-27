# Releasing bwid

How to cut a new version. Last used for v1.1.0.

## Choosing the version

bwid follows [semantic versioning](https://semver.org), and Go enforces it:

- **Patch** (1.1.x): bug fixes only, no new exported names, no behavior
  changes callers could notice.
- **Minor** (1.x.0): new exported functions or constants, or backward
  compatible behavior changes (e.g. the 1.1.0 token layout).
- **Major** (2.0.0): anything that breaks callers, such as changing a function
  signature, removing an export, or panicking on input that used to work. Go
  requires a new module path (`github.com/igsafe/bwid/v2`), so avoid this if at
  all possible.

Every exported name is a promise for the whole major version. Keep helpers
unexported unless there's a real caller for them; exporting later is easy,
removing is not.

## Steps

1. **Work on a branch**, e.g. `development-v1.2.0`, never directly on `main`.

2. **Check locally**, including on Linux (production runs on Linux, and macOS
   only has microsecond clock precision):

   ```sh
   gofmt -l .        # should print nothing
   go vet ./...
   go test ./...
   docker run --rm -v "$PWD":/src -w /src golang:1.18 go test ./...
   docker run --rm -v "$PWD":/src -w /src golang:latest go test ./...
   ```

   1.18 is the minimum in `go.mod`. If you raise it, update it in
   `.github/workflows/test.yml` too.

3. **Update `CHANGELOG.md`** with a new section at the top, marked
   `(unreleased)`, under Added / Changed / Fixed as needed. Say whether the
   release is backward compatible.

4. **Push the branch and open a pull request into `main`.** GitHub Actions
   (`.github/workflows/test.yml`) runs gofmt, vet, and the tests on Go 1.18
   and the latest stable Go, plus a short fuzz run. Wait for it to pass.

5. **Optional: test a release candidate** in the services that use bwid
   before the real release, especially if token layout or behavior changed:

   ```sh
   git tag -a v1.2.0-rc.1 -m "v1.2.0-rc.1"
   git push origin v1.2.0-rc.1
   # in a consuming service:
   go get github.com/igsafe/bwid@v1.2.0-rc.1
   ```

   `go get -u` never picks up pre-releases on its own, so no one else is
   affected.

6. **Set the release date** in `CHANGELOG.md` (replace `unreleased`) and push
   it to the PR branch.

7. **Merge the PR** on GitHub. CI runs again on `main`.

8. **Tag the merge commit on `main`** and push the tag:

   ```sh
   git switch main && git pull
   git tag -a v1.2.0 -m "v1.2.0"
   git push origin v1.2.0
   ```

9. **Create a GitHub Release** from the tag (Releases → Draft a new release →
   choose the tag). Use the version as the title and paste that version's
   `CHANGELOG.md` section as the notes.

10. **Publish to the Go proxy and pkg.go.dev**:

    ```sh
    GOPROXY=proxy.golang.org go list -m github.com/igsafe/bwid@v1.2.0
    ```

    pkg.go.dev picks up the new docs and examples within a few minutes.

11. **Upgrade the consuming services**:

    ```sh
    go get github.com/igsafe/bwid@v1.2.0
    go test ./...
    ```

## Never move or delete a pushed tag

The Go module proxy caches every tagged version permanently. If a release
turns out to be wrong, even minutes later, fix it on a branch and release the
next patch version (e.g. v1.2.1). Moving the tag would give different users
different code for the same version and break checksum verification.
