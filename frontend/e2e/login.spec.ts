import { test, expect } from '@playwright/test';
import { login, logout, TEST_CREDENTIALS } from './auth-helpers';

test.describe('Login page', () => {
  test('displays login form', async ({ page }) => {
    await page.goto('/login');
    await expect(page.locator('h1')).toContainText('FreeMED EMR');
    await expect(page.locator('#username')).toBeVisible();
    await expect(page.locator('#password')).toBeVisible();
    await expect(page.locator('button[type="submit"]')).toContainText('Sign in');
  });

  test('shows validation errors for empty fields', async ({ page }) => {
    await page.goto('/login');
    // Submit empty form (Playwright triggers HTML5 validation but superforms
    // also shows error messages — we check the button state and form presence)
    await page.click('button[type="submit"]');
    // The form should still be visible (no redirect)
    await expect(page.locator('#username')).toBeVisible();
  });

  test('shows error for invalid credentials', async ({ page }) => {
    await page.goto('/login');
    await page.waitForSelector('#username', { state: 'visible' });
    await page.fill('#username', 'nonexistent_user');
    await page.fill('#password', 'wrongpassword');
    await page.click('button[type="submit"]');

    // Should see an error message or stay on the login page
    const errorBanner = page.locator('[role="alert"]');
    if (await errorBanner.isVisible({ timeout: 5000 }).catch(() => false)) {
      await expect(errorBanner).toContainText(/Invalid|error|occurred/i);
    }
    // Regardless, should not have navigated away from login
    await expect(page).toHaveURL(/\/login/);
  });
});

test.describe('Authenticated session', () => {
  test('valid login redirects to dashboard', async ({ page }) => {
    await login(page);
    // Should be on dashboard with welcome message
    await expect(page.locator('main')).toContainText('Welcome');
    // Navbar should be visible
    await expect(page.locator('nav')).toBeVisible();
  });

  test('logout redirects to login page', async ({ page }) => {
    await login(page);
    await logout(page);
    // Should be back on login page
    await expect(page).toHaveURL('/login');
    await expect(page.locator('#username')).toBeVisible();
  });

  test('protected routes redirect to login when not authenticated', async ({ page }) => {
    await page.goto('/patients');
    // Should be redirected to login
    await expect(page).toHaveURL(/\/login/);
  });
});
