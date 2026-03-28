const { contextBridge, ipcRenderer } = require('electron');

contextBridge.exposeInMainWorld('electronAPI', {
  selectSourceFolder: () => ipcRenderer.invoke('dialog:select-source-folder'),
  selectDestinationFolder: () => ipcRenderer.invoke('dialog:select-destination-folder'),
  resolveDroppedPath: (droppedPath) => ipcRenderer.invoke('files:resolve-dropped-path', droppedPath),
  getLastOrganizationState: () => ipcRenderer.invoke('files:get-last-organization-state'),
  organizeFiles: (payload) => ipcRenderer.invoke('files:organize', payload),
  undoLastOrganization: () => ipcRenderer.invoke('files:undo-last-organization')
});
