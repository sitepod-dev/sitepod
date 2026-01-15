//! SitePod API client module.
//!
//! This module provides the API client for communicating with the SitePod server.
//! The functionality is organized into submodules:
//!
//! - `types`: Request and response type definitions
//! - `auth`: Authentication methods (anonymous_login, bind_email, delete_account)
//! - `deploy`: Deployment methods (plan, upload, commit, release, rollback, preview)
//! - `domain`: Domain management methods (add, list, verify, remove, rename)
//! - `query`: Query methods (history, current)

mod auth;
mod deploy;
mod domain;
mod query;
pub mod types;

use anyhow::Result;
use reqwest::Client;

use crate::config::Config;

// Re-export commonly used types for convenience
#[allow(unused_imports)]
pub use types::{
    AddDomainResponse, AnonymousAuthResponse, BindEmailResponse, CheckSubdomainResponse,
    CommitResponse, CurrentResponse, DeleteAccountResponse, DomainInfo, HistoryItem,
    HistoryResponse, ListDomainsResponse, MissingBlob, PlanResponse, PreviewResponse,
    ReleaseResponse, RollbackResponse, VerifyDomainResponse,
};

/// API client for SitePod server
pub struct ApiClient {
    pub(crate) client: Client,
    pub(crate) endpoint: String,
    pub(crate) token: String,
}

impl ApiClient {
    /// Create a new API client from configuration
    pub fn new(config: &Config) -> Result<Self> {
        let endpoint = config.endpoint()?;
        let token = config.token()?;

        Ok(Self {
            client: Client::new(),
            endpoint,
            token,
        })
    }

    /// Create a new API client from explicit endpoint and token
    #[allow(dead_code)]
    pub fn from_endpoint_token(endpoint: &str, token: &str) -> Self {
        Self {
            client: Client::new(),
            endpoint: endpoint.to_string(),
            token: token.to_string(),
        }
    }

    /// Create an unauthenticated client for login/auth operations
    pub fn unauthenticated(endpoint: &str) -> Self {
        Self {
            client: Client::new(),
            endpoint: endpoint.to_string(),
            token: String::new(),
        }
    }

    /// Build a URL for an API endpoint
    pub(crate) fn url(&self, path: &str) -> String {
        format!("{}/api/v1{}", self.endpoint.trim_end_matches('/'), path)
    }
}
