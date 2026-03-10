import { Given, When, Then, Before, After } from '@cucumber/cucumber';
import assert from 'assert';

// Toggle: om USE_REAL_AUTH=true använder vi riktiga Supabase-anrop.
const USE_REAL_AUTH = process.env.USE_REAL_AUTH === 'true';
const SUPABASE_URL = process.env.SUPABASE_URL;
const SUPABASE_KEY = process.env.SUPABASE_KEY;
const TEST_USER_EMAIL = process.env.TEST_USER_EMAIL || 'test@example.com';
const TEST_USER_PASSWORD = process.env.TEST_USER_PASSWORD || 'correctpassword';

interface User {
  email: string;
  password: string;
  isLoggedIn: boolean;
}

let users: Map<string, User>;
let currentUser: User | null;
let lastError: string | null;

Before(function ({ pickle }) {
  // reset state between scenarios
  users = new Map();
  currentUser = null;
  lastError = null;
  console.log(`Starting scenario: ${pickle?.name || 'unknown'}`);
});

After(function ({ pickle }) {
  console.log(`Finished scenario: ${pickle?.name || 'unknown'}`);
});

async function realSignIn(email: string, password: string) {
  if (!SUPABASE_URL || !SUPABASE_KEY) throw new Error('SUPABASE_URL and SUPABASE_KEY must be set to use real auth');
  // dynamic import to avoid requiring package when not used
  // eslint-disable-next-line @typescript-eslint/no-var-requires
  const { createClient } = require('@supabase/supabase-js');
  const supabase = createClient(SUPABASE_URL, SUPABASE_KEY);
  const res = await supabase.auth.signInWithPassword({ email, password });
  return res;
}

// Given: skapa testkonto (mock) eller bara markera att konto finns
async function createValidAccount() {
  users.set(TEST_USER_EMAIL, {
    email: TEST_USER_EMAIL,
    password: TEST_USER_PASSWORD,
    isLoggedIn: false
  });
}

Given('användaren har ett giltigt konto med e-post och lösenord', async function () {
  await createValidAccount();
});

// Alias - kort form som förekommer i vissa feature-filer
Given('användaren har ett giltigt konto', async function () {
  await createValidAccount();
});

Given('ingen användare finns med den angivna e-posten', function () {
  // leave users map empty for this scenario
});

Given('användaren är inloggad', async function () {
  if (USE_REAL_AUTH) {
    const r = await realSignIn(TEST_USER_EMAIL, TEST_USER_PASSWORD);
    if (r.error) throw r.error;
    // we won't store token here into mock state
    currentUser = { email: TEST_USER_EMAIL, password: TEST_USER_PASSWORD, isLoggedIn: true };
  } else {
    const user: User = { email: TEST_USER_EMAIL, password: TEST_USER_PASSWORD, isLoggedIn: true };
    users.set(user.email, user);
    currentUser = user;
  }
});

When('användaren anger korrekt e-post och lösenord', async function () {
  lastError = null;
  if (USE_REAL_AUTH) {
    const res = await realSignIn(TEST_USER_EMAIL, TEST_USER_PASSWORD);
    if (res.error) {
      lastError = res.error.message || 'auth error';
      return;
    }
    currentUser = { email: TEST_USER_EMAIL, password: TEST_USER_PASSWORD, isLoggedIn: true };
    return;
  }

  const user = users.get(TEST_USER_EMAIL);
  if (user && user.password === TEST_USER_PASSWORD) {
    user.isLoggedIn = true;
    currentUser = user;
  } else {
    lastError = 'Invalid login credentials';
  }
});

When('användaren anger fel e-post eller lösenord', async function () {
  lastError = null;
  if (USE_REAL_AUTH) {
    try {
      await realSignIn('wrong@example.com', 'wrongpassword');
      lastError = 'Expected login to fail';
    } catch (err: any) {
      lastError = err.message || String(err);
    }
    return;
  }

  const user = users.get(TEST_USER_EMAIL);
  if (user) {
    user.isLoggedIn = false;
    currentUser = null;
    lastError = 'Invalid login credentials';
  } else {
    lastError = 'Invalid login credentials';
  }
});

When('användaren försöker logga in med icke-existerande e-post', async function () {
  lastError = null;
  if (USE_REAL_AUTH) {
    try {
      await realSignIn(`nonexistent${Date.now()}@example.com`, 'any');
      lastError = 'Expected login to fail';
    } catch (err: any) {
      lastError = err.message || String(err);
    }
    return;
  }

  const user = users.get('nonexistent@example.com');
  if (!user) lastError = 'Invalid login credentials';
});

// Support both variants used across features
When('användaren loggar ut', function () {
  if (currentUser) {
    currentUser.isLoggedIn = false;
    currentUser = null;
  }
});

When('användaren klickar på logga ut', function () {
  // directly call the logout behaviour
  if (currentUser) {
    currentUser.isLoggedIn = false;
    currentUser = null;
  }
});

// Thens / assertions
Then('ska användaren loggas in framgångsrikt', function () {
  assert(currentUser !== null, 'Användaren borde vara inloggad');
  assert(currentUser?.isLoggedIn === true, 'isLoggedIn borde vara true');
});

Then('ska användaren loggas in', function () {
  // same as successful login check
  assert(currentUser !== null, 'Användaren borde vara inloggad');
  assert(currentUser?.isLoggedIn === true, 'isLoggedIn borde vara true');
});

Then('användaren ska omdirigeras till startsidan', function () {
  assert(currentUser !== null, 'Användaren borde vara inloggad för att omdirigeras');
});

Then('ska inloggningen misslyckas', function () {
  assert(currentUser === null, 'Användaren borde INTE vara inloggad');
});

// main error message assertions
Then('ett felmeddelande ska visas', function () {
  assert(lastError !== null, 'Ett felmeddelande borde finnas');
});

Then('ett felmeddelande visas', function () {
  assert(lastError !== null, 'Ett felmeddelande borde finnas');
});

Then('ska användaren loggas ut', function () {
  assert(currentUser === null, 'Användaren borde vara utloggad');
});

Then('ska omdirigeras till inloggningssidan', function () {
  assert(currentUser === null, 'Användaren borde vara utloggad för att omdirigeras');
});
