use anyhow::{Context, Result};
use dialoguer::Select;

use crate::api::ApiClient;
use crate::config::Config;
use crate::ui;

/// Run the rollback command
pub async fn run(
    config: &Config,
    project: Option<&str>,
    env: &str,
    image: Option<String>,
) -> Result<()> {
    let project_name = project.context(
        "Project name not specified. Run 'sitepod init' or use --name flag.",
    )?;

    let client = ApiClient::new(config)?;

    // Get image ID
    let image_id = if let Some(id) = image {
        id
    } else {
        // Interactive selection
        ui::step("Fetching history");

        let history = client.history(project_name, 10).await?;

        if history.items.is_empty() {
            anyhow::bail!("No deployment history found");
        }

        // Build selection list
        let items: Vec<String> = history
            .items
            .iter()
            .map(|item| {
                let commit = item
                    .git_commit
                    .as_ref()
                    .map(|c| format!(" ({})", &c[..7.min(c.len())]))
                    .unwrap_or_default();
                format!(
                    "{} - {}{}",
                    item.image_id,
                    item.created_at.format("%Y-%m-%d %H:%M"),
                    commit
                )
            })
            .collect();

        println!();
        let selection = Select::new()
            .with_prompt("Rollback to")
            .items(&items)
            .default(0)
            .interact()?;

        history.items[selection].image_id.clone()
    };

    println!();
    ui::step(&format!("Rollback {} → {}", env, image_id));

    let result = client.rollback(project_name, env, &image_id).await?;

    println!();
    ui::ok("Rollback complete");
    ui::kv("url", ui::accent(&result.url).underlined());
    ui::kv("previous", ui::dim(&result.previous_image_id));

    Ok(())
}
