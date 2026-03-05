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

// Validate cron expression (simple validation)
function isValidCronExpression(expr: string): boolean {
  const parts = expr.split(' ');
  if (parts.length !== 5) return false;
  // Check for obviously invalid expressions
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

  getAllJobs(): CronJob[] {
    return this.jobs;
  }

  getJobByID(id: number): CronJob | null {
    return this.jobs.find(j => j.id === id) || null;
  }

  updateJob(job: CronJob): { job: CronJob | null; error: Error | null } {
    if (!isValidCronExpression(job.cronExpression)) {
      return { job: null, error: new Error('invalid cron expression') };
    }
    const idx = this.jobs.findIndex(j => j.id === job.id);
    if (idx === -1) return { job: null, error: new Error('cron job not found') };
    this.jobs[idx] = { ...job, updatedAt: new Date() };
    return { job: this.jobs[idx], error: null };
  }

  deleteJob(id: number): Error | null {
    const idx = this.jobs.findIndex(j => j.id === id);
    if (idx === -1) return new Error('cron job not found');
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

Given('a cron job service with mock database', function () {
  db = new MockCronJobDB();
});

Given('the database has cron job records', function (table: any) {
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

When('I create a cron job with name {string}, expression {string}, search term IDs {string}, and active {word}', function (name: string, expr: string, termIDs: string, activeStr: string) {
  const active = activeStr === 'true';
  const job: CronJob = {
    id: 0,
    name,
    cronExpression: expr,
    searchTermIDs: parseSearchTermIDs(termIDs),
    isActive: active,
  };
  const result = db.createJob(job);
  resultJob = result.job;
  resultErr = result.error;
});

When('I get all cron jobs', function () {
  resultJobs = db.getAllJobs();
  resultJob = null;
  resultErr = null;
});

When('I update cron job {int} to name {string}, expression {string}, search term IDs {string}, and active {word}', function (id: number, name: string, expr: string, termIDs: string, activeStr: string) {
  const active = activeStr === 'true' || activeStr === 'false' ? activeStr === 'true' : Boolean(activeStr);
  const job: CronJob = {
    id,
    name,
    cronExpression: expr,
    searchTermIDs: parseSearchTermIDs(termIDs),
    isActive: active,
  };
  const result = db.updateJob(job);
  resultJob = result.job;
  resultErr = result.error;
});

When('I delete cron job {int}', function (id: number) {
  resultErr = db.deleteJob(id);
});

When('I toggle cron job {int} active status', function (id: number) {
  const job = db.getJobByID(id);
  if (!job) {
    resultErr = new Error('cron job not found');
    return;
  }
  const updated = { ...job, isActive: !job.isActive };
  const result = db.updateJob(updated);
  resultJob = result.job;
  resultErr = result.error;
});

When('I get cron job by ID {int}', function (id: number) {
  const job = db.getJobByID(id);
  if (job) {
    resultJobs = [job];
    resultJob = job;
  } else {
    resultJobs = [];
    resultJob = null;
  }
  resultErr = null;
});

Then('the cron job should be saved successfully', function () {
  assert(resultErr === null, `Expected no error, got: ${resultErr?.message}`);
});

Then('the cron job should have ID set', function () {
  assert(resultJob !== null && resultJob.id > 0, 'Expected ID to be set');
});

Then('the cron job name should be {string}', function (expected: string) {
  assert(resultJob !== null, 'Expected job to exist');
  assert.strictEqual(resultJob!.name, expected);
});

Then('the cron job expression should be {string}', function (expected: string) {
  assert(resultJob !== null, 'Expected job to exist');
  assert.strictEqual(resultJob!.cronExpression, expected);
});

Then('the cron job should be active', function () {
  assert(resultJob !== null, 'Expected job to exist');
  assert(resultJob!.isActive, 'Expected isActive to be true');
});

Then('the cron job should be inactive', function () {
  assert(resultJob !== null, 'Expected job to exist');
  assert(!resultJob!.isActive, 'Expected isActive to be false');
});

Then('I should receive {int} cron job records', function (count: number) {
  assert.strictEqual(resultJobs.length, count);
});

Then('the first cron job should have name {string}', function (expected: string) {
  assert(resultJobs.length > 0, 'Expected at least one job');
  assert.strictEqual(resultJobs[0].name, expected);
});

Then('the cron job should be updated successfully', function () {
  assert(resultErr === null, `Expected no error, got: ${resultErr?.message}`);
});

Then('the cron job should be deleted successfully', function () {
  assert(resultErr === null, `Expected no error, got: ${resultErr?.message}`);
});

Then('there should be {int} cron jobs in the database', function (count: number) {
  assert.strictEqual(db.jobs.length, count);
});

Then('the cron job should have empty search term IDs', function () {
  assert(resultJob !== null, 'Expected job to exist');
  assert.strictEqual(resultJob!.searchTermIDs.length, 0);
});

Then('the error message should contain {string}', function (expected: string) {
  assert(resultErr !== null, 'Expected an error');
  assert(resultErr!.message.includes(expected), `Expected error to contain "${expected}", got "${resultErr!.message}"`);
});

Then('an error should be returned', function () {
  assert(resultErr !== null, 'Expected an error but got none');
});

Then('I should receive {int} cron job record', function (count: number) {
  assert.strictEqual(resultJobs.length, count);
});
