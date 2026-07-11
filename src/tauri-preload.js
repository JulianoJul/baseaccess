(function() {
    // Si ya está definida (por ejemplo, en Electron), no sobrescribir
    if (window.electronAPI) return;

    let tauriAPI = null;

    Object.defineProperty(window, 'electronAPI', {
        get: function() {
            if (!window.__TAURI__) {
                return undefined;
            }
            if (!tauriAPI) {
                const invoke = window.__TAURI__.core.invoke;
                tauriAPI = {
                    isTauri: true,
                    saveDb: (dataBase64) => invoke('save_db', { dataBase64 }),
                    saveDbAs: (dataBase64) => invoke('save_db_as', { dataBase64 }),
                    setDbPath: (filePath) => invoke('set_db_path', { filePath }),
                    getDbPath: () => invoke('get_db_path'),
                    openDbDialog: async () => {
                        return await invoke('open_db_dialog');
                    },
                    openDbFilePath: (filePath) => invoke('open_db_file', { filePath }),
                    setBackupCopies: (n) => invoke('set_backup_copies', { n }),
                    getBackupCopies: () => invoke('get_backup_copies'),
                };
            }
            return tauriAPI;
        },
        configurable: true,
        enumerable: true
    });
})();
