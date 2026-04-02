# GitHub Release Guide for v3.1.0

**Quick Links:**
- [Create Release via Web UI (Recommended)](#option-1-create-release-via-github-web-ui)
- [Create Release via CLI](#option-2-create-release-via-github-cli)
- [Release Checklist](#release-checklist)

---

## Option 1: Create Release via GitHub Web UI (Recommended)

### Step-by-Step Instructions

1. **Navigate to GitHub Releases**
   - Go to: https://github.com/torresglauco/pganalytics-v3/releases
   - Click "Draft a new release" button

2. **Select or Create Tag**
   - Choose "v3.1.0" from the tag dropdown (already exists in git)
   - If not showing, click "Create a new tag"
   - Tag: `v3.1.0`
   - Target: `main` branch

3. **Fill in Release Title**
   ```
   pgAnalytics v3.1.0 - Wave 3 Complete: MCP Integration & PostgreSQL 14-18 Support
   ```

4. **Add Release Description**

   Copy the content from [`RELEASE_NOTES_v3.1.0.md`](RELEASE_NOTES_v3.1.0.md):

   ```markdown
   ## 🎉 Overview

   pgAnalytics v3.1.0 represents the complete implementation of **all three development waves**, delivering a comprehensive PostgreSQL monitoring and analysis platform...

   [Full release notes content]
   ```

   Or click "Generate release notes" and GitHub will auto-populate based on commits.

5. **Configure Release Options**
   - [ ] Check "This is a pre-release" (if applicable)
   - [ ] Check "Create a discussion for this release" (optional)
   - [ ] Add assets (binaries, Docker images, etc.) if available

6. **Publish Release**
   - Click "Publish release" button
   - Release is now live!

---

## Option 2: Create Release via GitHub CLI

### Prerequisites

```bash
# Ensure GitHub CLI is installed
gh --version
# Expected: gh version 2.x.x

# Authenticate with GitHub
gh auth login
# Follow prompts to authenticate (choose HTTPS or SSH)
# Use a personal access token if prompted
```

### Create Release Command

```bash
# Create release with release notes from file
gh release create v3.1.0 \
  --title "pgAnalytics v3.1.0 - Wave 3 Complete: MCP Integration & PostgreSQL 14-18 Support" \
  --notes-file RELEASE_NOTES_v3.1.0.md

# Or with inline notes
gh release create v3.1.0 \
  --title "pgAnalytics v3.1.0 - Wave 3 Complete" \
  --notes "See RELEASE_NOTES_v3.1.0.md for full details"
```

### Verify Release Created

```bash
# List releases
gh release list

# View specific release
gh release view v3.1.0
```

---

## Option 3: Create Release via GitHub API

### Using curl

```bash
# Set variables
OWNER="torresglauco"
REPO="pganalytics-v3"
TAG="v3.1.0"
TOKEN="ghp_your_personal_access_token_here"  # Get from GitHub Settings

# Create release
curl -X POST \
  -H "Authorization: token $TOKEN" \
  -H "Accept: application/vnd.github.v3+json" \
  https://api.github.com/repos/$OWNER/$REPO/releases \
  -d '{
    "tag_name": "'$TAG'",
    "name": "pgAnalytics v3.1.0 - Wave 3 Complete",
    "body": "See RELEASE_NOTES_v3.1.0.md for full release notes",
    "draft": false,
    "prerelease": false
  }'
```

---

## Release Checklist

Before publishing the release:

- [ ] **Code is ready**
  ```bash
  git status
  # Should show: nothing to commit, working tree clean
  ```

- [ ] **Tag exists locally**
  ```bash
  git tag -l | grep v3.1.0
  # Should output: v3.1.0
  ```

- [ ] **Tag is pushed to origin**
  ```bash
  git ls-remote origin | grep v3.1.0
  # Should show: refs/tags/v3.1.0
  ```

- [ ] **All commits merged to main**
  ```bash
  git log main --oneline | head -5
  # Should include recent commits
  ```

- [ ] **Release notes file exists**
  ```bash
  ls -lh RELEASE_NOTES_v3.1.0.md
  # Should show: 14K file with release notes
  ```

- [ ] **Production monitoring guide created**
  ```bash
  ls -lh PRODUCTION_MONITORING_GUIDE.md
  # Should show: monitoring guide
  ```

- [ ] **Documentation index updated in README**
  ```bash
  grep "RELEASE_NOTES_v3.1.0" README.md
  # Should show: release notes link
  ```

---

## What to Include in GitHub Release

### Essential Content

✅ **Release Title**
```
pgAnalytics v3.1.0 - Wave 3 Complete: MCP Integration & PostgreSQL 14-18 Support
```

✅ **Release Summary**
```markdown
## What's New in v3.1.0

- ✅ Wave 1: CLI tools for query/index/VACUUM analysis
- ✅ Wave 2: ML models for latency prediction & anomaly detection
- ✅ Wave 3: MCP protocol with JSON-RPC 2.0 transport
- ✅ PostgreSQL 14-18 full support with 100% feature parity
- ✅ 741+ tests with >85% code coverage
- ✅ Comprehensive documentation (58+ files, 800+ KB)
```

✅ **Key Metrics**
```markdown
## Release Metrics

- **Tests:** 741+ passing with >85% coverage
- **PostgreSQL Support:** Versions 14, 15, 16, 17, 18
- **Breaking Changes:** None - fully backward compatible
- **Commits:** 15 total, 2 new (release notes + monitoring guide)
```

✅ **Installation Instructions**
```markdown
## Quick Start

### Docker Compose
\`\`\`bash
git clone https://github.com/torresglauco/pganalytics-v3
cd pganalytics-v3
git checkout v3.1.0
docker-compose up -d
\`\`\`

### Kubernetes
\`\`\`bash
helm install pganalytics ./helm/pganalytics -n pganalytics
\`\`\`
```

✅ **Documentation Links**
```markdown
## Documentation

- [Release Notes](RELEASE_NOTES_v3.1.0.md) - Detailed feature list
- [Production Monitoring Guide](PRODUCTION_MONITORING_GUIDE.md) - Deployment validation
- [Installation Guide](DEPLOYMENT.md) - Deployment procedures
- [PostgreSQL Compatibility](POSTGRES_COMPATIBILITY.md) - Version support matrix
```

### Optional Enhancements

- 📦 **Attach Binaries**
  - Compiled collector binary
  - Docker image tarball
  - CLI tool executable

- 📊 **Add Screenshots**
  - Dashboard overview
  - CLI tool output
  - MCP integration example

- 📈 **Performance Metrics**
  - API response times
  - Resource usage
  - Load test results

---

## After Release Publication

### Share the Release

1. **Notify Team**
   ```bash
   # Slack message
   @here pgAnalytics v3.1.0 released! 🎉
   All 3 waves complete: CLI, ML/AI, & MCP integration
   PostgreSQL 14-18 full support validated
   See: https://github.com/torresglauco/pganalytics-v3/releases/tag/v3.1.0
   ```

2. **Update Documentation**
   - [ ] Add release link to main README
   - [ ] Update version numbers in docs
   - [ ] Add to version history table

3. **Create Deployment PR** (optional)
   ```bash
   # Create PR to deploy to staging/production
   git checkout -b deploy/v3.1.0
   # Update deployment manifests
   git push origin deploy/v3.1.0
   # Create PR for review
   ```

4. **Monitor Feedback**
   - Watch GitHub releases page for comments
   - Address any reported issues
   - Prepare hotfix if needed

---

## Release Assets

### Optional: Add Docker Image

```bash
# Build and push Docker image
docker build -t torresglauco/pganalytics:v3.1.0 .
docker push torresglauco/pganalytics:v3.1.0

# Tag as latest
docker tag torresglauco/pganalytics:v3.1.0 torresglauco/pganalytics:latest
docker push torresglauco/pganalytics:latest
```

### Optional: Create Binary Release

```bash
# Build CLI binary
cd backend/cmd/pganalytics-cli
go build -o pganalytics-cli

# Attach to release via GitHub web UI
# Or use gh CLI:
gh release upload v3.1.0 pganalytics-cli
```

---

## Troubleshooting

### Issue: Tag not found when creating release

**Solution:**
```bash
# Ensure tag exists locally
git tag -l | grep v3.1.0

# If missing, create it
git tag -a v3.1.0 -m "pgAnalytics v3.1.0 Release"

# Push to remote
git push origin v3.1.0
```

### Issue: GitHub CLI authentication fails

**Solution:**
```bash
# Remove cached credentials
rm ~/.config/gh/hosts.yml

# Re-authenticate
gh auth login

# Follow interactive prompts
# Choose: GitHub.com
# Choose: HTTPS
# Choose: Y for git credential helper
# Paste personal access token when prompted
```

### Issue: Release notes file too large

**Solution:**
```bash
# GitHub has a 125KB limit for release bodies
wc -c RELEASE_NOTES_v3.1.0.md

# If too large, create summary and link to full notes
# Use smaller summary in release, link to full file
```

---

## Next Steps After Release

1. ✅ **Monitor Deployment**
   - Follow [PRODUCTION_MONITORING_GUIDE.md](PRODUCTION_MONITORING_GUIDE.md)
   - Run validation checklist

2. ✅ **Track Issues**
   - Monitor GitHub issues for v3.1.0
   - Create hotfix branch if needed

3. ✅ **Plan Next Release**
   - Begin planning Wave 4 or v3.2.0
   - Gather feature requests

4. ✅ **Update Project Management**
   - Close completed tasks
   - Archive current milestone
   - Create next milestone

---

## Example Release Notes

See [`RELEASE_NOTES_v3.1.0.md`](RELEASE_NOTES_v3.1.0.md) for complete release notes content that can be copied directly into GitHub releases.

---

**🎉 Ready to Release!**

Choose your preferred method above and publish pgAnalytics v3.1.0 to the world.

For questions or assistance, refer to:
- [GitHub Releases Documentation](https://docs.github.com/en/repositories/releasing-projects-on-github/about-releases)
- [GitHub CLI Release Documentation](https://cli.github.com/manual/gh_release_create)

---

*Generated: April 2, 2026*
*Version: v3.1.0*
