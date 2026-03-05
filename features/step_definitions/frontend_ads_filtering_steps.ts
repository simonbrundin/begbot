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
      l.potentialProfit !== undefined && l.potentialProfit > 0 &&
      l.discountPercent !== undefined && l.discountPercent > 0
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

Given('frontend-annonser finns med varierande priser och värderingar', function () {
  frontendListings = [
    { id: 1, potentialProfit: 200, discountPercent: 20 },
    { id: 2, potentialProfit: -100, discountPercent: -10 },
    { id: 3 },
  ];
});

Given('inga frontend-annonser finns', function () {
  frontendListings = [];
});

Given('frontend-annonser med vinstdata:', function (table: any) {
  frontendListings = [];
  for (const row of table.rows()) {
    frontendListings.push({
      id: parseInt(row[0]),
      potentialProfit: parseInt(row[1]),
      discountPercent: parseInt(row[2]),
    });
  }
});

Given('en frontend-annons med id {int} och ingen vinstdata', function (id: number) {
  frontendListings = [{ id }];
});

When('jag applicerar frontend-filtret {string}', function (tab: string) {
  frontendFiltered = filterListingsByTab(frontendListings, tab);
});

Then('ska alla frontend-annonser returneras', function () {
  assert.strictEqual(frontendFiltered.length, frontendListings.length);
});

Then('ska {int} frontend-annonser returneras', function (expected: number) {
  assert.strictEqual(frontendFiltered.length, expected, `Förväntade ${expected} annonser, fick ${frontendFiltered.length}`);
});

Then('den returnerade frontend-annonsen ska ha id {int}', function (expectedId: number) {
  assert(frontendFiltered.length > 0, 'Förväntade minst en annons');
  assert.strictEqual(frontendFiltered[0].id, expectedId);
});

Then('ska {int} frontend-annons returneras', function (expected: number) {
  assert.strictEqual(frontendFiltered.length, expected, `Förväntade ${expected} annons, fick ${frontendFiltered.length}`);
});
