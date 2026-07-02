const { contextBridge, ipcRenderer } = require('electron');

contextBridge.exposeInMainWorld('electronAPI', {
  saveDb: (dataBase64) => ipcRenderer.invoke('save-db', dataBase64),
  saveDbAs: (dataBase64) => ipcRenderer.invoke('save-db-as', dataBase64),
  setDbPath: (filePath) => ipcRenderer.invoke('set-db-path', filePath),
  getDbPath: () => ipcRenderer.invoke('get-db-path'),
  openDbDialog: () => ipcRenderer.invoke('open-db-dialog'),
  openDbFilePath: (filePath) => ipcRenderer.invoke('open-db-file', filePath),
});
