use anyhow::{Context, Result};
use crate::config::Config;
use crate::ui;

/// Open the SitePod console in browser
pub fn run(config: &Config) -> Result<()> {
    let endpoint = config
        .server
        .endpoint
        .as_ref()
        .context("No endpoint configured. Run 'sitepod login' first.")?;

    // Parse the endpoint to get the domain
    let url = url::Url::parse(endpoint).context("Invalid endpoint URL")?;
    let host = url.host_str().unwrap_or("localhost");

    // Build console URL
    let console_url = if host.contains("localhost") || host.starts_with("127.") {
        // Local development - use subdomain
        let port = url.port().map(|p| format!(":{}", p)).unwrap_or_default();
        format!("{}://console.{}{}", url.scheme(), host, port)
    } else {
        // Production - use console subdomain
        format!("{}://console.{}", url.scheme(), host)
    };

    ui::info("Opening console");
    ui::kv("url", ui::accent(&console_url).underlined());

    // Open in browser
    if let Err(e) = open::that(&console_url) {
        println!();
        ui::warn(&format!("Open browser failed: {}", e));
        ui::kv("url", &console_url);
    }

    Ok(())
}
