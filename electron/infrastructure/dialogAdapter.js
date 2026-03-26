const { dialog } = require('electron');

async function selectDirectory(title) {
  const result = await dialog.showOpenDialog({
    properties: ['openDirectory'],
    title
  });

  if (result.canceled || result.filePaths.length === 0) {
    return null;
  }

  return result.filePaths[0];
}

module.exports = {
  selectDirectory
};