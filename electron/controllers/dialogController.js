const { selectDirectory } = require('../infrastructure/dialogAdapter');

function registerDialogHandlers(ipcMain) {
  ipcMain.handle('dialog:select-source-folder', async () => {
    return selectDirectory('Selecione a pasta com os arquivos para organizar');
  });

  ipcMain.handle('dialog:select-destination-folder', async () => {
    return selectDirectory('Selecione a pasta de destino para receber os arquivos organizados');
  });
}

module.exports = {
  registerDialogHandlers
};