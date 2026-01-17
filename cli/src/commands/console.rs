use crate::config::Config;
use crate::ui;
use anyhow::{Context, Result};

/// Open the SitePod console in browser
pub fn run(config: &Config) -> Result<()> {
    let endpoint = config
        .server
        .endpoint
        .as_ref()
        .context("No endpoint configured. Run 'sitepod login' first.")?;

    // Validate the endpoint URL
    let _ = url::Url::parse(endpoint).context("Invalid endpoint URL")?;

    // Console is now at the root domain (same as endpoint)
    let console_url = endpoint.trim_end_matches('/').to_string();

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
