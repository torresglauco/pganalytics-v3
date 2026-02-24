# Pull Request Creation Guide

A comprehensive guide for creating pull requests for the pgAnalytics-v3 project.

---

## Quick Start (30 seconds)

### Fastest Method: Direct GitHub Link

Click your branch link directly to create a PR:

```
https://github.com/torresglauco/pganalytics-v3/compare/master...YOUR_BRANCH
```

This will:
1. Pre-fill the base and compare branches
2. Show you all commits to be included
3. Allow you to create the PR with a single click

---

## Step-by-Step Instructions

### Method 1: Direct GitHub Link (Fastest - 30 seconds)

1. **Open the PR comparison page** using the link format above
2. **Click "Create pull request"** button
3. **Fill in the title** (see examples below)
4. **Fill in the description** (copy from PR_TEMPLATE.md or examples below)
5. **Click "Create pull request"** to submit

### Method 2: GitHub Web Interface (2 minutes)

1. Go to: https://github.com/torresglauco/pganalytics-v3
2. Click the **"Pull requests"** tab (top navigation)
3. Click the green **"New pull request"** button
4. Set the comparison:
   - **Base**: `master` (or appropriate base branch)
   - **Compare**: Your feature branch (e.g., `feature/phase3-collector-modernization`)
5. Click **"Create pull request"**
6. Fill in title and description from templates below
7. Click **"Create pull request"** to submit

### Method 3: GitHub CLI (If Authenticated)

If you have a GitHub personal access token:

```bash
# First time only (interactive):
gh auth login

# Then create the PR:
gh pr create \
  --title "Your PR Title Here" \
  --body-file PR_TEMPLATE.md \
  --base master
```

---

## PR Title Guidelines

Keep titles concise (50-70 characters) and descriptive:

**Good Examples:**
- `Phase 3.4: Complete End-to-End Testing Suite for pgAnalytics v3`
- `Phase 3.5: C/C++ Collector Modernization - Foundation Implementation`
- `Add query performance optimization features`

**Title Format:**
```
[Phase X.Y]: Brief description of what was implemented
```

---

## PR Description Template

The description should include:

```markdown
# Overview
Brief summary of what this PR accomplishes

## What's Included
- Feature 1
- Feature 2
- Feature 3

## Test Results
Include any relevant test metrics:
- Unit Tests: X/X passing
- Integration Tests: X/X passing
- Build: ✅ Successful

## Files Changed
| Category | Files | Status |
|----------|-------|--------|
| Source | X | Modified/New |
| Tests | X | New |
| Docs | X | New |

## Known Limitations
- Any limitations or future work
- Environmental dependencies

## How to Test
Instructions for testing the changes

## Next Steps
What comes after this PR is merged
```

---

## Pre-PR Checklist

Before creating your PR, verify:

- ✅ All code is committed: `git status` shows "working tree clean"
- ✅ All commits are pushed: `git push origin YOUR_BRANCH`
- ✅ Tests pass: `make test` or appropriate test command
- ✅ Build succeeds: `make build` or appropriate build command
- ✅ No hardcoded credentials or secrets
- ✅ Documentation is complete and accurate
- ✅ Branch is up-to-date with base branch (if needed)
- ✅ PR title is descriptive and concise
- ✅ PR description includes all relevant information

---

## Common PR Creation Issues

### Issue: "Compare across forks" error

**Solution:** Make sure you're in the correct repository and that your branch has been pushed.

### Issue: No commits shown in comparison

**Solution:** Verify your branch has commits: `git log master..YOUR_BRANCH`

### Issue: Can't authenticate with GitHub CLI

**Solution:** Use the direct GitHub link method instead, or run `gh auth login` first.

### Issue: Merge conflicts with base branch

**Solution:**
1. Update your local base: `git fetch origin`
2. Rebase or merge: `git rebase origin/master` or `git merge origin/master`
3. Resolve conflicts in your editor
4. Push: `git push origin YOUR_BRANCH`

---

## After Creating the PR

### Immediate Next Steps

1. **Verify PR created successfully**
   - Check that all commits are included
   - Confirm title and description are correct
   - Review the changed files list

2. **Monitor CI/CD checks** (if configured)
   - GitHub will run automated checks
   - Look for green ✅ or red ❌ indicators
   - Address any check failures

3. **Wait for code review**
   - Team members will review your changes
   - They may comment on specific lines
   - Respond to feedback and questions

### If Changes Are Requested

1. Make the requested changes locally
2. Commit with a clear message
3. Push to the same branch
4. PR will automatically update with new commits
5. Respond to review comments

### Merging the PR

When all reviews are approved and checks pass:

1. Click **"Merge pull request"** button
2. Choose merge strategy (usually "Create a merge commit")
3. Click **"Confirm merge"**
4. Optionally delete the feature branch

---

## Repository-Specific Guidelines

### Branch Naming Convention

Use descriptive feature branch names:

```
feature/phase-X-Y-description    # For feature implementation
fix/issue-description             # For bug fixes
docs/documentation-topic          # For documentation updates
refactor/component-name           # For refactoring work
```

Examples:
- `feature/phase3-collector-modernization`
- `feature/phase4-ml-optimization`
- `fix/query-performance-bug`
- `docs/deployment-guide`

### Commit Message Guidelines

Write clear, descriptive commit messages:

```
Phase X.Y: Brief summary of what changed

Optional detailed explanation of:
- What was changed and why
- How it works
- Testing performed
- Potential impacts

Co-Authored-By: Name <email> (if applicable)
```

### Documentation Requirements

Ensure your PR includes:

1. **Code comments** for complex logic
2. **README updates** if changing setup/deployment
3. **API documentation** if adding endpoints
4. **Configuration examples** for new options
5. **Migration guides** if breaking changes

---

## Advanced: Pull Request Templates

If the repository has pull request templates, they will auto-populate when you create a PR. The template is typically located in `.github/pull_request_template.md`.

## FAQ

**Q: What if I forgot to add something before creating the PR?**
A: You can edit the PR title and description after creation by clicking the three dots (⋯) menu.

**Q: Can I add more commits to the PR after creating it?**
A: Yes! Push additional commits to the same branch and the PR will automatically update.

**Q: What's the difference between my branch and master?**
A: Check with: `git log master..YOUR_BRANCH` (shows commits in your branch)

**Q: How often should I update my branch from master?**
A: Keep it updated regularly to avoid conflicts: `git rebase origin/master`

**Q: Can I create PRs to branches other than master?**
A: Yes, use any base branch. Common patterns: `develop`, `release-X.Y`, or version branches.

---

## Resources

- [GitHub PR Documentation](https://docs.github.com/en/pull-requests)
- [Git Documentation](https://git-scm.com/doc)
- Repository: https://github.com/torresglauco/pganalytics-v3

---

**Last Updated:** February 24, 2026
