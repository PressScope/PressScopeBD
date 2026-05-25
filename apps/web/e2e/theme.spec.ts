// data-slot = "theme-toggler-button"

import { test, expect } from "@playwright/test";

test("has theme toggler button", async ({ page }) => {
  await page.goto("/");

  await expect(
    page.locator("button[data-slot='theme-toggler-button']"),
  ).toBeVisible();
});
test("toggles theme on click", async ({ page }) => {
  await page.goto("/");

  const themeToggler = page.locator("button[data-slot='theme-toggler-button']");
  const html = page.locator("html");

  await expect(themeToggler).toBeVisible();

  await expect(html).toHaveClass(/light|dark/);

  const initial = await html.getAttribute("class");
  const initialTheme = initial?.includes("dark") ? "dark" : "light";
  const nextTheme = initialTheme === "dark" ? "light" : "dark";

  await themeToggler.click();
  await expect(html).toHaveClass(new RegExp(nextTheme), { timeout: 5000 });
  await themeToggler.click();
  await expect(html).toHaveClass(new RegExp(initialTheme), { timeout: 5000 });
});
