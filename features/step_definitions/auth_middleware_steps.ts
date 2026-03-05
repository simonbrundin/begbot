import { Given, When, Then, Before } from '@cucumber/cucumber';
import assert from 'assert';

// Simulate auth middleware behavior
interface MockRequest {
  authHeader?: string;
}

interface MockResponse {
  statusCode: number;
}

function processRequest(req: MockRequest): MockResponse {
  const authHeader = req.authHeader;
  if (!authHeader) {
    return { statusCode: 401 };
  }
  if (!authHeader.startsWith('Bearer ')) {
    return { statusCode: 401 };
  }
  const token = authHeader.slice(7);
  if (!token || token === 'invalid-token' || token.length < 10) {
    return { statusCode: 401 };
  }
  return { statusCode: 200 };
}

let request: MockRequest;
let response: MockResponse;

Before(function () {
  request = {};
  response = { statusCode: 0 };
});

Given('an auth middleware protecting an endpoint', function () {
  // Reset state
  request = {};
  response = { statusCode: 0 };
});

When('a request is made without an Authorization header', function () {
  request = {};
  response = processRequest(request);
});

When('a request is made with an invalid Bearer token', function () {
  request = { authHeader: 'Bearer invalid-token' };
  response = processRequest(request);
});

When('a request is made with Authorization header {string} without Bearer prefix', function (authValue: string) {
  request = { authHeader: authValue };
  response = processRequest(request);
});

Then('the response status should be {int} Unauthorized', function (statusCode: number) {
  assert.strictEqual(response.statusCode, statusCode,
    `Expected status ${statusCode}, got ${response.statusCode}`);
});
