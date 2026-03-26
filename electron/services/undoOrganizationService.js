const { getUniqueDestination } = require('../models/pathModel');

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

  for (const folderPath of lastOperation.createdFolders) {
    await fsAdapter.removeDirectoryIfEmpty(folderPath);
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