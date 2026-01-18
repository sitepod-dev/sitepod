# Simple Static Site Example

This is a minimal example demonstrating how to deploy a static HTML site with SitePod.

## Directory Structure

```
.
├── sitepod.toml    # SitePod configuration file
└── dist/           # The directory containing your build output
    └── index.html  # The actual file served to users
```

## How to Deploy

1. **Install SitePod** (if you haven't already):
   ```bash
   curl -sL sitepod.dev/install.sh | bash
   ```

2. **Initialize** (only once):
   ```bash
   # Create a new project in the console or CLI
   sitepod init
   ```

3. **Deploy**:
   ```bash
   sitepod deploy
   ```

4. **Preview**:
   ```bash
   sitepod preview
   ```

5. **Rollback**:
   ```bash
   sitepod rollback
   ```

6. **Update**:
   Modify `dist/index.html` and run `sitepod deploy` again. The changes will be live instantly.
