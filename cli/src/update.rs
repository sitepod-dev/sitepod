//! Version check module - checks for CLI updates from GitHub releases

use anyhow::Result;
use console::style;
use serde::Deserialize;
use std::fs;
use std::path::PathBuf;
use std::time::{Duration, SystemTime};

const GITHUB_REPO: &str = "sitepod-dev/sitepod";
const CHECK_INTERVAL: Duration = Duration::from_secs(24 * 60 * 60); // 24 hours
const CURRENT_VERSION: &str = env!("CARGO_PKG_VERSION");

#[derive(Debug, Deserialize)]
struct GitHubRelease {
    tag_name: String,
    #[allow(dead_code)]
    html_url: String,
}

#[derive(Debug, serde::Serialize, serde::Deserialize)]
struct VersionCache {
    latest_version: String,
    checked_at: u64, // Unix timestamp
}

/// Check for updates and print notification if available
pub async fn check_for_updates() {
    // Run in background, don't block the main command
    if let Err(e) = do_check().await {
        // Silently ignore errors - version check should not affect normal operation
        if std::env::var("SITEPOD_DEBUG").is_ok() {
            eprintln!("Version check failed: {}", e);
        }
    }
}

async fn do_check() -> Result<()> {
    let cache_path = get_cache_path()?;

    // Check if we have a recent cache
    if let Some(cache) = read_cache(&cache_path) {
        let now = SystemTime::now()
            .duration_since(SystemTime::UNIX_EPOCH)?
            .as_secs();

        if now - cache.checked_at < CHECK_INTERVAL.as_secs() {
            // Use cached version
            if is_newer_version(&cache.latest_version, CURRENT_VERSION) {
                print_update_notification(&cache.latest_version);
            }
            return Ok(());
        }
    }

    // Fetch latest version from GitHub
    let latest = fetch_latest_version().await?;

    // Update cache
    let cache = VersionCache {
        latest_version: latest.clone(),
        checked_at: SystemTime::now()
            .duration_since(SystemTime::UNIX_EPOCH)?
            .as_secs(),
    };
    write_cache(&cache_path, &cache)?;

    // Show notification if newer
    if is_newer_version(&latest, CURRENT_VERSION) {
        print_update_notification(&latest);
    }

    Ok(())
}

fn get_cache_path() -> Result<PathBuf> {
    let cache_dir = dirs::cache_dir()
        .or_else(dirs::home_dir)
        .ok_or_else(|| anyhow::anyhow!("Could not determine cache directory"))?;

    let sitepod_cache = cache_dir.join("sitepod");
    fs::create_dir_all(&sitepod_cache)?;

    Ok(sitepod_cache.join("version-check.json"))
}

fn read_cache(path: &PathBuf) -> Option<VersionCache> {
    let content = fs::read_to_string(path).ok()?;
    serde_json::from_str(&content).ok()
}

fn write_cache(path: &PathBuf, cache: &VersionCache) -> Result<()> {
    let content = serde_json::to_string(cache)?;
    fs::write(path, content)?;
    Ok(())
}

async fn fetch_latest_version() -> Result<String> {
    let url = format!(
        "https://api.github.com/repos/{}/releases/latest",
        GITHUB_REPO
    );

    let client = reqwest::Client::builder()
        .user_agent("sitepod-cli")
        .timeout(Duration::from_secs(5))
        .build()?;

    let resp = client.get(&url).send().await?;

    if !resp.status().is_success() {
        anyhow::bail!("GitHub API returned {}", resp.status());
    }

    let release: GitHubRelease = resp.json().await?;

    // Remove 'v' prefix if present
    let version = release.tag_name.trim_start_matches('v').to_string();

    Ok(version)
}

/// Compare versions (simple semver comparison)
fn is_newer_version(latest: &str, current: &str) -> bool {
    let parse_version =
        |v: &str| -> Vec<u32> { v.split('.').filter_map(|s| s.parse().ok()).collect() };

    let latest_parts = parse_version(latest);
    let current_parts = parse_version(current);

    for i in 0..3 {
        let l = latest_parts.get(i).copied().unwrap_or(0);
        let c = current_parts.get(i).copied().unwrap_or(0);

        if l > c {
            return true;
        } else if l < c {
            return false;
        }
    }

    false
}

fn print_update_notification(latest: &str) {
    eprintln!();
    eprintln!(
        "{} A new version of sitepod is available: {} -> {}",
        style("Update available!").yellow().bold(),
        style(CURRENT_VERSION).dim(),
        style(latest).green().bold()
    );
    eprintln!("  Run {} or visit:", style("npm install -g sitepod").cyan());
    eprintln!(
        "  {}",
        style(format!("https://github.com/{}/releases", GITHUB_REPO)).dim()
    );
    eprintln!();
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_version_comparison() {
        assert!(is_newer_version("1.0.1", "1.0.0"));
        assert!(is_newer_version("1.1.0", "1.0.0"));
        assert!(is_newer_version("2.0.0", "1.9.9"));
        assert!(!is_newer_version("1.0.0", "1.0.0"));
        assert!(!is_newer_version("1.0.0", "1.0.1"));
        assert!(!is_newer_version("0.9.0", "1.0.0"));
    }
}
