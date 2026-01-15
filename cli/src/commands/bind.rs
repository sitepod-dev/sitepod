use anyhow::Result;
use dialoguer::Input;

use crate::api::ApiClient;
use crate::config::Config;
use crate::ui;

/// Run the bind command to upgrade anonymous account
pub async fn run(config: &Config, email: Option<String>) -> Result<()> {
    ui::heading("Bind email");
    println!();

    // Check if we have a token
    if !config.has_token() {
        anyhow::bail!("No session. Run 'sitepod login' or 'sitepod deploy'.");
    }

    // Get email
    let email: String = if let Some(e) = email {
        e
    } else {
        Input::new()
            .with_prompt("Email")
            .interact_text()?
    };

    // Validate email format (basic check)
    if !email.contains('@') || !email.contains('.') {
        anyhow::bail!("Invalid email format");
    }

    println!();
    ui::step("Sending verification email");

    // Send bind request
    let client = ApiClient::new(config)?;
    let response = client.bind_email(&email).await?;

    println!();
    ui::ok(&response.message);
    println!();
    println!("Next:");
    println!("  - Check your inbox");
    println!("  - Click the verification link");
    println!("  - Account upgraded");
    println!();
    println!("  {}", ui::dim("After verification, deployments are permanent."));

    Ok(())
}
