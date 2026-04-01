# hevy-cli follow-up fixes

## Goal
Fix the 3 rough edges found during manual testing after Phase 3 shipped.

## Issues to fix

### 1. `hevy routine` help/usage is misleading
Current behavior:
- `hevy routine --json` fails with `accepts 1 arg(s), received 0`
- help text currently shows `Usage: hevy routine [flags]`

Expected:
- help/usage must clearly show that an ID is required, e.g. `hevy routine <routine-id>`
- error/help output should be consistent with that requirement

### 2. `hevy exercise` help/usage is misleading
Current behavior:
- `hevy exercise --json` fails with `accepts 1 arg(s), received 0`
- help text currently shows `Usage: hevy exercise [flags]`

Expected:
- help/usage must clearly show that an ID is required, e.g. `hevy exercise <exercise-id>`
- error/help output should be consistent with that requirement

### 3. `hevy history` ID behavior is confusing/broken
Current behavior:
- `hevy exercises --page-size 1 --json` returns template IDs like `3BC06AD3`
- `hevy history 3BC06AD3` returns an empty table even though command shape suggests this should work

Need to determine the correct fix:
- either `history` is using the wrong endpoint/ID type and should accept exercise template IDs from `hevy exercises`
- or the command/help/output must clearly state it expects a different ID source

Preferred outcome:
- `hevy history <id>` should work with the IDs surfaced by `hevy exercises`
- if that is impossible because of Hevy API constraints, make the UX explicit and helpful:
  - update help text
  - improve error/empty-state messaging
  - ideally suggest where to get the correct ID from

## Deliverables
1. Fix `routine` command usage/help
2. Fix `exercise` command usage/help
3. Fix `history` behavior and/or UX so it is not misleading
4. Add/adjust tests for all 3 fixes
5. Ensure help text examples stay accurate

## Verification
Run and report output for:
```bash
go test ./...
go vet ./...
hevy routine --help
hevy exercise --help
hevy history --help
export GO_HEVY_API_KEY=your-hevy-api-key
hevy exercises --page-size 1 --json
hevy history <id-from-hevy-exercises>
```

## Notes
- Keep this focused on the 3 issues above, no unrelated refactors
- Preserve existing working behavior for the rest of the CLI
- If `history` cannot support template IDs, make the output bluntly clear instead of silently printing an empty table

## Commit
Use a conventional commit message, likely:
```bash
fix(cli): clarify id-based commands and improve history UX
```