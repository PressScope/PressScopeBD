import { expect, test } from "@playwright/test";

test("theme toggler switches color-scheme", async ({ page }) => {
  await page.goto("/login");

  // The ThemeTogglerButton in header.tsx renders a button with data-slot="theme-toggler-button"
  const button = page.locator('[data-slot="theme-toggler-button"]');
  await expect(button).toBeVisible();

  // Capture the initial color-scheme from the documentElement classList
  const initial = await page.evaluate(() => {
    return document.documentElement.classList.contains("dark") ? "dark" : "light";
  });

  // Click to cycle to the next theme
  await button.click();
  await page.waitForTimeout(1000);

  // --- Capture post-click state ---
  await page.screenshot({ path: "test-results/theme-after-first-click.png" });

  const afterClick = await page.evaluate(() => {
    return document.documentElement.classList.contains("dark") ? "dark" : "light";
  });

  // The color-scheme must have Changed
  expect(afterClick).not.toBe(initial);

  // Click a second time to move away from the first value
  await button.click();
  await page.waitForTimeout(1000);

  await page.screenshot({ path: "test-results/theme-after-second-click.png" });

  const afterSecondClick = await page.evaluate(() => {
    return document.documentElement.classList.contains("dark") ? "dark" : "light";
  });

  expect(afterSecondClick).toBe(initial);
});
