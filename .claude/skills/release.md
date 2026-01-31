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

2. **Run tests**:
   ```bash
   go test ./...
   ```

3. **Check if release already exists** (for recreate scenario):
   ```bash
   gh release view <version> -R hholst80/glow 2>/dev/null && echo "EXISTS" || echo "NEW"
   ```

4. **If recreating, delete existing release and tag**:
   ```bash
   gh release delete <version> -R hholst80/glow --yes
   git tag -d <version> 2>/dev/null
   git push origin :refs/tags/<version>
   ```

5. **Get the previous release tag**:
   ```bash
   gh release list -R hholst80/glow --limit 1 --json tagName -q '.[0].tagName'
   ```

6. **Generate changelog** - commits since previous release:
   ```bash
   git log <previous_tag>..HEAD --oneline --format="- [%h](https://github.com/hholst80/glow/commit/%H) %s"
   ```

7. **Create the release**:
   ```bash
   gh release create <version> -R hholst80/glow --title "<version>" --notes "$(cat <<'EOF'
   ## Changelog

   <commit list here>
   EOF
   )"
   ```

8. Report the release URL to the user

## Important

- Always use `-R hholst80/glow` flag with all `gh release` commands
- Always run `go test ./...` before creating a release
- The release will automatically create a git tag with the same name
- Do NOT use `--generate-notes` flag - we want our custom format
- When recreating a release, delete both the release AND the tag before creating new one
