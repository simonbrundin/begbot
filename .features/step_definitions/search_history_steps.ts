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
      id: this.db.history.length + 1, searchTermID, searchTermDesc, url, resultsFound, newAdsFound, searchedAt: new Date(),
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
    return { history: this.db.history.slice(offset, Math.min(offset + pageSize, count)), count, error: null };
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

Given('en sökhistoriktjänst är tillgänglig', function () {
  mockDB = { history: [], error: null };
  service = new SearchHistoryService(mockDB);
});

Given('databasen är ansluten', function () {
  mockDB.error = null;
});

When('en användare söker efter {string} med URL:en {string}', function (termDesc: string, url: string) {
  const result = service.recordSearch(1, termDesc, url, 10, 3);
  resultRecord = result.record;
  resultError = result.error;
});

Then('ska sökningen sparas', function () {
  assert(resultError === null, `Förväntade inget fel, fick: ${resultError?.message}`);
});

Then('sökningen ska ha ett giltigt ID', function () {
  assert(resultRecord !== null && resultRecord.id > 0, 'Förväntade giltigt ID');
});

Then('sökbeskrivningen ska vara {string}', function (expected: string) {
  assert(resultRecord !== null, 'Förväntade att posten finns');
  assert.strictEqual(resultRecord!.searchTermDesc, expected);
});

Then('antalet hittade resultat ska vara {int}', function (expected: number) {
  assert(resultRecord !== null, 'Förväntade att posten finns');
  assert.strictEqual(resultRecord!.resultsFound, expected);
});

Then('antalet nya annonser ska vara {int}', function (expected: number) {
  assert(resultRecord !== null, 'Förväntade att posten finns');
  assert.strictEqual(resultRecord!.newAdsFound, expected);
});

Given('databasen har {int} sökposter', function (count: number) {
  mockDB.history = [];
  for (let i = 0; i < count; i++) {
    mockDB.history.push({
      id: i + 1, searchTermID: i + 1, searchTermDesc: `iPhone ${15 - i}`,
      url: `https://blocket.se/search${i}`, resultsFound: 10, newAdsFound: 1, searchedAt: new Date(),
    });
  }
});

When('användaren begär sökhistorik för sida {int} med {int} poster per sida', function (page: number, pageSize: number) {
  const result = service.getHistory(page, pageSize);
  resultHistory = result.history;
  resultCount = result.count;
  resultError = result.error;
});

Then('ska svaret innehålla {int} sökposter', function (expected: number) {
  assert.strictEqual(resultHistory.length, expected);
});

Then('det totala antalet ska vara {int}', function (expected: number) {
  assert.strictEqual(resultCount, expected);
});

Then('den första posten ska ha söktermen {string}', function (expected: string) {
  assert(resultHistory.length > 0, 'Inga historikposter');
  assert.strictEqual(resultHistory[0].searchTermDesc, expected);
});

Given('databasen har inga sökposter', function () {
  mockDB.history = [];
});

When('användaren begär sökhistorik', function () {
  const result = service.getHistory(1, 20);
  resultHistory = result.history;
  resultCount = result.count;
  resultError = result.error;
});

Then('ska svaret innehålla {int} poster', function (expected: number) {
  assert.strictEqual(resultHistory.length, expected);
});

Then('den första posten på sida {int} ska ha ID {int}', function (_page: number, expectedID: number) {
  assert(resultHistory.length > 0, 'Inga historikposter');
  assert.strictEqual(resultHistory[0].id, expectedID);
});

When('användaren begär sida {int} med {int} poster per sida', function (page: number, pageSize: number) {
  const result = service.getHistory(page, pageSize);
  resultHistory = result.history;
  resultCount = result.count;
  resultError = result.error;
});

When('användaren begär sida {int}', function (page: number) {
  const result = service.getHistory(page, 20);
  resultHistory = result.history;
  resultCount = result.count;
  resultError = result.error;
});

Then('antalet ska vara {int}', function (expected: number) {
  assert.strictEqual(resultCount, expected);
});

Given('databasen är otillgänglig', function () {
  mockDB.error = new Error('databasen är otillgänglig');
});

When('användaren försöker registrera en sökning', function () {
  const result = service.recordSearch(1, 'Test', 'https://...', 10, 2);
  resultRecord = result.record;
  resultError = result.error;
});

Then('ett sökhistorikfel ska returneras', function () {
  assert(resultError !== null, 'Förväntade ett fel men fick inget');
});

Then('ska förfrågan lyckas', function () {
  assert(resultError === null, `Förväntade inget fel, fick: ${resultError?.message}`);
});

Then('ska svaret innehålla {int} post', function (expected: number) {
  assert.strictEqual(resultHistory.length, expected, `Förväntade ${expected} post, fick ${resultHistory.length}`);
});

Then('sökningen hittar {int} resultat med {int} nya annonser', function (results: number, newAds: number) {
  assert(resultRecord !== null, 'Förväntade att sökposten finns');
  assert.strictEqual(resultRecord!.resultsFound, results);
  assert.strictEqual(resultRecord!.newAdsFound, newAds);
});
