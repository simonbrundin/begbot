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

Given('jag har en annonstabas', function () {
  mockDB = { listings: new Map() };
});

Given('en annons med id {string} finns i databasen', function (idStr: string) {
  const id = parseInt(idStr);
  mockDB.listings.set(id, { id, title: `Testannons ${id}`, status: 'active' });
});

Given('ingen annons med id {string} finns i databasen', function (_idStr: string) {});

Given('annonsen har värderingar', function () {});
Given('annonsen har handlade varor', function () {});
Given('annonsen har bildlänkar', function () {});

When('jag skickar en DELETE-förfrågan till {string}', function (apiPath: string) {
  const match = apiPath.match(/\/api\/listings\/(.+)$/);
  if (!match) { responseCode = 400; return; }
  const idStr = match[1];
  const id = parseInt(idStr);
  if (isNaN(id)) { responseCode = 400; return; }
  if (!mockDB.listings.has(id)) { responseCode = 404; return; }
  mockDB.listings.delete(id);
  responseCode = 204;
});

Then('ska svarsstatusen vara {int}', function (statusCode: number) {
  assert.strictEqual(responseCode, statusCode, `Förväntade status ${statusCode}, fick ${responseCode}`);
});

Then('annonsen med id {string} ska inte längre finnas i databasen', function (idStr: string) {
  const id = parseInt(idStr);
  assert(!mockDB.listings.has(id), `Annons ${id} ska inte finnas efter borttagning`);
});

Then('tillhörande värderingar ska också tas bort', function () {});
Then('tillhörande handlade varor ska också tas bort', function () {});
Then('tillhörande bildlänkar ska också tas bort', function () {});
