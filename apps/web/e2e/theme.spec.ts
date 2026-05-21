import { expect, test } from "@playwright/test";

test("theme toggler switches color-scheme", async ({ page }) => {
  await page.goto("/login");

  // --- Capture initial state -------------------------------------------------
  await page.screenshot({ path: "test-results/theme-before.png" });

  const button = page.locator('[data-slot="theme-toggler-button"]');
  await expect(button).toBeVisible();

  const initial = await page.evaluate(() => {
    return document.documentElement.classList.contains("dark") ? "dark" : "light";
  });

  // --- Click and wait for the transition to settle ---------------------------
  await button.click();
  await page.waitForTimeout(1000);

  // --- Capture post-click state ----------------------------------------------
  await page.screenshot({ path: "test-results/theme-after.png" });

  const afterClick = await page.evaluate(() => {
    return document.documentElement.classList.contains("dark") ? "dark" : "light";
  });

  expect(afterClick).not.toBe(initial);

  // --- Click a second time to move away from the first value -----------------
  await button.click();
  await page.waitForTimeout(1000);

  await page.screenshot({ path: "test-results/theme-after-second-click.png" });

  const afterSecondClick = await page.evaluate(() => {
    return document.documentElement.classList.contains("dark") ? "dark" : "light";
  });

  expect(afterSecondClick).toBe(initial);
});
