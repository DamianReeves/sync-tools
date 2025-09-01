# Homebrew Tap Setup Instructions

This document provides step-by-step instructions for setting up the `DamianReeves/homebrew-tap` repository for Homebrew distribution of sync-tools.

## Prerequisites

1. GitHub account with access to create repositories under `DamianReeves`
2. Personal Access Token with repo write permissions
3. GoReleaser configuration already in place (✅ completed)

## Step 1: Create homebrew-tap Repository

1. Go to GitHub and create a new repository:
   - **Name**: `homebrew-tap` 
   - **Full name**: `DamianReeves/homebrew-tap`
   - **Description**: "Homebrew tap for DamianReeves projects"
   - **Visibility**: Public (required for Homebrew taps)
   - **Initialize**: with README

2. Clone the repository locally:
   ```bash
   git clone https://github.com/DamianReeves/homebrew-tap.git
   cd homebrew-tap
   ```

## Step 2: Set Up Repository Structure

Create the required directory structure:

```bash
# Create Formula directory (required by Homebrew)
mkdir -p Formula

# Create initial README
cat > README.md << 'EOF'
# DamianReeves Homebrew Tap

This is a Homebrew tap for DamianReeves projects.

## Installation

```bash
# Add the tap
brew tap DamianReeves/tap

# Install sync-tools
brew install sync-tools
```

## Available Packages

- **sync-tools**: Go CLI wrapper around rsync with advanced filtering and SyncFile post-sync actions

## Automated Updates

This tap is automatically updated by GoReleaser when new releases are created.
EOF

# Commit initial structure
git add .
git commit -m "Initial tap setup with Formula directory and README"
git push origin main
```

## Step 3: Configure GitHub Token

You'll need a Personal Access Token (PAT) with repo permissions:

1. Go to GitHub Settings → Developer settings → Personal access tokens
2. Create a new token (classic) with these scopes:
   - `repo` (Full control of private repositories)
   - `workflow` (Update GitHub Action workflows) - optional but recommended

3. Add the token as a secret in the sync-tools repository:
   - Go to `DamianReeves/sync-tools` → Settings → Secrets and variables → Actions
   - Create a new repository secret:
     - **Name**: `HOMEBREW_TAP_GITHUB_TOKEN`
     - **Value**: Your PAT token

## Step 4: Test the Setup

The setup can be tested by creating a release:

1. **Create a test tag** (from sync-tools repository):
   ```bash
   cd /path/to/sync-tools
   git tag v0.4.1-M1
   git push origin v0.4.1-M1
   ```

2. **Monitor the release process**:
   - GitHub Actions will run in sync-tools repository
   - GoReleaser will build binaries and create GitHub Release
   - GoReleaser will automatically update homebrew-tap with new formula

3. **Verify the formula was created**:
   - Check `DamianReeves/homebrew-tap` repository
   - Look for `Formula/sync-tools.rb` file
   - Verify it contains correct version and download URLs

4. **Test installation** (after successful release):
   ```bash
   brew tap DamianReeves/tap
   brew install sync-tools
   sync-tools --version
   ```

## Step 5: Repository Configuration

Add these files to the homebrew-tap repository:

### .gitignore
```gitignore
.DS_Store
*.swp
*.swo
*~
```

### .github/FUNDING.yml (optional)
```yaml
github: DamianReeves
```

## Expected Workflow

Once set up, the release process will be:

1. **Developer creates release tag** in sync-tools:
   ```bash
   git tag v0.5.0
   git push origin v0.5.0
   ```

2. **GitHub Actions triggers** in sync-tools repository

3. **GoReleaser runs** and:
   - Builds cross-platform binaries
   - Creates GitHub Release with assets
   - Generates Homebrew formula
   - Commits formula to `DamianReeves/homebrew-tap`

4. **Users can install** immediately:
   ```bash
   brew upgrade sync-tools  # If already installed
   # OR
   brew install DamianReeves/tap/sync-tools  # First installation
   ```

## Troubleshooting

### Common Issues

**Formula not updating:**
- Check GitHub Actions logs in sync-tools repository
- Verify `HOMEBREW_TAP_GITHUB_TOKEN` secret is set correctly
- Ensure token has repo permissions for homebrew-tap

**Installation fails:**
- Verify formula syntax: `brew audit --formula DamianReeves/tap/sync-tools`
- Check download URLs are accessible
- Test formula manually: `brew install --build-from-source DamianReeves/tap/sync-tools`

**Token permission errors:**
- Ensure PAT has `repo` scope
- Verify token hasn't expired
- Check token has access to both repositories

### Manual Formula Update

If needed, you can manually update the formula:

1. Install GoReleaser locally:
   ```bash
   brew install goreleaser
   ```

2. Generate formula manually:
   ```bash
   cd /path/to/sync-tools
   goreleaser build --snapshot --clean
   # Formula will be generated in dist/
   ```

3. Copy to homebrew-tap and commit:
   ```bash
   cp dist/homebrew-tap/Formula/sync-tools.rb /path/to/homebrew-tap/Formula/
   cd /path/to/homebrew-tap
   git add Formula/sync-tools.rb
   git commit -m "Update sync-tools formula to vX.Y.Z"
   git push origin main
   ```

## Verification Checklist

- [ ] Repository `DamianReeves/homebrew-tap` created and public
- [ ] `Formula/` directory exists
- [ ] README.md with installation instructions
- [ ] GitHub token `HOMEBREW_TAP_GITHUB_TOKEN` configured in sync-tools repo
- [ ] GoReleaser configuration references correct homebrew-tap repository
- [ ] Test release creates formula automatically
- [ ] Manual installation works: `brew tap DamianReeves/tap && brew install sync-tools`
- [ ] Binary executes correctly: `sync-tools --version`

## Repository Links

- **Main Project**: https://github.com/DamianReeves/sync-tools
- **Homebrew Tap**: https://github.com/DamianReeves/homebrew-tap
- **Releases**: https://github.com/DamianReeves/sync-tools/releases

Once this setup is complete, sync-tools will have professional-grade Homebrew distribution with automated formula updates on every release.