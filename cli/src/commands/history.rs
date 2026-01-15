use anyhow::{Context, Result};
use crate::api::ApiClient;
use crate::config::Config;
use crate::ui;

/// Run the history command
pub async fn run(config: &Config, project: Option<&str>, limit: usize) -> Result<()> {
    let project_name = project.context(
        "Project name not specified. Run 'sitepod init' or use --name flag.",
    )?;

    let client = ApiClient::new(config)?;

    ui::heading(&format!("History: {}", project_name));
    println!();

    let history = client.history(project_name, limit).await?;

    if history.items.is_empty() {
        println!("{}", ui::dim("No deployments found."));
        return Ok(());
    }

    // Print table header
    println!(
        "{:<12} {:<20} {:<18} {}",
        console::style("IMAGE ID").bold(),
        console::style("CREATED").bold(),
        console::style("CONTENT HASH").bold(),
        console::style("GIT COMMIT").bold()
    );
    println!("{}", "-".repeat(80));

    // Print items
    for item in history.items {
        let commit = item
            .git_commit
            .map(|c| c[..7.min(c.len())].to_string())
            .unwrap_or_else(|| "-".to_string());

        let hash = if item.content_hash.len() > 16 {
            format!("{}...", &item.content_hash[..16])
        } else {
            item.content_hash
        };

        println!(
            "{:<12} {:<20} {:<18} {}",
            console::style(&item.image_id).green(),
            item.created_at.format("%Y-%m-%d %H:%M:%S"),
            console::style(hash).dim(),
            console::style(commit).dim()
        );
    }

    Ok(())
}
