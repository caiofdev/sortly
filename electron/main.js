const { app, BrowserWindow, ipcMain, dialog } = require('electron');
const path = require('path');
const fs = require('fs/promises');

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

ipcMain.handle('dialog:select-folder', async () => {
  const result = await dialog.showOpenDialog({
    properties: ['openDirectory'],
    title: 'Selecione uma pasta para organizar'
  });

  if (result.canceled || result.filePaths.length === 0) {
    return null;
  }

  return result.filePaths[0];
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

ipcMain.handle('files:organize', async (_, folderPath) => {
  if (!folderPath || typeof folderPath !== 'string') {
    throw new Error('Pasta inválida.');
  }

  const entries = await fs.readdir(folderPath, { withFileTypes: true });
  let processedFiles = 0;
  let movedFiles = 0;
  let ignoredWithoutExtension = 0;
  let ignoredFolders = 0;

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

    const extensionFolderPath = path.join(folderPath, extension);
    await fs.mkdir(extensionFolderPath, { recursive: true });

    const sourcePath = path.join(folderPath, entry.name);
    const destinationPath = path.join(extensionFolderPath, entry.name);
    const safeDestinationPath = await getUniqueDestination(destinationPath);

    await fs.rename(sourcePath, safeDestinationPath);
    movedFiles += 1;
  }

  return {
    processedFiles,
    movedFiles,
    ignoredWithoutExtension,
    ignoredFolders,
    message: `Organização concluída: ${movedFiles} arquivo(s) movido(s).`
  };
});
