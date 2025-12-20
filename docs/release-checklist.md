# Axiomod Release Checklist

Follow this checklist to push the initial commit and create the first release of the Axiomod framework.

## 1. Local Preparation

- [ ] **Verify Git Ignore**: Ensure `.gitignore` exists and covers secrets/binaries.
- [ ] **Clean Dependencies**: Run `go mod tidy` to remove unused dependencies.
- [ ] **Verify Build**: Run `make build` to ensure the binary compiles.
- [ ] **Run Tests**: Run `make test` to ensure all tests pass.
- [ ] **Lint Code**: Run `make lint` to check for style issues.

## 2. Git Initialization (First Time Only)

```bash
# Initialize repository
git init

# Add all files
git add .

# Create initial commit
git commit -m "feat: initial commit of axiomod framework v1.0.0"

# Rename branch to main
git branch -M main
```

## 3. Remote Configuration

- [ ] **Create Repository**: Create a new public/private repository on GitHub named `axiomod`.
- [ ] **Add Remote**:

```bash
git remote add origin https://github.com/axiomod/axiomod.git
```

## 4. Push & Release

- [ ] **Push Code**:

```bash
git push -u origin main
```

- [ ] **Tag Version**:

```bash
git tag v1.0.0
git push origin v1.0.0
```

## 5. GitHub Release (Optional but Recommended)

1. Go to the GitHub repository.
2. Click **Releases** > **Draft a new release**.
3. Choose tag `v1.0.0`.
4. Title: `v1.0.0 - Enterprise Ready`.
5. Description: Copy content from `readiness_assessment.md` or `metrics`.
6. Attach binaries (if applicable).
