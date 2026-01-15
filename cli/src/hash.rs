use anyhow::Result;
use base64::{engine::general_purpose::STANDARD, Engine};
use sha2::{Digest, Sha256};
use std::fs::File;
use std::io::{BufReader, Read};
use std::path::Path;

/// File hashes for upload
#[derive(Debug, Clone)]
pub struct FileHashes {
    /// BLAKE3 hash (hex) - used as CAS key
    pub blake3: String,
    /// SHA256 hash (base64) - used for S3 checksum verification
    pub sha256: String,
    /// File size in bytes
    pub size: u64,
}

/// Compute both BLAKE3 and SHA256 hashes for a file
pub fn compute_hashes(path: &Path) -> Result<FileHashes> {
    let file = File::open(path)?;
    let mut reader = BufReader::new(file);

    let mut blake3_hasher = blake3::Hasher::new();
    let mut sha256_hasher = Sha256::new();
    let mut size = 0u64;

    let mut buffer = [0u8; 64 * 1024]; // 64KB buffer
    loop {
        let n = reader.read(&mut buffer)?;
        if n == 0 {
            break;
        }
        blake3_hasher.update(&buffer[..n]);
        sha256_hasher.update(&buffer[..n]);
        size += n as u64;
    }

    Ok(FileHashes {
        blake3: blake3_hasher.finalize().to_hex().to_string(),
        sha256: STANDARD.encode(sha256_hasher.finalize()),
        size,
    })
}

#[cfg(test)]
mod tests {
    use super::*;
    use std::io::Write;
    use tempfile::NamedTempFile;

    #[test]
    fn test_compute_hashes() {
        let mut file = NamedTempFile::new().unwrap();
        file.write_all(b"hello world").unwrap();

        let hashes = compute_hashes(file.path()).unwrap();

        assert_eq!(hashes.size, 11);
        assert!(!hashes.blake3.is_empty());
        assert!(!hashes.sha256.is_empty());
    }
}
