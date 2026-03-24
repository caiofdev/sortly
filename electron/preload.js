const { contextBridge, ipcRenderer } = require('electron');

contextBridge.exposeInMainWorld('electronAPI', {
  selectSourceFolder: () => ipcRenderer.invoke('dialog:select-source-folder'),
  selectDestinationFolder: () => ipcRenderer.invoke('dialog:select-destination-folder'),
  organizeFiles: (payload) => ipcRenderer.invoke('files:organize', payload),
  undoLastOrganization: () => ipcRenderer.invoke('files:undo-last-organization')
});
