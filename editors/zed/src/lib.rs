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
        // Try to find kukicha-lsp in PATH
        let path = worktree
            .which("kukicha-lsp")
            .ok_or_else(|| "kukicha-lsp not found in PATH. Install it with: make install-lsp")?;

        Ok(Command {
            command: path,
            args: vec![],
            env: worktree.shell_env(),
        })
    }
}

zed::register_extension!(KukichaExtension);
