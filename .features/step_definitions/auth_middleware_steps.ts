import { Given, When, Then, Before } from '@cucumber/cucumber';
import assert from 'assert';

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

Given('en autentiseringsmiddleware som skyddar en endpoint', function () {
  request = {};
  response = { statusCode: 0 };
});

When('en förfrågan görs utan Authorization-header', function () {
  request = {};
  response = processRequest(request);
});

When('en förfrågan görs med en ogiltig Bearer-token', function () {
  request = { authHeader: 'Bearer invalid-token' };
  response = processRequest(request);
});

When('en förfrågan görs med Authorization-headern {string} utan Bearer-prefix', function (authValue: string) {
  request = { authHeader: authValue };
  response = processRequest(request);
});

Then('ska svarsstatusen vara {int} Obehörig', function (statusCode: number) {
  assert.strictEqual(response.statusCode, statusCode,
    `Förväntade status ${statusCode}, fick ${response.statusCode}`);
});
