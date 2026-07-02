const { app, BrowserWindow, ipcMain, dialog } = require('electron')
const path = require('path')
const fs = require('fs')

let win
let currentDbPath = null

function createWindow() {
  win = new BrowserWindow({
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

// IPC handlers
ipcMain.handle('pick-db-file', async () => {
  const result = await dialog.showOpenDialog(win, {
    filters: [{ name: 'SQLite DB', extensions: ['db', 'sqlite'] }],
    properties: ['openFile']
  })
  if (result.canceled || result.filePaths.length === 0) {
    return null
  }
  const filePath = result.filePaths[0]
  const data = fs.readFileSync(filePath)
  const dataBase64 = data.toString('base64')
  currentDbPath = filePath
  return { path: filePath, data_base64: dataBase64 }
})

ipcMain.handle('save-db', async (_event, dataBase64) => {
  if (!currentDbPath) throw new Error('No hay archivo abierto')
  const data = Buffer.from(dataBase64, 'base64')
  fs.writeFileSync(currentDbPath, data)
  return true
})

ipcMain.handle('save-db-as', async (_event, dataBase64) => {
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

ipcMain.handle('get-db-path', async () => {
  return currentDbPath
})

app.whenReady().then(createWindow)
app.on('window-all-closed', () => app.quit())
