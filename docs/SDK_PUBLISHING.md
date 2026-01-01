# SDK Publishing Guide

## Python SDK (PyPI)

### Prerequisites
- PyPI account at https://pypi.org
- API token (Settings > API tokens > Add token)

### Build
```bash
cd sdks/python
pip install build twine
python -m build
```

### Publish
```bash
# Test upload (optional)
twine upload --repository testpypi dist/*

# Production upload
twine upload dist/*
```

You'll be prompted for credentials:
- Username: `__token__`
- Password: Your API token (starts with `pypi-`)

### Verify
```bash
pip install gatewayops
python -c "from gatewayops import GatewayOps; print('Success!')"
```

---

## TypeScript SDK (npm)

### Prerequisites
- npm account at https://www.npmjs.com
- Logged in: `npm login`

### Build
```bash
cd sdks/typescript
npm install
npm run build
```

### Publish
```bash
# First time (scoped package)
npm publish --access public

# Updates
npm version patch  # or minor/major
npm publish
```

### Verify
```bash
npm install @gatewayops/sdk
node -e "const { GatewayOps } = require('@gatewayops/sdk'); console.log('Success!')"
```

---

## Automation (GitHub Actions)

Add `.github/workflows/publish-sdks.yml`:

```yaml
name: Publish SDKs

on:
  release:
    types: [published]

jobs:
  publish-python:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-python@v5
        with:
          python-version: '3.11'
      - name: Build and publish
        env:
          TWINE_USERNAME: __token__
          TWINE_PASSWORD: ${{ secrets.PYPI_API_TOKEN }}
        run: |
          cd sdks/python
          pip install build twine
          python -m build
          twine upload dist/*

  publish-typescript:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-node@v4
        with:
          node-version: '20'
          registry-url: 'https://registry.npmjs.org'
      - name: Build and publish
        env:
          NODE_AUTH_TOKEN: ${{ secrets.NPM_TOKEN }}
        run: |
          cd sdks/typescript
          npm install
          npm run build
          npm publish --access public
```

### Required Secrets
- `PYPI_API_TOKEN`: PyPI token
- `NPM_TOKEN`: npm token (Settings > Access Tokens)

---

## Version Management

### Bump Python version
Edit `sdks/python/pyproject.toml`:
```toml
version = "0.2.0"
```

### Bump TypeScript version
```bash
cd sdks/typescript
npm version minor
```

Or edit `package.json`:
```json
"version": "0.2.0"
```

---

## Current Versions

| SDK | Version | Status |
|-----|---------|--------|
| Python | 0.1.0 | Built |
| TypeScript | 0.1.0 | Built |

Built packages:
- `sdks/python/dist/gatewayops-0.1.0-py3-none-any.whl`
- `sdks/python/dist/gatewayops-0.1.0.tar.gz`
- `sdks/typescript/dist/` (ESM + CJS + types)
