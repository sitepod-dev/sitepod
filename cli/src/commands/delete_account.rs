use anyhow::Result;
use dialoguer::Confirm;
use std::io::IsTerminal;

use crate::api::ApiClient;
use crate::config::Config;
use crate::ui;

pub async fn run(config: &Config, skip_confirm: bool) -> Result<()> {
    // Check if we have a token
    if !config.has_token() {
        anyhow::bail!("No login session found. Run 'sitepod login' first.");
    }

    ui::warn("This will delete your account and all projects");
    println!();

    // Confirm unless skipped
    if !skip_confirm {
        if !std::io::stdin().is_terminal() {
            anyhow::bail!("Cannot confirm in non-interactive mode. Use --yes to skip confirmation.");
        }

        if !Confirm::new()
            .with_prompt("Delete account?")
            .default(false)
            .interact()?
        {
            ui::info("Cancelled");
            return Ok(());
        }
    }

    let client = ApiClient::new(config)?;

    println!();
    ui::step("Deleting account");

    let result = client.delete_account().await?;

    println!();
    ui::ok("Account deleted");
    ui::kv("projects", result.deleted_projects);

    // Clear local config
    if let Some(config_path) = Config::global_config_path() {
        if config_path.exists() {
            std::fs::remove_file(&config_path).ok();
            ui::kv("config", "cleared");
        }
    }

    Ok(())
}
