# Claude Command: iCafe

This command helps you create iCafe cards based on current Git code changes, automatically analyzing modifications and generating titles, content, and types.

## Usage

Create an iCafe card:
```
/icafe
```

With options:
```
/icafe --type=bug
/icafe --type=story
/icafe --title="Custom title"
```

**Note**: This command only supports two card types: `bug` and `story`.

## What This Command Does

1. **Analyze Git Status and Changes**:
   - Execute `git status` to get current repository state
   - **Smart File Selection Logic**:
     - **If files are already staged**: Only analyzes the manually staged files
     - **If no files are staged**: Analyzes all modified, deleted, and untracked files
   - Get both staged and unstaged changes based on selection

2. **Smart Content Generation**:
   - Execute `git diff --cached` for staged files or `git diff` for unstaged files
   - Intelligently determine card type (bug or story) based on changes
   - Auto-generate clear titles and description content
   - Extract key information like affected modules and features

3. **Card Type Detection Logic**:
   - **Bug Type**: Detects error fixes, patches, rollbacks with keywords
   - **Story Type**: Detects new features, enhancements, refactoring with keywords
   - **Smart Analysis**: Infers type based on code diff patterns

4. **iCafe MCP Integration**:
   - Call iCafe MCP service to create cards
   - Pass generated title, content, and type
   - Return creation result and card link

## Card Content Generation Rules

### Title Generation
- Generate concise titles based on main changed modules and features
- Format: `[Module] Feature Description`
- Examples:
  - `[User Auth] Fix login failure issue`
  - `[API] Add user management endpoints`
  - `[Frontend] Optimize responsive layout`

### Content Generation
Contains structured information:
- **Change Overview**: Brief description of the change purpose
- **Affected Files**: List main modified files
- **Key Changes**: Detailed description of core changes
- **Technical Details**: Important implementation points

### Type Detection Rules

**Bug Type Triggers**:
- Code changes include error fix patterns
- Commit messages contain "fix", "bug", "patch", "hotfix" keywords
- Deleted or modified exception handling logic
- Fixed test cases

**Story Type Triggers**:
- New files or substantial new code
- Commit messages contain "feat", "add", "enhance", "improve" keywords
- New API endpoints or feature modules
- New test cases added

## Use Cases

1. **Record After Development**:
   ```bash
   # After completing feature development
   git add .
   /icafe
   ```

2. **Bug Fix Recording**:
   ```bash
   # After fixing issues
   git add fixed_files.go
   /icafe --type=bug
   ```

3. **Before Code Review**:
   ```bash
   # Preparing for code review
   /icafe --title="[Code Review] User permission module refactor"
   ```

## Command Options

- `--type=<bug|story>`: Manually specify card type, skip auto-detection
- `--title="<title>"`: Custom card title, skip auto-generation
- `--content="<content>"`: Custom card content, skip auto-generation
- `--no-analysis`: Skip Git change analysis, use manual input

## Smart Analysis Examples

### Bug Type Example
```
Title: [Auth Module] Fix JWT token expiration handling issue
Content:
Change Overview: Fix issue where user JWT tokens cannot refresh properly after expiration

Affected Files:
- src/auth/jwt.go: Fix token refresh logic
- src/middleware/auth.go: Improve expiration detection

Key Changes:
- Fix timestamp calculation error in token refresh
- Add automatic refresh mechanism before expiration
- Improve error handling and logging

Technical Details: Use time.Now().Unix() to replace incorrect time calculation
```

### Story Type Example
```
Title: [User Module] Add user preference settings feature
Content:
Change Overview: Implement user personal preference settings supporting themes, languages, etc.

Affected Files:
- src/models/user_preference.go: New user preference data model
- src/api/preference.go: Implement preference settings API
- src/frontend/settings.tsx: Add settings page component

Key Changes:
- Design and implement user preference data structure
- Create CRUD API interfaces
- Implement frontend settings interface
- Add database migration files

Technical Details: Use JSON fields for flexible configuration with dynamic extension support
```

## MCP Integration

This command depends on iCafe MCP service for card creation:

1. **Connection Verification**: Check if iCafe MCP service is available
2. **Data Transfer**: Send generated content to iCafe
3. **Result Handling**: Return creation result and card access link
4. **Error Handling**: Handle network errors and service exceptions

## Best Practices

1. **Timely Recording**: Create cards immediately after development or fixes
2. **Clear Commits**: Ensure Git changes reflect actual work content
3. **Reasonable Splitting**: Consider multiple commits and records for large changes
4. **Regular Organization**: Use cards for work summary and review

## Important Notes

- Command must be executed in Git repository
- Recommend using when there are clear code changes, avoid empty changes
- Auto-generated content can be overridden with options
- Supports Chinese and English code change analysis
- Card creation failure won't affect local Git status

## Examples

Simple usage:
```bash
# Auto-analyze and create card
/icafe
```

Specify type:
```bash
# Explicitly specify bug type
/icafe --type=bug
```

Custom title:
```bash
# Use custom title
/icafe --title="[Performance] Database query optimization"
```

Fully customized:
```bash
# Fully manual content specification
/icafe --type=story --title="New feature development" --content="Detailed feature description..."
```