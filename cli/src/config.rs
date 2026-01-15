use anyhow::{Context, Result};
use serde::{Deserialize, Serialize};
use std::fs;
use std::path::PathBuf;

/// Global CLI configuration (stored in ~/.sitepod/config.toml)
#[derive(Debug, Default, Serialize, Deserialize)]
pub struct Config {
    #[serde(default)]
    pub server: ServerConfig,

    #[serde(default)]
    pub auth: AuthConfig,

    #[serde(default)]
    pub project: ProjectConfig,

    #[serde(default)]
    pub build: BuildConfig,

    #[serde(default)]
    pub deploy: DeployConfig,
}

#[derive(Debug, Default, Serialize, Deserialize)]
pub struct ServerConfig {
    pub endpoint: Option<String>,
}

#[derive(Debug, Default, Serialize, Deserialize)]
pub struct AuthConfig {
    pub token: Option<String>,
}

#[derive(Debug, Default, Clone, Serialize, Deserialize)]
pub struct ProjectConfig {
    pub name: Option<String>,
    pub subdomain: Option<String>,
    #[serde(default)]
    pub routing_mode: RoutingMode,
}

#[derive(Debug, Default, Clone, Serialize, Deserialize, PartialEq)]
#[serde(rename_all = "lowercase")]
pub enum RoutingMode {
    #[default]
    Subdomain,
    Path,
}

#[derive(Debug, Default, Clone, Serialize, Deserialize)]
pub struct RoutingConfig {
    pub domain: Option<String>,
    pub slug: Option<String>,
}

#[derive(Debug, Default, Serialize, Deserialize)]
pub struct BuildConfig {
    #[serde(default = "default_directory")]
    pub directory: String,
}

fn default_directory() -> String {
    "./dist".to_string()
}

#[derive(Debug, Serialize, Deserialize)]
pub struct DeployConfig {
    #[serde(default = "default_ignore")]
    pub ignore: Vec<String>,

    #[serde(default = "default_concurrent")]
    pub concurrent: usize,
}

impl Default for DeployConfig {
    fn default() -> Self {
        Self {
            ignore: default_ignore(),
            concurrent: default_concurrent(),
        }
    }
}

fn default_ignore() -> Vec<String> {
    vec![
        "**/*.map".to_string(),
        ".*".to_string(),
        "node_modules/**".to_string(),
    ]
}

fn default_concurrent() -> usize {
    20
}

impl Config {
    /// Load configuration from global and local sources
    pub fn load() -> Result<Self> {
        let mut config = Config::default();

        // Load global config first
        if let Some(global_path) = Self::global_config_path() {
            if global_path.exists() {
                let content = fs::read_to_string(&global_path)
                    .with_context(|| format!("Failed to read {}", global_path.display()))?;
                config = toml::from_str(&content)?;
            }
        }

        // Merge with local config (sitepod.toml)
        let local_path = PathBuf::from("sitepod.toml");
        if local_path.exists() {
            let content = fs::read_to_string(&local_path)?;
            let local: ProjectToml = toml::from_str(&content)?;

            // Merge local project config into global
            config.project = local.project;
            config.build = local.build;
            config.deploy = DeployConfig {
                ignore: local.deploy.ignore,
                concurrent: local.deploy.concurrent,
            };
        }

        Ok(config)
    }

    /// Check if sitepod.toml exists in current directory
    pub fn has_local_config() -> bool {
        PathBuf::from("sitepod.toml").exists()
    }

    /// Check if we have valid auth token
    pub fn has_token(&self) -> bool {
        self.auth.token.is_some()
    }

    /// Save token to global config
    pub fn save_token(endpoint: &str, token: &str) -> Result<()> {
        let config_dir = Self::config_dir().context("Failed to determine config directory")?;
        fs::create_dir_all(&config_dir)?;

        let config_path = config_dir.join("config.toml");

        let mut config = if config_path.exists() {
            let content = fs::read_to_string(&config_path)?;
            toml::from_str(&content)?
        } else {
            Config::default()
        };

        config.server.endpoint = Some(endpoint.to_string());
        config.auth.token = Some(token.to_string());

        let content = toml::to_string_pretty(&config)?;
        fs::write(&config_path, content)?;

        Ok(())
    }

    /// Get the global config directory path
    pub fn config_dir() -> Option<PathBuf> {
        dirs::home_dir().map(|h| h.join(".sitepod"))
    }

    /// Get the global config file path
    pub fn global_config_path() -> Option<PathBuf> {
        Self::config_dir().map(|d| d.join("config.toml"))
    }

    /// Get the API endpoint
    pub fn endpoint(&self) -> Result<String> {
        self.server
            .endpoint
            .clone()
            .context("No endpoint configured. Run 'sitepod login' first.")
    }

    /// Get the auth token
    pub fn token(&self) -> Result<String> {
        self.auth
            .token
            .clone()
            .context("No token configured. Run 'sitepod login' first.")
    }
}

/// Project configuration (sitepod.toml)
#[derive(Debug, Default, Serialize, Deserialize)]
pub struct ProjectToml {
    pub project: ProjectConfig,
    pub build: BuildConfig,
    #[serde(default)]
    pub deploy: DeployToml,
}

#[derive(Debug, Default, Serialize, Deserialize)]
pub struct DeployToml {
    #[serde(default = "default_ignore")]
    pub ignore: Vec<String>,

    #[serde(default = "default_concurrent")]
    pub concurrent: usize,

    #[serde(default, skip_serializing_if = "Option::is_none")]
    pub routing: Option<RoutingConfig>,
}

impl ProjectToml {
    #[allow(dead_code)]
    pub fn new(name: &str, directory: &str) -> Self {
        Self {
            project: ProjectConfig {
                name: Some(name.to_string()),
                subdomain: None,
                routing_mode: RoutingMode::Subdomain,
            },
            build: BuildConfig {
                directory: directory.to_string(),
            },
            deploy: DeployToml::default(),
        }
    }

    pub fn with_subdomain(name: &str, subdomain: &str, directory: &str) -> Self {
        Self {
            project: ProjectConfig {
                name: Some(name.to_string()),
                subdomain: Some(subdomain.to_string()),
                routing_mode: RoutingMode::Subdomain,
            },
            build: BuildConfig {
                directory: directory.to_string(),
            },
            deploy: DeployToml::default(),
        }
    }

    pub fn with_path_mode(name: &str, domain: &str, slug: &str, directory: &str) -> Self {
        Self {
            project: ProjectConfig {
                name: Some(name.to_string()),
                subdomain: None,
                routing_mode: RoutingMode::Path,
            },
            build: BuildConfig {
                directory: directory.to_string(),
            },
            deploy: DeployToml {
                ignore: default_ignore(),
                concurrent: default_concurrent(),
                routing: Some(RoutingConfig {
                    domain: Some(domain.to_string()),
                    slug: Some(slug.to_string()),
                }),
            },
        }
    }

    pub fn save(&self) -> Result<()> {
        let content = toml::to_string_pretty(self)?;
        fs::write("sitepod.toml", content)?;
        Ok(())
    }

    #[allow(dead_code)]
    pub fn load() -> Result<Option<Self>> {
        let path = PathBuf::from("sitepod.toml");
        if !path.exists() {
            return Ok(None);
        }
        let content = fs::read_to_string(&path)?;
        let config: Self = toml::from_str(&content)?;
        Ok(Some(config))
    }
}
