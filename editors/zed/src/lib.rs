use zed_extension_api::{self as zed, Command, LanguageServerId, Result, Worktree};

struct KukichaExtension;

impl zed::Extension for KukichaExtension {
    fn new() -> Self {
        KukichaExtension
    }

    fn language_server_command(
        &mut self,
        _language_server_id: &LanguageServerId,
        worktree: &Worktree,
    ) -> Result<Command> {
        // KUKICHA_LSP_PATH override (may not be available in WASM sandbox)
        if let Ok(explicit_path) = std::env::var("KUKICHA_LSP_PATH") {
            let explicit_path = explicit_path.trim();
            if !explicit_path.is_empty() {
                return Ok(Command {
                    command: explicit_path.to_string(),
                    args: vec![],
                    env: worktree.shell_env(),
                });
            }
        }

        let path = worktree
            .which("kukicha-lsp")
            .ok_or_else(|| {
                "kukicha-lsp not found. Set KUKICHA_LSP_PATH or install with: make install-lsp"
            })?;

        Ok(Command {
            command: path,
            args: vec![],
            env: worktree.shell_env(),
        })
    }
}

zed::register_extension!(KukichaExtension);
// test
// test
// test
