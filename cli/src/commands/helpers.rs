use reqwest::Client;
use serde::Deserialize;

#[derive(Deserialize)]
struct ConfigResponse {
    domain: Option<String>,
}

pub async fn fetch_base_domain(endpoint: Option<&str>) -> Option<String> {
    let endpoint = endpoint?;
    let url = format!("{}/api/v1/config", endpoint.trim_end_matches('/'));
    let resp = Client::new().get(url).send().await.ok()?;
    if !resp.status().is_success() {
        return None;
    }
    let data: ConfigResponse = resp.json().await.ok()?;
    data.domain
}

pub fn format_subdomain(subdomain: &str, base_domain: Option<&str>) -> String {
    match base_domain {
        Some(domain) if !domain.is_empty() => format!("{}.{}", subdomain, domain),
        _ => subdomain.to_string(),
    }
}
