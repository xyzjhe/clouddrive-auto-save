import { test as setup, expect } from '@playwright/test';
import { add139Account, addQuarkAccount } from '../fixtures/account.fixture';

const authFile = 'playwright/.auth/user.json';

setup('预置测试账号', async ({ page }) => {
  await add139Account(page);
  await addQuarkAccount(page);
  await page.context().storageState({ path: authFile });
});
