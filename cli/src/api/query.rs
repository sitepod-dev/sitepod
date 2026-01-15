//! Query methods for the SitePod API.

use anyhow::{Context, Result};

use super::types::{CurrentResponse, HistoryResponse};
use super::ApiClient;

impl ApiClient {
    /// Get deployment history
    pub async fn history(&self, project: &str, limit: usize) -> Result<HistoryResponse> {
        let url = format!(
            "{}/api/v1/history?project={}&limit={}",
            self.endpoint.trim_end_matches('/'),
            project,
            limit
        );

        let resp = self
            .client
            .get(&url)
            .bearer_auth(&self.token)
            .send()
            .await
            .context("Failed to send history request")?;

        if !resp.status().is_success() {
            let status = resp.status();
            let text = resp.text().await.unwrap_or_default();
            anyhow::bail!("History failed ({}): {}", status, text);
        }

        resp.json()
            .await
            .context("Failed to parse history response")
    }

    /// Get current deployment
    #[allow(dead_code)]
    pub async fn current(&self, project: &str, environment: &str) -> Result<CurrentResponse> {
        let url = format!(
            "{}/api/v1/current?project={}&environment={}",
            self.endpoint.trim_end_matches('/'),
            project,
            environment
        );

        let resp = self
            .client
            .get(&url)
            .bearer_auth(&self.token)
            .send()
            .await
            .context("Failed to send current request")?;

        if !resp.status().is_success() {
            let status = resp.status();
            let text = resp.text().await.unwrap_or_default();
            anyhow::bail!("Current failed ({}): {}", status, text);
        }

        resp.json()
            .await
            .context("Failed to parse current response")
    }
}
