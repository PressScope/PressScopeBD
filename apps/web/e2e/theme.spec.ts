import { expect, test } from "@playwright/test";

test("theme toggler switches color-scheme", async ({ page }) => {
  await page.goto("/login", { waitUntil: "networkidle" });

  // Dump the HTML to help diagnose CI failures
  const html = await page.evaluate(() => document.documentElement.outerHTML);
  console.log(html.substring(0, 2000));

  // Wait up to 15 s for the button to appear (handles SSG/hydration delay in CI)
  const button = page.locator('[data-slot="theme-toggler-button"]');
  await button.waitFor({ state: "attached", timeout: 15_000 });

  // Capture the initial color-scheme
  const initial = await page.evaluate(() => {
    return document.documentElement.classList.contains("dark") ? "dark" : "light";
  });
  console.log("Initial color-scheme:", initial);

  // Click to cycle to the next theme
  await button.click();
  await page.waitForTimeout(1500);

  // --- Capture post-click state ---
  await page.screenshot({ path: "test-results/theme-after-first-click.png" });

  const afterClick = await page.evaluate(() => {
    return document.documentElement.classList.contains("dark") ? "dark" : "light";
  });
  console.log("After click color-scheme:", afterClick);

  // The color-scheme must have Changed
  expect(afterClick).not.toBe(initial);

  // Click a second time to move away from the first value
  await button.click();
  await page.waitForTimeout(1500);

  await page.screenshot({ path: "test-results/theme-after-second-click.png" });

  const afterSecondClick = await page.evaluate(() => {
    return document.documentElement.classList.contains("dark") ? "dark" : "light";
  });
  console.log("After second click color-scheme:", afterSecondClick);

  expect(afterSecondClick).toBe(initial);
});
