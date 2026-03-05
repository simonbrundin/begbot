import { Given, When, Then, Before } from '@cucumber/cucumber';
import assert from 'assert';

interface ListingWithDetails {
  id: number;
  potentialProfit?: number;
  discountPercent?: number;
}

function filterListingsByTab(listings: ListingWithDetails[], tab: string): ListingWithDetails[] {
  if (tab === 'potential') {
    return listings.filter(l =>
      l.potentialProfit !== undefined &&
      l.potentialProfit > 0 &&
      l.discountPercent !== undefined &&
      l.discountPercent > 0
    );
  }
  return listings;
}

let frontendListings: ListingWithDetails[];
let frontendFiltered: ListingWithDetails[];

Before(function () {
  frontendListings = [];
  frontendFiltered = [];
});

Given('frontend listings exist with various prices and valuations', function () {
  frontendListings = [
    { id: 1, potentialProfit: 200, discountPercent: 20 },
    { id: 2, potentialProfit: -100, discountPercent: -10 },
    { id: 3 },
  ];
});

Given('no frontend listings exist', function () {
  frontendListings = [];
});

Given('frontend listings with profit data:', function (table: any) {
  frontendListings = [];
  for (const row of table.rows()) {
    frontendListings.push({
      id: parseInt(row[0]),
      potentialProfit: parseInt(row[1]),
      discountPercent: parseInt(row[2]),
    });
  }
});

Given('a frontend listing with id {int} and no potential profit data', function (id: number) {
  frontendListings = [{ id }];
});

When('I apply the frontend filter {string}', function (tab: string) {
  frontendFiltered = filterListingsByTab(frontendListings, tab);
});

Then('all frontend listings should be returned', function () {
  assert.strictEqual(frontendFiltered.length, frontendListings.length);
});

Then('{int} frontend listings should be returned', function (expected: number) {
  assert.strictEqual(frontendFiltered.length, expected, `Expected ${expected} listings, got ${frontendFiltered.length}`);
});

Then('the returned frontend listing should have id {int}', function (expectedId: number) {
  assert(frontendFiltered.length > 0, 'Expected at least one listing');
  assert.strictEqual(frontendFiltered[0].id, expectedId);
});

Then('{int} frontend listing should be returned', function (expected: number) {
  assert.strictEqual(frontendFiltered.length, expected, `Expected ${expected} listing, got ${frontendFiltered.length}`);
});
