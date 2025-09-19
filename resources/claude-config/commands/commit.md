Analyze the current code changes and perform the following Git operations carefully:

1. First, thoroughly examine the repository status with `git status` to identify:
   - Modified files
   - Deleted files
   - Untracked files (categorize them by relevance)

2. For untracked files, analyze their relationship to the current changes:
   - Determine which untracked files are logically related to the modified files
   - Identify dependencies between tracked changes and new files
   - Exclude unrelated untracked files (e.g., temporary files, logs)

3. Only add the relevant untracked files using selective commands:
   - `git add path/to/specific_file` for individual files
   - `git add dir/related_files/` for directories of related files
   - Avoid using `git add .` or `git add *` to prevent over-inclusion

4. Verify unit test coverage for modified files:
   - Check that all modified Go files have corresponding test files
   - Run `go test ./...` to ensure all tests pass
   - For new functionality, ensure test coverage exists before committing
   - If tests are missing, add them or document why they're not needed

5. Create a detailed commit message that:
   - Clearly describes the purpose of the changes
   - Lists both modified and newly added relevant files
   - Explains the relationship between changes and added files
   - Mentions test coverage status for modified files
   - Example: "feat: add user profile image support\n\n- Modified user model to handle image URLs\n- Added new image upload service (new file)\n- Updated API endpoints to support images\n- Added comprehensive unit tests for all changes"

6. Verify the staged changes with `git diff --cached` to ensure:
   - Only relevant changes are included
   - No unrelated files are accidentally staged
   - All necessary dependencies are accounted for

7. Commit with `git commit -m "[descriptive message]"`

8. Push to remote repository with `git push origin [branch-name]`

## Usage Examples

### Basic commit workflow
```bash
/commit
```
Automatically analyzes current changes and generates intelligent commit message.

### Commit with additional reference
```bash
/commit "[SQCEE-DOX-8949] dox add makefile"
```
When additional reference is provided:
- System still auto-generates commit title and detailed description
- The provided reference is appended to the end of the commit content
- All validation steps (tests, linting, file analysis) are still performed

### Example commit message format:
```
feat: add Makefile support for build automation

- Added Makefile with build, test, and lint targets
- Updated documentation for build process
- Added make install target for easy deployment

Test coverage: All modified files covered
All tests passing: ✓

[SQCEE-DOX-8949] dox add makefile
```

Key considerations:
- Be judicious about which untracked files to include
- Maintain clear relationships between changes
- Document added files in commit message
- Ensure comprehensive test coverage for all code changes
- Double-check staged changes before committing
- When using `/commit [message]`, the provided message is appended to the auto-generated commit content