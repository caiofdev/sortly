const { ipcMain } = require('electron');
const { bootstrapApp } = require('./bootstrap/appBootstrap');
const { registerDialogHandlers } = require('./controllers/dialogController');
const { registerFilesHandlers } = require('./controllers/filesController');

registerDialogHandlers(ipcMain);
registerFilesHandlers(ipcMain);
bootstrapApp();