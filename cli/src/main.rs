use anyhow::Result;
use clap::{Parser, Subcommand};

mod api;
mod commands;
mod config;
mod hash;
mod scanner;
mod ui;
mod update;

use commands::{console, deploy, domain, history, init, login, preview, rollback};

/// SitePod — Self-hosted static deployments
#[derive(Parser)]
#[command(name = "sitepod")]
#[command(author, version, about, long_about = None)]
struct Cli {
    #[command(subcommand)]
    command: Commands,

    /// API endpoint (overrides config)
    #[arg(long, env = "SITEPOD_ENDPOINT")]
    endpoint: Option<String>,

    /// Auth token (overrides config)
    #[arg(long, env = "SITEPOD_TOKEN")]
    token: Option<String>,

    /// Skip update check
    #[arg(long, env = "SITEPOD_SKIP_UPDATE_CHECK")]
    skip_update_check: bool,
}

#[derive(Subcommand)]
enum Commands {
    /// Login to SitePod server
    Login {
        /// Server endpoint URL
        #[arg(long)]
        endpoint: Option<String>,
    },

    /// Initialize project configuration
    Init {
        /// Project name
        #[arg(short, long)]
        name: Option<String>,

        /// Build directory (default: ./dist)
        #[arg(short, long)]
        directory: Option<String>,

        /// Subdomain (for subdomain mode)
        #[arg(short, long)]
        subdomain: Option<String>,
    },

    /// Deploy to an environment
    Deploy {
        /// Source directory to deploy
        #[arg(default_value = ".")]
        source: String,

        /// Deploy to production environment
        #[arg(long)]
        prod: bool,

        /// Project name (overrides config)
        #[arg(short, long)]
        name: Option<String>,

        /// Number of concurrent uploads
        #[arg(short, long, default_value = "20")]
        concurrent: usize,

        /// Skip confirmation prompts (for CI/CD)
        #[arg(short, long)]
        yes: bool,
    },

    /// Create a preview deployment
    Preview {
        /// Source directory to deploy
        #[arg(default_value = ".")]
        source: String,

        /// Custom preview slug
        #[arg(short, long)]
        slug: Option<String>,

        /// Expiration time in seconds (default: 86400 = 24h)
        #[arg(short, long)]
        expires: Option<u64>,
    },

    /// Rollback to a previous version
    Rollback {
        /// Project name
        #[arg(short, long)]
        name: Option<String>,

        /// Environment (prod or beta)
        #[arg(short, long, default_value = "prod")]
        env: String,

        /// Image ID to rollback to (interactive if not specified)
        #[arg(long)]
        image: Option<String>,
    },

    /// View deployment history
    History {
        /// Project name
        #[arg(short, long)]
        name: Option<String>,

        /// Maximum number of entries
        #[arg(short, long, default_value = "20")]
        limit: usize,
    },

    /// Open SitePod console in browser
    Console,

    /// Delete your account and all projects
    DeleteAccount {
        /// Skip confirmation prompt
        #[arg(short, long)]
        yes: bool,
    },

    /// Manage custom domains
    Domain {
        #[command(subcommand)]
        subcommand: DomainCommands,
    },
}

#[derive(Subcommand)]
enum DomainCommands {
    /// Add a custom domain
    Add {
        /// Domain to add (e.g., example.com)
        domain: String,

        /// Project name
        #[arg(short, long)]
        project: Option<String>,

        /// URL path slug (default: /)
        #[arg(short, long)]
        slug: Option<String>,
    },

    /// List domains for a project
    List {
        /// Project name
        #[arg(short, long)]
        project: Option<String>,
    },

    /// Verify domain ownership via DNS
    Verify {
        /// Domain to verify
        domain: String,
    },

    /// Remove a domain
    Remove {
        /// Domain to remove
        domain: String,
    },

    /// Rename the subdomain for a project
    Rename {
        /// New subdomain name
        new_subdomain: String,

        /// Project name
        #[arg(short, long)]
        project: Option<String>,
    },
}

#[tokio::main]
async fn main() -> Result<()> {
    ui::init();
    let cli = Cli::parse();

    // Check for updates (in background, non-blocking)
    if !cli.skip_update_check {
        tokio::spawn(async {
            update::check_for_updates().await;
        });
    }

    // Load config
    let mut cfg = config::Config::load()?;

    // Override with CLI args
    if let Some(endpoint) = cli.endpoint {
        cfg.server.endpoint = Some(endpoint);
    }
    if let Some(token) = cli.token {
        cfg.auth.token = Some(token);
    }

    match cli.command {
        Commands::Login { endpoint } => {
            login::run(endpoint.or(cfg.server.endpoint)).await?;
        }
        Commands::Init {
            name,
            directory,
            subdomain,
        } => {
            init::run(&cfg, name, directory, subdomain).await?;
        }
        Commands::Deploy {
            source,
            prod,
            name,
            concurrent,
            yes,
        } => {
            let project = name.or(cfg.project.name.clone());
            let env = if prod { "prod" } else { "beta" };
            deploy::run(&mut cfg, &source, project.as_deref(), env, concurrent, yes).await?;
        }
        Commands::Preview {
            source,
            slug,
            expires,
        } => {
            preview::run(&cfg, &source, slug, expires).await?;
        }
        Commands::Rollback { name, env, image } => {
            let project = name.or(cfg.project.name.clone());
            rollback::run(&cfg, project.as_deref(), &env, image).await?;
        }
        Commands::History { name, limit } => {
            let project = name.or(cfg.project.name.clone());
            history::run(&cfg, project.as_deref(), limit).await?;
        }
        Commands::Console => {
            console::run(&cfg)?;
        }
        Commands::DeleteAccount { yes } => {
            commands::delete_account::run(&cfg, yes).await?;
        }
        Commands::Domain { subcommand } => match subcommand {
            DomainCommands::Add {
                domain: dom,
                project,
                slug,
            } => {
                domain::add(&cfg, project.as_deref(), &dom, slug.as_deref()).await?;
            }
            DomainCommands::List { project } => {
                domain::list(&cfg, project.as_deref()).await?;
            }
            DomainCommands::Verify { domain: dom } => {
                domain::verify(&cfg, &dom).await?;
            }
            DomainCommands::Remove { domain: dom } => {
                domain::remove(&cfg, &dom).await?;
            }
            DomainCommands::Rename {
                new_subdomain,
                project,
            } => {
                domain::rename(&cfg, project.as_deref(), &new_subdomain).await?;
            }
        },
    }

    Ok(())
}
