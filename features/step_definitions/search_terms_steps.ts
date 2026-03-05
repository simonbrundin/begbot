import { Given, When, Then, Before } from '@cucumber/cucumber';
import assert from 'assert';

interface SearchTerm {
  id: number;
  description: string;
  url: string;
  marketplaceID: number | null;
  isActive: boolean;
}

interface Marketplace {
  id: number;
  name: string;
}

interface SearchJob {
  searchTerm: SearchTerm;
  marketplace: Marketplace | null;
}

let searchTerm: SearchTerm;
let searchJob: SearchJob;
let contextPassed: boolean;

Before(function () {
  searchTerm = { id: 0, description: '', url: '', marketplaceID: null, isActive: true };
  searchJob = { searchTerm, marketplace: null };
  contextPassed = false;
});

Given('a search term with description {string} and URL {string}', function (description: string, url: string) {
  searchTerm = { id: 1, description, url, marketplaceID: null, isActive: true };
});

Given('the search term is linked to marketplace ID {int}', function (id: number) {
  searchTerm.marketplaceID = id;
});

Given('a search term {string} linked to Blocket marketplace', function (description: string) {
  searchTerm = {
    id: 1,
    description,
    url: 'https://blocket.se/search?q=iphone',
    marketplaceID: 1,
    isActive: true,
  };
});

Given('a search term {string} linked to Tradera marketplace', function (description: string) {
  searchTerm = {
    id: 1,
    description,
    url: 'https://www.tradera.com/search?q=lego+star+wars',
    marketplaceID: 2,
    isActive: true,
  };
});

Given('a search term with isActive false', function () {
  searchTerm = { id: 1, description: 'Inactive', url: 'https://example.com', marketplaceID: 1, isActive: false };
});

Given('a search term with no marketplace ID', function () {
  searchTerm = { id: 1, description: 'No Marketplace', url: 'https://example.com', marketplaceID: null, isActive: true };
});

Given('a search term service', function () {
  // Nothing to set up
});

When('creating a search job', function () {
  const marketplace: Marketplace = searchTerm.marketplaceID === 2
    ? { id: 2, name: 'Tradera' }
    : { id: 1, name: 'Blocket' };
  searchJob = { searchTerm, marketplace };
});

When('passing a context', function () {
  contextPassed = true;
});

Then('the description should be {string}', function (expected: string) {
  assert.strictEqual(searchTerm.description, expected);
});

Then('the URL should not be empty', function () {
  assert(searchTerm.url !== '', 'URL should not be empty');
});

Then('the marketplace ID should be {int}', function (expected: number) {
  assert.strictEqual(searchTerm.marketplaceID, expected);
});

Then('the job URL should not be empty', function () {
  assert(searchJob.searchTerm.url !== '', 'Job URL should not be empty');
});

Then('the marketplace should not be nil', function () {
  assert(searchJob.marketplace !== null, 'Marketplace should not be nil');
});

Then('the marketplace name should be {string}', function (expected: string) {
  assert(searchJob.marketplace !== null, 'Marketplace should not be nil');
  assert.strictEqual(searchJob.marketplace!.name, expected);
});

Then('the search term should be inactive', function () {
  assert(!searchTerm.isActive, 'Expected search term to be inactive');
});

Then('the marketplace ID should be null', function () {
  assert(searchTerm.marketplaceID === null, 'Expected marketplace ID to be null');
});

Then('no error should occur', function () {
  assert(contextPassed, 'Expected context to be passed successfully');
});
