const path = require('path');
const { getUniqueDestination } = require('../models/pathModel');

function normalizePathForCompare(filePath) {
  const resolvedPath = path.resolve(filePath);
  return process.platform === 'win32' ? resolvedPath.toLowerCase() : resolvedPath;
}

function isPathInsideOrEqual(childPath, rootPath) {
  const normalizedChild = normalizePathForCompare(childPath);
  const normalizedRoot = normalizePathForCompare(rootPath);

  return (
    normalizedChild === normalizedRoot ||
    normalizedChild.startsWith(`${normalizedRoot}${path.sep}`)
  );
}

async function removeEmptyAncestors(startPath, stopAtPath, fsAdapter) {
  let currentPath = path.resolve(startPath);
  const stopPath = path.resolve(stopAtPath);

  while (isPathInsideOrEqual(currentPath, stopPath) && currentPath !== stopPath) {
    await fsAdapter.removeDirectoryIfEmpty(currentPath);

    const parentPath = path.dirname(currentPath);
    if (parentPath === currentPath) {
      break;
    }

    currentPath = parentPath;
  }
}

async function undoLastOrganization(dependencies) {
  const { fsAdapter, lastOperationRepository } = dependencies;
  const lastOperation = lastOperationRepository.getLastOperation();

  if (!lastOperation || lastOperation.movedItems.length === 0) {
    throw new Error('Nenhuma separação recente para desfazer.');
  }

  let restoredFiles = 0;
  let renamedOnRestore = 0;
  let skippedMissing = 0;

  for (const move of [...lastOperation.movedItems].reverse()) {
    if (!(await fsAdapter.pathExists(move.to))) {
      skippedMissing += 1;
      continue;
    }

    const safeRestorePath = await getUniqueDestination(move.from, fsAdapter.pathExists);
    await fsAdapter.moveFile(move.to, safeRestorePath);

    if (safeRestorePath !== move.from) {
      renamedOnRestore += 1;
    }

    restoredFiles += 1;
  }

  const createdFolders = Array.isArray(lastOperation.createdFolders)
    ? lastOperation.createdFolders
    : [];

  const cleanupCandidates = new Set(createdFolders);
  for (const move of lastOperation.movedItems) {
    cleanupCandidates.add(path.dirname(move.to));
  }

  const cleanupRoot = lastOperation.destinationFolderPath || lastOperation.sourceFolderPath;

  for (const folderPath of cleanupCandidates) {
    if (!cleanupRoot) {
      await fsAdapter.removeDirectoryIfEmpty(folderPath);
      continue;
    }

    await removeEmptyAncestors(folderPath, cleanupRoot, fsAdapter);
  }

  lastOperationRepository.clearLastOperation();

  return {
    restoredFiles,
    renamedOnRestore,
    skippedMissing,
    message: `Desfazer concluído: ${restoredFiles} arquivo(s) restaurado(s).`,
    canUndo: false
  };
}

module.exports = {
  undoLastOrganization
};