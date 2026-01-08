Generate a comprehesive and informative Git commit message based on the staged changes.
Follow the Conventional Commits specification: `<type>(<scope>): <description>`.
 
- **Type:** Must be one of `feature`, `bug`, `docs`, `refactor`, `test`, `build`, `cicd`, `revert`.
- **Scope (optional):** A noun describing the section of the codebase affected.
- **Description:** A brief, imperative summary of the change.
- **Body:** A detailed bullet list explanation of the changes, including motivation and context.
- **Footer (optional):** Reference issues (e.g., `Closes #123`).
 
Analyze the output of `git diff --staged` to understand the changes.

NEVER ever mention a co-authored-by or similar aspects. In particular, never mention the tool used to create the commit message or PR.
