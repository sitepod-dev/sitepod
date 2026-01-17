//! Authentication methods for the SitePod API.

use anyhow::{Context, Result};

use super::types::DeleteAccountResponse;
use super::ApiClient;

impl ApiClient {
    /// Delete the current user's account
    pub async fn delete_account(&self) -> Result<DeleteAccountResponse> {
        let resp = self
            .client
            .delete(self.url("/account"))
            .bearer_auth(&self.token)
            .send()
            .await
            .context("Failed to send delete account request")?;

        if !resp.status().is_success() {
            let status = resp.status();
            let text = resp.text().await.unwrap_or_default();
            anyhow::bail!("Delete account failed ({}): {}", status, text);
        }

        resp.json()
            .await
            .context("Failed to parse delete account response")
    }
}
