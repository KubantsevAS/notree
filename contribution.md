# Contribution

Rules for commit messages and branch names to maintain consistency and readability in our project.

## Branch Naming Conventions

We use a prefix based on the type of work, the task ID from our task tracker, and a short description.
Format: `<type>/task-XXX-<short-description>`

Where:

* **`<type>`** One of the following types:
  * **`feature`**  For new features.
  * **`refactor`** For code rewrites/restructuring.
  * **`build`**    For changes in build tools, CI pipeline, dependencies.
  * **`test`**     For testing a hypothesis.
  * **`docs`**     For documentation-only changes.
  * **`fix`**      For bug fixes.
  * **`ops`**      For operational changes (infrastructure, deployment, etc.).

**Example Branch Names:**

* **`feature/task-001-add-user-auth`**
* **`refactor/task-015-update-react-19`**
* **`build/task-021-update-deploy-scripts`**
* **`test/task-008-will-it-work-or-wont`**
* **`docs/task-005-update-contribution`**
* **`fix/task-042-resolve-login-bug`**
* **`ops/task-033-backup-utils`**

**Invalid Branch Names:**

Avoid vague or ambiguous branch names.  Examples of invalid branch names:

* `my-branch`
* `work-in-progress`
* `bugfix`

## Commit Message Conventions

Commit messages should be clear, concise, and informative.  They should follow the following format:

`<type>(Optional: <scope>): <short description>`

Where:

* **`<type>`** One of the following types:
  * **`feat`**     A new feature for the user.
  * **`fix`**      A bug fix for the user.
  * **`hotfix`**   A 🔥hot🔥 fix.
  * **`docs`**     Changes to the documentation.
  * **`style`**    Changes that don't affect the functionality (e.g., formatting, whitespace).
  * **`refactor`** A code change that neither fixes a bug nor adds a feature.
  * **`test`**     Adding missing tests or refactoring existing tests.
  * **`build`**    Add command in build script.

* **`<scope>`** (Optional) A short description of the area of the codebase affected by the change.  Example: `auth`, `database`, `ui`.

* **`<short description>`:** A concise summary of the change (50 characters or less).

**Example Commit Messages:**

* `feat(auth): Add user authentication`
* `fix(ui): Resolve button alignment issue`
* `docs(readme): Update installation instructions`
* `style: Fix formatting inconsistencies and eslint fixes`
* `refactor(database): Improve database query performance`
* `test(api): Add unit tests for API endpoints`

## References

[Conventional Commit Messages^](https://gist.github.com/qoomon/5dfcdf8eec66a051ecd85625518cfd13)
