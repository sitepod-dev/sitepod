use console::{style, StyledObject};
use std::io::IsTerminal;

pub fn init() {
    let no_color = std::env::var_os("NO_COLOR").is_some();
    if no_color || !std::io::stdout().is_terminal() {
        console::set_colors_enabled(false);
    }
}

pub fn heading(text: &str) {
    println!("{}", style(text).bold().cyan());
}

pub fn step(text: &str) {
    println!("{} {}", icon_step(), text);
}

pub fn info(text: &str) {
    println!("{} {}", icon_info(), text);
}

pub fn ok(text: &str) {
    println!("{} {}", icon_ok(), text);
}

pub fn warn(text: &str) {
    println!("{} {}", icon_warn(), text);
}

pub fn err(text: &str) {
    println!("{} {}", icon_err(), text);
}

pub fn kv(key: &str, value: impl std::fmt::Display) {
    println!("  {:<8} {}", style(format!("{key}:")).dim(), value);
}

pub fn cmd(text: &str) -> StyledObject<&str> {
    style(text).cyan()
}

pub fn dim(text: &str) -> StyledObject<&str> {
    style(text).dim()
}

pub fn accent(text: &str) -> StyledObject<&str> {
    style(text).cyan()
}

pub fn icon_ok() -> StyledObject<&'static str> {
    style("✓").green()
}

pub fn icon_warn() -> StyledObject<&'static str> {
    style("⚠").yellow()
}

pub fn icon_err() -> StyledObject<&'static str> {
    style("✗").red()
}

pub fn icon_info() -> StyledObject<&'static str> {
    style("→").cyan()
}

pub fn icon_step() -> StyledObject<&'static str> {
    style("◐").cyan()
}

pub fn icon_note() -> StyledObject<&'static str> {
    style("ℹ").blue()
}
