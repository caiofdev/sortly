const fsAdapter = require('../infrastructure/fsAdapter');
const lastOperationRepository = require('../repositories/lastOperationRepository');
const { organizeFiles } = require('../services/organizeFilesService');
const { undoLastOrganization } = require('../services/undoOrganizationService');
const { resolveDroppedPath } = require('../services/resolveDroppedPathService');

function registerFilesHandlers(ipcMain) {
  ipcMain.handle('files:get-last-organization-state', async () => {
    const lastOperation = lastOperationRepository.getLastOperation();

    const hasUndo = Boolean(lastOperation?.movedItems?.length);

    return {
      hasUndo,
      sourceFolderPath: hasUndo ? lastOperation.sourceFolderPath || '' : '',
      destinationFolderPath: hasUndo ? lastOperation.destinationFolderPath || '' : ''
    };
  });

  ipcMain.handle('files:organize', async (_, payload) => {
    return organizeFiles(payload, {
      fsAdapter,
      lastOperationRepository
    });
  });

  ipcMain.handle('files:undo-last-organization', async () => {
    return undoLastOrganization({
      fsAdapter,
      lastOperationRepository
    });
  });

  ipcMain.handle('files:resolve-dropped-path', async (_, droppedPath) => {
    return resolveDroppedPath(droppedPath, {
      fsAdapter
    });
  });
}

module.exports = {
  registerFilesHandlers
};