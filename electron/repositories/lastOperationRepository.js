const fs = require('fs');
const os = require('os');
const path = require('path');

let lastOperation = null;
let wasLoaded = false;

const cacheDir = path.join(os.homedir(), '.sortly');
const cacheFile = path.join(cacheDir, 'last-operation.json');

function loadFromDiskIfNeeded() {
  if (wasLoaded) {
    return;
  }

  wasLoaded = true;

  try {
    if (!fs.existsSync(cacheFile)) {
      return;
    }

    const raw = fs.readFileSync(cacheFile, 'utf8');
    lastOperation = JSON.parse(raw);
  } catch {
    lastOperation = null;
  }
}

function persistToDisk() {
  try {
    if (!fs.existsSync(cacheDir)) {
      fs.mkdirSync(cacheDir, { recursive: true });
    }

    fs.writeFileSync(cacheFile, JSON.stringify(lastOperation), 'utf8');
  } catch {
    // Ignora falhas de persistencia para nao bloquear fluxo principal.
  }
}

function getLastOperation() {
  loadFromDiskIfNeeded();
  return lastOperation;
}

function setLastOperation(operation) {
  loadFromDiskIfNeeded();
  lastOperation = operation;
  persistToDisk();
}

function clearLastOperation() {
  loadFromDiskIfNeeded();
  lastOperation = null;

  try {
    if (fs.existsSync(cacheFile)) {
      fs.unlinkSync(cacheFile);
    }
  } catch {
    // Ignora falhas ao limpar cache em disco.
  }
}

module.exports = {
  getLastOperation,
  setLastOperation,
  clearLastOperation
};