use anyhow::{Context, Result};
use dialoguer::{Input, Password};
use std::env;

use crate::config::Config;
use crate::ui;

/// Run the login command
/// 
/// Supports non-interactive mode via environment variables:
/// - SITEPOD_EMAIL: Account email
/// - SITEPOD_PASSWORD: Account password
pub async fn run(endpoint: Option<String>) -> Result<()> {
    // Check for non-interactive mode (CI/CD)
    let env_email = env::var("SITEPOD_EMAIL").ok();
    let env_password = env::var("SITEPOD_PASSWORD").ok();
    let non_interactive = env_email.is_some() && env_password.is_some();

    if !non_interactive {
        ui::heading("Login");
        println!();
    }

    // Get endpoint
    let endpoint: String = if let Some(ep) = endpoint {
        ep
    } else if non_interactive {
        // In CI, endpoint must be provided via --endpoint or SITEPOD_ENDPOINT
        anyhow::bail!("Endpoint required in non-interactive mode (use --endpoint or SITEPOD_ENDPOINT)");
    } else {
        Input::new()
            .with_prompt("Server endpoint")
            .default("http://localhost:8080".to_string())
            .interact_text()?
    };

    // Get email and password (from env or interactive)
    let email: String = if let Some(e) = env_email {
        e
    } else {
        Input::new().with_prompt("Email").interact_text()?
    };
    
    let password: String = if let Some(p) = env_password {
        p
    } else {
        Password::new().with_prompt("Password").interact()?
    };

    if !non_interactive {
        println!();
    }
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

    if non_interactive {
        // Minimal output for CI
        if created {
            println!("Account created");
        } else {
            println!("Logged in");
        }
    } else {
        println!();
        if created {
            ui::ok("Account created");
        } else {
            ui::ok("Logged in");
        }
        let config_path = Config::global_config_path().unwrap().display().to_string();
        ui::kv("config", ui::dim(&config_path));
    }

    Ok(())
}
