const { contextBridge, ipcRenderer } = require('electron');

contextBridge.exposeInMainWorld('electronAPI', {
  pickDbFile: () => ipcRenderer.invoke('pick-db-file'),
  saveDb: (dataBase64) => ipcRenderer.invoke('save-db', dataBase64),
  saveDbAs: (dataBase64) => ipcRenderer.invoke('save-db-as', dataBase64),
  getDbPath: () => ipcRenderer.invoke('get-db-path'),
});
