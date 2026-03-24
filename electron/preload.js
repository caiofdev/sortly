const { contextBridge, ipcRenderer } = require('electron');

contextBridge.exposeInMainWorld('electronAPI', {
  selectFolder: () => ipcRenderer.invoke('dialog:select-folder'),
  organizeFiles: (folderPath) => ipcRenderer.invoke('files:organize', folderPath)
});
