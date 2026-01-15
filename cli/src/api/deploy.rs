//! Deployment methods for the SitePod API.

use anyhow::{Context, Result};

use super::types::{
    CommitRequest, CommitResponse, GitInfo, PlanFileEntry, PlanRequest, PlanResponse,
    PreviewRequest, PreviewResponse, ReleaseRequest, ReleaseResponse, RollbackRequest,
    RollbackResponse,
};
use super::ApiClient;
use crate::scanner::ScannedFile;

impl ApiClient {
    /// Create a deployment plan
    pub async fn plan(&self, project: &str, files: &[ScannedFile]) -> Result<PlanResponse> {
        let entries: Vec<PlanFileEntry> = files
            .iter()
            .map(|f| PlanFileEntry {
                path: f.path.clone(),
                blake3: f.hashes.blake3.clone(),
                sha256: f.hashes.sha256.clone(),
                size: f.hashes.size,
                content_type: f.content_type.clone(),
            })
            .collect();

        let req = PlanRequest {
            project: project.to_string(),
            files: entries,
            git: get_git_info(),
        };

        let resp = self
            .client
            .post(self.url("/plan"))
            .bearer_auth(&self.token)
            .json(&req)
            .send()
            .await
            .context("Failed to send plan request")?;

        if !resp.status().is_success() {
            let status = resp.status();
            let text = resp.text().await.unwrap_or_default();
            anyhow::bail!("Plan failed ({}): {}", status, text);
        }

        resp.json().await.context("Failed to parse plan response")
    }

    /// Upload a blob directly (for local storage mode)
    pub async fn upload_blob(&self, plan_id: &str, hash: &str, data: Vec<u8>) -> Result<()> {
        let url = self.url(&format!("/upload/{}/{}", plan_id, hash));

        let resp = self
            .client
            .post(&url)
            .bearer_auth(&self.token)
            .header("Content-Type", "application/octet-stream")
            .body(data)
            .send()
            .await
            .context("Failed to upload blob")?;

        if !resp.status().is_success() {
            let status = resp.status();
            let text = resp.text().await.unwrap_or_default();
            anyhow::bail!("Upload failed ({}): {}", status, text);
        }

        Ok(())
    }

    /// Upload a blob to presigned URL (for S3 storage mode)
    pub async fn upload_to_presigned(&self, url: &str, data: Vec<u8>) -> Result<()> {
        let resp = self
            .client
            .put(url)
            .header("Content-Type", "application/octet-stream")
            .body(data)
            .send()
            .await
            .context("Failed to upload to presigned URL")?;

        if !resp.status().is_success() {
            let status = resp.status();
            let text = resp.text().await.unwrap_or_default();
            anyhow::bail!("Presigned upload failed ({}): {}", status, text);
        }

        Ok(())
    }

    /// Commit a deployment plan
    pub async fn commit(&self, plan_id: &str) -> Result<CommitResponse> {
        let req = CommitRequest {
            plan_id: plan_id.to_string(),
        };

        let resp = self
            .client
            .post(self.url("/commit"))
            .bearer_auth(&self.token)
            .json(&req)
            .send()
            .await
            .context("Failed to send commit request")?;

        if !resp.status().is_success() {
            let status = resp.status();
            let text = resp.text().await.unwrap_or_default();
            anyhow::bail!("Commit failed ({}): {}", status, text);
        }

        resp.json().await.context("Failed to parse commit response")
    }

    /// Release an image to an environment
    pub async fn release(
        &self,
        project: &str,
        environment: &str,
        image_id: Option<&str>,
    ) -> Result<ReleaseResponse> {
        let req = ReleaseRequest {
            project: project.to_string(),
            environment: environment.to_string(),
            image_id: image_id.map(String::from),
        };

        let resp = self
            .client
            .post(self.url("/release"))
            .bearer_auth(&self.token)
            .json(&req)
            .send()
            .await
            .context("Failed to send release request")?;

        if !resp.status().is_success() {
            let status = resp.status();
            let text = resp.text().await.unwrap_or_default();
            anyhow::bail!("Release failed ({}): {}", status, text);
        }

        resp.json()
            .await
            .context("Failed to parse release response")
    }

    /// Rollback to a previous image
    pub async fn rollback(
        &self,
        project: &str,
        environment: &str,
        image_id: &str,
    ) -> Result<RollbackResponse> {
        let req = RollbackRequest {
            project: project.to_string(),
            environment: environment.to_string(),
            image_id: image_id.to_string(),
        };

        let resp = self
            .client
            .post(self.url("/rollback"))
            .bearer_auth(&self.token)
            .json(&req)
            .send()
            .await
            .context("Failed to send rollback request")?;

        if !resp.status().is_success() {
            let status = resp.status();
            let text = resp.text().await.unwrap_or_default();
            anyhow::bail!("Rollback failed ({}): {}", status, text);
        }

        resp.json()
            .await
            .context("Failed to parse rollback response")
    }

    /// Create a preview deployment
    pub async fn preview(
        &self,
        project: &str,
        image_id: &str,
        slug: Option<&str>,
        expires_in: Option<u64>,
    ) -> Result<PreviewResponse> {
        let req = PreviewRequest {
            project: project.to_string(),
            image_id: image_id.to_string(),
            slug: slug.map(String::from),
            expires_in,
        };

        let resp = self
            .client
            .post(self.url("/preview"))
            .bearer_auth(&self.token)
            .json(&req)
            .send()
            .await
            .context("Failed to send preview request")?;

        if !resp.status().is_success() {
            let status = resp.status();
            let text = resp.text().await.unwrap_or_default();
            anyhow::bail!("Preview failed ({}): {}", status, text);
        }

        resp.json()
            .await
            .context("Failed to parse preview response")
    }
}

/// Try to get git information from the current repository
fn get_git_info() -> Option<GitInfo> {
    // Try to get git commit
    let commit = std::process::Command::new("git")
        .args(["rev-parse", "HEAD"])
        .output()
        .ok()
        .and_then(|o| {
            if o.status.success() {
                String::from_utf8(o.stdout).ok().map(|s| s.trim().to_string())
            } else {
                None
            }
        });

    let branch = std::process::Command::new("git")
        .args(["rev-parse", "--abbrev-ref", "HEAD"])
        .output()
        .ok()
        .and_then(|o| {
            if o.status.success() {
                String::from_utf8(o.stdout).ok().map(|s| s.trim().to_string())
            } else {
                None
            }
        });

    let message = std::process::Command::new("git")
        .args(["log", "-1", "--pretty=%s"])
        .output()
        .ok()
        .and_then(|o| {
            if o.status.success() {
                String::from_utf8(o.stdout).ok().map(|s| s.trim().to_string())
            } else {
                None
            }
        });

    if commit.is_some() || branch.is_some() || message.is_some() {
        Some(GitInfo {
            commit,
            branch,
            message,
        })
    } else {
        None
    }
}
