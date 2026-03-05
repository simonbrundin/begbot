import { Given, When, Then, Before } from '@cucumber/cucumber';
import assert from 'assert';

// State for ads filtering
interface ListingItem {
  id: number;
  price: number | null;
  valuation: number | null;
}

let listings: ListingItem[] = [];
let filtered: ListingItem[] = [];

function filterListings(items: ListingItem[], tab: string): ListingItem[] {
  if (tab === 'all') return items;
  if (tab === 'good-value') {
    return items.filter(l => l.price !== null && l.valuation !== null && l.price < l.valuation);
  }
  return items;
}

Before(function () {
  listings = [];
  filtered = [];
});

Given('I have a listing filter service', function () {
  listings = [];
  filtered = [];
});

Given('the following listings:', function (table: any) {
  listings = [];
  for (const row of table.rows()) {
    const id = parseInt(row[0]);
    const price = row[1] === '' ? null : parseInt(row[1]);
    const valuation = row[2] === '' ? null : parseInt(row[2]);
    listings.push({ id, price, valuation });
  }
});

Given('there are no listings', function () {
  listings = [];
});

When('I filter by {string} tab', function (tab: string) {
  filtered = filterListings(listings, tab);
});

Then('I should receive {int} listings', function (count: number) {
  assert.strictEqual(filtered.length, count, `Expected ${count} listings, got ${filtered.length}`);
});
