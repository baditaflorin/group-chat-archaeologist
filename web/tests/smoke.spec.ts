import { expect, test } from '@playwright/test';

test('loads the static chat archaeology explorer and map', async ({ page }) => {
  const errors: string[] = [];
  page.on('console', (message) => {
    if (message.type() === 'error') {
      errors.push(message.text());
    }
  });

  await page.goto('./');
  await expect(page.getByRole('heading', { name: 'Group Chat Archaeologist' })).toBeVisible();
  await expect(page.getByRole('link', { name: 'Star on GitHub' })).toHaveAttribute(
    'href',
    'https://github.com/baditaflorin/group-chat-archaeologist'
  );
  await expect(page.getByRole('link', { name: 'Support' })).toHaveAttribute(
    'href',
    'https://www.paypal.com/paypalme/florinbadita'
  );
  await page.getByRole('button', { name: 'Map' }).click();
  await expect(page.getByRole('heading', { name: 'Who Introduced Whom' })).toBeVisible();
  await expect(page.getByAltText('GraphViz map showing who introduced whom in the group chat')).toBeVisible();
  await expect(page.getByText(/Version v0\.1\.0/)).toBeVisible();
  expect(errors).toEqual([]);
});
