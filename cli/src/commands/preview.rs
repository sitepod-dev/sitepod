use anyhow::{Context, Result};
use futures::stream::{self, StreamExt};
use indicatif::{ProgressBar, ProgressStyle};
use std::collections::HashMap;
use std::sync::Arc;
use tokio::sync::Semaphore;

use crate::api::ApiClient;
use crate::config::Config;
use crate::scanner::{get_source_dir, Scanner};
use crate::ui;

/// Run the preview command
pub async fn run(
    config: &Config,
    source: &str,
    slug: Option<String>,
    expires: Option<u64>,
) -> Result<()> {
    let project_name = config
        .project
        .name
        .as_ref()
        .context("Project name not specified. Run 'sitepod init' or use --name flag.")?;

    // Get source directory
    let source_dir = get_source_dir(source, Some(&config.build.directory));

    ui::step(&format!("Scanning {}", source_dir.display()));

    // Scan files
    let scanner = Scanner::new(&source_dir, &config.deploy.ignore)?;
    let files = scanner.scan()?;

    if files.is_empty() {
        anyhow::bail!("No files found in {}", source_dir.display());
    }

    ui::ok(&format!("Found {} files", files.len()));

    println!();
    ui::step("Creating preview");
    ui::kv("project", ui::accent(project_name));

    // Create API client
    let client = ApiClient::new(config)?;

    // Create plan
    let plan = client.plan(project_name, &files).await?;

    // Upload missing blobs if any
    if !plan.missing.is_empty() {
        let file_map: HashMap<String, _> = files
            .into_iter()
            .map(|f| (f.hashes.blake3.clone(), f))
            .collect();

        let pb = ProgressBar::new(plan.missing.len() as u64);
        pb.set_style(
            ProgressStyle::default_bar()
                .template("  Uploading [{bar:40.cyan/blue}] {pos}/{len}")
                .unwrap()
                .progress_chars("=>-"),
        );

        let semaphore = Arc::new(Semaphore::new(config.deploy.concurrent));
        let client = Arc::new(client);
        let plan_id = plan.plan_id.clone();
        let upload_mode = plan.upload_mode.clone();

        let uploads = stream::iter(plan.missing.clone())
            .map(|blob| {
                let sem = semaphore.clone();
                let client = client.clone();
                let file_map = &file_map;
                let plan_id = plan_id.clone();
                let upload_mode = upload_mode.clone();
                let pb = pb.clone();

                async move {
                    let _permit = sem.acquire().await.unwrap();

                    let file = file_map.get(&blob.hash).context("File not found")?;
                    let data = std::fs::read(&file.absolute_path)?;

                    if upload_mode == "presigned" {
                        client.upload_to_presigned(&blob.upload_url, data).await?;
                    } else {
                        client.upload_blob(&plan_id, &blob.hash, data).await?;
                    }

                    pb.inc(1);
                    Ok::<_, anyhow::Error>(())
                }
            })
            .buffer_unordered(config.deploy.concurrent)
            .collect::<Vec<_>>()
            .await;

        pb.finish_and_clear();

        for result in uploads {
            result?;
        }

        // Commit
        let client = ApiClient::new(config)?;
        let commit = client.commit(&plan_id).await?;

        // Create preview
        let preview = client
            .preview(project_name, &commit.image_id, slug.as_deref(), expires)
            .await?;

        println!();
        ui::ok("Preview ready");
        ui::kv("url", ui::accent(&preview.url).underlined());
        let expires = preview.expires_at.format("%Y-%m-%d %H:%M:%S UTC").to_string();
        ui::kv("expires", ui::dim(&expires));
    } else {
        // Use existing image
        let history = client.history(project_name, 1).await?;
        if history.items.is_empty() {
            anyhow::bail!("No existing images found. Deploy first.");
        }

        let image_id = &history.items[0].image_id;
        let preview = client
            .preview(project_name, image_id, slug.as_deref(), expires)
            .await?;

        println!();
        ui::ok("Preview ready");
        ui::kv("url", ui::accent(&preview.url).underlined());
        let expires = preview.expires_at.format("%Y-%m-%d %H:%M:%S UTC").to_string();
        ui::kv("expires", ui::dim(&expires));
    }

    Ok(())
}
