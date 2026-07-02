const { contextBridge, ipcRenderer } = require('electron');

console.log('[PRELOAD] Iniciando preload script...');

contextBridge.exposeInMainWorld('electronAPI', {
  saveDb: (dataBase64) => ipcRenderer.invoke('save-db', dataBase64),
  saveDbAs: (dataBase64) => ipcRenderer.invoke('save-db-as', dataBase64),
  setDbPath: (filePath) => ipcRenderer.invoke('set-db-path', filePath),
  getDbPath: () => ipcRenderer.invoke('get-db-path'),
  openDbDialog: async () => {
    console.log('[PRELOAD] Llamando a ipcRenderer.invoke para open-db-dialog');
    try {
      const result = await ipcRenderer.invoke('open-db-dialog');
      console.log('[PRELOAD] Resultado recibido:', result);
      return result;
    } catch (err) {
      console.error('[PRELOAD] Error en openDbDialog:', err);
      throw err;
    }
  },
  openDbFilePath: (filePath) => ipcRenderer.invoke('open-db-file', filePath),
});

console.log('[PRELOAD] contextBridge configurado correctamente');
