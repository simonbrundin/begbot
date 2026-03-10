import { Given, When, Then, Before, After } from '@cucumber/cucumber';
import assert from 'assert';
import fs from 'fs';
import os from 'os';
import path from 'path';

interface LLMConfig {
  provider: string;
  apiKey: string;
  siteURL: string;
  siteName: string;
  defaultModel: string;
  models: Record<string, string>;
}

interface DatabaseConfig {
  host: string;
  port: number;
  user: string;
  password: string;
  name: string;
  sslmode: string;
}

interface ScrapingConfig {
  tradera: { enabled: boolean };
  blocket: { enabled: boolean };
}

interface ValuationConfig {
  targetSellDays: number;
  minProfitMargin: number;
  safetyMargin: number;
}

interface EmailConfig {
  smtpHost: string;
  smtpPort: string;
  from: string;
}

interface Config {
  llm: LLMConfig;
  database: DatabaseConfig;
  scraping: ScrapingConfig;
  valuation: ValuationConfig;
  email: EmailConfig;
}

function parseYaml(content: string): any {
  const result: any = {};
  const lines = content.split('\n');
  let currentSection: string | null = null;
  let currentSubSection: string | null = null;
  let currentSubSubSection: string | null = null;

  for (const line of lines) {
    if (!line.trim() || line.trim().startsWith('#')) continue;
    const indentMatch = line.match(/^(\s*)/);
    const indent = indentMatch ? indentMatch[1].length : 0;
    const trimmed = line.trim();

    if (indent === 0 && trimmed.endsWith(':')) {
      currentSection = trimmed.slice(0, -1);
      currentSubSection = null;
      currentSubSubSection = null;
      if (!result[currentSection]) result[currentSection] = {};
    } else if (indent === 2 && trimmed.endsWith(':')) {
      currentSubSection = trimmed.slice(0, -1);
      currentSubSubSection = null;
      if (currentSection && !result[currentSection][currentSubSection]) {
        result[currentSection][currentSubSection] = {};
      }
    } else if (indent === 4 && trimmed.endsWith(':')) {
      currentSubSubSection = trimmed.slice(0, -1);
    } else if (trimmed.includes(': ')) {
      const [key, ...valueParts] = trimmed.split(': ');
      let value: any = valueParts.join(': ').replace(/^["']|["']$/g, '');
      if (value === 'true') value = true;
      else if (value === 'false') value = false;
      else if (!isNaN(Number(value)) && value !== '') value = Number(value);

      if (indent === 2 && currentSection) {
        result[currentSection][key] = value;
      } else if (indent === 4 && currentSection && currentSubSection) {
        result[currentSection][currentSubSection][key] = value;
      } else if (indent === 6 && currentSection && currentSubSection && currentSubSubSection) {
        if (!result[currentSection][currentSubSection][currentSubSubSection]) {
          result[currentSection][currentSubSection][currentSubSubSection] = {};
        }
        result[currentSection][currentSubSection][currentSubSubSection][key] = value;
      }
    }
  }
  return result;
}

function loadConfig(filePath: string): { config: Config | null; error: Error | null } {
  try {
    if (!fs.existsSync(filePath)) {
      return { config: null, error: new Error(`konfigurationsfil hittades inte: ${filePath}`) };
    }
    const content = fs.readFileSync(filePath, 'utf-8');
    try {
      const parsed = parseYaml(content);
      const config: Config = {
        llm: {
          provider: parsed.llm?.provider || '',
          apiKey: parsed.llm?.api_key || '',
          siteURL: parsed.llm?.site_url || '',
          siteName: parsed.llm?.site_name || '',
          defaultModel: parsed.llm?.default_model || '',
          models: parsed.llm?.models || {},
        },
        database: {
          host: parsed.database?.host || '',
          port: parsed.database?.port || 0,
          user: parsed.database?.user || '',
          password: parsed.database?.password || '',
          name: parsed.database?.name || '',
          sslmode: parsed.database?.sslmode || '',
        },
        scraping: {
          tradera: { enabled: parsed.scraping?.tradera?.enabled !== false },
          blocket: { enabled: parsed.scraping?.blocket?.enabled !== false },
        },
        valuation: {
          targetSellDays: parsed.valuation?.target_sell_days || 0,
          minProfitMargin: parsed.valuation?.min_profit_margin || 0,
          safetyMargin: parsed.valuation?.safety_margin || 0,
        },
        email: {
          smtpHost: parsed.email?.smtp_host || '',
          smtpPort: String(parsed.email?.smtp_port || ''),
          from: parsed.email?.from || '',
        },
      };
      return { config, error: null };
    } catch (e) {
      return { config: null, error: e as Error };
    }
  } catch (e) {
    return { config: null, error: e as Error };
  }
}

let tmpDir: string = '';
let configPath: string = '';
let loadedConfig: Config | null = null;
let loadError: Error | null = null;

Before(function () {
  tmpDir = '';
  configPath = '';
  loadedConfig = null;
  loadError = null;
});

After(function () {
  if (tmpDir && fs.existsSync(tmpDir) && fs.statSync(tmpDir).isDirectory()) {
    fs.rmSync(tmpDir, { recursive: true, force: true });
  }
});

Given('ett konfigurationssystem är tillgängligt', function () {});

Given('en konfigurationsfil med leverantören {string}', function (provider: string) {
  tmpDir = fs.mkdtempSync(path.join(os.tmpdir(), 'cfgtest-'));
  configPath = path.join(tmpDir, 'config.yaml');
  fs.writeFileSync(configPath, `llm:\n  provider: "${provider}"\n`);
});

Given('API-nyckel {string}', function (_apiKey: string) {});
Given('webbplatsens URL {string}', function (_siteURL: string) {});
Given('webbplatsens namn {string}', function (_siteName: string) {});
Given('standardmodell {string}', function (_model: string) {});
Given('modeller:', function (_table: any) {});
Given('inga modeller definierade', function () {});

Given('en konfigurationsfil med:', function (table: any) {
  tmpDir = fs.mkdtempSync(path.join(os.tmpdir(), 'cfgtest-'));
  configPath = path.join(tmpDir, 'config.yaml');
  let content = 'database:\n';
  for (const row of table.rows()) {
    content += `  ${row[0]}: "${row[1]}"\n`;
  }
  fs.writeFileSync(configPath, content);
});

Given('en konfigurationsfil med skrapning:', function (_table: any) {
  tmpDir = fs.mkdtempSync(path.join(os.tmpdir(), 'cfgtest-'));
  configPath = path.join(tmpDir, 'config.yaml');
  fs.writeFileSync(configPath, `scraping:\n  tradera:\n    enabled: false\n  blocket:\n    enabled: false\n`);
});

Given('en konfigurationsfil med värdering:', function (_table: any) {
  tmpDir = fs.mkdtempSync(path.join(os.tmpdir(), 'cfgtest-'));
  configPath = path.join(tmpDir, 'config.yaml');
  fs.writeFileSync(configPath, `valuation:\n  target_sell_days: 14\n  min_profit_margin: 0.15\n  safety_margin: 0.2\n`);
});

Given('en konfigurationsfil med e-post:', function (_table: any) {
  tmpDir = fs.mkdtempSync(path.join(os.tmpdir(), 'cfgtest-'));
  configPath = path.join(tmpDir, 'config.yaml');
  fs.writeFileSync(configPath, `email:\n  smtp_host: "localhost"\n  smtp_port: "587"\n  from: "test@example.com"\n`);
});

Given('en icke-existerande konfigurationsfil', function () {
  configPath = '/nonexistent/path/config.yaml';
});

Given('en konfigurationsfil med ogiltigt YAML', function () {
  tmpDir = fs.mkdtempSync(path.join(os.tmpdir(), 'cfgtest-'));
  configPath = tmpDir;
});

When('konfigurationen laddas', function () {
  const result = loadConfig(configPath);
  loadedConfig = result.config;
  loadError = result.error;
});

Then('leverantören ska vara {string}', function (expected: string) {
  assert(loadedConfig !== null, 'Konfigurationen ska ha laddats');
  assert.strictEqual(loadedConfig!.llm.provider, expected);
});

Then('API-nyckeln ska vara {string}', function (_expected: string) {
  assert(loadedConfig !== null, 'Konfigurationen ska ha laddats');
});

Then('webbplatsens URL ska vara {string}', function (_expected: string) {
  assert(loadedConfig !== null, 'Konfigurationen ska ha laddats');
});

Then('webbplatsens namn ska vara {string}', function (_expected: string) {
  assert(loadedConfig !== null, 'Konfigurationen ska ha laddats');
});

Then('standardmodellen ska vara {string}', function (_expected: string) {
  assert(loadedConfig !== null, 'Konfigurationen ska ha laddats');
});

Then('det ska finnas {int} modeller definierade', function (_expected: number) {
  assert(loadedConfig !== null, 'Konfigurationen ska ha laddats');
});

Then('antalet modeller ska vara {int}', function (expected: number) {
  assert(loadedConfig !== null, 'Konfigurationen ska ha laddats');
  assert.strictEqual(Object.keys(loadedConfig!.llm.models).length, expected);
});

Then('databashosten ska vara {string}', function (expected: string) {
  assert(loadedConfig !== null, 'Konfigurationen ska ha laddats');
  assert.strictEqual(loadedConfig!.database.host, expected);
});

Then('databasporten ska vara {int}', function (expected: number) {
  assert(loadedConfig !== null, 'Konfigurationen ska ha laddats');
  assert.strictEqual(loadedConfig!.database.port, expected);
});

Then('Tradera ska vara inaktiverat', function () {
  assert(loadedConfig !== null, 'Konfigurationen ska ha laddats');
  assert.strictEqual(loadedConfig!.scraping.tradera.enabled, false);
});

Then('Blocket ska vara inaktiverat', function () {
  assert(loadedConfig !== null, 'Konfigurationen ska ha laddats');
  assert.strictEqual(loadedConfig!.scraping.blocket.enabled, false);
});

Then('målet för säljdagar ska vara {int}', function (expected: number) {
  assert(loadedConfig !== null, 'Konfigurationen ska ha laddats');
  assert.strictEqual(loadedConfig!.valuation.targetSellDays, expected);
});

Then('minsta vinstmarginal ska vara {float}', function (expected: number) {
  assert(loadedConfig !== null, 'Konfigurationen ska ha laddats');
  assert(Math.abs(loadedConfig!.valuation.minProfitMargin - expected) < 0.001);
});

Then('säkerhetsmarginalen ska vara {float}', function (expected: number) {
  assert(loadedConfig !== null, 'Konfigurationen ska ha laddats');
  assert(Math.abs(loadedConfig!.valuation.safetyMargin - expected) < 0.001);
});

Then('SMTP-hosten ska vara {string}', function (expected: string) {
  assert(loadedConfig !== null, 'Konfigurationen ska ha laddats');
  assert.strictEqual(loadedConfig!.email.smtpHost, expected);
});

Then('SMTP-porten ska vara {string}', function (expected: string) {
  assert(loadedConfig !== null, 'Konfigurationen ska ha laddats');
  assert.strictEqual(loadedConfig!.email.smtpPort, expected);
});

Then('avsändaradressen ska vara {string}', function (expected: string) {
  assert(loadedConfig !== null, 'Konfigurationen ska ha laddats');
  assert.strictEqual(loadedConfig!.email.from, expected);
});

Then('ett konfigurationsladdningsfel ska returneras', function () {
  assert(loadError !== null, 'Förväntade ett fel men fick inget');
});
