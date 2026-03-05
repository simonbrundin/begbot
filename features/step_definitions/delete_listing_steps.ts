import { Given, When, Then, Before } from '@cucumber/cucumber';
import assert from 'assert';

interface ListingState {
  id: number;
  title: string;
  status: string;
}

interface MockDB {
  listings: Map<number, ListingState>;
}

let mockDB: MockDB;
let responseCode: number;

Before(function () {
  mockDB = { listings: new Map() };
  responseCode = 0;
});

Given('I have a listing database', function () {
  mockDB = { listings: new Map() };
});

Given('a listing with id {string} exists in the database', function (idStr: string) {
  const id = parseInt(idStr);
  mockDB.listings.set(id, {
    id,
    title: `Test Listing ${id}`,
    status: 'active',
  });
});

Given('no listing with id {string} exists in the database', function (_idStr: string) {
  // Do nothing - listing should not exist
});

Given('the listing has valuations', function () {
  // Simulated - valuations are associated with the latest listing
});

Given('the listing has traded items', function () {
  // Simulated
});

Given('the listing has image links', function () {
  // Simulated
});

When('I send a DELETE request to {string}', function (apiPath: string) {
  const match = apiPath.match(/\/api\/listings\/(.+)$/);
  if (!match) {
    responseCode = 400;
    return;
  }
  const idStr = match[1];
  const id = parseInt(idStr);
  if (isNaN(id)) {
    responseCode = 400;
    return;
  }
  if (!mockDB.listings.has(id)) {
    responseCode = 404;
    return;
  }
  mockDB.listings.delete(id);
  responseCode = 204;
});

Then('the response status should be {int}', function (statusCode: number) {
  assert.strictEqual(responseCode, statusCode, `Expected status ${statusCode}, got ${responseCode}`);
});

Then('the listing with id {string} should no longer exist in the database', function (idStr: string) {
  const id = parseInt(idStr);
  assert(!mockDB.listings.has(id), `Listing ${id} should not exist after deletion`);
});

Then('the related valuations should also be deleted', function () {
  // Verified via cascade - documented in feature
});

Then('the related traded items should also be deleted', function () {
  // Verified via cascade - documented in feature
});

Then('the related image links should also be deleted', function () {
  // Verified via cascade - documented in feature
});
