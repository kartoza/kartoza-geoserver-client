"""Playwright configuration for E2E tests."""

# This configuration is used by pytest-playwright
# Run tests with: pytest tests/e2e -v

# Browser settings
PLAYWRIGHT_BROWSERS = ["chromium"]

# Test settings
PLAYWRIGHT_HEADLESS = True
PLAYWRIGHT_TIMEOUT = 30000  # 30 seconds

# Video recording (for debugging failures)
PLAYWRIGHT_VIDEO = "retain-on-failure"

# Screenshot settings
PLAYWRIGHT_SCREENSHOT = "only-on-failure"

# Trace settings (for debugging)
PLAYWRIGHT_TRACE = "retain-on-failure"
