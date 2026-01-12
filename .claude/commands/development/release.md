Automate the complete software release process including committing changes, bumping version, building release binary, and pushing tags to origin.

## Process Flow

1. **Pre-release cleanup and validation**:
   - Run `make clean` to remove build artifacts
   - Run `make vet` to check for Go code issues
   - Run `make fmt` to format Go code
   - Check if fmt made any changes with `git diff --exit-code`
   - If fmt made changes, STOP and report that code needs formatting
   - Run `make test-all` to execute all test suites
   - If ANY tests fail, STOP and report the failures
   - If vet reports issues, STOP and report the issues
   - Show success message only after all checks pass

2. **Commit staged changes** (if any exist):
   - Check if there are staged changes with `git diff --staged --quiet`
   - If staged changes exist, run the commit workflow (same as /commit command)
   - Generate comprehensive commit message following Conventional Commits
   - Execute `git commit` with the generated message

3. **Bump version and create git tag**:
   - Default (no args or "patch"): Run `make bump` (bumps patch version)
   - If "major" argument: Run `make bump-major` (bumps major version, resets minor and patch to 0)
   - If "minor" argument: Run `make bump-minor` (bumps minor version, resets patch to 0)
   - Capture the new tag version created by the make command

4. **Build release binary**:
   - Run `make release` to build binary with version information injected

5. **Validate release binary**:
   - Run `./note --help` and capture the output
   - Verify the version matches the tag created (without 'v' prefix)
   - Verify the version is NOT "dev" or "0.0.0" (these indicate errors)
   - Verify Build Date is present and valid (not "not set")
   - Verify Commit SHA is present and valid (not "not set")
   - If any validation fails, STOP and report the error - DO NOT push the tag
   - Show the validated release information to the user

6. **Push tag to origin**:
   - Only after successful validation, push the newly created tag to origin: `git push origin <TAG>`
   - Where TAG is the version tag created in step 3

## Arguments

- No argument or `patch`: Bump patch version (default)
- `major`: Bump major version
- `minor`: Bump minor version

## Examples

- `/release` - Commits, bumps patch (0.1.3 → 0.1.4), builds, and pushes tag
- `/release patch` - Same as above
- `/release minor` - Commits, bumps minor (0.1.3 → 0.2.0), builds, and pushes tag
- `/release major` - Commits, bumps major (0.1.3 → 1.0.0), builds, and pushes tag

## Important Notes

- CRITICAL: Run cleanup and tests FIRST - make clean, vet, fmt, test-all must all pass
- CRITICAL: If `make fmt` changes any files, STOP - code must be formatted before release
- CRITICAL: If any tests fail, STOP - all tests must pass before release
- Always show what version is being created before executing
- Show clear progress indicators for each step
- If any step fails, stop the process and report the error
- CRITICAL: Validate the built binary before pushing tags - check version, build date, and commit SHA
- CRITICAL: Never push a tag if version is "dev" or "0.0.0" - these indicate build errors
- After successful validation, confirm the release information before pushing
- After completion, confirm the new version and that it was pushed to origin
- CRITICAL: You MUST actually execute all the commands, not just describe them
