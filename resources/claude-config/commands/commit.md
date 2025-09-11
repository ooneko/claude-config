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

4. Create a detailed commit message that:
   - Clearly describes the purpose of the changes
   - Lists both modified and newly added relevant files
   - Explains the relationship between changes and added files
   - Example: "feat: add user profile image support\n\n- Modified user model to handle image URLs\n- Added new image upload service (new file)\n- Updated API endpoints to support images"

5. Verify the staged changes with `git diff --cached` to ensure:
   - Only relevant changes are included
   - No unrelated files are accidentally staged
   - All necessary dependencies are accounted for

6. Commit with `git commit -m "[descriptive message]"`

7. Push to remote repository with `git push origin [branch-name]`

Key considerations:
- Be judicious about which untracked files to include
- Maintain clear relationships between changes
- Document added files in commit message
- Double-check staged changes before committing