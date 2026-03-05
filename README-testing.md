Testing guide for BDD features

Overview
--------
We run Cucumber.js for feature tests. Steps live in `features/step_definitions` and shared helpers in `features/support`.

Run tests (mocked, safe):

```bash
npm run cucumber
```

Use real auth (only in a test environment):

```bash
export USE_REAL_AUTH=true
export SUPABASE_URL=... 
export SUPABASE_KEY=...
export TEST_USER_EMAIL=... 
export TEST_USER_PASSWORD=...
npm run cucumber
```

CI guidance
-----------
- Run mocked tests (`npm run cucumber`) in CI by default.
- If you need real auth tests, run them in a separate job with only against a dedicated test Supabase instance.
