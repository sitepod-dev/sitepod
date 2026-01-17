//! Authentication methods for the SitePod API.

use anyhow::{Context, Result};

use super::types::{
    AnonymousAuthResponse, BindEmailRequest, BindEmailResponse, DeleteAccountResponse,
};
use super::ApiClient;

impl ApiClient {
    /// Create an anonymous session
    pub async fn anonymous_login(&self) -> Result<AnonymousAuthResponse> {
        let resp = self
            .client
            .post(self.url("/auth/anonymous"))
            .send()
            .await
            .context("Failed to connect to server")?;

        if !resp.status().is_success() {
            let status = resp.status();
            let text = resp.text().await.unwrap_or_default();
            anyhow::bail!("Anonymous auth failed ({}): {}", status, text);
        }

        resp.json()
            .await
            .context("Failed to parse anonymous auth response")
    }

    /// Bind email to upgrade anonymous account
    pub async fn bind_email(&self, email: &str) -> Result<BindEmailResponse> {
        let req = BindEmailRequest {
            email: email.to_string(),
        };

        let resp = self
            .client
            .post(self.url("/auth/bind"))
            .bearer_auth(&self.token)
            .json(&req)
            .send()
            .await
            .context("Failed to send bind request")?;

        if !resp.status().is_success() {
            let status = resp.status();
            let text = resp.text().await.unwrap_or_default();
            anyhow::bail!("Bind failed ({}): {}", status, text);
        }

        resp.json().await.context("Failed to parse bind response")
    }

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
