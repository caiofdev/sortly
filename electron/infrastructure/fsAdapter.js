const fs = require('fs/promises');

async function readDirectoryWithTypes(directoryPath) {
  return fs.readdir(directoryPath, { withFileTypes: true });
}

async function createDirectory(directoryPath) {
  await fs.mkdir(directoryPath, { recursive: true });
}

async function moveFile(sourcePath, destinationPath) {
  await fs.rename(sourcePath, destinationPath);
}

async function removeDirectoryIfEmpty(directoryPath) {
  try {
    await fs.rmdir(directoryPath);
  } catch {
    // Ignora pastas não vazias ou já removidas.
  }
}

async function pathExists(filePath) {
  try {
    await fs.access(filePath);
    return true;
  } catch {
    return false;
  }
}

module.exports = {
  readDirectoryWithTypes,
  createDirectory,
  moveFile,
  removeDirectoryIfEmpty,
  pathExists
};