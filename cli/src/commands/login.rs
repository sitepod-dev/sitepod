use anyhow::{Context, Result};
use dialoguer::{Input, Password};

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

    // Get email and password
    let email: String = Input::new().with_prompt("Email").interact_text()?;
    let password: String = Password::new().with_prompt("Password").interact()?;

    println!();
    ui::step("Authenticating");

    let client = reqwest::Client::new();

    // Call register-or-login endpoint
    let auth_url = format!("{}/api/v1/auth/login", endpoint.trim_end_matches('/'));

    let resp = client
        .post(&auth_url)
        .json(&serde_json::json!({
            "email": email,
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

    let token = auth_response["token"]
        .as_str()
        .context("No token in response")?
        .to_string();

    let created = auth_response["created"].as_bool().unwrap_or(false);

    // Save to config
    Config::save_token(&endpoint, &token)?;

    println!();
    if created {
        ui::ok("Account created");
    } else {
        ui::ok("Logged in");
    }
    let config_path = Config::global_config_path().unwrap().display().to_string();
    ui::kv("config", ui::dim(&config_path));

    Ok(())
}
