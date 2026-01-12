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

3. **Generate and commit release notes**:
   - Determine the next version number based on bump type (patch/minor/major)
   - Get the current latest tag with `git describe --tags --abbrev=0` (e.g., v0.1.5)
   - Calculate what the next tag will be based on bump type
   - Get all commits since the last tag with `git log <last_tag>..HEAD --pretty=format:"%h %s"`
   - Parse commits following Conventional Commits format (type(scope): description)
   - Generate release notes in markdown format with sections:
     * Version header with date
     * "What's New" section with categorized changes (Features, Bug Fixes, Documentation, etc.)
     * List of commits grouped by type
   - Read existing RELEASE.md file
   - Insert new release notes at the top (after the main title, before previous releases)
   - Write updated RELEASE.md back to file
   - Stage RELEASE.md: `git add RELEASE.md`
   - Commit with message: `docs(release): add release notes for v<VERSION>`
   - Show the generated release notes to the user

4. **Bump version and create git tag**:
   - Default (no args or "patch"): Run `make bump` (bumps patch version)
   - If "major" argument: Run `make bump-major` (bumps major version, resets minor and patch to 0)
   - If "minor" argument: Run `make bump-minor` (bumps minor version, resets patch to 0)
   - Capture the new tag version created by the make command
   - IMPORTANT: This tag should match the version used in step 3 for the release notes

5. **Build release binary**:
   - Run `make release` to build binary with version information injected

6. **Validate release binary**:
   - Run `./note --help` and capture the output
   - Verify the version matches the tag created (without 'v' prefix)
   - Verify the version is NOT "dev" or "0.0.0" (these indicate errors)
   - Verify Build Date is present and valid (not "not set")
   - Verify Commit SHA is present and valid (not "not set")
   - If any validation fails, STOP and report the error - DO NOT push the tag
   - Show the validated release information to the user

7. **Push tag to origin**:
   - Only after successful validation, push the newly created tag to origin: `git push origin <TAG>`
   - Where TAG is the version tag created in step 4

## Arguments

- No argument or `patch`: Bump patch version (default)
- `major`: Bump major version
- `minor`: Bump minor version

## Examples

- `/release` - Commits, generates release notes, bumps patch (0.1.3 → 0.1.4), builds, and pushes tag
- `/release patch` - Same as above
- `/release minor` - Commits, generates release notes, bumps minor (0.1.3 → 0.2.0), builds, and pushes tag
- `/release major` - Commits, generates release notes, bumps major (0.1.3 → 1.0.0), builds, and pushes tag

## Important Notes

- CRITICAL: Run cleanup and tests FIRST - make clean, vet, fmt, test-all must all pass
- CRITICAL: If `make fmt` changes any files, STOP - code must be formatted before release
- CRITICAL: If any tests fail, STOP - all tests must pass before release
- CRITICAL: Release notes are automatically generated from git commits since the last tag
  * Commits should follow Conventional Commits format: `type(scope): description`
  * Release notes will categorize changes by type (feature, bug, docs, refactor, etc.)
  * The new version's release notes are inserted at the top of RELEASE.md
  * RELEASE.md is committed before creating the version tag
- Always show what version is being created before executing
- Show clear progress indicators for each step
- If any step fails, stop the process and report the error
- CRITICAL: Validate the built binary before pushing tags - check version, build date, and commit SHA
- CRITICAL: Never push a tag if version is "dev" or "0.0.0" - these indicate build errors
- After successful validation, confirm the release information before pushing
- After completion, confirm the new version and that it was pushed to origin
- CRITICAL: You MUST actually execute all the commands, not just describe them
