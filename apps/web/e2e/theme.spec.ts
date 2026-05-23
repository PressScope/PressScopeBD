import { expect, test } from "@playwright/test";

test("theme toggler switches color-scheme", async ({ page }) => {
  await page.goto("/login", {
    waitUntil: "networkidle",
    timeout: 60_000,
  });

  await page.waitForTimeout(1000);

  await page.screenshot({
    path: "test-results/theme-s0-clean.png",
  });

  const button = page.locator('[data-slot="theme-toggler-button"]');

  const getTheme = async () =>
    page.evaluate(() => {
      const root = document.documentElement;

      if (root.classList.contains("dark")) return "dark";
      if (root.classList.contains("light")) return "light";

      return window.matchMedia("(prefers-color-scheme: dark)").matches
        ? "dark"
        : "light";
    });

  const initial = await getTheme();

  console.log("Initial color-scheme:", initial);

  // First toggle
  await button.click();
  await page.waitForTimeout(2000);

  await page.screenshot({
    path: "test-results/theme-s1-after-1st-click.png",
  });

  const after1 = await getTheme();

  expect(after1).not.toBe(initial);

  // Second toggle
  await button.click();
  await page.waitForTimeout(2000);

  await page.screenshot({
    path: "test-results/theme-s2-after-2nd-click.png",
  });

  const after2 = await getTheme();

  expect(after2).toBe(initial);
});
