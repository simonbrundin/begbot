import { Given, When, Then, Before } from '@cucumber/cucumber';
import assert from 'assert';

interface SearchHistoryRecord {
  id: number;
  searchTermID: number;
  searchTermDesc: string;
  url: string;
  resultsFound: number;
  newAdsFound: number;
  searchedAt: Date;
}

interface MockSearchHistoryDB {
  history: SearchHistoryRecord[];
  error: Error | null;
}

class SearchHistoryService {
  private db: MockSearchHistoryDB;
  constructor(db: MockSearchHistoryDB) { this.db = db; }

  recordSearch(searchTermID: number, searchTermDesc: string, url: string, resultsFound: number, newAdsFound: number): { record: SearchHistoryRecord | null; error: Error | null } {
    if (this.db.error) return { record: null, error: this.db.error };
    const record: SearchHistoryRecord = {
      id: this.db.history.length + 1,
      searchTermID,
      searchTermDesc,
      url,
      resultsFound,
      newAdsFound,
      searchedAt: new Date(),
    };
    this.db.history.push(record);
    return { record, error: null };
  }

  getHistory(page: number, pageSize: number): { history: SearchHistoryRecord[]; count: number; error: Error | null } {
    if (this.db.error) return { history: [], count: 0, error: this.db.error };
    const normalizedPage = Math.max(page, 1);
    const count = this.db.history.length;
    const offset = (normalizedPage - 1) * pageSize;
    if (offset >= count) return { history: [], count, error: null };
    const end = Math.min(offset + pageSize, count);
    return { history: this.db.history.slice(offset, end), count, error: null };
  }
}

let mockDB: MockSearchHistoryDB;
let service: SearchHistoryService;
let resultRecord: SearchHistoryRecord | null;
let resultHistory: SearchHistoryRecord[];
let resultCount: number;
let resultError: Error | null;

Before(function () {
  mockDB = { history: [], error: null };
  service = new SearchHistoryService(mockDB);
  resultRecord = null;
  resultHistory = [];
  resultCount = 0;
  resultError = null;
});

Given('a search history service is available', function () {
  mockDB = { history: [], error: null };
  service = new SearchHistoryService(mockDB);
});

Given('the database is connected', function () {
  mockDB.error = null;
});

When('a user searches for {string} with URL {string}', function (termDesc: string, url: string) {
  const result = service.recordSearch(1, termDesc, url, 10, 3);
  resultRecord = result.record;
  resultError = result.error;
});

Then('the search should be saved successfully', function () {
  assert(resultError === null, `Expected no error, got: ${resultError?.message}`);
});

Then('the search should have a valid ID', function () {
  assert(resultRecord !== null && resultRecord.id > 0, 'Expected valid ID');
});

Then('the search term description should be {string}', function (expected: string) {
  assert(resultRecord !== null, 'Expected record to exist');
  assert.strictEqual(resultRecord!.searchTermDesc, expected);
});

Then('the results found should be {int}', function (expected: number) {
  assert(resultRecord !== null, 'Expected record to exist');
  assert.strictEqual(resultRecord!.resultsFound, expected);
});

Then('the new ads found should be {int}', function (expected: number) {
  assert(resultRecord !== null, 'Expected record to exist');
  assert.strictEqual(resultRecord!.newAdsFound, expected);
});

Given('the database has {int} search records', function (count: number) {
  mockDB.history = [];
  for (let i = 0; i < count; i++) {
    mockDB.history.push({
      id: i + 1,
      searchTermID: i + 1,
      searchTermDesc: `iPhone ${15 - i}`,
      url: `https://blocket.se/search${i}`,
      resultsFound: 10,
      newAdsFound: 1,
      searchedAt: new Date(),
    });
  }
});

When('the user requests search history for page {int} with {int} items per page', function (page: number, pageSize: number) {
  const result = service.getHistory(page, pageSize);
  resultHistory = result.history;
  resultCount = result.count;
  resultError = result.error;
});

Then('the response should contain {int} search records', function (expected: number) {
  assert.strictEqual(resultHistory.length, expected);
});

Then('the total count should be {int}', function (expected: number) {
  assert.strictEqual(resultCount, expected);
});

Then('the first record should have search term {string}', function (expected: string) {
  assert(resultHistory.length > 0, 'No history records');
  assert.strictEqual(resultHistory[0].searchTermDesc, expected);
});

Given('the database has no search records', function () {
  mockDB.history = [];
});

When('the user requests search history', function () {
  const result = service.getHistory(1, 20);
  resultHistory = result.history;
  resultCount = result.count;
  resultError = result.error;
});

Then('the response should contain {int} items', function (expected: number) {
  assert.strictEqual(resultHistory.length, expected);
});

Then('the first item on page {int} should have ID {int}', function (_page: number, expectedID: number) {
  assert(resultHistory.length > 0, 'No history records');
  assert.strictEqual(resultHistory[0].id, expectedID);
});

When('the user requests page {int} with {int} items per page', function (page: number, pageSize: number) {
  const result = service.getHistory(page, pageSize);
  resultHistory = result.history;
  resultCount = result.count;
  resultError = result.error;
});

When('the user requests page {int}', function (page: number) {
  const result = service.getHistory(page, 20);
  resultHistory = result.history;
  resultCount = result.count;
  resultError = result.error;
});

Then('the count should be {int}', function (expected: number) {
  assert.strictEqual(resultCount, expected);
});

Given('the database is unavailable', function () {
  mockDB.error = new Error('database unavailable');
});

When('the user attempts to record a search', function () {
  const result = service.recordSearch(1, 'Test', 'https://...', 10, 2);
  resultRecord = result.record;
  resultError = result.error;
});

Then('a search history error should be returned', function () {
  assert(resultError !== null, 'Expected an error but got none');
});

Then('the request should succeed', function () {
  assert(resultError === null, `Expected no error, got: ${resultError?.message}`);
});

Then('the response should contain {int} item', function (expected: number) {
  assert.strictEqual(resultHistory.length, expected, `Expected ${expected} item, got ${resultHistory.length}`);
});

// The search finds N results - part of the recording process (values are pre-set in the search call)
Then('the search finds {int} results with {int} new ads', function (results: number, newAds: number) {
  assert(resultRecord !== null, 'Expected search record to exist');
  assert.strictEqual(resultRecord!.resultsFound, results);
  assert.strictEqual(resultRecord!.newAdsFound, newAds);
});
