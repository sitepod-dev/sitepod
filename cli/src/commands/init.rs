use anyhow::Result;
use dialoguer::{Input, Select};
use rand::Rng;
use std::path::Path;

use crate::api::ApiClient;
use crate::commands::helpers;
use crate::config::{Config, ProjectToml, RoutingMode};
use crate::ui;

/// Run the init command
pub async fn run(
    config: &Config,
    name: Option<String>,
    directory: Option<String>,
    subdomain: Option<String>,
) -> Result<()> {
    ui::heading("Init project");
    println!();

    // Check if config already exists
    if Path::new("sitepod.toml").exists() {
        ui::warn("sitepod.toml exists. Overwrite? (y/N)");
        let confirm: String = Input::new().default("n".to_string()).interact_text()?;
        if confirm.to_lowercase() != "y" {
            ui::info("Cancelled");
            return Ok(());
        }
    }

    // Get project name
    let default_name = std::env::current_dir()
        .ok()
        .and_then(|p| p.file_name().map(|n| n.to_string_lossy().to_string()))
        .unwrap_or_else(|| "my-project".to_string());

    let project_name: String = if let Some(name) = name {
        name
    } else {
        Input::new()
            .with_prompt("Project name")
            .default(default_name.clone())
            .interact_text()?
    };

    // Choose routing mode
    let modes = vec![
        "Subdomain mode (e.g., my-app.yourdomain.com)",
        "Path mode (e.g., domain.com/my-app/)",
    ];

    let mode_selection = Select::new()
        .with_prompt("Routing mode")
        .items(&modes)
        .default(0)
        .interact()?;

    let routing_mode = if mode_selection == 0 {
        RoutingMode::Subdomain
    } else {
        RoutingMode::Path
    };

    let base_domain = helpers::fetch_base_domain(config.server.endpoint.as_deref()).await;

    // Get build directory
    let default_dir = detect_build_directory();
    let build_dir: String = if let Some(dir) = directory {
        dir
    } else {
        Input::new()
            .with_prompt("Build directory")
            .default(default_dir)
            .interact_text()?
    };

    // Handle routing mode specific configuration
    let project_toml = if routing_mode == RoutingMode::Subdomain {
        // Subdomain mode: check subdomain availability
        let chosen_subdomain = if let Some(sub) = subdomain {
            sub
        } else {
            choose_subdomain(config, &project_name, base_domain.as_deref()).await?
        };

        ProjectToml::with_subdomain(&project_name, &chosen_subdomain, &build_dir)
    } else {
        // Path mode: get domain and slug
        let domain: String = Input::new()
            .with_prompt("Domain")
            .default("h5.example.com".to_string())
            .interact_text()?;

        let slug: String = Input::new()
            .with_prompt("Path prefix")
            .default(format!("/{}", project_name))
            .interact_text()?;

        ProjectToml::with_path_mode(&project_name, &domain, &slug, &build_dir)
    };

    project_toml.save()?;

    println!();
    ui::ok("Created sitepod.toml");

    if routing_mode == RoutingMode::Subdomain {
        if let Some(subdomain) = &project_toml.project.subdomain {
            ui::kv(
                "subdomain",
                ui::accent(&helpers::format_subdomain(
                    subdomain,
                    base_domain.as_deref(),
                )),
            );
        }
    } else if let Some(routing) = &project_toml.deploy.routing {
        if let (Some(domain), Some(slug)) = (&routing.domain, &routing.slug) {
            ui::kv("url", ui::accent(&format!("{}{}", domain, slug)));
        }
    }

    println!();
    println!("Next:");
    println!("  - Build your project");
    println!("  - {}", ui::cmd("sitepod deploy"));
    println!("  - {}", ui::cmd("sitepod deploy --prod"));

    Ok(())
}

/// Interactive subdomain selection with availability checking
async fn choose_subdomain(
    config: &Config,
    default: &str,
    base_domain: Option<&str>,
) -> Result<String> {
    // Normalize default to be URL-safe
    let normalized_default = normalize_subdomain(default);

    loop {
        let input: String = Input::new()
            .with_prompt(format!(
                "Subdomain (enter '-' for random ID, default: {})",
                normalized_default
            ))
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
                    let label = helpers::format_subdomain(&subdomain, base_domain);
                    println!("{} {} available", ui::icon_ok(), label);
                    return Ok(subdomain);
                }
                Ok(false) => {
                    let label = helpers::format_subdomain(&subdomain, base_domain);
                    println!("{} {} taken", ui::icon_err(), label);
                    // Continue loop to ask again
                }
                Err(e) => {
                    // If check fails (e.g., network error), allow proceeding
                    ui::warn(&format!("Availability check failed: {}", e));
                    ui::kv(
                        "subdomain",
                        helpers::format_subdomain(&subdomain, base_domain),
                    );
                    return Ok(subdomain);
                }
            }
        } else {
            // No token yet, can't check availability
            println!(
                "{} Subdomain will be verified on first deploy",
                ui::icon_note()
            );
            return Ok(subdomain);
        }
    }
}

/// Check subdomain availability with the server
async fn check_subdomain_availability(config: &Config, subdomain: &str) -> Result<bool> {
    let client = ApiClient::new(config)?;
    let response = client.check_subdomain(subdomain).await?;
    Ok(response.available)
}

/// Normalize a string to be a valid subdomain
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

/// Generate a random subdomain with a 4-character suffix
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

/// Detect the build directory based on common patterns
fn detect_build_directory() -> String {
    let candidates = ["dist", "build", "out", "public", ".next", ".output"];

    for candidate in candidates {
        if Path::new(candidate).is_dir() {
            return format!("./{}", candidate);
        }
    }

    // Check package.json for hints
    if let Ok(content) = std::fs::read_to_string("package.json") {
        if let Ok(pkg) = serde_json::from_str::<serde_json::Value>(&content) {
            // Check for Next.js
            if pkg
                .get("dependencies")
                .is_some_and(|d| d.get("next").is_some())
            {
                return "./.next".to_string();
            }
            // Check for Vite
            if pkg
                .get("devDependencies")
                .is_some_and(|d| d.get("vite").is_some())
            {
                return "./dist".to_string();
            }
        }
    }

    "./dist".to_string()
}
