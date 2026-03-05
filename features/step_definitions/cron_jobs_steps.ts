import { Given, When, Then, Before } from '@cucumber/cucumber';
import assert from 'assert';

interface CronJob {
  id: number;
  name: string;
  cronExpression: string;
  searchTermIDs: number[];
  isActive: boolean;
  createdAt?: Date;
  updatedAt?: Date;
}

function isValidCronExpression(expr: string): boolean {
  const parts = expr.split(' ');
  if (parts.length !== 5) return false;
  if (expr === 'invalid') return false;
  for (const part of parts) {
    if (!/^[\d*,\-\/]+$/.test(part)) return false;
  }
  return true;
}

class MockCronJobDB {
  jobs: CronJob[] = [];

  createJob(job: CronJob): { job: CronJob | null; error: Error | null } {
    if (!isValidCronExpression(job.cronExpression)) {
      return { job: null, error: new Error('invalid cron expression') };
    }
    const newJob = { ...job, id: this.jobs.length + 1, createdAt: new Date(), updatedAt: new Date() };
    this.jobs.push(newJob);
    return { job: newJob, error: null };
  }

  getAllJobs(): CronJob[] { return this.jobs; }

  getJobByID(id: number): CronJob | null {
    return this.jobs.find(j => j.id === id) || null;
  }

  updateJob(job: CronJob): { job: CronJob | null; error: Error | null } {
    if (!isValidCronExpression(job.cronExpression)) {
      return { job: null, error: new Error('invalid cron expression') };
    }
    const idx = this.jobs.findIndex(j => j.id === job.id);
    if (idx === -1) return { job: null, error: new Error('schemalagt jobb hittades inte') };
    this.jobs[idx] = { ...job, updatedAt: new Date() };
    return { job: this.jobs[idx], error: null };
  }

  deleteJob(id: number): Error | null {
    const idx = this.jobs.findIndex(j => j.id === id);
    if (idx === -1) return new Error('schemalagt jobb hittades inte');
    this.jobs.splice(idx, 1);
    return null;
  }
}

function parseSearchTermIDs(str: string): number[] {
  if (!str || str === '[]') return [];
  return (str.match(/\d+/g) || []).map(Number);
}

let db: MockCronJobDB;
let resultJob: CronJob | null;
let resultJobs: CronJob[];
let resultErr: Error | null;

Before(function () {
  db = new MockCronJobDB();
  resultJob = null;
  resultJobs = [];
  resultErr = null;
});

Given('en schemalagd-jobb-tjänst med mockdatabas', function () {
  db = new MockCronJobDB();
});

Given('databasen har schemalagda jobbposter', function (table: any) {
  for (const row of table.rows()) {
    const id = parseInt(row[0]);
    const isActive = row[4] === 'true';
    const job: CronJob = {
      id,
      name: row[1],
      cronExpression: row[2],
      searchTermIDs: parseSearchTermIDs(row[3]),
      isActive,
      createdAt: new Date(),
      updatedAt: new Date(),
    };
    db.jobs.push(job);
  }
});

When('jag skapar ett schemalagt jobb med namn {string}, uttryck {string}, sökterms-ID:n {string}, och aktivt {word}', function (name: string, expr: string, termIDs: string, activeStr: string) {
  const active = activeStr === 'true';
  const job: CronJob = { id: 0, name, cronExpression: expr, searchTermIDs: parseSearchTermIDs(termIDs), isActive: active };
  const result = db.createJob(job);
  resultJob = result.job;
  resultErr = result.error;
});

When('jag hämtar alla schemalagda jobb', function () {
  resultJobs = db.getAllJobs();
  resultJob = null;
  resultErr = null;
});

When('jag uppdaterar schemalagt jobb {int} till namn {string}, uttryck {string}, sökterms-ID:n {string}, och aktivt {word}', function (id: number, name: string, expr: string, termIDs: string, activeStr: string) {
  const active = activeStr === 'true';
  const job: CronJob = { id, name, cronExpression: expr, searchTermIDs: parseSearchTermIDs(termIDs), isActive: active };
  const result = db.updateJob(job);
  resultJob = result.job;
  resultErr = result.error;
});

When('jag tar bort schemalagt jobb {int}', function (id: number) {
  resultErr = db.deleteJob(id);
});

When('jag växlar aktivt tillstånd för schemalagt jobb {int}', function (id: number) {
  const job = db.getJobByID(id);
  if (!job) { resultErr = new Error('schemalagt jobb hittades inte'); return; }
  const updated = { ...job, isActive: !job.isActive };
  const result = db.updateJob(updated);
  resultJob = result.job;
  resultErr = result.error;
});

When('jag hämtar schemalagt jobb med ID {int}', function (id: number) {
  const job = db.getJobByID(id);
  if (job) { resultJobs = [job]; resultJob = job; }
  else { resultJobs = []; resultJob = null; }
  resultErr = null;
});

Then('det schemalagda jobbet ska sparas', function () {
  assert(resultErr === null, `Förväntade inget fel, fick: ${resultErr?.message}`);
});

Then('det schemalagda jobbet ska ha ett ID', function () {
  assert(resultJob !== null && resultJob.id > 0, 'Förväntade att ID är satt');
});

Then('jobbnamnet ska vara {string}', function (expected: string) {
  assert(resultJob !== null, 'Förväntade att jobbet finns');
  assert.strictEqual(resultJob!.name, expected);
});

Then('jobbuttrycket ska vara {string}', function (expected: string) {
  assert(resultJob !== null, 'Förväntade att jobbet finns');
  assert.strictEqual(resultJob!.cronExpression, expected);
});

Then('jobbet ska vara aktivt', function () {
  assert(resultJob !== null, 'Förväntade att jobbet finns');
  assert(resultJob!.isActive, 'Förväntade isActive att vara true');
});

Then('jobbet ska vara inaktivt', function () {
  assert(resultJob !== null, 'Förväntade att jobbet finns');
  assert(!resultJob!.isActive, 'Förväntade isActive att vara false');
});

Then('ska jag få {int} schemalagda jobbposter', function (count: number) {
  assert.strictEqual(resultJobs.length, count);
});

Then('det första schemalagda jobbet ska ha namn {string}', function (expected: string) {
  assert(resultJobs.length > 0, 'Förväntade minst ett jobb');
  assert.strictEqual(resultJobs[0].name, expected);
});

Then('det schemalagda jobbet ska uppdateras', function () {
  assert(resultErr === null, `Förväntade inget fel, fick: ${resultErr?.message}`);
});

Then('det schemalagda jobbet ska tas bort', function () {
  assert(resultErr === null, `Förväntade inget fel, fick: ${resultErr?.message}`);
});

Then('det ska finnas {int} schemalagda jobb i databasen', function (count: number) {
  assert.strictEqual(db.jobs.length, count);
});

Then('sökterms-ID:na ska vara tomma', function () {
  assert(resultJob !== null, 'Förväntade att jobbet finns');
  assert.strictEqual(resultJob!.searchTermIDs.length, 0);
});

Then('ett fel ska returneras', function () {
  assert(resultErr !== null, 'Förväntade ett fel men fick inget');
});

Then('felmeddelandet ska innehålla {string}', function (expected: string) {
  assert(resultErr !== null, 'Förväntade ett fel');
  assert(resultErr!.message.includes(expected), `Förväntade att felmeddelandet innehåller "${expected}", fick "${resultErr!.message}"`);
});

Then('ska jag få {int} schemalagd jobbpost', function (count: number) {
  assert.strictEqual(resultJobs.length, count);
});
