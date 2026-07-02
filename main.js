const { app, BrowserWindow, ipcMain, dialog } = require('electron')
const path = require('path')
const fs = require('fs')

console.log('[MAIN] Iniciando proceso principal...')

let currentDbPath = null

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

// IPC handlers
ipcMain.handle('save-db', async (_event, dataBase64) => {
  if (!currentDbPath) throw new Error('No hay archivo abierto')
  const data = Buffer.from(dataBase64, 'base64')
  fs.writeFileSync(currentDbPath, data)
  return true
})

ipcMain.handle('save-db-as', async (event, dataBase64) => {
  const win = BrowserWindow.fromWebContents(event.sender)
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

ipcMain.handle('open-db-file', async (_event, filePath) => {
  currentDbPath = filePath
  const data = fs.readFileSync(filePath)
  return data.toString('base64')
})

ipcMain.handle('open-db-dialog', async (event) => {
  console.log('[MAIN] Recibida solicitud open-db-dialog')
  const win = BrowserWindow.fromWebContents(event.sender)
  if (!win) {
    console.error('[MAIN] No se pudo obtener BrowserWindow desde el sender')
    return null
  }
  console.log('[MAIN] Ventana encontrada, abriendo diálogo...')
  const result = await dialog.showOpenDialog(win, {
    filters: [{ name: 'SQLite DB', extensions: ['db', 'sqlite'] }],
    properties: ['openFile'],
    title: 'Seleccionar Base de Datos',
  })
  console.log('[MAIN] Resultado del diálogo:', result)
  if (result.canceled || result.filePaths.length === 0) return null
  const filePath = result.filePaths[0]
  console.log('[MAIN] Archivo seleccionado:', filePath)
  currentDbPath = filePath
  const data = fs.readFileSync(filePath)
  return { path: filePath, data: data.toString('base64') }
})

app.whenReady().then(createWindow)
app.on('window-all-closed', () => app.quit())
