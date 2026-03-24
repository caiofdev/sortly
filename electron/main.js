const { app, BrowserWindow, ipcMain, dialog } = require('electron');
const path = require('path');
const fs = require('fs/promises');

let lastOperation = null;

function createWindow() {
  const window = new BrowserWindow({
    width: 980,
    height: 700,
    minWidth: 820,
    minHeight: 600,
    show: false,
    title: 'Organizador de Arquivos',
    backgroundColor: '#0f172a',
    webPreferences: {
      preload: path.join(__dirname, 'preload.js'),
      contextIsolation: true,
      nodeIntegration: false,
      sandbox: false
    }
  });

  const devServerUrl = process.env.VITE_DEV_SERVER_URL;

  if (devServerUrl) {
    window.loadURL(devServerUrl);
  } else {
    window.loadFile(path.join(__dirname, '../dist/index.html'));
  }

  window.once('ready-to-show', () => {
    window.show();
  });
}

app.whenReady().then(() => {
  createWindow();

  app.on('activate', () => {
    if (BrowserWindow.getAllWindows().length === 0) {
      createWindow();
    }
  });
});

app.on('window-all-closed', () => {
  if (process.platform !== 'darwin') {
    app.quit();
  }
});

async function openDirectoryDialog(title) {
  const result = await dialog.showOpenDialog({
    properties: ['openDirectory'],
    title
  });

  if (result.canceled || result.filePaths.length === 0) {
    return null;
  }

  return result.filePaths[0];
}

ipcMain.handle('dialog:select-source-folder', async () => {
  return openDirectoryDialog('Selecione a pasta com os arquivos para organizar');
});

ipcMain.handle('dialog:select-destination-folder', async () => {
  return openDirectoryDialog('Selecione a pasta de destino para receber os arquivos organizados');
});

async function fileExists(filePath) {
  try {
    await fs.access(filePath);
    return true;
  } catch {
    return false;
  }
}

async function getUniqueDestination(destinationPath) {
  if (!(await fileExists(destinationPath))) {
    return destinationPath;
  }

  const parentDir = path.dirname(destinationPath);
  const extension = path.extname(destinationPath);
  const baseName = path.basename(destinationPath, extension);
  let counter = 1;

  while (true) {
    const candidatePath = path.join(parentDir, `${baseName} (${counter})${extension}`);
    if (!(await fileExists(candidatePath))) {
      return candidatePath;
    }
    counter += 1;
  }
}

ipcMain.handle('files:organize', async (_, payload) => {
  const sourceFolderPath = payload?.sourceFolderPath;
  const destinationFolderPath = payload?.destinationFolderPath || sourceFolderPath;

  if (!sourceFolderPath || typeof sourceFolderPath !== 'string') {
    throw new Error('Pasta inválida.');
  }

  if (!destinationFolderPath || typeof destinationFolderPath !== 'string') {
    throw new Error('Pasta de destino inválida.');
  }

  const entries = await fs.readdir(sourceFolderPath, { withFileTypes: true });
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
    await fs.mkdir(extensionFolderPath, { recursive: true });
    createdFolders.add(extensionFolderPath);

    const sourcePath = path.join(sourceFolderPath, entry.name);
    const destinationPath = path.join(extensionFolderPath, entry.name);
    const safeDestinationPath = await getUniqueDestination(destinationPath);

    await fs.rename(sourcePath, safeDestinationPath);
    movedItems.push({ from: sourcePath, to: safeDestinationPath });
    movedFiles += 1;
  }

  lastOperation = {
    sourceFolderPath,
    destinationFolderPath,
    movedItems,
    createdFolders: [...createdFolders]
  };

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
});

ipcMain.handle('files:undo-last-organization', async () => {
  if (!lastOperation || lastOperation.movedItems.length === 0) {
    throw new Error('Nenhuma separação recente para desfazer.');
  }

  let restoredFiles = 0;
  let renamedOnRestore = 0;
  let skippedMissing = 0;

  for (const move of [...lastOperation.movedItems].reverse()) {
    if (!(await fileExists(move.to))) {
      skippedMissing += 1;
      continue;
    }

    const safeRestorePath = await getUniqueDestination(move.from);
    await fs.rename(move.to, safeRestorePath);

    if (safeRestorePath !== move.from) {
      renamedOnRestore += 1;
    }

    restoredFiles += 1;
  }

  for (const folderPath of lastOperation.createdFolders) {
    try {
      await fs.rmdir(folderPath);
    } catch {
      // Ignora pastas não vazias ou já removidas.
    }
  }

  lastOperation = null;

  return {
    restoredFiles,
    renamedOnRestore,
    skippedMissing,
    message: `Desfazer concluído: ${restoredFiles} arquivo(s) restaurado(s).`,
    canUndo: false
  };
});
