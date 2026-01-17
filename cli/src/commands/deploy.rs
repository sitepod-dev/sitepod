use anyhow::{Context, Result};
use console::style;
use dialoguer::{Confirm, Input};
use futures::stream::{self, StreamExt};
use indicatif::{MultiProgress, ProgressBar, ProgressStyle};
use qrcode::QrCode;
use rand::Rng;
use std::collections::HashMap;
use std::path::Path;
use std::sync::Arc;
use tokio::sync::Semaphore;

use crate::api::ApiClient;
use crate::config::{Config, ProjectToml};
use crate::scanner::{get_source_dir, Scanner};
use crate::ui;

/// Run the deploy command with smart flow
/// 1. Check if logged in, prompt to login if not
/// 2. Auto-init if sitepod.toml doesn't exist
/// 3. Deploy to the specified environment
pub async fn run(
    config: &mut Config,
    source: &str,
    project: Option<&str>,
    env: &str,
    concurrent: usize,
    skip_confirm: bool,
) -> Result<()> {
    // Step 1: Ensure we have authentication
    if !config.has_token() {
        anyhow::bail!("Not logged in. Run 'sitepod login' first.");
    }

    // Step 2: Ensure we have project config
    if !Config::has_local_config() {
        ui::step("Initializing project");
        auto_init(config).await?;
        println!();

        // Reload config after init
        *config = Config::load()?;
    }

    // Now proceed with actual deployment
    // If auth fails, tell user to re-login
    match do_deploy(config, source, project, env, concurrent, skip_confirm).await {
        Ok(()) => Ok(()),
        Err(e) => {
            let err_str = e.to_string();
            // Check if it's an auth error (401 or "authentication required")
            if err_str.contains("401") || err_str.contains("authentication required") {
                println!();
                ui::warn("Session expired");
                println!();
                anyhow::bail!("Session expired. Run 'sitepod login' to authenticate.");
            } else {
                Err(e)
            }
        }
    }
}

/// Auto-initialize project with minimal interaction
async fn auto_init(config: &Config) -> Result<()> {
    // Get project name from directory
    let default_name = std::env::current_dir()
        .ok()
        .and_then(|p| p.file_name().map(|n| n.to_string_lossy().to_string()))
        .unwrap_or_else(|| "my-project".to_string());

    let project_name: String = Input::new()
        .with_prompt("Project name")
        .default(default_name.clone())
        .interact_text()?;

    // Get subdomain with availability check
    let subdomain = choose_subdomain_quick(config, &project_name).await?;

    // Detect build directory
    let build_dir = detect_build_directory();
    ui::kv("build", format!("{} {}", build_dir, ui::dim("(detected)")));

    // Create and save config
    let project_toml = ProjectToml::with_subdomain(&project_name, &subdomain, &build_dir);
    project_toml.save()?;

    ui::ok("Created sitepod.toml");

    Ok(())
}

/// Quick subdomain selection for auto-init flow
async fn choose_subdomain_quick(config: &Config, default: &str) -> Result<String> {
    let normalized_default = normalize_subdomain(default);

    let input: String = Input::new()
        .with_prompt(format!("Subdomain (default: {})", normalized_default))
        .default(normalized_default.clone())
        .interact_text()?;

    // Handle random ID request
    let subdomain = if input == "-" {
        generate_random_subdomain(&normalized_default)
    } else {
        normalize_subdomain(&input)
    };

    // Check availability if we have a token
    if config.has_token() {
        match check_subdomain_availability(config, &subdomain).await {
            Ok(true) => {
                println!(
                    "  {}.sitepod.dev {}",
                    ui::accent(&subdomain),
                    ui::dim("(available)")
                );
            }
            Ok(false) => {
                ui::warn(&format!(
                    "{}.sitepod.dev taken. Using random suffix.",
                    subdomain
                ));
                let random_subdomain = generate_random_subdomain(&subdomain);
                ui::kv(
                    "subdomain",
                    ui::accent(&format!("{}.sitepod.dev", random_subdomain)),
                );
                return Ok(random_subdomain);
            }
            Err(_) => {
                // Network error, proceed anyway
                ui::kv(
                    "subdomain",
                    format!(
                        "{}.sitepod.dev {}",
                        subdomain,
                        ui::dim("(will verify on deploy)")
                    ),
                );
            }
        }
    }

    Ok(subdomain)
}

/// Check subdomain availability
async fn check_subdomain_availability(config: &Config, subdomain: &str) -> Result<bool> {
    let client = ApiClient::new(config)?;
    let response = client.check_subdomain(subdomain).await?;
    Ok(response.available)
}

/// Normalize string to valid subdomain
fn normalize_subdomain(s: &str) -> String {
    s.to_lowercase()
        .chars()
        .map(|c| {
            if c.is_alphanumeric() || c == '-' {
                c
            } else {
                '-'
            }
        })
        .collect::<String>()
        .trim_matches('-')
        .to_string()
}

/// Generate random subdomain with 4-char suffix
fn generate_random_subdomain(base: &str) -> String {
    let mut rng = rand::thread_rng();
    let suffix: String = (0..4)
        .map(|_| {
            let idx = rng.gen_range(0..36);
            if idx < 10 {
                (b'0' + idx) as char
            } else {
                (b'a' + idx - 10) as char
            }
        })
        .collect();

    format!("{}-{}", base, suffix)
}

/// Detect build directory
fn detect_build_directory() -> String {
    let candidates = ["dist", "build", "out", "public", ".next", ".output"];

    for candidate in candidates {
        if Path::new(candidate).is_dir() {
            return format!("./{}", candidate);
        }
    }

    "./dist".to_string()
}

/// The actual deployment logic
async fn do_deploy(
    config: &Config,
    source: &str,
    project: Option<&str>,
    env: &str,
    concurrent: usize,
    skip_confirm: bool,
) -> Result<()> {
    // Validate project name
    let project_name = project
        .or(config.project.name.as_deref())
        .context("Project name not specified. Run 'sitepod init' or use --name flag.")?;

    // Confirm production deployments
    if env == "prod" && !skip_confirm {
        ui::warn("Deploy to prod");
        if !Confirm::new()
            .with_prompt("Deploy to prod?")
            .default(false)
            .interact()?
        {
            ui::info("Cancelled");
            return Ok(());
        }
        println!();
    }

    // Get source directory
    let source_dir = get_source_dir(source, Some(&config.build.directory));

    ui::step(&format!("Scanning {}", source_dir.display()));

    // Scan files
    let scanner = Scanner::new(&source_dir, &config.deploy.ignore)?;
    let files = scanner.scan()?;

    if files.is_empty() {
        anyhow::bail!("No files found in {}", source_dir.display());
    }

    println!();
    ui::ok(&format!("Found {} files", files.len()));
    println!();
    ui::step("Planning deployment");
    ui::kv("project", ui::accent(project_name));
    ui::kv("env", ui::accent(env));
    ui::kv("files", files.len());

    // Create API client
    let client = ApiClient::new(config)?;

    // Create plan
    let plan = client.plan(project_name, &files).await?;

    let new_count = plan.missing.len();
    let reused = plan.reusable as usize;
    let total = new_count + reused;
    let reuse_pct = if total > 0 {
        (reused as f64 / total as f64 * 100.0) as u32
    } else {
        0
    };

    ui::ok("Plan ready");
    println!(
        "  {} {} new, {} reused ({}%)",
        ui::icon_info(),
        style(new_count).yellow(),
        style(reused).green(),
        reuse_pct
    );

    // Upload missing blobs
    if !plan.missing.is_empty() {
        println!();
        ui::step(&format!("Uploading {} files", new_count));

        // Build file map for lookup
        let file_map: HashMap<String, _> = files
            .into_iter()
            .map(|f| (f.hashes.blake3.clone(), f))
            .collect();

        // Setup progress
        let multi = MultiProgress::new();
        let pb = multi.add(ProgressBar::new(new_count as u64));
        pb.set_style(
            ProgressStyle::default_bar()
                .template("  [{bar:40.cyan/blue}] {pos}/{len} ({bytes_per_sec})")
                .unwrap()
                .progress_chars("=>-"),
        );

        // Upload with concurrency control
        let semaphore = Arc::new(Semaphore::new(concurrent));
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
            .buffer_unordered(concurrent)
            .collect::<Vec<_>>()
            .await;

        pb.finish();

        // Check for errors
        for result in uploads {
            result?;
        }

        ui::ok("Upload complete");

        // Get client back from Arc (we only have immutable reference now)
        let client = ApiClient::new(config)?;

        // Commit and release
        commit_and_release(&client, &plan_id, project_name, env, config).await?;
    } else {
        // No new files to upload, but still need to commit to create image
        println!();
        ui::info("No new files");

        // Commit and release
        commit_and_release(&client, &plan.plan_id, project_name, env, config).await?;
    }

    Ok(())
}

/// Commit and release the deployment
async fn commit_and_release(
    client: &ApiClient,
    plan_id: &str,
    project_name: &str,
    env: &str,
    _config: &Config,
) -> Result<()> {
    // Commit
    ui::step("Committing");

    let commit = client.commit(plan_id).await?;
    ui::ok("Commit ready");
    ui::kv("image", style(&commit.image_id).green());
    let short_hash = format!("{}...", &commit.content_hash[..16]);
    ui::kv("hash", style(short_hash).dim());

    // Release
    println!();
    ui::step(&format!("Releasing to {}", env));

    let release = client
        .release(project_name, env, Some(&commit.image_id))
        .await?;

    println!();
    ui::ok(&format!("Released to {}", env));
    ui::kv("image", style(&commit.image_id).green());
    ui::kv("url", ui::accent(&release.url).underlined());

    // Show QR code
    print_qr_code(&release.url);

    Ok(())
}

/// Print QR code for the URL in terminal
fn print_qr_code(url: &str) {
    let Ok(code) = QrCode::new(url.as_bytes()) else {
        return;
    };

    println!();

    // Use Unicode block characters for compact QR code
    let string = code
        .render::<char>()
        .quiet_zone(true)
        .module_dimensions(2, 1)
        .build();

    for line in string.lines() {
        println!("  {}", line);
    }
}
