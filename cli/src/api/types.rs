//! Request and response types for the SitePod API.

use chrono::{DateTime, Utc};
use serde::{Deserialize, Serialize};

// ============================================================================
// Auth types
// ============================================================================

#[derive(Debug, Deserialize)]
pub struct AnonymousAuthResponse {
    pub token: String,
    #[allow(dead_code)]
    pub user_id: String,
    #[allow(dead_code)]
    pub expires_at: String,
}

#[derive(Debug, Serialize)]
pub struct BindEmailRequest {
    pub email: String,
}

#[derive(Debug, Deserialize)]
pub struct BindEmailResponse {
    pub message: String,
}

#[derive(Debug, Deserialize)]
pub struct DeleteAccountResponse {
    pub message: String,
    pub deleted_projects: i32,
}

// ============================================================================
// Subdomain types
// ============================================================================

#[derive(Debug, Serialize)]
#[allow(dead_code)]
pub struct CheckSubdomainRequest {
    pub subdomain: String,
}

#[derive(Debug, Deserialize)]
pub struct CheckSubdomainResponse {
    pub available: bool,
    #[allow(dead_code)]
    pub suggestion: Option<String>,
}

// ============================================================================
// Domain types
// ============================================================================

#[derive(Debug, Serialize)]
pub struct AddDomainRequest {
    pub project: String,
    pub domain: String,
    pub slug: String,
}

#[derive(Debug, Deserialize)]
pub struct AddDomainResponse {
    pub domain: String,
    #[allow(dead_code)]
    pub slug: String,
    pub status: String,
    #[allow(dead_code)]
    pub verification_token: Option<String>,
    pub verification_txt: Option<String>,
}

#[derive(Debug, Deserialize)]
pub struct ListDomainsResponse {
    pub domains: Vec<DomainInfo>,
}

#[derive(Debug, Deserialize, Clone)]
pub struct DomainInfo {
    pub domain: String,
    pub slug: String,
    #[serde(rename = "type")]
    pub domain_type: String,
    pub status: String,
    pub is_primary: bool,
    #[allow(dead_code)]
    pub created_at: String,
}

#[derive(Debug, Deserialize)]
pub struct VerifyDomainResponse {
    #[allow(dead_code)]
    pub domain: String,
    #[allow(dead_code)]
    pub status: String,
    pub verified: bool,
    pub message: String,
}

#[derive(Debug, Serialize)]
pub struct RenameDomainRequest {
    pub new_subdomain: String,
}

// ============================================================================
// Plan/Deploy types
// ============================================================================

#[derive(Debug, Serialize)]
pub struct PlanRequest {
    pub project: String,
    pub files: Vec<PlanFileEntry>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub git: Option<GitInfo>,
}

#[derive(Debug, Serialize)]
pub struct PlanFileEntry {
    pub path: String,
    pub blake3: String,
    pub sha256: String,
    pub size: u64,
    pub content_type: String,
}

#[derive(Debug, Serialize)]
pub struct GitInfo {
    pub commit: Option<String>,
    pub branch: Option<String>,
    pub message: Option<String>,
}

#[derive(Debug, Deserialize)]
pub struct PlanResponse {
    pub plan_id: String,
    #[allow(dead_code)]
    pub content_hash: String,
    pub upload_mode: String,
    #[serde(default)]
    pub missing: Vec<MissingBlob>,
    pub reusable: i32,
}

#[derive(Debug, Deserialize, Clone)]
pub struct MissingBlob {
    #[allow(dead_code)]
    pub path: String,
    pub hash: String,
    #[allow(dead_code)]
    pub size: u64,
    pub upload_url: String,
}

#[derive(Debug, Serialize)]
pub struct CommitRequest {
    pub plan_id: String,
}

#[derive(Debug, Deserialize)]
pub struct CommitResponse {
    pub image_id: String,
    pub content_hash: String,
}

#[derive(Debug, Serialize)]
pub struct ReleaseRequest {
    pub project: String,
    pub environment: String,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub image_id: Option<String>,
}

#[derive(Debug, Deserialize)]
pub struct ReleaseResponse {
    pub url: String,
}

#[derive(Debug, Serialize)]
pub struct RollbackRequest {
    pub project: String,
    pub environment: String,
    pub image_id: String,
}

#[derive(Debug, Deserialize)]
pub struct RollbackResponse {
    pub url: String,
    pub previous_image_id: String,
}

#[derive(Debug, Serialize)]
pub struct PreviewRequest {
    pub project: String,
    pub image_id: String,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub slug: Option<String>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub expires_in: Option<u64>,
}

#[derive(Debug, Deserialize)]
pub struct PreviewResponse {
    pub url: String,
    pub expires_at: DateTime<Utc>,
}

// ============================================================================
// Query types
// ============================================================================

#[derive(Debug, Deserialize)]
#[allow(dead_code)]
pub struct CurrentResponse {
    pub image_id: String,
    pub content_hash: String,
    pub deployed_at: DateTime<Utc>,
}

#[derive(Debug, Deserialize)]
pub struct HistoryResponse {
    pub items: Vec<HistoryItem>,
}

#[derive(Debug, Deserialize, Clone)]
pub struct HistoryItem {
    pub image_id: String,
    pub content_hash: String,
    pub created_at: DateTime<Utc>,
    pub git_commit: Option<String>,
}

#[derive(Debug, Deserialize)]
#[allow(dead_code)]
pub struct ApiError {
    pub code: String,
    pub message: String,
}
