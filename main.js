const { app, BrowserWindow, ipcMain, dialog } = require('electron')
const path = require('path')
const fs = require('fs')

let currentDbPath = null

function getWindow() {
  return BrowserWindow.getFocusedWindow() || BrowserWindow.getAllWindows()[0]
}

function createWindow() {
  const win = new BrowserWindow({
    width: 1400,
    height: 900,
    title: 'Gestión de Expedientes con Historial',
    webPreferences: {
      preload: path.join(__dirname, 'preload.js'),
      nodeIntegration: false,
      contextIsolation: true,
    }
  })

  win.loadFile('index.html')
  win.setMenuBarVisibility(false)
}

// IPC handlers (solo para guardar)
ipcMain.handle('save-db', async (_event, dataBase64) => {
  if (!currentDbPath) throw new Error('No hay archivo abierto')
  const data = Buffer.from(dataBase64, 'base64')
  fs.writeFileSync(currentDbPath, data)
  return true
})

ipcMain.handle('save-db-as', async (_event, dataBase64) => {
  const win = getWindow()
  if (!win) return null
  const result = await dialog.showSaveDialog(win, {
    filters: [{ name: 'SQLite DB', extensions: ['db', 'sqlite'] }],
  })
  if (result.canceled || !result.filePath) {
    return null
  }
  const data = Buffer.from(dataBase64, 'base64')
  fs.writeFileSync(result.filePath, data)
  currentDbPath = result.filePath
  return result.filePath
})

ipcMain.handle('set-db-path', async (_event, filePath) => {
  currentDbPath = filePath
  return true
})

ipcMain.handle('get-db-path', async () => {
  return currentDbPath
})

// Sincronizar currentDbPath desde el frontend cuando abre un archivo
ipcMain.handle('open-db-file', async (_event, filePath) => {
  currentDbPath = filePath
  const data = fs.readFileSync(filePath)
  return data.toString('base64')
})

app.whenReady().then(createWindow)
app.on('window-all-closed', () => app.quit())
