// Controllers IPC finos

const fsAdapter = require('../infrastructure/fsAdapter');
const lastOperationRepository = require('../repositories/lastOperationRepository');
const { organizeFiles } = require('../services/organizeFilesService');
const { undoLastOrganization } = require('../services/undoOrganizationService');

function registerFilesHandlers(ipcMain) {
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
}

module.exports = {
  registerFilesHandlers
};