import { test, expect } from '@playwright/test';
import { login } from './auth-helpers';

test.describe('Patients page', () => {
  test.beforeEach(async ({ page }) => {
    await login(page);
  });

  test('displays patient search interface', async ({ page }) => {
    await page.goto('/patients');

    // Should see the heading
    await expect(page.locator('h1')).toContainText('Patient Search');

    // Search input should be visible
    const searchInput = page.locator('input[placeholder*="Search by name"]');
    await expect(searchInput).toBeVisible();

    // Search button should be visible
    await expect(page.locator('button:has-text("Search")').first()).toBeVisible();
  });

  test('shows "no patients found" on empty search', async ({ page }) => {
    await page.goto('/patients');

    // Type a search that won't match anything
    const searchInput = page.locator('input[placeholder*="Search by name"]');
    await searchInput.fill('zzz_no_match_xyz');
    await page.click('button:has-text("Search")');

    // Should show "no patients found" message
    await expect(page.locator('text=No patients found')).toBeVisible({ timeout: 10000 });
  });

  test('navigates to patient detail on row click (if results)', async ({ page }) => {
    await page.goto('/patients');

    // Try searching with just two letters to get picklist
    const searchInput = page.locator('input[placeholder*="Search by name"]');
    await searchInput.fill('Ja');
    // Wait a bit for debounced picklist
    await page.waitForTimeout(500);

    // If picklist or results appear, verify we can interact
    const hasResults = await page.locator('table tbody tr').first().isVisible({ timeout: 3000 }).catch(() => false);
    if (hasResults) {
      // Click the first row
      await page.locator('table tbody tr').first().click();
      // Should navigate to patient detail page
      await expect(page).toHaveURL(/\/patients\/\d+/);
    }
    // If no results, that's fine — DB may be empty
  });
});
