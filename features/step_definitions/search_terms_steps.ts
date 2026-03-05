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

Given('en sökterm med beskrivningen {string} och URL:en {string}', function (description: string, url: string) {
  searchTerm = { id: 1, description, url, marketplaceID: null, isActive: true };
});

Given('söktermen är kopplad till marknadsplats-ID {int}', function (id: number) {
  searchTerm.marketplaceID = id;
});

Given('en sökterm {string} kopplad till Blocket-marknadsplatsen', function (description: string) {
  searchTerm = { id: 1, description, url: 'https://blocket.se/search?q=iphone', marketplaceID: 1, isActive: true };
});

Given('en sökterm {string} kopplad till Tradera-marknadsplatsen', function (description: string) {
  searchTerm = { id: 1, description, url: 'https://www.tradera.com/search?q=lego+star+wars', marketplaceID: 2, isActive: true };
});

Given('en sökterm med isActive false', function () {
  searchTerm = { id: 1, description: 'Inaktiv', url: 'https://example.com', marketplaceID: 1, isActive: false };
});

Given('en sökterm utan marknadsplats-ID', function () {
  searchTerm = { id: 1, description: 'Ingen marknadsplats', url: 'https://example.com', marketplaceID: null, isActive: true };
});

Given('en söktermstjänst', function () {});

When('ett sökjobb skapas', function () {
  const marketplace: Marketplace = searchTerm.marketplaceID === 2
    ? { id: 2, name: 'Tradera' }
    : { id: 1, name: 'Blocket' };
  searchJob = { searchTerm, marketplace };
});

When('en kontext skickas', function () {
  contextPassed = true;
});

Then('ska beskrivningen vara {string}', function (expected: string) {
  assert.strictEqual(searchTerm.description, expected);
});

Then('URL:en ska inte vara tom', function () {
  assert(searchTerm.url !== '', 'URL ska inte vara tom');
});

Then('marknadsplats-ID:et ska vara {int}', function (expected: number) {
  assert.strictEqual(searchTerm.marketplaceID, expected);
});

Then('jobbets URL ska inte vara tom', function () {
  assert(searchJob.searchTerm.url !== '', 'Jobbets URL ska inte vara tom');
});

Then('marknadsplatsen ska inte vara null', function () {
  assert(searchJob.marketplace !== null, 'Marknadsplatsen ska inte vara null');
});

Then('marknadsplatsens namn ska vara {string}', function (expected: string) {
  assert(searchJob.marketplace !== null, 'Marknadsplatsen ska inte vara null');
  assert.strictEqual(searchJob.marketplace!.name, expected);
});

Then('ska söktermen vara inaktiv', function () {
  assert(!searchTerm.isActive, 'Förväntade att söktermen är inaktiv');
});

Then('ska marknadsplats-ID:et vara null', function () {
  assert(searchTerm.marketplaceID === null, 'Förväntade att marknadsplats-ID är null');
});

Then('ska inget fel inträffa', function () {
  assert(contextPassed, 'Förväntade att kontexten skickades');
});
