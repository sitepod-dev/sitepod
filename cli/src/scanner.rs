use anyhow::{Context, Result};
use glob::Pattern;
use indicatif::{ProgressBar, ProgressStyle};
use std::collections::HashMap;
use std::path::{Path, PathBuf};
use walkdir::WalkDir;

use crate::hash::{self, FileHashes};

/// Scanned file entry
#[derive(Debug, Clone)]
pub struct ScannedFile {
    /// Relative path from source root
    pub path: String,
    /// Absolute path on disk
    pub absolute_path: PathBuf,
    /// File hashes
    pub hashes: FileHashes,
    /// MIME type
    pub content_type: String,
}

/// Scanner for discovering and hashing files
pub struct Scanner {
    source_dir: PathBuf,
    ignore_patterns: Vec<Pattern>,
}

impl Scanner {
    pub fn new(source_dir: &Path, ignore: &[String]) -> Result<Self> {
        let source_dir = source_dir
            .canonicalize()
            .with_context(|| format!("Source directory not found: {}", source_dir.display()))?;

        let ignore_patterns: Vec<Pattern> = ignore
            .iter()
            .filter_map(|p| Pattern::new(p).ok())
            .collect();

        Ok(Self {
            source_dir,
            ignore_patterns,
        })
    }

    /// Scan directory and compute hashes for all files
    pub fn scan(&self) -> Result<Vec<ScannedFile>> {
        // First pass: collect all file paths
        let paths: Vec<PathBuf> = WalkDir::new(&self.source_dir)
            .into_iter()
            .filter_map(|e| e.ok())
            .filter(|e| e.file_type().is_file())
            .map(|e| e.path().to_path_buf())
            .filter(|p| !self.should_ignore(p))
            .collect();

        let total = paths.len() as u64;

        // Progress bar
        let pb = ProgressBar::new(total);
        pb.set_style(
            ProgressStyle::default_bar()
                .template("{spinner:.green} [{bar:40.cyan/blue}] {pos}/{len} {msg}")
                .unwrap()
                .progress_chars("=>-"),
        );
        pb.set_message("Computing hashes...");

        // Second pass: compute hashes
        let mut files = Vec::with_capacity(paths.len());

        for path in paths {
            let rel_path = path
                .strip_prefix(&self.source_dir)
                .unwrap()
                .to_string_lossy()
                .replace('\\', "/"); // Normalize path separators

            let hashes = hash::compute_hashes(&path)?;

            let content_type = mime_guess::from_path(&path)
                .first()
                .map(|m| m.to_string())
                .unwrap_or_else(|| "application/octet-stream".to_string());

            files.push(ScannedFile {
                path: rel_path,
                absolute_path: path,
                hashes,
                content_type,
            });

            pb.inc(1);
        }

        pb.finish_with_message("Done!");

        Ok(files)
    }

    /// Create a map of path -> ScannedFile for quick lookup
    #[allow(dead_code)]
    pub fn scan_as_map(&self) -> Result<HashMap<String, ScannedFile>> {
        let files = self.scan()?;
        Ok(files.into_iter().map(|f| (f.path.clone(), f)).collect())
    }

    fn should_ignore(&self, path: &Path) -> bool {
        let rel_path = path
            .strip_prefix(&self.source_dir)
            .unwrap_or(path)
            .to_string_lossy();

        for pattern in &self.ignore_patterns {
            if pattern.matches(&rel_path) {
                return true;
            }
        }

        // Also ignore hidden files by default
        if let Some(name) = path.file_name() {
            if name.to_string_lossy().starts_with('.') {
                return true;
            }
        }

        false
    }
}

/// Get build directory from config or default
pub fn get_source_dir(source: &str, config_dir: Option<&str>) -> PathBuf {
    if source != "." {
        PathBuf::from(source)
    } else if let Some(dir) = config_dir {
        PathBuf::from(dir)
    } else {
        PathBuf::from("./dist")
    }
}
