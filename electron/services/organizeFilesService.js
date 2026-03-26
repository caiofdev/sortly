const path = require('path');
const { getUniqueDestination } = require('../models/pathModel');

async function organizeFiles(payload, dependencies) {
  const { fsAdapter, lastOperationRepository } = dependencies;

  const sourceFolderPath = payload?.sourceFolderPath;
  const destinationFolderPath = payload?.destinationFolderPath || sourceFolderPath;

  if (!sourceFolderPath || typeof sourceFolderPath !== 'string') {
    throw new Error('Pasta inválida.');
  }

  if (!destinationFolderPath || typeof destinationFolderPath !== 'string') {
    throw new Error('Pasta de destino inválida.');
  }

  const entries = await fsAdapter.readDirectoryWithTypes(sourceFolderPath);
  let processedFiles = 0;
  let movedFiles = 0;
  let ignoredWithoutExtension = 0;
  let ignoredFolders = 0;
  const movedItems = [];
  const createdFolders = new Set();

  for (const entry of entries) {
    if (entry.isDirectory()) {
      ignoredFolders += 1;
      continue;
    }

    if (!entry.isFile()) {
      continue;
    }

    processedFiles += 1;

    const extension = path.extname(entry.name).slice(1).toLowerCase();
    if (!extension) {
      ignoredWithoutExtension += 1;
      continue;
    }

    const extensionFolderPath = path.join(destinationFolderPath, extension);
    await fsAdapter.createDirectory(extensionFolderPath);
    createdFolders.add(extensionFolderPath);

    const sourcePath = path.join(sourceFolderPath, entry.name);
    const destinationPath = path.join(extensionFolderPath, entry.name);
    const safeDestinationPath = await getUniqueDestination(destinationPath, fsAdapter.pathExists);

    await fsAdapter.moveFile(sourcePath, safeDestinationPath);
    movedItems.push({ from: sourcePath, to: safeDestinationPath });
    movedFiles += 1;
  }

  lastOperationRepository.setLastOperation({
    sourceFolderPath,
    destinationFolderPath,
    movedItems,
    createdFolders: [...createdFolders]
  });

  return {
    sourceFolderPath,
    destinationFolderPath,
    processedFiles,
    movedFiles,
    ignoredWithoutExtension,
    ignoredFolders,
    message: `Organização concluída: ${movedFiles} arquivo(s) movido(s).`,
    canUndo: movedItems.length > 0
  };
}

module.exports = {
  organizeFiles
};