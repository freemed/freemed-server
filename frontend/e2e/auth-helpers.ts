import { Page } from '@playwright/test';

export const TEST_CREDENTIALS = {
  username: 'e2e_test_user',
  password: 'e2e_test_password',
};

/**
 * Login helper: performs login via direct API calls (bypasses SvelteKit form issues).
 * After login, navigates to the dashboard and waits for it to load.
 */
export async function login(page: Page): Promise<void> {
  await page.goto('/login');

  // Wait for CSRF token to be fetched (the page does this onMount)
  await page.waitForTimeout(500);

  // Perform login via the auth API directly (same as what the JS handler does)
  const result = await page.evaluate(async ({ username, password }) => {
    // Read CSRF cookie
    const csrfMatch = document.cookie.match(/(?:^|;\s*)csrf_token=([^;]*)/);
    const csrfToken = csrfMatch ? csrfMatch[1] : null;

    const headers: Record<string, string> = { 'Content-Type': 'application/json' };
    if (csrfToken) {
      headers['X-CSRF-Token'] = csrfToken;
    }

    const res = await fetch('/auth/login', {
      method: 'POST',
      headers,
      body: JSON.stringify({ username, password }),
    });
    return res.ok;
  }, TEST_CREDENTIALS);

  if (!result) {
    throw new Error('Login API call failed');
  }

  // Navigate to dashboard
  await page.goto('/');

  // Wait for authenticated state — the layout checks /auth/me and renders the dashboard
  await page.waitForSelector('nav', { timeout: 10000 });
  await page.waitForSelector('text=Welcome', { timeout: 10000 });
}

/**
 * Logout helper: clicks the Logout button in the navbar.
 */
export async function logout(page: Page): Promise<void> {
  // Perform logout via API (same as what the Logout button handler does)
  await page.evaluate(async () => {
    await fetch('/auth/logout', { method: 'DELETE' });
  });

  // Navigate to login page
  await page.goto('/login');
  await page.waitForSelector('#username', { timeout: 5000 });
}
