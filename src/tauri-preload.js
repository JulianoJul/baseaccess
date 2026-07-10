(function() {
    const invoke = window.__TAURI__.core.invoke;

    window.electronAPI = {
        saveDb: (dataBase64) => invoke('save_db', { dataBase64 }),
        saveDbAs: (dataBase64) => invoke('save_db_as', { dataBase64 }),
        setDbPath: (filePath) => invoke('set_db_path', { filePath }),
        getDbPath: () => invoke('get_db_path'),
        openDbDialog: async () => {
            const result = await invoke('open_db_dialog');
            return result;
        },
        openDbFilePath: (filePath) => invoke('open_db_file', { filePath }),
        setBackupCopies: (n) => invoke('set_backup_copies', { n }),
        getBackupCopies: () => invoke('get_backup_copies'),
    };
})();
