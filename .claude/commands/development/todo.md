# TODO Command

Manage project TODO list in `docs/TODO.md`

## Usage

- `/todo ADD <task description>` - Add a new task to the TODO list
- `/todo DONE <task description or keyword>` - Mark a task as completed
- `/todo LIST` - Show all uncompleted tasks

## Examples

```
/todo ADD Refactor UI codebase
/todo ADD Implement user authentication
/todo DONE UI Refactor done
/todo LIST
```

## Implementation

When this command is used:

### File Location Strategy
1. First check if `docs/TODO.md` exists (current project structure)
2. If not found, check if `./TODO.md` exists (CLAUDE.md specification)
3. If neither exists, create `docs/TODO.md` with proper structure

### ADD Command
1. Locate the TODO.md file using the file location strategy
2. Read the current TODO.md file
3. Add the new task under "# Backlog" section with `[ ]` checkbox (matching current format)
4. Use the format: `- [ ] <task description>`
5. Save the file
6. Confirm the task was added

### DONE Command
1. Locate the TODO.md file using the file location strategy
2. Read the current TODO.md file
3. Find the matching task in the "# Backlog" section (partial match on description)
4. Change the checkbox from `[ ]` to `[x]`
5. Add completion timestamp: `- [x] <task description> (Completed: YYYY-MM-DD)`
6. Keep completed tasks in the same location (no moving to separate section)
7. Save the file
8. Confirm the task was marked as done

### LIST Command
1. Locate the TODO.md file using the file location strategy
2. Read the current TODO.md file
3. Extract all tasks under "# Backlog" that have `[ ]` (uncompleted)
4. Display them in a clean numbered list
5. Show count of active tasks

## File Format

The TODO.md file uses this structure (matching current project format):

```markdown
# Backlog

- [ ] Task description 1
- [ ] Task description 2
- [ ] Another task
- [x] Completed task (Completed: YYYY-MM-DD)
```

## Error Handling

- Check multiple file locations before creating new file
- If task not found for DONE command, show available tasks and ask for clarification
- If no TODO.md file exists in either location, create `docs/TODO.md` with proper structure
- Handle malformed TODO.md files gracefully
- Validate command format and provide usage help if incorrect

## File Location Priority

1. `docs/TODO.md` (current project location)
2. `./TODO.md` (CLAUDE.md specification)
3. Create `docs/TODO.md` if neither exists

## Notes

- Only tracks new feature additions and major development tasks
- Bug fixes and formatting errors are not tracked here
- Completed tasks remain in same section with timestamp
- Partial matching is used for DONE command to be more user-friendly
- Adapts to existing project TODO.md structure (uses "# Backlog" heading)
