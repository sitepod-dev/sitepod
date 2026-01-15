//! Domain management methods for the SitePod API.

use anyhow::{Context, Result};

use super::types::{
    AddDomainRequest, AddDomainResponse, CheckSubdomainResponse, ListDomainsResponse,
    RenameDomainRequest, VerifyDomainResponse,
};
use super::ApiClient;

impl ApiClient {
    /// Check if a subdomain is available
    pub async fn check_subdomain(&self, subdomain: &str) -> Result<CheckSubdomainResponse> {
        let url = format!(
            "{}/api/v1/subdomain/check?subdomain={}",
            self.endpoint.trim_end_matches('/'),
            urlencoding::encode(subdomain)
        );

        let resp = self
            .client
            .get(&url)
            .bearer_auth(&self.token)
            .send()
            .await
            .context("Failed to check subdomain")?;

        if !resp.status().is_success() {
            let status = resp.status();
            let text = resp.text().await.unwrap_or_default();
            anyhow::bail!("Subdomain check failed ({}): {}", status, text);
        }

        resp.json()
            .await
            .context("Failed to parse subdomain check response")
    }

    /// Add a domain to a project
    pub async fn add_domain(
        &self,
        project: &str,
        domain: &str,
        slug: &str,
    ) -> Result<AddDomainResponse> {
        let req = AddDomainRequest {
            project: project.to_string(),
            domain: domain.to_string(),
            slug: slug.to_string(),
        };

        let resp = self
            .client
            .post(self.url("/domains"))
            .bearer_auth(&self.token)
            .json(&req)
            .send()
            .await
            .context("Failed to send add domain request")?;

        if !resp.status().is_success() {
            let status = resp.status();
            let text = resp.text().await.unwrap_or_default();
            anyhow::bail!("Add domain failed ({}): {}", status, text);
        }

        resp.json()
            .await
            .context("Failed to parse add domain response")
    }

    /// List domains for a project
    pub async fn list_domains(&self, project: &str) -> Result<ListDomainsResponse> {
        let url = format!(
            "{}/api/v1/domains?project={}",
            self.endpoint.trim_end_matches('/'),
            urlencoding::encode(project)
        );

        let resp = self
            .client
            .get(&url)
            .bearer_auth(&self.token)
            .send()
            .await
            .context("Failed to send list domains request")?;

        if !resp.status().is_success() {
            let status = resp.status();
            let text = resp.text().await.unwrap_or_default();
            anyhow::bail!("List domains failed ({}): {}", status, text);
        }

        resp.json()
            .await
            .context("Failed to parse list domains response")
    }

    /// Verify domain ownership
    pub async fn verify_domain(&self, domain: &str) -> Result<VerifyDomainResponse> {
        let url = format!(
            "{}/api/v1/domains/{}/verify",
            self.endpoint.trim_end_matches('/'),
            urlencoding::encode(domain)
        );

        let resp = self
            .client
            .post(&url)
            .bearer_auth(&self.token)
            .send()
            .await
            .context("Failed to send verify domain request")?;

        if !resp.status().is_success() {
            let status = resp.status();
            let text = resp.text().await.unwrap_or_default();
            anyhow::bail!("Verify domain failed ({}): {}", status, text);
        }

        resp.json()
            .await
            .context("Failed to parse verify domain response")
    }

    /// Remove a domain
    pub async fn remove_domain(&self, domain: &str) -> Result<()> {
        let url = format!(
            "{}/api/v1/domains/{}",
            self.endpoint.trim_end_matches('/'),
            urlencoding::encode(domain)
        );

        let resp = self
            .client
            .delete(&url)
            .bearer_auth(&self.token)
            .send()
            .await
            .context("Failed to send remove domain request")?;

        if !resp.status().is_success() {
            let status = resp.status();
            let text = resp.text().await.unwrap_or_default();
            anyhow::bail!("Remove domain failed ({}): {}", status, text);
        }

        Ok(())
    }

    /// Rename subdomain for a project
    pub async fn rename_subdomain(&self, project: &str, new_subdomain: &str) -> Result<()> {
        let url = format!(
            "{}/api/v1/domains/rename?project={}",
            self.endpoint.trim_end_matches('/'),
            urlencoding::encode(project)
        );

        let req = RenameDomainRequest {
            new_subdomain: new_subdomain.to_string(),
        };

        let resp = self
            .client
            .put(&url)
            .bearer_auth(&self.token)
            .json(&req)
            .send()
            .await
            .context("Failed to send rename subdomain request")?;

        if !resp.status().is_success() {
            let status = resp.status();
            let text = resp.text().await.unwrap_or_default();
            anyhow::bail!("Rename subdomain failed ({}): {}", status, text);
        }

        Ok(())
    }
}
