---
description: Create a new GitHub release with auto-generated changelog
argument-hint: "[optional: version override]"
allowed-tools:
  - Bash
  - Read
---

# Release Skill

Create a new GitHub release for hholst80/glow with an auto-generated changelog.

## Version Format

This repo uses date-based versioning: `YYYY.MM.DD`

If no version is provided as argument, use today's date.

## Steps

1. **Determine version**: Use `$ARGUMENTS` if provided, otherwise use today's date in `YYYY.MM.DD` format

2. **Get the previous release tag**:
   ```bash
   gh release list -R hholst80/glow --limit 1 --json tagName -q '.[0].tagName'
   ```

3. **Generate changelog**: Get commits since the previous release:
   ```bash
   git log <previous_tag>..HEAD --oneline --format="- [%h](https://github.com/hholst80/glow/commit/%H) %s"
   ```

4. **Create the release** using `gh release create`:
   ```bash
   gh release create <version> --title "<version>" --notes "$(cat <<'EOF'
   ## Changelog

   <commit list here>
   EOF
   )"
   ```

5. Report the release URL to the user

## Important

- Always run `go test ./...` before creating a release to ensure tests pass
- The release will automatically create a tag with the same name as the version
- Do NOT create releases with `--generate-notes` flag - we want our custom format
