use anyhow::{Context, Result};
use crate::api::ApiClient;
use crate::config::Config;
use crate::ui;

/// Add a custom domain to a project
pub async fn add(
    config: &Config,
    project: Option<&str>,
    domain: &str,
    slug: Option<&str>,
) -> Result<()> {
    let project_name = project
        .map(String::from)
        .or_else(|| config.project.name.clone())
        .context("Project name required. Use --project or run from a project directory.")?;

    let slug = slug.unwrap_or("/");

    let client = ApiClient::new(config)?;
    let resp = client.add_domain(&project_name, domain, slug).await?;

    ui::ok("Domain added");
    ui::kv("domain", ui::accent(&resp.domain));
    ui::kv("project", ui::accent(&project_name));

    if let Some(txt) = resp.verification_txt {
        println!();
        ui::warn("Verification required");
        println!("Add DNS TXT record:");
        println!();
        println!("  {}", ui::accent(&txt));
        println!();
        println!(
            "{} Run {} to verify",
            ui::icon_info(),
            ui::cmd(&format!("sitepod domain verify {}", resp.domain))
        );
    } else {
        let status_style = match resp.status.as_str() {
            "active" => console::style(&resp.status).green(),
            "pending" => console::style(&resp.status).yellow(),
            _ => console::style(&resp.status).dim(),
        };
        ui::kv("status", status_style);
    }

    Ok(())
}

/// List all domains for a project
pub async fn list(config: &Config, project: Option<&str>) -> Result<()> {
    let project_name = project
        .map(String::from)
        .or_else(|| config.project.name.clone())
        .context("Project name required. Use --project or run from a project directory.")?;

    let client = ApiClient::new(config)?;
    let resp = client.list_domains(&project_name).await?;

    if resp.domains.is_empty() {
        println!(
            "No domains configured for {}",
            ui::accent(&project_name)
        );
        return Ok(());
    }

    ui::heading(&format!("Domains: {}", project_name));
    println!();

    for domain in &resp.domains {
        let status_style = match domain.status.as_str() {
            "active" => console::style(&domain.status).green(),
            "pending" => console::style(&domain.status).yellow(),
            _ => console::style(&domain.status).dim(),
        };

        let primary_marker = if domain.is_primary { " (primary)" } else { "" };
        let type_marker = if domain.domain_type == "system" {
            console::style(" [system]").dim()
        } else {
            console::style(" [custom]").blue()
        };

        println!(
            "  {} {} → {}{}{}",
            status_style,
            console::style(&domain.domain).cyan(),
            console::style(&domain.slug).dim(),
            type_marker,
            console::style(primary_marker).yellow()
        );
    }

    Ok(())
}

/// Verify domain ownership via DNS
pub async fn verify(config: &Config, domain: &str) -> Result<()> {
    let client = ApiClient::new(config)?;
    let resp = client.verify_domain(domain).await?;

    if resp.verified {
        ui::ok("Domain verified");
        ui::kv("domain", ui::accent(domain));
    } else {
        ui::err("Domain not verified");
        ui::kv("domain", ui::accent(domain));
        println!();
        println!("{}", resp.message);
    }

    Ok(())
}

/// Remove a domain from a project
pub async fn remove(config: &Config, domain: &str) -> Result<()> {
    let client = ApiClient::new(config)?;
    client.remove_domain(domain).await?;

    ui::ok("Domain removed");
    ui::kv("domain", ui::accent(domain));

    Ok(())
}

/// Rename the subdomain for a project (subdomain mode only)
pub async fn rename(config: &Config, project: Option<&str>, new_subdomain: &str) -> Result<()> {
    let project_name = project
        .map(String::from)
        .or_else(|| config.project.name.clone())
        .context("Project name required. Use --project or run from a project directory.")?;

    let client = ApiClient::new(config)?;
    client.rename_subdomain(&project_name, new_subdomain).await?;

    ui::ok("Subdomain updated");
    ui::kv("subdomain", ui::accent(new_subdomain));

    Ok(())
}
