use base64::engine::general_purpose::STANDARD as BASE64;
use base64::Engine;
use serde::{Deserialize, Serialize};
use std::sync::Mutex;
use tauri::State;
use tauri_plugin_dialog::{DialogExt, FilePath};
use tokio::time::{timeout, Duration};

struct AppState {
    current_db_path: Mutex<Option<String>>,
    backup_max_copies: Mutex<u32>,
}

#[derive(Serialize, Deserialize)]
struct FileResult {
    path: String,
    data: String,
}

fn crear_backup_rotativo(file_path: &str, max_copies: u32) {
    let oldest = format!("{}.bak.{}", file_path, max_copies);
    if std::path::Path::new(&oldest).exists() {
        let _ = std::fs::remove_file(&oldest);
    }
    for i in (1..max_copies).rev() {
        let src = format!("{}.bak.{}", file_path, i);
        let dst = format!("{}.bak.{}", file_path, i + 1);
        if std::path::Path::new(&src).exists() {
            let _ = std::fs::rename(&src, &dst);
        }
    }
    let bak1 = format!("{}.bak.{}", file_path, 1);
    let _ = std::fs::copy(file_path, &bak1);
}

#[tauri::command]
fn save_db(state: State<AppState>, data_base64: String) -> Result<(), String> {
    let path = state
        .current_db_path
        .lock()
        .map_err(|e| e.to_string())?
        .clone()
        .ok_or_else(|| "No hay archivo abierto".to_string())?;

    let max_copies = *state.backup_max_copies.lock().map_err(|e| e.to_string())?;
    crear_backup_rotativo(&path, max_copies);

    let data = BASE64.decode(&data_base64).map_err(|e| e.to_string())?;
    std::fs::write(&path, &data).map_err(|e| e.to_string())?;
    Ok(())
}

#[tauri::command]
async fn save_db_as(
    app: tauri::AppHandle,
    state: State<'_, AppState>,
    data_base64: String,
) -> Result<Option<String>, String> {
    eprintln!("[RUST] save_db_as: abriendo dialogo Guardar Como...");
    let (tx, rx) = tokio::sync::oneshot::channel();
    app.dialog()
        .file()
        .add_filter("SQLite DB", &["db", "sqlite"])
        .save_file(move |file_path| {
            eprintln!("[RUST] save_db_as: callback recibido: {:?}", file_path);
            let _ = tx.send(file_path);
        });
    let file = timeout(Duration::from_secs(60), rx)
        .await
        .map_err(|_| "Timeout: el dialogo no respondio en 60s".to_string())?
        .map_err(|e| format!("Error en oneshot: {}", e))?;
    eprintln!("[RUST] save_db_as: resultado: {:?}", file);

    match file {
        Some(FilePath::Path(path_buf)) => {
            let path_str = path_buf.to_string_lossy().to_string();
            let data = BASE64.decode(&data_base64).map_err(|e| e.to_string())?;
            std::fs::write(&path_str, &data).map_err(|e| e.to_string())?;
            *state.current_db_path.lock().map_err(|e| e.to_string())? = Some(path_str.clone());
            Ok(Some(path_str))
        }
        Some(FilePath::Url(_)) => Err("URL no soportada".to_string()),
        None => Ok(None),
    }
}

#[tauri::command]
fn set_db_path(state: State<AppState>, file_path: String) -> Result<(), String> {
    *state.current_db_path.lock().map_err(|e| e.to_string())? = Some(file_path);
    Ok(())
}

#[tauri::command]
fn get_db_path(state: State<AppState>) -> Result<Option<String>, String> {
    Ok(state
        .current_db_path
        .lock()
        .map_err(|e| e.to_string())?
        .clone())
}

#[tauri::command]
fn open_db_file(state: State<AppState>, file_path: String) -> Result<String, String> {
    let data = std::fs::read(&file_path).map_err(|e| e.to_string())?;
    *state.current_db_path.lock().map_err(|e| e.to_string())? = Some(file_path);
    Ok(BASE64.encode(&data))
}

#[tauri::command]
async fn open_db_dialog(
    app: tauri::AppHandle,
    state: State<'_, AppState>,
) -> Result<Option<FileResult>, String> {
    eprintln!("[RUST] open_db_dialog: abriendo dialogo...");
    let (tx, rx) = tokio::sync::oneshot::channel();
    app.dialog()
        .file()
        .add_filter("SQLite DB", &["db", "sqlite"])
        .pick_file(move |file_path| {
            eprintln!("[RUST] open_db_dialog: callback recibido: {:?}", file_path);
            let _ = tx.send(file_path);
        });
    let file = timeout(Duration::from_secs(60), rx)
        .await
        .map_err(|_| "Timeout: el dialogo no respondio en 60s".to_string())?
        .map_err(|e| format!("Error en oneshot: {}", e))?;
    eprintln!("[RUST] open_db_dialog: resultado: {:?}", file);

    match file {
        Some(FilePath::Path(path_buf)) => {
            let path_str = path_buf.to_string_lossy().to_string();
            let data = std::fs::read(&path_str).map_err(|e| e.to_string())?;
            *state.current_db_path.lock().map_err(|e| e.to_string())? = Some(path_str.clone());
            Ok(Some(FileResult {
                path: path_str,
                data: BASE64.encode(&data),
            }))
        }
        Some(FilePath::Url(_)) => Err("URL no soportada".to_string()),
        None => Ok(None),
    }
}

#[tauri::command]
fn set_backup_copies(state: State<AppState>, n: u32) -> Result<(), String> {
    if n > 0 && n <= 20 {
        *state.backup_max_copies.lock().map_err(|e| e.to_string())? = n;
    }
    Ok(())
}

#[tauri::command]
fn get_backup_copies(state: State<AppState>) -> Result<u32, String> {
    Ok(*state.backup_max_copies.lock().map_err(|e| e.to_string())?)
}

pub fn run() {
    tauri::Builder::default()
        .plugin(tauri_plugin_dialog::init())
        .manage(AppState {
            current_db_path: Mutex::new(None),
            backup_max_copies: Mutex::new(2),
        })
        .invoke_handler(tauri::generate_handler![
            save_db,
            save_db_as,
            set_db_path,
            get_db_path,
            open_db_file,
            open_db_dialog,
            set_backup_copies,
            get_backup_copies,
        ])
        .run(tauri::generate_context!())
        .expect("error al ejecutar la aplicación Tauri");
}
