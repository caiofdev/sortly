const { app, BrowserWindow } = require('electron');
const path = require('path');

const appIconPath = path.join(__dirname, '../assets/app-logo.ico');
const windowsAppId = 'com.sortly.app';

function createWindow() {
  const window = new BrowserWindow({
    width: 980,
    height: 700,
    minWidth: 820,
    minHeight: 600,
    show: false,
    title: 'Sortly',
    icon: appIconPath,
    backgroundColor: '#0f172a',
    webPreferences: {
      preload: path.join(__dirname, '../preload.js'),
      contextIsolation: true,
      nodeIntegration: false,
      sandbox: false
    }
  });

  const devServerUrl = process.env.VITE_DEV_SERVER_URL;

  if (devServerUrl) {
    window.loadURL(devServerUrl);
  } else {
    window.loadFile(path.join(__dirname, '../../dist/index.html'));
  }

  window.once('ready-to-show', () => {
    window.setTitle('Sortly');
    window.show();
  });
}

function bootstrapApp() {
  if (process.platform === 'win32') {
    app.setAppUserModelId(windowsAppId);
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
}

module.exports = {
  bootstrapApp
};