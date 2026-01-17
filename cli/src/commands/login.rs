use anyhow::{Context, Result};
use dialoguer::{Input, Password, Select};

use crate::config::Config;
use crate::ui;

/// Run the login command
pub async fn run(endpoint: Option<String>) -> Result<()> {
    ui::heading("Login");
    println!();

    // Get endpoint
    let endpoint: String = if let Some(ep) = endpoint {
        ep
    } else {
        Input::new()
            .with_prompt("Server endpoint")
            .default("http://localhost:8080".to_string())
            .interact_text()?
    };

    // Choose login method
    let methods = vec!["Anonymous (quick start, 24h limit)", "Email & Password"];

    let selection = Select::new()
        .with_prompt("Login method")
        .items(&methods)
        .default(0)
        .interact()?;

    let client = reqwest::Client::new();

    let token = if selection == 0 {
        // Anonymous login
        println!();
        ui::step("Creating anonymous session");

        let anon_url = format!("{}/api/v1/auth/anonymous", endpoint.trim_end_matches('/'));

        let resp = client
            .post(&anon_url)
            .send()
            .await
            .context("Failed to connect to server")?;

        if !resp.status().is_success() {
            let status = resp.status();
            let text = resp.text().await.unwrap_or_default();
            anyhow::bail!("Anonymous auth failed ({}): {}", status, text);
        }

        let auth_response: serde_json::Value = resp.json().await?;

        let token = auth_response["token"]
            .as_str()
            .context("No token in response")?
            .to_string();

        let expires_at = auth_response["expires_at"].as_str().unwrap_or("24 hours");

        println!();
        ui::warn("Anonymous session");
        ui::kv("expires", ui::dim(expires_at));
        ui::kv("next", ui::cmd("sitepod bind"));

        token
    } else {
        // Email & password login
        let email: String = Input::new().with_prompt("Email").interact_text()?;

        let password: String = Password::new().with_prompt("Password").interact()?;

        println!();
        ui::step("Authenticating");

        let auth_url = format!(
            "{}/api/collections/users/auth-with-password",
            endpoint.trim_end_matches('/')
        );

        let resp = client
            .post(&auth_url)
            .json(&serde_json::json!({
                "identity": email,
                "password": password
            }))
            .send()
            .await
            .context("Failed to connect to server")?;

        if !resp.status().is_success() {
            let status = resp.status();
            let text = resp.text().await.unwrap_or_default();
            anyhow::bail!("Authentication failed ({}): {}", status, text);
        }

        let auth_response: serde_json::Value = resp.json().await?;
        auth_response["token"]
            .as_str()
            .context("No token in response")?
            .to_string()
    };

    // Save to config
    Config::save_token(&endpoint, &token)?;

    println!();
    ui::ok("Logged in");
    let config_path = Config::global_config_path().unwrap().display().to_string();
    ui::kv("config", ui::dim(&config_path));

    Ok(())
}
