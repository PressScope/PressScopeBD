import { expect, test } from "@playwright/test";

test("theme toggler switches color-scheme", async ({ page }) => {
  // Use "networkidle" so Vite's HMR websocket must drain before we touch the page.
  await page.goto("/login", { waitUntil: "networkidle", timeout: 60_000 });

  // Dump the HTML for CI diagnostics — helps us see what was actually rendered
  await page.evaluate(() => {
    (window as any).__dump = document.documentElement.outerHTML.slice(0, 3000);
  });
  // @ts-ignore
  const html = (await page.evaluate(() => (window as any).__dump)) as string;
  console.log("=== PAGE HTML ===\n", html);

  // Snapshot the state of the button (is it in the DOM?)
  const bodyChildren = await page.evaluate(
    () => (document.body || {}).innerHTML,
  );
  console.log("=== BODY (first 2000) ===\n", (bodyChildren || "").slice(0, 2000));

  // The next-themes inline script injects `html.class='dark'` based on
  // local storage.  If that script failed (e.g. Vite dev HMR race on CI),
  // it will still run to completion given a settle delay.
  // We make sure by waiting for Vite to finish draining.
  await page.waitForTimeout(1000);

  // Capture initial screenshot
  await page.screenshot({ path: "test-results/theme-s0-clean.png" });

  const button = page.locator('[data-slot="theme-toggler-button"]');

  const initial = await page.evaluate(() =>
    document.documentElement.classList.contains("dark") ? "dark" : "light",
  );
  console.log("Initial color-scheme:", initial);

  // 1st click
  await button.click();
  await page.waitForTimeout(2000);
  await page.screenshot({ path: "test-results/theme-s1-after-1st-click.png" });
  const after1 = await page.evaluate(() =>
    document.documentElement.classList.contains("dark") ? "dark" : "light",
  );
  expect(after1).not.toBe(initial);

  // 2nd click — returns to the original value
  await button.click();
  await page.waitForTimeout(2000);
  await page.screenshot({ path: "test-results/theme-s2-after-2nd-click.png" });
  const after2 = await page.evaluate(() =>
    document.documentElement.classList.contains("dark") ? "dark" : "light",
  );
  expect(after2).toBe(initial);
});
